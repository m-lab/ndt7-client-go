// Package client implements a minimal ndt7 client.
package client

import (
	"net/http"
	"net/url"
	"time"

	"github.com/apex/log"
	"github.com/gorilla/websocket"
)

// minMeasurementInterval is the minimum value of the interval betwen
// two consecutive measurements performed by either party. An implementation
// MAY choose to close the connection if it is receiving too frequent
// Measurement messages from the other endpoint.
const minMeasurementInterval = 250 * time.Millisecond

// defaultTimeout is the default I/O timeout.
const defaultTimeout = 7 * time.Second

// minMaxMessageSize is the minimum value of the maximum message size
// that an implementation MAY want to configure. Messages smaller than this
// threshold MUST always be accepted by an implementation.
const minMaxMessageSize = 1 << 17

// secWebSocketProtocol is the WebSocket subprotocol used by ndt7.
const secWebSocketProtocol = "net.measurementlab.ndt.v7"

// downloadURLPath selects the download subtest.
const downloadURLPath = "/ndt/v7/download"

// Client is a ndt7 client.
type Client struct {
	// Dialer is the websocket dialer.
	Dialer websocket.Dialer

	// URL is the URL to use.
	URL url.URL
}

type websocketConn interface {
	Close() error
	ReadMessage()(int, []byte, error)
	SetReadLimit(int64)
	SetReadDeadline(time.Time) error
}

type dependencies interface {
	Dial(websocket.Dialer, string, http.Header)(websocketConn, *http.Response, error)
}

type defaultDependencies struct {
}

func (defaultDependencies) Dial(dialer websocket.Dialer, URL string, header http.Header)(websocketConn, *http.Response, error) {
	return dialer.Dial(URL, header)
}

// Download runs a ndt7 download test.
func (cl Client) Download() error {
	return cl.downloadWithDeps(defaultDependencies{})
}

func (cl Client) downloadWithDeps(deps dependencies) error {
	cl.URL.Path = downloadURLPath
	log.Debugf("Conncting to: %s", cl.URL.String())
	headers := http.Header{}
	headers.Add("Sec-WebSocket-Protocol", secWebSocketProtocol)
	cl.Dialer.HandshakeTimeout = defaultTimeout
	conn, _, err := deps.Dial(cl.Dialer, cl.URL.String(), headers)
	if err != nil {
		log.WithError(err).Warn("Connecting failed")
		return err
	}
	// We discard the return value of Close. In the download context this is
	// fine. We either wait for the close message or don't care. When we care,
	// it's consistent to return nil because we're in the good path. In all
	// the other cases, we already have an error to return.
	defer conn.Close()
	conn.SetReadLimit(minMaxMessageSize)
	log.Debug("Starting download")
	for {
		err = conn.SetReadDeadline(time.Now().Add(defaultTimeout))
		if err != nil {
			log.WithError(err).Warn("Cannot set read deadline")
			return err
		}
		mtype, mdata, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.WithError(err).Warn("Download failed")
				return err
			}
			break
		}
		if mtype != websocket.TextMessage {
			continue
		}
		log.Infof("%s", mdata)
	}
	log.Debug("Download complete")
	return nil
}
