// Package client implements a minimal ndt7 client.
package client

import (
	"encoding/json"
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

// secWebSocketProtocol is the WebSocket subprotocol used by ndt7.
const secWebSocketProtocol = "net.measurementlab.ndt.v7"

// downloadURLPath selects the download subtest.
const downloadURLPath = "/ndt/v7/download"

// defaultTimeout is the default I/O timeout.
const defaultTimeout = 7 * time.Second

// Client is a ndt7 client.
type Client struct {
	// Dialer is the websocket dialer.
	Dialer websocket.Dialer

	// URL is the URL to use.
	URL url.URL
}

// NewClient creates a new client that will use the specified
// hostname and port to create a ndt7 URL.
func NewClient(hostname, port string) Client {
	var clnt Client
	clnt.URL.Scheme = "wss"
	clnt.URL.Host = hostname + ":" + port
	return clnt
}

// dial allows to inject failures when running tests
var dial = func(dialer websocket.Dialer, URL string, header http.Header) (*websocket.Conn, *http.Response, error) {
	return dialer.Dial(URL, header)
}

// setReadDeadline allows to inject failures when running tests
var setReadDeadline = func(conn *websocket.Conn, time time.Time) error {
	return conn.SetReadDeadline(time)
}

// readMessage allows to inject failures when running tests
var readMessage = func(conn *websocket.Conn) (int, []byte, error) {
	return conn.ReadMessage()
}

// readresult is the result of read
type readresult struct {
	// kind is the message type
	kind int

	// data is the message content
	data []byte

	// err is the error
	err error
}

// reader runs read in a background goroutine
func reader(conn *websocket.Conn) <-chan readresult {
	out := make(chan readresult)
	go func() {
		defer close(out)
		for {
			err := setReadDeadline(conn, time.Now().Add(defaultTimeout))
			if err != nil {
				out <- readresult{err: err}
				return
			}
			kind, data, err := readMessage(conn)
			out <- readresult{kind: kind, data: data, err: err}
			if err != nil {
				return
			}
		}
	}()
	return out
}

// setWriteDeadline allows to inject failures when running tests
var setWriteDeadline = func(conn *websocket.Conn, time time.Time) error {
	return conn.SetWriteDeadline(time)
}

// writeMessage allows to inject failures when running tests
var writeMessage = func(c *websocket.Conn, k int, d []byte) error {
	return c.WriteMessage(k, d)
}

// writeinfo contains information for write
type writeinfo struct {
	// kind is the websocket message type
	kind int

	// data is the message content
	data []byte
}

// writer runs write in a background goroutine.
func writer(conn *websocket.Conn, in <-chan writeinfo) {
	go func() {
		for {
			info, ok := <-in
			if !ok {
				return
			}
			err := setWriteDeadline(conn, time.Now().Add(defaultTimeout))
			if err != nil {
				break // must drain first
			}
			err = writeMessage(conn, info.kind, info.data)
			if err != nil {
				break // must drain first
			}
		}
		for range in {
			// DRAIN
		}
	}()
}

// dial creates and configures a websocket connection.
func (cl Client) dial(urlpath string) (*websocket.Conn, error) {
	URL := cl.URL
	URL.Path = urlpath
	log.Debugf("Connecting to: %s", URL.String())
	headers := http.Header{}
	headers.Add("Sec-WebSocket-Protocol", secWebSocketProtocol)
	dialer := cl.Dialer
	dialer.HandshakeTimeout = defaultTimeout
	conn, _, err := dial(dialer, URL.String(), headers)
	if err != nil {
		return nil, err
	}
	conn.SetReadLimit(1 << 17)
	return conn, nil
}

// appinfo contains application level measurements
type appinfo struct {
	// NumBytes is the number of bytes transferred so far
	NumBytes int64 `json:"num_bytes"`
}

// measurement contains a measurement
type measurement struct {
	// Elapsed is the number of elapsed seconds
	Elapsed float64 `json:"elapsed"`

	// AppInfo contains optional application level info
	AppInfo *appinfo `json:"app_info,omitempty"`
}

// Download runs a ndt7 download test.
func (cl Client) Download() error {
	conn, err := cl.dial("/ndt/v7/download")
	if err != nil {
		return err
	}
	defer conn.Close()
	out := make(chan writeinfo)
	defer close(out)
	writer(conn, out)
	in := reader(conn)
	ticker := time.NewTicker(250 * time.Millisecond)
	timer := time.NewTimer(10 * time.Second)
	begin := time.Now()
	var count int64
	for {
		select {
		case <-timer.C: // Enough time has passed
			return nil
		case rr := <-in: // A read completed
			if rr.err != nil {
				return rr.err
			}
			count += int64(len(rr.data))
			if rr.kind == websocket.TextMessage {
				log.Infof("%s", rr.data)
			}
		case t := <-ticker.C:
			var mm measurement
			mm.Elapsed = float64(t.Sub(begin)) / float64(time.Second)
			mm.AppInfo = &appinfo{NumBytes: count}
			data, err := json.Marshal(mm)
			if err != nil {
				return err
			}
			var wi writeinfo
			wi.kind = websocket.TextMessage
			wi.data = data
			out <- wi
		}
	}
}
