package apiclient

type EmptyArg struct{}

type ApiRelayEntry struct {
	Id                string
	IsConnected       bool
	PublicKey         string
	IsTrusted         bool
	TotalConnections  uint64
	TotalBytesSend    uint64
	TotalBytesRecived uint64
	PingMS            uint64
	BandwithKBs       uint64
}
