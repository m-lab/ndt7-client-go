// Package ndt7 contains a ndt7 client.
//
// The client will automatically discover a suitable server to use
// by default. However, you can also manually discover a server and
// configure the client accordingly.
//
// See cmdt/ndt7-client for a complete usage example.
package ndt7

import (
	"context"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt7-client-go/download"
	"github.com/m-lab/ndt7-client-go/mlabns"
	"github.com/m-lab/ndt7-client-go/spec"
	"github.com/m-lab/ndt7-client-go/upload"
)

// Client is a ndt7 client.
type Client struct {
	// FQDN is the server FQDN.
	FQDN string

	// ctx is the client context.
	ctx context.Context
}

// NewClient creates a new client with the specified context.
func NewClient(ctx context.Context) *Client {
	return &Client{
		ctx: ctx,
	}
}

// UserAgent is the user agent used by this client. You can change it
// if you want to specify a different user agent.
const UserAgent = "ndt7-client-go/0.1.0"

// DiscoverServer discovers and returns the closest mlab server.
func DiscoverServer(ctx context.Context) (string, error) {
	config := mlabns.NewConfig("ndt_ssl", UserAgent)
	return mlabns.Query(ctx, config)
}

// connect establishes a websocket connection.
func connect(ctx context.Context, FQDN, URLPath string) (*websocket.Conn, error) {
	URL := url.URL{}
	URL.Scheme = "wss"
	URL.Host = FQDN
	URL.Path = URLPath
	dialer := websocket.Dialer{}
	headers := http.Header{}
	headers.Add("Sec-WebSocket-Protocol", spec.SecWebSocketProtocol)
	headers.Add("User-Agent", UserAgent)
	conn, _, err := dialer.DialContext(ctx, URL.String(), headers)
	return conn, err
}

// startFunc is the function that starts a nettest.
type startFunc = func(context.Context, *websocket.Conn, chan<- spec.Measurement)

// start is the function for starting a subtest.
func (c *Client) start(f startFunc, p string) (<-chan spec.Measurement, error) {
	if c.FQDN == "" {
		fqdn, err := DiscoverServer(c.ctx)
		if err != nil {
			return nil, err
		}
		c.FQDN = fqdn
	}
	conn, err := connect(c.ctx, c.FQDN, p)
	if err != nil {
		return nil, err
	}
	ch := make(chan spec.Measurement)
	go f(c.ctx, conn, ch)
	return ch, nil
}

// StartDownload discovers a ndt7 server (if needed) and starts a download. On
// success it returns a channel where measurements are emitted. This channel is
// closed when the download ends. On failure, the error is non nil and you
// should not attempt using the channel.
func (c *Client) StartDownload() (<-chan spec.Measurement, error) {
	return c.start(download.Run, spec.DownloadURLPath)
}

// StartUpload is like StartDownload but for the upload.
func (c *Client) StartUpload() (<-chan spec.Measurement, error) {
	return c.start(upload.Run, spec.UploadURLPath)
}
