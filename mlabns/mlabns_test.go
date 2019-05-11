package mlabns

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/m-lab/ndt7-client-go/internal/mocks"
)

// TestQueryCommonCase tests the common case.
func TestQueryCommonCase(t *testing.T) {
	const expectedFQDN = "ndt7-mlab1-nai01.measurementlab.org"
	client := NewClient("ndt_ssl", "ndt7-client-go")
	client.requestor = mocks.NewHTTPRequestor(
		200, []byte(fmt.Sprintf(`{"fqdn":"%s"}`, expectedFQDN)), nil,
	)
	fqdn, err := client.Query(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if fqdn != expectedFQDN {
		t.Fatal("Not the FQDN we were expecting")
	}
}

// TestQueryURLError ensures we deal with an invalid URL.
func TestQueryURLError(t *testing.T) {
	client := NewClient("ndt_ssl", "ndt7-client-go")
	client.BaseURL = "\t" // breaks the parser
	_, err := client.Query(context.Background())
	if err == nil {
		t.Fatal("We were expecting an error here")
	}
}

// TestQueryNewRequestError ensures we deal
// with an http.NewRequest errors.
func TestQueryNewRequestError(t *testing.T) {
	mockedError := errors.New("mocked error")
	client := NewClient("ndt_ssl", "ndt7-client-go")
	client.requestMaker = func(
		method, url string, body io.Reader) (*http.Request, error,
	) {
		return nil, mockedError
	}
	_, err := client.Query(context.Background())
	if err != mockedError {
		t.Fatal("Not the error we were expecting")
	}
}

// TestQueryNetworkError ensures we deal with network errors.
func TestQueryNetworkError(t *testing.T) {
	mockedError := errors.New("mocked error")
	client := NewClient("ndt_ssl", "ndt7-client-go")
	client.requestor = mocks.NewHTTPRequestor(
		0, []byte{}, mockedError,
	)
	_, err := client.Query(context.Background())
	if err != mockedError {
		t.Fatal("Not the error we were expecting")
	}
}

// TestQueryInvalidStatusCode ensures we deal with
// a non 200 HTTP status code.
func TestQueryInvalidStatusCode(t *testing.T) {
	client := NewClient("ndt_ssl", "ndt7-client-go")
	client.requestor = mocks.NewHTTPRequestor(
		500, []byte{}, nil,
	)
	_, err := client.Query(context.Background())
	if err != ErrQueryFailed {
		t.Fatal("Not the error we were expecting")
	}
}

// TestQueryJSONParseError ensures we deal with
// a JSON parse error.
func TestQueryJSONParseError(t *testing.T) {
	client := NewClient("ndt_ssl", "ndt7-client-go")
	client.requestor = mocks.NewHTTPRequestor(
		200, []byte("{"), nil,
	)
	_, err := client.Query(context.Background())
	if err == nil {
		t.Fatal("We expected an error here")
	}
}

// TestQueryNoServer ensures we deal with the case
// where no servers are returned.
func TestQueryNoServers(t *testing.T) {
	client := NewClient("ndt_ssl", "ndt7-client-go")
	client.requestor = mocks.NewHTTPRequestor(
		204, []byte(""), nil,
	)
	_, err := client.Query(context.Background())
	if err != ErrNoAvailableServers {
		t.Fatal("Not the error we were expecting")
	}
}
