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

const userAgent = "mocked/0.1.0"

// newMockedClient returns a mocked client that does nothing
// except pretending it is doing something.
func newMockedClient(ctx context.Context) *Client {
	client := NewClient(userAgent)
	// Override locate to return a fake IP address
	client.LocateFn = func(ctx context.Context, c *mlabns.Client) (string, error) {
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
	ctx := context.Background()
	client := newMockedClient(ctx)
	ch, err := client.StartDownload(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for range ch {
		t.Fatal("did not expect to see an event here")
	}
}

// TestUploadCase tests the download case.
func TestUploadCase(t *testing.T) {
	ctx := context.Background()
	client := newMockedClient(ctx)
	ch, err := client.StartUpload(ctx)
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
	ctx := context.Background()
	client := NewClient(userAgent)
	client.MLabNSClient.BaseURL = "\t" // cause URL parse to fail
	_, err := client.start(ctx, nil, "")
	if err == nil {
		t.Fatal("We expected an error here")
	}
}

// TestStartConnectError ensures that we deal
// with an error when connecting.
func TestStartConnectError(t *testing.T) {
	ctx := context.Background()
	client := NewClient(userAgent)
	client.FQDN = "\t" // cause URL parse to fail
	_, err := client.start(ctx, nil, "")
	if err == nil {
		t.Fatal("We expected an error here")
	}
}
