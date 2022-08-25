package emitter

// ValueUnitPair represents a {"Value": ..., "Unit": ...} pair.
type ValueUnitPair struct {
	Value float64
	Unit  string
}

type SubtestSummary struct {
	UUID       string
	Throughput ValueUnitPair
	// Latency is the MinRTT value of the latest measurement, in milliseconds.
	// For uploads, this is provided by the server.
	Latency ValueUnitPair
	// Retransmission is BytesRetrans / BytesSent from TCPInfo
	Retransmission ValueUnitPair
}

// Summary is a struct containing the values displayed to the user at
// the end of an ndt7 test.
type Summary struct {
	// ServerFQDN is the FQDN of the server used for this test.
	ServerFQDN string

	// ServerIP is the (v4 or v6) IP address of the server.
	ServerIP string

	// ClientIP is the (v4 or v6) IP address of the client.
	ClientIP string

	Download *SubtestSummary
	Upload   *SubtestSummary
}

// NewSummary returns a new Summary struct for a given FQDN.
func NewSummary(FQDN string) *Summary {
	return &Summary{
		ServerFQDN: FQDN,
	}
}
