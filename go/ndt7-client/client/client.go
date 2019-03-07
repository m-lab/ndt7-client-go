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

// received is the result of receiving a websocket message.
type received struct {
	// err indicates whether the message contains an error.
	err error

	// kind is the websocket message type (e.g. websocket.TextMessage).
	kind int

	// data contains the websocket message data.
	data []byte
}

// setReadDeadline allows to inject failures when running tests
var setReadDeadline = func(conn *websocket.Conn, time time.Time) error {
	return conn.SetReadDeadline(time)
}

// readMessage allows to inject failures when running tests
var readMessage = func(conn *websocket.Conn) (int, []byte, error) {
	return conn.ReadMessage()
}

// internalreceiver is a generator that receives all the possible messages
// from the websocket and emits them on a channel. The channel will be
// closed if the connection is cleanly closed or receiving fails. Unless
// the connection is cleanly closed, we also return the error. The consumer
// of the channel must drain it until completion.
func internalreceiver(conn *websocket.Conn) <-chan received {
	out := make(chan received)
	go func() {
		log.Debug("internalreceiver: start")
		defer log.Debug("internalreceiver: stop")
		defer close(out) // signal the receiver we're done
		for {
			err := setReadDeadline(conn, time.Now().Add(defaultTimeout))
			if err != nil {
				out <- received{err: err}
				return
			}
			kind, data, err := readMessage(conn)
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					out <- received{err: err} // no error on clean close
				}
				return
			}
			out <- received{kind: kind, data: data}
		}
	}()
	return out
}

// logmeasurement logs a measurement received from the server. We trim the
// trailing newlines such that the logs are more compact.
func logmeasurement(data []byte) {
	log.Infof("%s", strings.TrimRight(string(data), "\n"))
}

// logreceiver is a filter that drains the input channel and writes all
// the messages to the output channel. If any message contains a measurement
// from the server, this message is also logged. The consumer of the
// returned channel is supposed to drain all messages from it.
func logreceiver(in <-chan received) <-chan received {
	out := make(chan received)
	go func() {
		log.Debug("logreceiver: start")
		defer log.Debug("logreceiver: stop")
		defer close(out) // signal the reader we're done
		for received := range in {
			if received.err == nil && received.kind == websocket.TextMessage {
				logmeasurement(received.data)
			}
			out <- received
		}
	}()
	return out
}

// measured is an application level measurement we performed
type measured struct {
	// err indicates that a previous stage failed
	err error

	// elapsed is the time elapsed since the beginning
	elapsed time.Duration

	// count is the number of bytes read so far
	count int64
}

// measurementInterval is the minimum interval between measurements
const measurementInterval = 250 * time.Millisecond

// measuringreceiver is a filter that drains the input channel to compute
// application level measurements. Such measurements will be emitted on
// the output channel. The consumer of this channel must fully drain such
// channel because otherwise this goroutine will block.
func measuringreceiver(in <-chan received) <-chan measured {
	out := make(chan measured)
	go func() {
		log.Debug("measuringreceiver: start")
		defer log.Debug("measuringreceiver: stop")
		defer close(out) // signal the reader we're done
		defer func() {
			for range in {
				// to be strict drain and ignore possible subsequent messages
				// even though this should not be required
			}
		}()
		prev := time.Now()
		var count int64
		for received := range in {
			if received.err != nil {
				out <- measured{err: received.err}
				return // in channel is drained
			}
			now := time.Now()
			elapsed := now.Sub(prev)
			count += int64(len(received.data)) // overflow here is unlikely
			if elapsed < measurementInterval {
				continue
			}
			prev = now
			out <- measured{elapsed: elapsed, count: count}
		}
	}()
	return out
}

// appinfo is a field of measurement
type appinfo struct {
	// NumBytes indicates the number of bytes read so far
	NumBytes int64 `json:"num_bytes"`
}

// measurement is an application level measurement formatted according
// to the specification of ndt7
type measurement struct {
	// Elapsed is the number of seconds elapsed
	Elapsed float64 `json:"elapsed"`

	// AppInfo contains application level measurements
	AppInfo appinfo `json:"app_info"`
}

// measurementupload is a reducer that fully drains the input channel
// and sends the application level measurements to the server. In
// case of any marshalling or sending error, we will still continue
// draining the input channel. Yet, in such case, our error will
// take precedence over the error that the channel may report.
func measurementuploader(conn *websocket.Conn, in <-chan measured) error {
	log.Debug("measurementuploader: start")
	defer log.Debug("measurementuploader: stop")
	defer func() {
		for range in {
			// Make sure we drain the input channel. This is necessary if
			// we leave this context early because of a marshalling or I/O
			// error below. The common case should be that we fully drain
			// in because the download is done.
		}
	}()
	for measured := range in {
		if measured.err != nil {
			return measured.err // in channel is drained
		}
		// Prepare a ndt7 application level measurement message
		var mm measurement
		mm.Elapsed = float64(measured.elapsed) / float64(time.Second)
		mm.AppInfo.NumBytes = measured.count
		data, err := json.Marshal(mm)
		if err != nil {
			return err // in channel is drained
		}
		// Submit the message to the server
		err = conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			return err // in channel is drained
		}
	}
	return nil
}

// closeandwarn will warn if closing a closer causes a failure
func closeandwarn(closer io.Closer, message string) {
	err := closer.Close()
	if err != nil {
		log.WithError(err).Warn(message)
	}
}

func writeclose(conn *websocket.Conn) error {
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	deadline := time.Now().Add(defaultTimeout)
	return conn.WriteControl(websocket.CloseMessage, msg, deadline)
}

// Download runs a ndt7 download test.
func (cl Client) Download() error {
	conn, err := cl.dial("/ndt/v7/download")
	if err != nil {
		return err
	}
	// TODO(bassosimone): EXPLAIN EXPLAIN EXPLAIN!
	conn.SetCloseHandler(func(int, string) error {
		log.Debug("Got CLOSE message; defer replying until we stop sending")
		return nil
	})
	defer closeandwarn(conn, "Ignored error when closing download connection")
	err = measurementuploader(conn, measuringreceiver(
		logreceiver(internalreceiver(conn))))
	if err != nil {
		return err
	}
	return writeclose(conn)
}

// newprepared creates a new random prepared message.
func newprepared() (*websocket.PreparedMessage, error) {
	const size = 1 << 13
	data := make([]byte, size)
	rand.Read(data)
	return websocket.NewPreparedMessage(websocket.BinaryMessage, data)
}

// binaryuploader will send binary messages for the expected upload time
// and returns whether there was an error. The connection will still be
// open when this function returns and you shall close it.
func binaryuploader(conn *websocket.Conn) error {
	log.Debug("binaryuploader: start")
	defer log.Debug("binaryuploader: stop")
	prepared, err := newprepared()
	if err != nil {
		return err
	}
	timer := time.NewTimer(10 * time.Second)
	for {
		select {
		case <-timer.C:
			return nil
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

// Upload runs a ndt7 upload test.
func (cl Client) Upload() error {
	conn, err := cl.dial("/ndt/v7/upload")
	if err != nil {
		return err
	}
	defer closeandwarn(conn, "Ignored error when closing upload connection")
	go func() {
		// Just make sure we drain the receiver channel
		for received := range logreceiver(internalreceiver(conn)) {
			if received.err != nil {
				// Error not really actionable by us, so use the Debug level
				log.WithError(received.err).Debug("Read error")
			}
		}
	}()
	err = binaryuploader(conn)
	if err != nil {
		return err
	}
	return writeclose(conn)
}
