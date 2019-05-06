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
)

// HTTPRequestor is the interface of the implementation that
// performs a mlabns HTTP request for us.
type HTTPRequestor interface {
	// Do performs the request and returns either a response or
	// a non-nil error to the caller.
	Do(req *http.Request) (*http.Response, error)
}

// HTTPRequestMaker is the type of the function that
// creates a new HTTP request for us.
type HTTPRequestMaker = func(
	method, url string, body io.Reader) (*http.Request, error,
)

// Client is an mlabns client.
type Client struct {
	// BaseURL is the optional base URL for contacting mlabns. This is
	// initialized in NewClient, but you may override it.
	BaseURL string

	// Ctx is the context to use.
	Ctx context.Context

	// RequestMaker is the function that creates a request. This is
	// initialized in NewClient, but you may override it.
	RequestMaker HTTPRequestMaker

	// Requestor is the implementation that performs the request. This is
	// initialized in NewClient, but you may override it.
	Requestor HTTPRequestor

	// Tool is the mandatory tool to use.
	Tool string

	// UserAgent is the mandatory user agent to be used.
	UserAgent string
}

// baseURL is the default base URL.
//
// TODO(bassosimone): when ndt7 is deployed on the whole platform, we can
// stop using the staging mlabns service and use the production one.
const baseURL = "https://locate-dot-mlab-staging.appspot.com/"

// NewClient creates a new Client instance with mandatory userAgent, and tool
// name. For running ndt7, use "ndt7" as the tool name.
func NewClient(ctx context.Context, tool, userAgent string) *Client {
	return &Client{
		BaseURL:      baseURL,
		Ctx:          ctx,
		RequestMaker: http.NewRequest,
		Requestor:    http.DefaultClient,
		Tool:         tool,
		UserAgent:    userAgent,
	}
}

// serverEntry describes a mlab server.
type serverEntry struct {
	// FQDN is the the FQDN of the server.
	FQDN string `json:"fqdn"`
}

// ErrQueryFailed indicates a non-200 status code.
var ErrQueryFailed = errors.New("mlabns returned non-200 status code")

// doGET is an internal function used to perform the request.
func (c *Client) doGET(URL string) ([]byte, error) {
	request, err := c.RequestMaker("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", c.UserAgent)
	request = request.WithContext(c.Ctx)
	response, err := c.Requestor.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, ErrQueryFailed
	}
	return ioutil.ReadAll(response.Body)
}

// ErrNoAvailableServers is returned when there are no available servers. A
// background client should treat this error specially and schedule retrying
// after an exponentially distributed number of seconds.
var ErrNoAvailableServers = errors.New("No available M-Lab servers")

// Query returns the FQDN of a nearby mlab server. Returns an error on
// failure and the server FQDN on success.
func (c *Client) Query() (string, error) {
	URL, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", err
	}
	URL.Path = c.Tool
	data, err := c.doGET(URL.String())
	if err != nil {
		return "", err
	}
	var server serverEntry
	err = json.Unmarshal(data, &server)
	if err != nil {
		return "", err
	}
	if server.FQDN == "" {
		return "", ErrNoAvailableServers
	}
	return server.FQDN, nil
}
