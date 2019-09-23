package mlabns

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/m-lab/ndt7-client-go/internal/mocks"
)

const (
	toolName  = "ndt7"
	userAgent = "ndt7-client-go/0.1.0"
)

func TestQueryCommonCase(t *testing.T) {
	const expectedFQDN = "ndt7-mlab1-nai01.measurementlab.org"
	client := NewClient(toolName, userAgent)
	client.HTTPClient = mocks.NewHTTPClient(
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

func TestQueryURLError(t *testing.T) {
	client := NewClient(toolName, userAgent)
	client.BaseURL = "\t" // breaks the parser
	_, err := client.Query(context.Background())
	if err == nil {
		t.Fatal("We were expecting an error here")
	}
}

func TestQueryNewRequestError(t *testing.T) {
	mockedError := errors.New("mocked error")
	client := NewClient(toolName, userAgent)
	client.RequestMaker = func(
		method, url string, body io.Reader) (*http.Request, error,
	) {
		return nil, mockedError
	}
	_, err := client.Query(context.Background())
	if err != mockedError {
		t.Fatal("Not the error we were expecting")
	}
}

func TestQueryNetworkError(t *testing.T) {
	mockedError := errors.New("mocked error")
	client := NewClient(toolName, userAgent)
	client.HTTPClient = mocks.NewHTTPClient(
		0, []byte{}, mockedError,
	)
	_, err := client.Query(context.Background())
	// According to Go docs, the return value of http.Client.Do is always
	// of type `*url.Error` and wraps the original error.
	if err.(*url.Error).Err != mockedError {
		t.Fatal("Not the error we were expecting")
	}
}

func TestQueryInvalidStatusCode(t *testing.T) {
	client := NewClient(toolName, userAgent)
	client.HTTPClient = mocks.NewHTTPClient(
		500, []byte{}, nil,
	)
	_, err := client.Query(context.Background())
	if err != ErrQueryFailed {
		t.Fatal("Not the error we were expecting")
	}
}

func TestQueryJSONParseError(t *testing.T) {
	client := NewClient(toolName, userAgent)
	client.HTTPClient = mocks.NewHTTPClient(
		200, []byte("{"), nil,
	)
	_, err := client.Query(context.Background())
	if err == nil {
		t.Fatal("We expected an error here")
	}
}

func TestQueryNoServers(t *testing.T) {
	client := NewClient(toolName, userAgent)
	client.HTTPClient = mocks.NewHTTPClient(
		204, []byte(""), nil,
	)
	_, err := client.Query(context.Background())
	if err != ErrNoAvailableServers {
		t.Fatal("Not the error we were expecting")
	}
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	client := NewClient(toolName, userAgent)
	fqdn, err := client.Query(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if fqdn == "" {
		t.Fatal("unexpected empty fqdn")
	}
}
