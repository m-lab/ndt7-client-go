package emitter

// ValueUnitPair represents a {"Value": ..., "Unit": ...} pair.
type ValueUnitPair struct {
	Value float64
	Unit  string
}

// Summary is a struct containing the values displayed to the user at
// the end of an ndt7 test.
type Summary struct {
	// Server is the FQDN of the server used for this test.
	Server string

	// Client is the IP address of the client.
	Client string

	// Download is the download speed, in Mbit/s. This is measured at the
	// receiver.
	Download ValueUnitPair

	// Upload is the upload speed, in Mbit/s. This is measured at the sender.
	Upload ValueUnitPair

	// DownloadRetrans is the retransmission rate. This is based on the TCPInfo
	// values provided by the server during a download test.
	DownloadRetrans ValueUnitPair

	// RTT is the round-trip time of the latest measurement, in milliseconds.
	// This is provided by the server during a download test.
	RTT ValueUnitPair
}

// NewSummary returns a new Summary struct for a given FQDN.
func NewSummary(FQDN string) *Summary {
	return &Summary{
		Server: FQDN,
	}
}
