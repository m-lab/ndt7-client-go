package internal

// Summary is a struct containing the values displayed to the user at
// the end of an ndt7 test.
type Summary struct {
	// Download speed, in Bytes/s. This is measured at the receiver.
	Download float64
	// Upload speed, in Bytes/s. This is measured at the receiver.
	Upload float64
	// Retransmission rate. This is based on the TCPInfo values provided
	// by the server during a download test.
	DownloadRetrans float64
	// Round-trip time of the latest measurement, in microseconds.
	// This is provided by the server during a download test.
	RTT int64
}
