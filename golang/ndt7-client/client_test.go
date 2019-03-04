package main

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type dialFailure struct {
}

func (dialFailure) Dial(dialer websocket.Dialer, URL string, header http.Header)(websocketConn, *http.Response, error) {
	return nil, nil, errors.New("Cannot dial at this time")
}

func TestDialFailure(t *testing.T) {
	var cl Client
	err := cl.downloadWithDeps(dialFailure{})
	if err == nil {
		t.Fatal("this test should fail")
	}
}

type setReadDeadlineFailure struct {
}

type setReadDeadlineFailureConn struct {
}

func (setReadDeadlineFailureConn) Close() error {
	return nil
}
func (setReadDeadlineFailureConn) ReadMessage()(int, []byte, error) {
	panic("Should not be called")
}
func (setReadDeadlineFailureConn) SetReadLimit(int64) {
}
func (setReadDeadlineFailureConn) SetReadDeadline(time.Time) error {
	return errors.New("We cannot set the read deadline")
}

func (setReadDeadlineFailure) Dial(dialer websocket.Dialer, URL string, header http.Header)(websocketConn, *http.Response, error) {
	return setReadDeadlineFailureConn{}, nil, nil
}

func TestSetReadDeadlineFailure(t *testing.T) {
	var cl Client
	err := cl.downloadWithDeps(setReadDeadlineFailure{})
	if err == nil {
		t.Fatal("this test should fail")
	}
}

type readMessageFailure struct {
}

type readMessageFailureConn struct {
}

func (readMessageFailureConn) Close() error {
	return nil
}
func (readMessageFailureConn) ReadMessage()(int, []byte, error) {
	return 0, nil, errors.New("Oh, noes, cannot read a message")
}
func (readMessageFailureConn) SetReadLimit(int64) {
}
func (readMessageFailureConn) SetReadDeadline(time.Time) error {
	return nil
}

func (readMessageFailure) Dial(dialer websocket.Dialer, URL string, header http.Header)(websocketConn, *http.Response, error) {
	return readMessageFailureConn{}, nil, nil
}

func TestReadMessageFailure(t *testing.T) {
	var cl Client
	err := cl.downloadWithDeps(readMessageFailure{})
	if err == nil {
		t.Fatal("this test should fail")
	}
}

type commonCase struct {
}

type commonCaseConn struct {
	Begin time.Time
	Counter int
}

func (commonCaseConn) Close() error {
	return nil
}
func (ccc commonCaseConn) ReadMessage()(int, []byte, error) {
	// Implementation note: no need to mock running for ten seconds.
	if time.Since(ccc.Begin) > 1 * time.Second {
		return websocket.CloseMessage, nil, &websocket.CloseError{
			Code: websocket.CloseNormalClosure,
		}
	}
	ccc.Counter += 1
	const arbitrary = 25
	if ccc.Counter > arbitrary {
		ccc.Counter = 0  // reset
		return websocket.TextMessage, []byte("{}"), nil
	}
	return websocket.BinaryMessage, []byte("AAA"), nil
}
func (commonCaseConn) SetReadLimit(int64) {
}
func (commonCaseConn) SetReadDeadline(time.Time) error {
	return nil
}

func (commonCase) Dial(dialer websocket.Dialer, URL string, header http.Header)(websocketConn, *http.Response, error) {
	return commonCaseConn{
		Begin: time.Now(),
		Counter: 0,
	}, nil, nil
}

func TestCommonCase(t *testing.T) {
	var cl Client
	err := cl.downloadWithDeps(commonCase{})
	if err != nil {
		t.Fatal("this test should not fail")
	}
}
