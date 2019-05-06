// Package spec contains constants and structs. See also the spec:
// https://github.com/m-lab/ndt-server/blob/master/spec/ndt7-protocol.md
package spec

import (
	"time"
)

// SecWebSocketProtocol is the value of the Sec-WebSocket-Protocol header.
const SecWebSocketProtocol = "net.measurementlab.ndt.v7"

// MaxMessageSize is the maximum accepted message size.
const MaxMessageSize = 1 << 20

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

// BulkMessageSize is the size of uploaded messages
const BulkMessageSize = 1 << 13

// UpdateInterval is the interval between client side upload measurements.
const UpdateInterval = 250 * time.Millisecond

// OriginKind indicates the origin of a measurement.
type OriginKind string

// DirectionKind indicates the direction of a measurement.
type DirectionKind string

const (
	// OriginClient indicates that the measurement origin is the client.
	OriginClient = OriginKind("client")

	// OriginServer indicates that the measurement origin is the server.
	OriginServer = OriginKind("server")

	// DirectionDownload indicates that this is a download.
	DirectionDownload = DirectionKind("download")

	// DirectionUpload indicates that this is a upload.
	DirectionUpload = DirectionKind("upload")
)

// AppInfo contains an application level measurement. This message is
// consistent with v0.7.0 of the ndt7 spec.
type AppInfo struct {
	// NumBytes is the number of bytes transferred so far.
	NumBytes int64 `json:"num_bytes"`
}

// The TCPInfo struct contains information measured using TCP_INFO. This
// message is consistent with v0.7.0 of the ndt7 spec.
type TCPInfo struct {
	// SmoothedRTT is the smoothed RTT in milliseconds.
	SmoothedRTT float64 `json:"smoothed_rtt"`

	// RTTVar is the RTT variance in milliseconds.
	RTTVar float64 `json:"rtt_var"`
}

// The BBRInfo struct contains information measured using BBR. This
// message is consistent with v0.7.0 of the ndt7 spec.
type BBRInfo struct {
	// MaxBandwidth is the max bandwidth measured by BBR in bits per second.
	MaxBandwidth int64 `json:"max_bandwidth"`

	// MinRTT is the min RTT measured by BBR in milliseconds.
	MinRTT float64 `json:"min_rtt"`
}

// The Measurement struct contains measurement results. This message is
// an extension of the one inside of v0.7.0 of the ndt7 spec.
type Measurement struct {
	// AppInfo contains application level measurements.
	AppInfo AppInfo `json:"app_info"`

	// BBRInfo is the data measured using TCP BBR instrumentation.
	BBRInfo BBRInfo `json:"bbr_info"`

	// Direction indicates the measurement direction. This field is
	// an extension with respect to the spec.
	Direction DirectionKind `json:"direction"`

	// Elapsed is the number of seconds elapsed since the beginning.
	Elapsed float64 `json:"elapsed"`

	// Origin is either OriginClient or OriginServer. This file is
	// an extension with respect to the spec.
	Origin OriginKind `json:"origin"`

	// TCPInfo contains metrics measured using TCP_INFO instrumentation.
	TCPInfo TCPInfo `json:"tcp_info"`
}
