package main

import (
	"net/url"
	"os"
	"testing"

	"github.com/m-lab/go/testingx"
	"github.com/m-lab/locate/api/locate"
	"github.com/m-lab/ndt-server/ndt7/ndt7test"
	"github.com/m-lab/ndt7-client-go/internal/params"
)

func TestNormalUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	// Create local ndt7test server.
	h, fs := ndt7test.NewNDT7Server(t)
	defer os.RemoveAll(h.DataDir)
	defer fs.Close()
	u, err := url.Parse(fs.URL)
	testingx.Must(t, err, "failed to parse ndt7test server url")
	// Setup flags to use the service-url option.
	flagScheme.Value = "ws"
	flagService.URL = &url.URL{
		Scheme: "ws",
		Host:   u.Host,
		Path:   params.DownloadURLPath,
	}

	exitval := 0
	savedFunc := osExit
	osExit = func(code int) {
		exitval = code
	}
	savedArgs := osArgs
	osArgs = []string{"ndt7-client"}
	main()
	flagService.URL = nil
	osExit = savedFunc
	osArgs = savedArgs
	if exitval != 0 {
		t.Fatal("expected zero return code here")
	}
}

func TestQuietUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	// Create local ndt7test server.
	h, fs := ndt7test.NewNDT7Server(t)
	defer os.RemoveAll(h.DataDir)
	defer fs.Close()
	u, err := url.Parse(fs.URL)
	testingx.Must(t, err, "failed to parse ndt7test server url")
	// Setup flags to use the service-url option.
	flagScheme.Value = "ws"
	flagService.URL = &url.URL{
		Scheme: "ws",
		Host:   u.Host,
		Path:   params.UploadURLPath,
	}

	exitval := 0
	savedFunc := osExit
	osExit = func(code int) {
		exitval = code
	}
	savedArgs := osArgs
	osArgs = []string{"ndt7-client"}
	*flagQuiet = true
	main()
	flagService.URL = nil
	*flagQuiet = false
	osExit = savedFunc
	osArgs = savedArgs
	if exitval != 0 {
		t.Fatal("expected zero return code here")
	}
}

func TestBatchUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	// Create local ndt7test server.
	h, fs := ndt7test.NewNDT7Server(t)
	defer os.RemoveAll(h.DataDir)
	defer fs.Close()
	u, err := url.Parse(fs.URL)
	testingx.Must(t, err, "failed to parse ndt7test server url")
	// Setup flags to use the service-url option.
	flagService.URL = &url.URL{
		Scheme: "ws",
		Host:   u.Host,
		Path:   "this-is-a-bad-path",
	}
	flagScheme.Value = "ws"
	*flagServer = u.Host
	exitval := 0
	savedFunc := osExit
	osExit = func(code int) {
		exitval = code
	}
	savedArgs := osArgs
	osArgs = []string{"ndt7-client"}
	*flagBatch = true
	main()
	*flagBatch = false
	flagService.URL = nil
	osExit = savedFunc
	osArgs = savedArgs
	if exitval != 0 {
		t.Fatal("expected zero return code here")
	}
}

func TestDownloadError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	exitval := 0
	savedFunc := osExit
	osExit = func(code int) {
		exitval = code
	}
	savedArgs := osArgs
	osArgs = []string{"ndt7-client"}
	*flagServer = "\t" // fail parsing
	main()
	*flagServer = ""
	osExit = savedFunc
	osArgs = savedArgs
	if exitval == 0 {
		t.Fatal("expected nonzero return code here")
	}
}

func TestClientFactory_WithTokenNoURL(t *testing.T) {
	// Just save the two flags we're testing
	origToken := *flagLocateToken
	origURL := *flagLocateURL
	defer func() {
		*flagLocateToken = origToken
		*flagLocateURL = origURL
	}()

	*flagLocateToken = "test-jwt"
	*flagLocateURL = ""

	c := clientFactory()

	// Type assert to get the concrete locate.Client
	loc, ok := c.Locate.(*locate.Client)
	if !ok {
		t.Fatalf("expected *locate.Client, got %T", c.Locate)
	}

	if loc.Authorization != "test-jwt" {
		t.Errorf("got auth %q, want %q", loc.Authorization, "test-jwt")
	}
	if loc.BaseURL.Path != "/v2/priority/nearest" {
		t.Errorf("got path %q, want %q", loc.BaseURL.Path, "/v2/priority/nearest")
	}
}

func TestClientFactory_NoTokenNoURL(t *testing.T) {
	origToken := *flagLocateToken
	origURL := *flagLocateURL
	defer func() {
		*flagLocateToken = origToken
		*flagLocateURL = origURL
	}()

	*flagLocateToken = ""
	*flagLocateURL = ""

	c := clientFactory()

	// Type assert to get the concrete locate.Client
	loc, ok := c.Locate.(*locate.Client)
	if !ok {
		t.Fatalf("expected *locate.Client, got %T", c.Locate)
	}

	if loc.Authorization != "" {
		t.Errorf("got auth %q, want empty", loc.Authorization)
	}
	if loc.BaseURL.Path != "/v2/nearest" {
		t.Errorf("got path %q, want %q", loc.BaseURL.Path, "/v2/nearest")
	}
}

func TestClientFactory_WithCustomURL(t *testing.T) {
	origToken := *flagLocateToken
	origURL := *flagLocateURL
	defer func() {
		*flagLocateToken = origToken
		*flagLocateURL = origURL
	}()

	*flagLocateToken = "test-jwt"
	*flagLocateURL = "http://custom.example.com/my/path"

	c := clientFactory()

	// Type assert to get the concrete locate.Client
	loc, ok := c.Locate.(*locate.Client)
	if !ok {
		t.Fatalf("expected *locate.Client, got %T", c.Locate)
	}

	if loc.Authorization != "test-jwt" {
		t.Errorf("got auth %q, want %q", loc.Authorization, "test-jwt")
	}
	if loc.BaseURL.Path != "/my/path" {
		t.Errorf("got path %q, want %q", loc.BaseURL.Path, "/my/path")
	}
}
