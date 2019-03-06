// Package client implements a minimal ndt7 client.
package client

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"io"
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

// closeandwarn will warn if closing a close causes a failure
func closeandwarn(closer io.Closer, message string) {
	err := closer.Close()
	if err != nil {
		log.WithError(err).Warn(message)
	}
}

// ddappinfo is a ddata field sent to the server
type ddappinfo struct {
	// NumBytes is the number of bytes received so far
	NumBytes int64 `json:"num_bytes"`
}

// ddata contains application level download measurements
type ddata struct {
	// Elapsed is the time elapsed since we started
	Elapsed float64 `json:"elapsed"`

	// AppInfo contains the real measurement
	AppInfo ddappinfo `json:"app_info"`
}

// textwriter is a writer sending out textual data
func textwriter(conn *websocket.Conn, ch <-chan ddata) {
	go func() {
		for {
			ddata, good := <-ch
			if !good {
				return // No more messages to send will come
			}
			data, err := json.Marshal(ddata)
			if err != nil {
				log.WithError(err).Debug(
					"Error when marshalling JSON for server; stopping writing",
				)
				break // must drain the channel
			}
			err = conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.WithError(err).Debug(
					"Error when writing to the server; stopping writing",
				)
				break // must drain the channel
			}
		}
		for range ch {
			// do nothing and just drain
		}
	}()
}

// downloaderreader is a reader that also records how much we have
// received so far and attempts to update the server.
func downloaderreader(conn *websocket.Conn) <-chan readerinfo {
	out := make(chan readerinfo)
	go func() {
		begin := time.Now()
		var count int64
		in := reader(conn)
		ticker := time.NewTicker(250 * time.Millisecond)
		ddch := make(chan ddata)
		defer close(ddch)
		textwriter(conn, ddch)
		for {
			select {
			case now := <-ticker.C:
				var dd ddata
				dd.Elapsed = float64(now.Sub(begin)) / float64(time.Second)
				dd.AppInfo.NumBytes = count
				select {
				case ddch <- dd:
					// nothing
				default:
					log.Debug("cannot submit to the sender channel") // actionable?
				}
			case rinfo := <-in:
				if rinfo.err == nil {
					count += int64(len(rinfo.data))
				}
				out <- rinfo
				if rinfo.err != nil {
					return // Detach outself from the pipeline
				}
			}
		}
	}()
	return out
}

// Download runs a ndt7 download test.
func (cl Client) Download() error {
	conn, err := cl.dial("/ndt/v7/download")
	if err != nil {
		return err
	}
	defer closeandwarn(conn, "Ignored error when closing download connection")
	for rinfo := range downloaderreader(conn) {
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
				// Implementation note: using Debug here because this error
				// really isn't very actionable by us. If there is a real
				// network error, we'll see if _also_ when writing. If the
				// server doesn't send us anything, we'll just see that
				// after the timeout we'll stop reading here.
				log.WithError(rinfo.err).Debug(
					"Ignored error when reading messages during upload",
				)
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
	defer closeandwarn(conn, "Ignored error when closing upload connection")
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
