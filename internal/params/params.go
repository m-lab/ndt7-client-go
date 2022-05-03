// Package params contains private constants and structs. See also the spec:
// https://github.com/m-lab/ndt-server/blob/master/spec/ndt7-protocol.md
package params

import (
	"time"
)

// SecWebSocketProtocol is the value of the Sec-WebSocket-Protocol header.
const SecWebSocketProtocol = "net.measurementlab.ndt.v7"

// InitialMessageSize is initial size of uploaded messages.
const InitialMessageSize = 1 << 13

// MaxMessageSize is the maximum accepted message size.
const MaxMessageSize = 1 << 20

// ScalingFraction sets the threshold for scaling binary messages. When
// the current binary message size is <= than 1/scalingFactor of the
// amount of bytes sent so far, we scale the message. This is documented
// in the appendix of the ndt7 specification.
const ScalingFraction = 16

// DownloadTimeout is the time after which the download must stop.
const DownloadTimeout = 15 * time.Second

// IOTimeout is the timeout for I/O operations.
const IOTimeout = 7 * time.Second

// DownloadURLPath is the URL path used for the download.
const DownloadURLPath = "/ndt/v7/download"

// UploadURLPath is the URL path used for the download.
const UploadURLPath = "/ndt/v7/upload"

// UploadTimeout is the time after which the upload must stop.
const UploadTimeout = 10 * time.Second

// UpdateInterval is the interval between client side upload measurements.
const UpdateInterval = 250 * time.Millisecond
