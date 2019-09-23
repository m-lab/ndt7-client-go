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

const (
	clientName    = "ndt7-client-go-tests"
	clientVersion = "0.1.0"
)

func newMockedClient(ctx context.Context) *Client {
	client := NewClient(clientName, clientVersion)
	// Override locate to return a fake IP address
	client.locate = func(ctx context.Context, c *mlabns.Client) (string, error) {
		return "127.0.0.1", nil
	}
	// Override connect to return a fake websocket connection
	client.connect = func(
		dialer websocket.Dialer, ctx context.Context, urlStr string,
		requestHeader http.Header) (*websocket.Conn, *http.Response, error,
	) {
		return &websocket.Conn{}, &http.Response{}, nil
	}
	// Override the download function to basically do nothing
	client.download = func(
		ctx context.Context, conn websocketx.Conn, ch chan<- spec.Measurement,
	) error {
		close(ch)
		// Note that we cannot close the websocket connection because
		// it's just a zero initialized connection (see above)
		return nil
	}
	client.upload = client.download
	return client
}

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

func TestStartDiscoverServerError(t *testing.T) {
	ctx := context.Background()
	client := NewClient(clientName, clientVersion)
	client.MLabNSClient.BaseURL = "\t" // cause URL parse to fail
	_, err := client.start(ctx, nil, "")
	if err == nil {
		t.Fatal("We expected an error here")
	}
}

func TestStartConnectError(t *testing.T) {
	ctx := context.Background()
	client := NewClient(clientName, clientVersion)
	client.FQDN = "\t" // cause URL parse to fail
	_, err := client.start(ctx, nil, "")
	if err == nil {
		t.Fatal("We expected an error here")
	}
}

func TestIntegrationDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	client := NewClient(clientName, clientVersion)
	ch, err := client.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	tot := 0
	for range ch {
		// We already test that measurements follow our expectations
		// in internal/download/download_test.go.
		tot++
	}
	if tot <= 0 {
		t.Fatal("Expected at least a measurement")
	}
}

func TestIntegrationUpload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	client := NewClient(clientName, clientVersion)
	ch, err := client.StartUpload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	tot := 0
	for range ch {
		tot++
		// We already test that measurements follow our expectations
		// in internal/upload/upload_test.go.
	}
	if tot <= 0 {
		t.Fatal("Expected at least a measurement")
	}
}
