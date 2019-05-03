package mlabns

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
)

// TestQueryIntegration tests the common case.
func TestGeoOptionsIntegration(t *testing.T) {
	config := NewConfig("ndt_ssl", "ndt7-client-go")
	FQDN, err := Query(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(FQDN)
}

// TestQueryURLError ensures we deal with an invalid URL.
func TestQueryURLError(t *testing.T) {
	config := NewConfig("ndt_ssl", "ndt7-client-go")
	config.BaseURL = "\t" // breaks the parser
	_, err := Query(context.Background(), config)
	if err == nil {
		t.Fatal("We were expecting an error here")
	}
}

// TestQueryNewRequestError ensures we deal
// with an http.NewRequest errors.
func TestQueryNewRequestError(t *testing.T) {
	savedFunc := httpNewRequest
	mockedError := errors.New("mocked error")
	httpNewRequest = func(method, url string, body io.Reader) (*http.Request, error) {
		return nil, mockedError
	}
	config := NewConfig("ndt_ssl", "ndt7-client-go")
	_, err := Query(context.Background(), config)
	if err != mockedError {
		t.Fatal("Not the error we were expecting")
	}
	httpNewRequest = savedFunc
}

// TestQueryNetworkError ensures we deal with network errors.
func TestQueryNetworkError(t *testing.T) {
	savedFunc := httpClientDo
	mockedError := errors.New("mocked error")
	httpClientDo = func(client *http.Client, req *http.Request) (*http.Response, error) {
		return nil, mockedError
	}
	config := NewConfig("ndt_ssl", "ndt7-client-go")
	_, err := Query(context.Background(), config)
	if err != mockedError {
		t.Fatal("Not the error we were expecting")
	}
	httpClientDo = savedFunc
}

// TestQueryInvalidStatusCode ensures we deal with
// a non 200 HTTP status code.
func TestQueryInvalidStatusCode(t *testing.T) {
	config := NewConfig("nonexistent", "ndt7-client-go")
	_, err := Query(context.Background(), config)
	if err != ErrQueryFailed {
		t.Fatal("Not the error we were expecting")
	}
}

// TestQueryJSONParseError ensures we deal with
// a JSON parse error.
func TestQueryJSONParseError(t *testing.T) {
	savedFunc := doGET
	doGET = func(ctx context.Context, URL, userAgent string) ([]byte, error) {
		return []byte("{"), nil
	}
	config := NewConfig("ndt_ssl", "ndt7-client-go")
	_, err := Query(context.Background(), config)
	if err == nil {
		t.Fatal("We were expecting an error here")
	}
	doGET = savedFunc
}
