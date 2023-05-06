package apiclient

type EmptyArg struct{}

type ApiRelayConnection struct {
	Id              string
	Protocol        string
	SessionPkey     string
	InboundOutbound uint8
	TxBytes         uint64
	RxBytes         uint64
	Ping            uint64
}

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
	Connections       []ApiRelayConnection
}
