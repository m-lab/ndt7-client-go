// Package spec contains constants and structs. See also the spec:
// https://github.com/m-lab/ndt-server/blob/master/spec/ndt7-protocol.md
package spec

import (
	"github.com/m-lab/ndt-server/ndt7/model"
)

type (
	// AppInfo contains an application level measurement.
	AppInfo model.AppInfo

	// BBRInfo contains a BBR measurement.
	BBRInfo model.BBRInfo

	// ConnectionInfo contains info on this connection.
	ConnectionInfo model.ConnectionInfo

	// OriginKind indicates the origin of a measurement.
	OriginKind string

	// TestKind indicates the direction of a measurement.
	TestKind string

	// TCPInfo contains a TCP_INFO measurement.
	TCPInfo model.TCPInfo
)

const (
	// OriginClient indicates that the measurement origin is the client.
	OriginClient = OriginKind("client")

	// OriginServer indicates that the measurement origin is the server.
	OriginServer = OriginKind("server")

	// TestDownload indicates that this is a download.
	TestDownload = TestKind("download")

	// TestUpload indicates that this is an upload.
	TestUpload = TestKind("upload")
)

// The Measurement struct contains measurement results. This message is
// an extension of the one inside of v0.7.0 of the ndt7 spec.
type Measurement struct {
	// AppInfo contains application level measurements.
	AppInfo *AppInfo `json:",omitempty"`

	// BBRInfo is the data measured using TCP BBR instrumentation.
	BBRInfo *BBRInfo `json:",omitempty"`

	// ConnectionInfo contains info on the connection.
	ConnectionInfo *ConnectionInfo `json:",omitempty"`

	// Origin indicates who performed this measurement.
	Origin OriginKind `json:",omitempty"`

	// Test contains the test name.
	Test TestKind `json:",omitempty"`

	// TCPInfo contains metrics measured using TCP_INFO instrumentation.
	TCPInfo *TCPInfo `json:",omitempty"`
}
