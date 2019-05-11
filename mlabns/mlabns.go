// Package mlabns implements a simple mlab-ns client.
package mlabns

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// httpRequestor is the interface of the implementation that
// performs a mlabns HTTP request for us.
type httpRequestor interface {
	// Do performs the request and returns either a response or
	// a non-nil error to the caller.
	Do(req *http.Request) (*http.Response, error)
}

// httpRequestMaker is the type of the function that
// creates a new HTTP request for us.
type httpRequestMaker = func(
	method, url string, body io.Reader) (*http.Request, error,
)

// DefaultTimeout is the default value for Client.Timeout
const DefaultTimeout = 14 * time.Second

// Client is an mlabns client.
type Client struct {
	// BaseURL is the optional base URL for contacting mlabns. This is
	// initialized in NewClient, but you may override it.
	BaseURL string

	// Timeout is the optional maximum amount of time we're willing to wait
	// for mlabns to respond. This setting is initialized by NewClient to its
	// default value, but you may override it.
	Timeout time.Duration

	// Tool is the mandatory tool to use. This is initialize by NewClient.
	Tool string

	// UserAgent is the mandatory user agent to be used. Also this
	// field is initialized by NewClient.
	UserAgent string

	// requestMaker is the function that creates a request. This is
	// initialized in NewClient, but you may override it.
	requestMaker httpRequestMaker

	// requestor is the implementation that performs the request. This is
	// initialized in NewClient, but you may override it.
	requestor httpRequestor
}

// baseURL is the default base URL.
//
// TODO(bassosimone): when ndt7 is deployed on the whole platform, we can
// stop using the staging mlabns service and use the production one.
const baseURL = "https://locate-dot-mlab-staging.appspot.com/"

// NewClient creates a new Client instance with mandatory userAgent, and tool
// name. For running ndt7, use "ndt7" as the tool name.
func NewClient(tool, userAgent string) *Client {
	return &Client{
		BaseURL:      baseURL,
		Timeout:      DefaultTimeout,
		requestMaker: http.NewRequest,
		requestor:    http.DefaultClient,
		Tool:         tool,
		UserAgent:    userAgent,
	}
}

// serverEntry describes a mlab server.
type serverEntry struct {
	// FQDN is the the FQDN of the server.
	FQDN string `json:"fqdn"`
}

// ErrNoAvailableServers is returned when there are no available servers. A
// background client should treat this error specially and schedule retrying
// after an exponentially distributed number of seconds.
var ErrNoAvailableServers = errors.New("No available M-Lab servers")

// ErrQueryFailed indicates a non-200 status code.
var ErrQueryFailed = errors.New("mlabns returned non-200 status code")

// doGET is an internal function used to perform the request.
func (c *Client) doGET(ctx context.Context, URL string) ([]byte, error) {
	request, err := c.requestMaker("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", c.UserAgent)
	requestctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()
	request = request.WithContext(requestctx)
	response, err := c.requestor.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode == 204 {
		return nil, ErrNoAvailableServers
	}
	if response.StatusCode != 200 {
		return nil, ErrQueryFailed
	}
	return ioutil.ReadAll(response.Body)
}

// Query returns the FQDN of a nearby mlab server. Returns an error on
// failure and the server FQDN on success.
func (c *Client) Query(ctx context.Context) (string, error) {
	URL, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", err
	}
	URL.Path = c.Tool
	data, err := c.doGET(ctx, URL.String())
	if err != nil {
		return "", err
	}
	var server serverEntry
	err = json.Unmarshal(data, &server)
	if err != nil {
		return "", err
	}
	return server.FQDN, nil
}
