package internal

import (
	"strings"

	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/spec"
)

// valueUnitPair represents a {"Value": ..., "Unit": ...} pair.
type valueUnitPair struct {
	Value float64
	Unit  string
}

// Summary is a struct containing the values displayed to the user at
// the end of an ndt7 test.
type Summary struct {
	// FQDN of the server used for this test.
	Server string

	// IP address of the client.
	Client string

	// Download speed, in Mbit/s. This is measured at the receiver.
	Download valueUnitPair

	// Upload speed, in Mbit/s. This is measured at the sender.
	Upload valueUnitPair

	// Retransmission rate. This is based on the TCPInfo values provided
	// by the server during a download test.
	DownloadRetrans valueUnitPair

	// Round-trip time of the latest measurement, in milliseconds.
	// This is provided by the server during a download test.
	RTT valueUnitPair
}

// NewSummary creates a new Summary struct based on the ndt7 test results
// provided by the Client.
func NewSummary(FQDN string,
	results map[spec.TestKind]*ndt7.TestData) *Summary {

	s := &Summary{
		Server: FQDN,
	}

	if results[spec.TestDownload].ConnectionInfo != nil {
		endpoint := strings.Split(
			results[spec.TestDownload].ConnectionInfo.Client, ":")
		s.Client = endpoint[0]
	}

	// Download comes from the client-side Measurement during the download
	// test. DownloadRetrans and RTT come from the server-side Measurement,
	// if it includes a TCPInfo object.
	if dl, ok := results[spec.TestDownload]; ok {
		if dl.Client.AppInfo != nil && dl.Client.AppInfo.ElapsedTime > 0 {
			elapsed := float64(dl.Client.AppInfo.ElapsedTime) / 1e06
			s.Download = valueUnitPair{
				Value: (8.0 * float64(dl.Client.AppInfo.NumBytes)) /
					elapsed / (1000.0 * 1000.0),
				Unit: "Mbit/s",
			}
		}
		if dl.Server.TCPInfo != nil {
			if dl.Server.TCPInfo.BytesSent > 0 {
				s.DownloadRetrans = valueUnitPair{
					Value: float64(dl.Server.TCPInfo.BytesRetrans) / float64(dl.Server.TCPInfo.BytesSent) * 100,
					Unit:  "%",
				}
			}
			s.RTT = valueUnitPair{
				Value: float64(dl.Server.TCPInfo.RTT) / 1000,
				Unit:  "ms",
			}
		}
	}
	// Upload comes from the client-side Measurement during the upload test.
	if ul, ok := results[spec.TestUpload]; ok {
		if ul.Client.AppInfo != nil && ul.Client.AppInfo.ElapsedTime > 0 {
			elapsed := float64(ul.Client.AppInfo.ElapsedTime) / 1e06
			s.Upload = valueUnitPair{
				Value: (8.0 * float64(ul.Client.AppInfo.NumBytes)) /
					elapsed / (1000.0 * 1000.0),
				Unit: "Mbit/s",
			}
		}
	}

	return s
}
