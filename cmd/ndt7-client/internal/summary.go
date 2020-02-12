package internal

import (
	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/spec"
)

// Summary is a struct containing the values displayed to the user at
// the end of an ndt7 test.
type Summary struct {
	ServerFQDN string `json:"server"`
	// Download speed, in Mbit/s. This is measured at the receiver.
	Download float64 `json:"download,omitempty"`
	// Upload speed, in Mbit/s. This is measured at the receiver.
	Upload float64 `json:"upload,omitempty"`
	// Retransmission rate. This is based on the TCPInfo values provided
	// by the server during a download test.
	DownloadRetrans float64 `json:"downloadRetrans,omitempty"`
	// Round-trip time of the latest measurement, in microseconds.
	// This is provided by the server during a download test.
	RTT uint32 `json:"rtt,omitempty"`
}

// NewSummary creates a new Summary struct based on the ndt7 test results
// provided by the Client.
func NewSummary(FQDN string,
	results map[spec.TestKind]*ndt7.MeasurementPair) *Summary {

	s := &Summary{
		ServerFQDN: FQDN,
	}

	// Download comes from the client-side Measurement during the download
	// test. DownloadRetrans and RTT come from the server-side Measurement,
	// if it includes a TCPInfo object.
	if dl, ok := results[spec.TestDownload]; ok {
		if dl.Client.AppInfo != nil && dl.Client.AppInfo.ElapsedTime > 0 {
			elapsed := float64(dl.Client.AppInfo.ElapsedTime) / 1e06
			s.Download = (8.0 * float64(dl.Client.AppInfo.NumBytes)) /
				elapsed / (1000.0 * 1000.0)
		}
		if dl.Server.TCPInfo != nil && dl.Server.TCPInfo.BytesSent > 0 {
			s.DownloadRetrans = float64(dl.Server.TCPInfo.BytesRetrans) / float64(dl.Server.TCPInfo.BytesSent)
		}
		s.RTT = dl.Server.TCPInfo.RTT
	}
	// Upload comes from the client-side Measurement during the upload test.
	if ul, ok := results[spec.TestUpload]; ok {
		if ul.Client.AppInfo != nil && ul.Client.AppInfo.ElapsedTime > 0 {
			elapsed := float64(ul.Client.AppInfo.ElapsedTime) / 1e06
			s.Upload = (8.0 * float64(ul.Client.AppInfo.NumBytes)) /
				elapsed / (1000.0 * 1000.0)
		}
	}

	return s
}
