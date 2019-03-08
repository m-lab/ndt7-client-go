package main

import (
	"errors"
	"net/http"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt7-clients/go/ndt7-client/common"
	"github.com/m-lab/ndt7-clients/go/ndt7-client/sink"
	"github.com/m-lab/ndt7-clients/go/ndt7-client/source"
)

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
	conn.SetReadLimit(1 << 17) // consistent with the spec
	return conn, nil
}

// upload performs the ndt7 upload
func upload(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade(w, r)
	if err != nil {
		return
	}
	err = common.Closer(conn, sink.Writer(conn, sink.Measurer(sink.Reader(conn))))
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
	err = common.Closer(conn, source.Reader(conn, source.Writer(conn)))
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
