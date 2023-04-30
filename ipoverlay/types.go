package ipoverlay

type TransportPackageType uint8

const (
	Ping = TransportPackageType(0)
	Pong = TransportPackageType(1)
	Data = TransportPackageType(2)
)
