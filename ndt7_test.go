package ndt7

import (
	"context"
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt7-client-go/internal/websocketx"
	"github.com/m-lab/ndt7-client-go/mlabns"
	"github.com/m-lab/ndt7-client-go/spec"
)

// newMockedClient returns a mocked client that does nothing
// except pretending it is doing something.
func newMockedClient() *Client {
	client := NewClient(context.Background())
	// Override locate to return a fake IP address
	client.LocateFn = func(c *mlabns.Client) (string, error) {
		return "127.0.0.1", nil
	}
	// Override connect to return a fake websocket connection
	client.connectFn = func(
		dialer websocket.Dialer, ctx context.Context, urlStr string,
		requestHeader http.Header) (*websocket.Conn, *http.Response, error,
	) {
		return &websocket.Conn{}, &http.Response{}, nil
	}
	// Override the download function to basically do nothing
	client.downloadFn = func(
		ctx context.Context, conn websocketx.Conn, ch chan<- spec.Measurement,
	) {
		close(ch)
		// Note that we cannot close the websocket connection because
		// it's just a zero initialized connection (see above)
	}
	client.uploadFn = client.downloadFn
	return client
}

// TestDownloadCase tests the download case.
func TestDownloadCase(t *testing.T) {
	client := newMockedClient()
	ch, err := client.StartDownload()
	if err != nil {
		t.Fatal(err)
	}
	for range ch {
		t.Fatal("did not expect to see an event here")
	}
}

// TestUploadCase tests the download case.
func TestUploadCase(t *testing.T) {
	client := newMockedClient()
	ch, err := client.StartUpload()
	if err != nil {
		t.Fatal(err)
	}
	for range ch {
		t.Fatal("did not expect to see an event here")
	}
}

// TestStartDiscoverServerError ensures that we deal
// with an error when discovering a server.
func TestStartDiscoverServerError(t *testing.T) {
	client := NewClient(context.Background())
	client.MlabNSBaseURL = "\t" // cause URL parse to fail
	_, err := client.start(nil, "")
	if err == nil {
		t.Fatal("We expected an error here")
	}
}

// TestStartConnectError ensures that we deal
// with an error when connecting.
func TestStartConnectError(t *testing.T) {
	client := NewClient(context.Background())
	client.FQDN = "\t" // cause URL parse to fail
	_, err := client.start(nil, "")
	if err == nil {
		t.Fatal("We expected an error here")
	}
}
