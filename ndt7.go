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
type LocateFn = func(ctx context.Context, client *mlabns.Client) (string, error)

// connectFn is the type of the function used to create
// a new *websocket.Conn connection.
type connectFn = func(
	dialer websocket.Dialer,
	ctx context.Context, urlStr string,
	requestHeader http.Header,
) (*websocket.Conn, *http.Response, error)

// subtestFn is the type of the function running a subtest.
type subtestFn = func(
	ctx context.Context, conn websocketx.Conn, ch chan<- spec.Measurement,
)

// DefaultUserAgent is the default user agent used by this client.
const DefaultUserAgent = "ndt7-client-go/0.1.0"

// Client is a ndt7 client.
type Client struct {
	// Dialer is the optional websocket Dialer. It's set to its
	// default value by NewClient; you may override it.
	Dialer websocket.Dialer

	// FQDN is the optional server FQDN. We will discover the FQDN of
	// a nearby M-Lab server for you if this field is empty.
	FQDN string

	// LocateFn is the optional function to locate a ndt7 server using
	// the mlab-ns service. This function is set to its default value
	// by NewClient, but you may want to override it.
	LocateFn LocateFn

	// MlabNSBaseURL is the optional base URL for mlab-ns. We will use
	// the default URL if this field is empty.
	MlabNSBaseURL string

	// UserAgent is the user-agent that will be used. It's set by
	// NewClient; you may want to change this value.
	UserAgent string

	// connectFn is the function for connecting a specific
	// websocket cnnection. It's set to its default value by
	// NewClient, but you may override it.
	connectFn connectFn

	// downloadFn is the function running the download subtest. We
	// set it in NewClient and you may override it.
	downloadFn subtestFn

	// uploadFn is like DownloadFn but for the upload subtest.
	uploadFn subtestFn
}

// NewClient creates a new client instance.
func NewClient() *Client {
	return &Client{
		connectFn: func(
			dialer websocket.Dialer, ctx context.Context, urlStr string,
			requestHeader http.Header) (*websocket.Conn, *http.Response, error,
		) {
			return dialer.DialContext(ctx, urlStr, requestHeader)
		},
		downloadFn: download.Run,
		LocateFn: func(ctx context.Context, c *mlabns.Client) (string, error) {
			return c.Query(ctx)
		},
		uploadFn:  upload.Run,
		UserAgent: DefaultUserAgent,
	}
}

// discoverServer discovers and returns the closest mlab server.
func (c *Client) discoverServer(ctx context.Context) (string, error) {
	client := mlabns.NewClient("ndt_ssl", c.UserAgent)
	if c.MlabNSBaseURL != "" {
		client.BaseURL = c.MlabNSBaseURL
	}
	return c.LocateFn(ctx, client)
}

// connect establishes a websocket connection.
func (c *Client) connect(ctx context.Context, URLPath string) (*websocket.Conn, error) {
	URL := url.URL{}
	URL.Scheme = "wss"
	URL.Host = c.FQDN
	URL.Path = URLPath
	headers := http.Header{}
	headers.Add("Sec-WebSocket-Protocol", spec.SecWebSocketProtocol)
	headers.Add("User-Agent", c.UserAgent)
	conn, _, err := c.connectFn(c.Dialer, ctx, URL.String(), headers)
	return conn, err
}

// start is the function for starting a subtest.
func (c *Client) start(ctx context.Context, f subtestFn, p string) (<-chan spec.Measurement, error) {
	if c.FQDN == "" {
		fqdn, err := c.discoverServer(ctx)
		if err != nil {
			return nil, err
		}
		c.FQDN = fqdn
	}
	conn, err := c.connect(ctx, p)
	if err != nil {
		return nil, err
	}
	ch := make(chan spec.Measurement)
	go f(ctx, conn, ch)
	return ch, nil
}

// StartDownload discovers a ndt7 server (if needed) and starts a download. On
// success it returns a channel where measurements are emitted. This channel is
// closed when the download ends. On failure, the error is non nil and you
// should not attempt using the channel. A side effect of starting the download
// is that, if you did not specify a server FQDN, we will discover a server
// for you and store that value into the c.FQDN field.
func (c *Client) StartDownload(ctx context.Context) (<-chan spec.Measurement, error) {
	return c.start(ctx, c.downloadFn, spec.DownloadURLPath)
}

// StartUpload is like StartDownload but for the upload.
func (c *Client) StartUpload(ctx context.Context) (<-chan spec.Measurement, error) {
	return c.start(ctx, c.uploadFn, spec.UploadURLPath)
}
