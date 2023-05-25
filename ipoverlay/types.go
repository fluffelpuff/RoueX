package ipoverlay

// Gibt den Typen des Paketes an
type TransportPackageType uint8

// Gibt den Protokolltypen an Client Server Protokolles an
type ClientServerP2PProtocol uint8

// Stellt einen Protokoll Flag dar
type ProtocolFlag uint8

// Definiert alle verfügbaren Typen
const (
	// Definiert alle Transportdatentypen
	Ping = TransportPackageType(0)
	Pong = TransportPackageType(1)
	Data = TransportPackageType(2)

	// Definiert alle ClientServerP2P Protokolle
	WS_TCP_V4  = ClientServerP2PProtocol(0)
	WS_TCP_V6  = ClientServerP2PProtocol(1)
	WS_QUIC_v4 = ClientServerP2PProtocol(2)
	WS_QUIC_v6 = ClientServerP2PProtocol(3)
	QUIC_V4    = ClientServerP2PProtocol(4)
	QUIC_V6    = ClientServerP2PProtocol(5)
	TCP_V4     = ClientServerP2PProtocol(6)
	TCP_V6     = ClientServerP2PProtocol(7)
	TORv3      = ClientServerP2PProtocol(8)
	I2PED      = ClientServerP2PProtocol(9)
)
