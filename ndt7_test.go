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

// TestIntegrationDownload is an integration test for the download.
func TestIntegrationDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	client := NewClient(userAgent)
	ch, err := client.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	prev := spec.Measurement{}
	tot := 0
	for m := range ch {
		tot++
		if m.Origin != spec.OriginServer {
			t.Fatal("Invalid origin")
		}
		if m.Direction != spec.DirectionDownload {
			t.Fatal("Invalid direction")
		}
		if m.Elapsed <= prev.Elapsed {
			t.Fatal("The time is not increasing")
		}
		if m.BBRInfo.MaxBandwidth <= 0 {
			t.Fatal("Unexpected max bandwidth")
		}
		if m.BBRInfo.MinRTT <= 0.0 {
			t.Fatal("Unexpected min RTT")
		}
		if m.TCPInfo.RTTVar <= 0.0 {
			t.Fatal("Unexpected RTT var")
		}
		if m.TCPInfo.SmoothedRTT <= 0.0 {
			t.Fatal("Unexpected smoothed RTT")
		}
		prev = m
	}
	if tot <= 0 {
		t.Fatal("Expected at least a measurement")
	}
}

// TestIntegrationUpload is an integration test for the upload.
func TestIntegrationUpload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	client := NewClient(userAgent)
	ch, err := client.StartUpload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	prev := spec.Measurement{}
	tot := 0
	for m := range ch {
		tot++
		if m.Origin != spec.OriginClient {
			t.Fatal("Invalid origin")
		}
		if m.Direction != spec.DirectionUpload {
			t.Fatal("Invalid direction")
		}
		if m.Elapsed <= prev.Elapsed {
			t.Fatal("The time is not increasing")
		}
		// Note: it can stay constant when we're servicing
		// a TCP timeout longer than the update interval
		if m.AppInfo.NumBytes < prev.AppInfo.NumBytes {
			t.Fatal("Num bytes is decreasing")
		}
		prev = m
	}
	if tot <= 0 {
		t.Fatal("Expected at least a measurement")
	}
}
