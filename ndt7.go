// Package ndt7 contains a ndt7 client.
//
// The client will automatically discover a suitable server to use
// by default. However, you can also manually discover a server and
// configure the client accordingly.
package ndt7

import (
	"context"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt7-client-go/internal/download"
	"github.com/m-lab/ndt7-client-go/internal/upload"
	"github.com/m-lab/ndt7-client-go/internal/websocketx"
	"github.com/m-lab/ndt7-client-go/mlabns"
	"github.com/m-lab/ndt7-client-go/spec"
)

// LocateFn is the type of function used to locate a server.
type LocateFn = func(ctx context.Context, config mlabns.Config) (string, error)

// SubtestFn is the type of the function running a subtest.
type SubtestFn = func(
	ctx context.Context, conn websocketx.Conn, ch chan<- spec.Measurement,
)

// defaultUserAgent is the default user agent used by this client.
const defaultUserAgent = "ndt7-client-go/0.1.0"

// Client is a ndt7 client.
type Client struct {
	// Ctx is the client context.
	Ctx context.Context

	// Dialer is the optional websocket Dialer. It's set to its
	// default value by NewClient; you may override it.
	Dialer websocket.Dialer

	// DownloadFn is the function running the download subtest. We
	// set it in NewClient and you may override it.
	DownloadFn SubtestFn

	// FQDN is the optional server FQDN. We will discover the FQDN of
	// a nearby M-Lab server for you if this field is empty.
	FQDN string

	// Locate is the optional function to locate a ndt7 server using
	// the mlab-ns service. This function is set to its default value
	// by NewClient, but you may want to override it.
	Locate LocateFn

	// MlabNSBaseURL is the optional base URL for mlab-ns. We will use
	// the default URL if this field is empty.
	MlabNSBaseURL string

	// UploadFn is like DownloadFn but for the upload subtest.
	UploadFn SubtestFn

	// UserAgent is the user-agent that will be used. It's set by
	// NewClient; you may want to change this value.
	UserAgent string
}

// NewClient creates a new client with the specified context.
func NewClient(ctx context.Context) *Client {
	return &Client{
		Ctx:        ctx,
		DownloadFn: download.Run,
		Locate:     mlabns.Query,
		UploadFn:   upload.Run,
		UserAgent:  defaultUserAgent,
	}
}

// discoverServer discovers and returns the closest mlab server.
func (c *Client) discoverServer() (string, error) {
	config := mlabns.NewConfig("ndt_ssl", c.UserAgent)
	if c.MlabNSBaseURL != "" {
		config.BaseURL = c.MlabNSBaseURL
	}
	return c.Locate(c.Ctx, config)
}

// connect establishes a websocket connection.
func (c *Client) connect(URLPath string) (*websocket.Conn, error) {
	URL := url.URL{}
	URL.Scheme = "wss"
	URL.Host = c.FQDN
	URL.Path = URLPath
	headers := http.Header{}
	headers.Add("Sec-WebSocket-Protocol", spec.SecWebSocketProtocol)
	headers.Add("User-Agent", c.UserAgent)
	conn, _, err := c.Dialer.DialContext(c.Ctx, URL.String(), headers)
	return conn, err
}

// startFunc is the function that starts a nettest.
type startFunc = func(context.Context, websocketx.Conn, chan<- spec.Measurement)

// start is the function for starting a subtest.
func (c *Client) start(f startFunc, p string) (<-chan spec.Measurement, error) {
	if c.FQDN == "" {
		fqdn, err := c.discoverServer()
		if err != nil {
			return nil, err
		}
		c.FQDN = fqdn
	}
	conn, err := c.connect(p)
	if err != nil {
		return nil, err
	}
	ch := make(chan spec.Measurement)
	go f(c.Ctx, conn, ch)
	return ch, nil
}

// StartDownload discovers a ndt7 server (if needed) and starts a download. On
// success it returns a channel where measurements are emitted. This channel is
// closed when the download ends. On failure, the error is non nil and you
// should not attempt using the channel. A side effect of starting the download
// is that, if you did not specify a server FQDN, we will discover a server
// for you and store that value into the c.FQDN field.
func (c *Client) StartDownload() (<-chan spec.Measurement, error) {
	return c.start(c.DownloadFn, spec.DownloadURLPath)
}

// StartUpload is like StartDownload but for the upload.
func (c *Client) StartUpload() (<-chan spec.Measurement, error) {
	return c.start(c.UploadFn, spec.UploadURLPath)
}
