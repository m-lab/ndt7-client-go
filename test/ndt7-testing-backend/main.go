package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/gorilla/websocket"
)

// closeandwarn will warn if closing a closer causes a failure
func closeandwarn(closer io.Closer, message string) {
	err := closer.Close()
	if err != nil {
		log.WithError(err).Warn(message)
	}
}

// upgrade upgrades the connection to websocket
func upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	const protocol = "net.measurementlab.ndt.v7"
	if r.Header.Get("Sec-WebSocket-Protocol") != protocol {
		w.WriteHeader(http.StatusBadRequest)
		return nil, errors.New("Missing WebSocket subprotocol")
	}
	headers := http.Header{}
	headers.Add("Sec-WebSocket-Protocol", protocol)
	var u websocket.Upgrader
	conn, err := u.Upgrade(w, r, headers)
	if err != nil {
		log.WithError(err).Warn("Upgrade failed")
		return nil, err
	}
	return conn, nil
}

// newprepared creates a new random prepared message.
func newprepared() (*websocket.PreparedMessage, error) {
	const size = 1 << 13
	data := make([]byte, size)
	rand.Read(data)
	return websocket.NewPreparedMessage(websocket.BinaryMessage, data)
}

// writeclose writes the websocket.CloseMessage on the connection
func writeclose(conn *websocket.Conn) error {
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	deadline := time.Now().Add(defaultTimeout)
	return conn.WriteControl(websocket.CloseMessage, msg, deadline)
}

// logmeasurement logs a measurement received from the server. We trim the
// trailing newlines such that the logs are more compact.
func logmeasurement(data []byte) {
	log.Infof("%s", strings.TrimRight(string(data), "\n"))
}

// defaultTimeout is the default I/O timeout
const defaultTimeout = 7 * time.Second

// connreader reads the connection and posts the size of the messages
// it reads onto the returned channel. Additionally, this function will
// also log the measurements that it receives. The channel is closed
// silently in case the connection is closed or lost.
func connreader(conn *websocket.Conn) <-chan int64 {
	out := make(chan int64)
	go func() {
		log.Debug("connreader: start")
		defer log.Debug("connreader: stop")
		defer close(out)
		for {
			err := conn.SetReadDeadline(time.Now().Add(defaultTimeout))
			if err != nil {
				log.WithError(err).Warn("SetReadDeadline failed")
				return
			}
			kind, data, err := conn.ReadMessage()
			if err != nil {
				log.WithError(err).Warn("ReadMessage failed")
				return
			}
			if kind == websocket.TextMessage {
				logmeasurement(data)
			}
			out <- int64(len(data))
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

// bytecounter is a filter that drains in and periodically returns
// to the caller the corresponding measurement.
func bytecounter(in <-chan int64) <-chan measurement {
	out := make(chan measurement)
	go func() {
		log.Debug("bytecounter: start")
		defer log.Debug("bytecounter: stop")
		defer close(out)
		const interval = 250 * time.Millisecond
		var total int64
		prev := time.Now()
		begin := prev
		for n := range in {
			t := time.Now()
			total += n // int64 overflow is unlikely here
			if t.Sub(prev) < interval {
				continue
			}
			prev = t
			var m measurement
			m.Elapsed = float64(t.Sub(begin)) / float64(time.Second)
			m.AppInfo.NumBytes = total
			out <- m
		}
	}()
	return out
}

// counterflowsender is a reducer that receives measurements and
// submits such measurements to the client.
func counterflowsender(conn *websocket.Conn, in <-chan measurement) error {
	defer func() {
		for range in {
			// make sure we drain the channel
		}
	}()
	for m := range in {
		data, err := json.Marshal(m)
		if err != nil {
			log.WithError(err).Warn("Marshal failed")
			return err
		}
		err = conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.WithError(err).Warn("WriteMessage failed")
			return err
		}
	}
	return nil
}

// upload performs the ndt7 upload
func upload(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade(w, r)
	if err != nil {
		return
	}
	defer closeandwarn(conn, "Ignored upload connection error")
	// We need to defer responding immediately with a close because the peer
	// may still be sending us some measurements.
	conn.SetCloseHandler(func(int, string) error {
		log.Debug("Get CLOSE message; defer out response until we stop sending")
		return nil
	})
	err = counterflowsender(conn, bytecounter(connreader(conn)))
	if err != nil {
		return
	}
	err = writeclose(conn)
	if err != nil {
		log.WithError(err).Warn("writeclose failed")
		return
	}
}

// logdiscarder reads incoming messages and returns when the connection
// is lost or when there is an error on the connection.
func logdiscarder(conn *websocket.Conn) <-chan error {
	out := make(chan error)
	go func() {
		log.Debug("logdiscarder: start")
		defer log.Debug("logdiscarder: stop")
		defer close(out)
		for {
			err := conn.SetReadDeadline(time.Now().Add(defaultTimeout))
			if err != nil {
				out <- err
				return
			}
			kind, data, err := conn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					out <- err
				}
				return
			}
			if kind == websocket.TextMessage {
				logmeasurement(data)
			}
		}
	}()
	return out
}

// download performs the ndt7 download
func download(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade(w, r)
	if err != nil {
		return
	}
	defer closeandwarn(conn, "Ignored download connection error")
	prepared, err := newprepared()
	if err != nil {
		log.WithError(err).Warn("newprepared failed")
		return
	}
	ch := logdiscarder(conn)
	timer := time.NewTimer(10 * time.Second)
Loop:
	for {
		select {
		case <-timer.C:
			break Loop
		case err, ok := <-ch:
			if !ok {
				break Loop
			}
			log.WithError(err).Warn("ReadMessage failed")
			return
		default:
			err := conn.SetWriteDeadline(time.Now().Add(defaultTimeout))
			if err != nil {
				log.WithError(err).Warn("SetWriteDeadline failed")
				return
			}
			err = conn.WritePreparedMessage(prepared)
			if err != nil {
				log.WithError(err).Warn("WritePreparedMessage failed")
				return
			}
		}
	}
	// If we arrive here, we have sent data for enough time and we did
	// not receive any error when reading. Now, let's write the close
	// message. The logdiscarder may upload some more queued measurements and
	// then it will finally send us back a close message. Receiving a
	// close message, or an error, will cause the logdiscarder to exit.
	err = writeclose(conn)
	if err != nil {
		log.WithError(err).Warn("writeclose failed")
		return
	}
	err = <-ch
	if err != nil {
		log.WithError(err).Warn("ReadMessage failed")
		return
	}
}

func main() {
	log.SetHandler(cli.Default)
	log.SetLevel(log.DebugLevel)
	http.HandleFunc("/ndt/v7/download", download)
	http.HandleFunc("/ndt/v7/upload", upload)
	err := http.ListenAndServeTLS(
		"127.0.0.1:4443", "cert.pem", "key.pem", nil,
	)
	if err != nil {
		log.WithError(err).Fatal("ListenAndServeTLS failed")
	}
}
