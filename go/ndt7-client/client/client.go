// Package client implements a minimal ndt7 client.
package client

import (
	"crypto/rand"
	"crypto/tls"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/gorilla/websocket"
)

// defaultTimeout is the default I/O timeout.
const defaultTimeout = 7 * time.Second

// Client is a ndt7 client.
type Client struct {
	// Hostname is the hostname to use
	Hostname string

	// Port is the port to use
	Port string

	// Insecure controls whether to skip TLS verification
	Insecure bool
}

// dial allows to inject failures when running tests
var dial = func(dialer websocket.Dialer, URL string, header http.Header) (*websocket.Conn, *http.Response, error) {
	return dialer.Dial(URL, header)
}

// dial creates and configures the websocket connection
func (cl Client) dial(urlpath string) (*websocket.Conn, error) {
	var URL url.URL
	URL.Scheme = "wss"
	URL.Path = urlpath
	URL.Host = cl.Hostname + ":" + cl.Port
	var dialer websocket.Dialer
	if cl.Insecure {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	log.Debugf("Connecting to: %s", URL.String())
	headers := http.Header{}
	headers.Add("Sec-WebSocket-Protocol", "net.measurementlab.ndt.v7")
	dialer.HandshakeTimeout = defaultTimeout
	conn, _, err := dial(dialer, URL.String(), headers)
	if err != nil {
		return nil, err
	}
	// According to the specification we must be prepared to read messages
	// that are smaller than the following value.
	conn.SetReadLimit(1 << 17)
	return conn, nil
}

// readrinfo contains the result of reading a websocket message
type readerinfo struct {
	// kind is the message type
	kind int

	// data contains the message data
	data []byte

	// err is the error
	err error
}

// setReadDeadline allows to inject failures when running tests
var setReadDeadline = func(conn *websocket.Conn, time time.Time) error {
	return conn.SetReadDeadline(time)
}

// readMessage allows to inject failures when running tests
var readMessage = func(conn *websocket.Conn) (int, []byte, error) {
	return conn.ReadMessage()
}

// reader posts read websocket messages in the returned channel.
func reader(conn *websocket.Conn) <-chan readerinfo {
	out := make(chan readerinfo)
	go func() {
		for {
			err := setReadDeadline(conn, time.Now().Add(defaultTimeout))
			if err != nil {
				out <- readerinfo{err: err}
				return
			}
			kind, data, err := readMessage(conn)
			out <- readerinfo{kind: kind, data: data, err: err}
			if err != nil {
				return
			}
		}
	}()
	return out
}

// logmeasurement logs a measurement received from the server
func logmeasurement(data []byte) {
	log.Infof("%s", strings.TrimRight(string(data), "\n"))
}

// Download runs a ndt7 download test.
func (cl Client) Download() error {
	conn, err := cl.dial("/ndt/v7/download")
	if err != nil {
		return err
	}
	// We discard the return value of Close. In the download context this is
	// fine. We either wait for the close message or don't care. When we care,
	// it's consistent to return nil because we're in the good path. In all
	// the other cases, we already have an error to return.
	defer conn.Close()
	for rinfo := range reader(conn) {
		if rinfo.err != nil {
			if !websocket.IsCloseError(rinfo.err, websocket.CloseNormalClosure) {
				return rinfo.err
			}
			break
		}
		if rinfo.kind == websocket.TextMessage {
			logmeasurement(rinfo.data)
		}
	}
	return nil
}

// uploaderreader drains incoming messages and logs them
func uploaderreader(conn *websocket.Conn) {
	go func() {
		for rinfo := range reader(conn) {
			if rinfo.err != nil {
				return
			}
			if rinfo.kind == websocket.TextMessage {
				logmeasurement(rinfo.data)
			}
		}
	}()
}

// newprepared creates a new random prepared message.
func newprepared() (*websocket.PreparedMessage, error) {
	const size = 1 << 13
	data := make([]byte, size)
	rand.Read(data)
	return websocket.NewPreparedMessage(websocket.BinaryMessage, data)
}

// Upload runs a ndt7 upload test.
func (cl Client) Upload() error {
	prepared, err := newprepared()
	if err != nil {
		return err
	}
	conn, err := cl.dial("/ndt/v7/upload")
	if err != nil {
		return err
	}
	// We discard the return value of Close. In the download context this is
	// fine. We either wait for the close message or don't care. When we care,
	// it's consistent to return nil because we're in the good path. In all
	// the other cases, we already have an error to return.
	defer conn.Close()
	uploaderreader(conn)
	timer := time.NewTimer(10 * time.Second)
	for {
		select {
		case <-timer.C:
			msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
			deadline := time.Now().Add(defaultTimeout)
			return conn.WriteControl(websocket.CloseMessage, msg, deadline)
		default:
			err := conn.SetWriteDeadline(time.Now().Add(defaultTimeout))
			if err != nil {
				return err
			}
			err = conn.WritePreparedMessage(prepared)
			if err != nil {
				return err
			}
		}
	}
}
