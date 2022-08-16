package main

import (
	"net/url"
	"os"
	"testing"

	"github.com/m-lab/go/testingx"
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
	main()
	flagService.URL = nil
	osExit = savedFunc
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
	*flagQuiet = true
	main()
	flagService.URL = nil
	*flagQuiet = false
	osExit = savedFunc
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
	*flagBatch = true
	main()
	*flagBatch = false
	flagService.URL = nil
	osExit = savedFunc
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
	*flagServer = "\t" // fail parsing
	main()
	*flagServer = ""
	osExit = savedFunc
	if exitval == 0 {
		t.Fatal("expected nonzero return code here")
	}
}

