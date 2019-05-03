// Package mlabns contains an mlabns implementation.
package mlabns

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Config contains mlabns settings.
type Config struct {
	// BaseURL is the optional base URL for contacting mlabns.
	BaseURL string

	// Tool is the mandatory tool to use.
	Tool string

	// UserAgent is the mandatory user agent to be used.
	UserAgent string
}

// BaseURL is the default base URL.
//
// TODO(bassosimone): when ndt7 is deployed on the whole platform, we can
// stop using the staging mlabns service and use the production one.
const BaseURL = "https://locate-dot-mlab-staging.appspot.com/"

// NewConfig creates a new Config instance with mandatory userAgent.
// name. For running ndt7, use "ndt7" as the tool name.
func NewConfig(tool, userAgent string) Config {
	return Config{
		BaseURL:   BaseURL,
		Tool:      tool,
		UserAgent: userAgent,
	}
}

// serverEntry describes a mlab server.
type serverEntry struct {
	// FQDN is the the FQDN of the server.
	FQDN string `json:"fqdn"`
}

// ErrQueryFailed indicates a non-200 status code.
var ErrQueryFailed = errors.New("mlabns returned non-200 status code")

// httpNewRequest allows to mock http.NewRequest
var httpNewRequest = http.NewRequest

// httpClientDo allows to mock httpClient.Do
var httpClientDo = func(client *http.Client, req *http.Request) (*http.Response, error) {
	return client.Do(req)
}

// doGET is an internal function used to perform the request.
var doGET = func(ctx context.Context, URL, userAgent string) ([]byte, error) {
	request, err := httpNewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", userAgent)
	request = request.WithContext(ctx)
	response, err := httpClientDo(http.DefaultClient, request)
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
// failure and the FQDN on success. Note that the FQDN may be an empty string
// when mlab is overloaded. So, make sure you handle this case.
func Query(ctx context.Context, config Config) (string, error) {
	URL, err := url.Parse(config.BaseURL)
	if err != nil {
		return "", err
	}
	URL.Path = config.Tool
	data, err := doGET(ctx, URL.String(), config.UserAgent)
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
