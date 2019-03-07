package main

import (
	"errors"
	"io"
	"net/http"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt7-clients/go/ndt7-client/protocol"
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

// upload performs the ndt7 upload
func upload(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade(w, r)
	if err != nil {
		return
	}
	defer closeandwarn(conn, "Ignored error when closing connection")
	err = protocol.Counterflow(conn, protocol.Measurer(protocol.Reader(conn)))
	if err != nil {
		return
	}
}

// download performs the ndt7 download
func download(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade(w, r)
	if err != nil {
		return
	}
	defer closeandwarn(conn, "Ignored error when closing connection")
	go func() {
		for range protocol.Reader(conn) {
			// discard
		}
	}()
	err = <-protocol.Writer(conn)
	if err != nil {
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
