package ndt7

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/m-lab/go/testingx"
	"github.com/m-lab/locate/api/locate"
	v2 "github.com/m-lab/locate/api/v2"
	"github.com/m-lab/locate/locatetest"
	"github.com/m-lab/ndt-server/ndt7/ndt7test"
	"github.com/m-lab/ndt7-client-go/internal/params"
	"github.com/m-lab/ndt7-client-go/internal/websocketx"
	"github.com/m-lab/ndt7-client-go/spec"
)

const (
	clientName    = "ndt7-client-go-tests"
	clientVersion = "0.1.0"
)

type f struct{}

func (x *f) Nearest(ctx context.Context, service string) ([]v2.Target, error) {
	return []v2.Target{
		{
			Machine: "127.0.0.1",
			URLs:    map[string]string{},
		},
	}, nil
}

func newMockedClient(ctx context.Context) *Client {
	client := NewClient(clientName, clientVersion)
	// Override locate to return a fake IP address
	client.Locate = &f{}
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
	testingx.Must(t, err, "failed to start download")
	for range ch {
		t.Fatal("did not expect to see an event here")
	}
}

func TestUploadCase(t *testing.T) {
	ctx := context.Background()
	client := newMockedClient(ctx)
	ch, err := client.StartUpload(ctx)
	testingx.Must(t, err, "failed to start upload")
	for range ch {
		t.Fatal("did not expect to see an event here")
	}
}

func TestStartDiscoverServerError(t *testing.T) {
	ctx := context.Background()
	client := NewClient(clientName, clientVersion)
	badURL := &url.URL{Path: "\t"}
	l := locate.NewClient(makeUserAgent(clientName, clientVersion))
	l.BaseURL = badURL
	client.Locate = l // cause URL parse to fail
	_, err := client.start(ctx, nil, "")
	if err == nil {
		t.Fatal("We expected an error here")
	}
}

func TestStartConnectError(t *testing.T) {
	ctx := context.Background()
	client := NewClient(clientName, clientVersion)
	client.Server = "\t" // cause URL parse to fail
	_, err := client.start(ctx, nil, params.DownloadURLPath)
	if err == nil {
		t.Fatal("We expected an error here")
	}
}

// newLocator returns a locate.Client that returns the given server URLs.
func newLocator(serverURLs []string) *locatetest.Locator {
	s := []string{}
	for _, serverURL := range serverURLs {
		u, err := url.Parse(serverURL)
		if err != nil {
			panic(err)
		}
		s = append(s, u.Host)
	}
	return &locatetest.Locator{
		Servers: s,
	}
}

func TestIntegrationDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	h, fs := ndt7test.NewNDT7Server(t)
	defer os.RemoveAll(h.DataDir)
	defer fs.Close()
	l := locatetest.NewLocateServer(newLocator([]string{fs.URL}))
	client := NewClient(clientName, clientVersion)
	client.Scheme = "ws"
	u, err := url.Parse(l.URL + "/v2/nearest")
	testingx.Must(t, err, "failed to parse locatetest url")

	loc := locate.NewClient(makeUserAgent(clientName, clientVersion))
	loc.BaseURL = u
	client.Locate = loc

	ch, err := client.StartDownload(context.Background())
	testingx.Must(t, err, "download failed to start")

	tot := 0
	for range ch {
		// We already test that measurements follow our expectations
		// in internal/download/download_test.go.
		tot++
	}
	if tot <= 0 {
		t.Fatal("Expected at least a measurement")
	}
	if len(client.Results()) == 0 {
		t.Fatal("Failed to collect any results")
	}
}

func TestIntegrationUpload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	h, fs := ndt7test.NewNDT7Server(t)
	defer os.RemoveAll(h.DataDir)
	defer fs.Close()
	l := locatetest.NewLocateServer(newLocator([]string{fs.URL}))
	client := NewClient(clientName, clientVersion)
	client.Scheme = "ws"
	u, err := url.Parse(l.URL + "/v2/nearest")
	testingx.Must(t, err, "failed to parse locate URL: %s", l.URL+"/v2/nearest")

	loc := locate.NewClient(makeUserAgent(clientName, clientVersion))
	loc.BaseURL = u
	client.Locate = loc

	ch, err := client.StartUpload(context.Background())
	testingx.Must(t, err, "upload failed to start")

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

func TestIntegrationDownloadServiceURL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	h, fs := ndt7test.NewNDT7Server(t)
	defer os.RemoveAll(h.DataDir)
	defer fs.Close()

	u, err := url.Parse(fs.URL)
	testingx.Must(t, err, "failed to parse ndt7test server url")

	t.Run("success", func(t *testing.T) {
		client := NewClient(clientName, clientVersion)
		client.Scheme = "ws"
		client.ServiceURL = &url.URL{
			Scheme: "ws",
			Host:   u.Host,
			Path:   params.DownloadURLPath,
		}

		ch, err := client.StartDownload(context.Background())
		testingx.Must(t, err, "download failed to start")

		tot := 0
		for range ch {
			// We already test that measurements follow our expectations
			// in internal/download/download_test.go.
			tot++
		}
		if tot <= 0 {
			t.Fatal("Expected at least a measurement")
		}
	})
	t.Run("unknown-service", func(t *testing.T) {
		client := NewClient(clientName, clientVersion)
		client.Scheme = "ws"
		client.ServiceURL = &url.URL{
			Scheme: "ws",
			Host:   u.Host,
			Path:   "unknown-service-path",
		}

		_, err := client.StartDownload(context.Background())
		if err != ErrServiceUnsupported {
			t.Fatal("Expected error with unknown-service-path", err)
		}
	})
}

func TestDownloadNoTargets(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	h, fs := ndt7test.NewNDT7Server(t)
	defer os.RemoveAll(h.DataDir)
	defer fs.Close()
	// The first URL is intentionally invalid to test the retry loop.
	l := locatetest.NewLocateServer(newLocator([]string{"https://invalid", fs.URL}))
	client := NewClient(clientName, clientVersion)
	client.Scheme = "ws"
	u, err := url.Parse(l.URL + "/v2/nearest")
	testingx.Must(t, err, "failed to parse locate URL: %s", l.URL+"/v2/nearest")
	loc := locate.NewClient(makeUserAgent(clientName, clientVersion))
	loc.BaseURL = u
	client.Locate = loc

	// This should succeed by using the second URL since the first one fails.
	ch, err := client.StartDownload(context.Background())
	testingx.Must(t, err, "failed to download first attempt")
	tot := 0
	for range ch {
		tot++
	}
	if tot <= 0 {
		t.Fatal("Expected at least a measurement")
	}
	// Second attempt should return ErrNoTargets since all the available
	// servers have been tried.
	_, err = client.StartDownload(context.Background())
	if err != ErrNoTargets {
		t.Fatalf("Expected no target error: %v", err)
	}
}

func TestUploadBadConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	h, fs := ndt7test.NewNDT7Server(t)
	defer os.RemoveAll(h.DataDir)
	defer fs.Close()

	client := NewClient(clientName, clientVersion)
	client.Scheme = "ws"
	u, err := url.Parse(fs.URL)
	testingx.Must(t, err, "failed to parse ndt7test server URL: %s", fs.URL)
	client.Server = u.Host

	// Shutdown the server so the client fails to connect.
	fs.Close()

	// First attempt should succeed.
	_, err = client.StartDownload(context.Background())
	if err == nil {
		t.Fatal("expected error downloading from closed ndt7test server")
	}
}
