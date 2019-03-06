package client

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestDialFailure(t *testing.T) {
	orig := dial
	dial = func(d websocket.Dialer, u string, h http.Header) (*websocket.Conn, *http.Response, error) {
		return nil, nil, errors.New("Cannot dial at this time")
	}
	var cl Client
	err := cl.Download()
	if err == nil {
		t.Fatal("this test should fail")
	}
	dial = orig
}

func TestSetReadDeadlineFailure(t *testing.T) {
	orig := setReadDeadline
	setReadDeadline = func(c *websocket.Conn, t time.Time) error {
		return errors.New("Cannot set the read deadline")
	}
	var cl Client
	err := cl.Download()
	if err == nil {
		t.Fatal("this test should fail")
	}
	setReadDeadline = orig
}

func TestReadMessageFailure(t *testing.T) {
	orig := readMessage
	readMessage = func(c *websocket.Conn) (int, []byte, error) {
		return 0, nil, errors.New("Oh, noes, cannot read a message")
	}
	var cl Client
	err := cl.Download()
	if err == nil {
		t.Fatal("this test should fail")
	}
	readMessage = orig
}
