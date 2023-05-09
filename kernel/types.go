package kernel

import (
	"github.com/btcsuite/btcd/btcec/v2"
)

// Stellt das Gerüst für ein Server Modul dar
type ServerModule interface {
	RegisterKernel(kernel *Kernel) error
	GetObjectId() string
	GetProtocol() string
	Start() error
	Shutdown()
}

// Gibt den Datentypen an, welcher von einem Client Module zurückgegeben wird
type ClientModuleMetaData struct {
}

// Stellt Proxy Einstellung dar
type ProxyConfig struct {
	Host string
}

// Stellt ein Client Modul dar
type ClientModule interface {
	GetObjectId() string
	RegisterKernel(kernel *Kernel) error
	GetMetaDataInfo() ClientModuleMetaData
	ConnectTo(string, *btcec.PublicKey, *ProxyConfig) error
	GetProtocol() string
	Shutdown()
	Serve()
}

// Stellt eine Verbindung dar
type RelayConnection interface {
	RegisterKernel(kernel *Kernel) error
	GetTxRxBytes() (uint64, uint64)
	GetSessionPKey() (*btcec.PublicKey, error)
	GetIOType() ConnectionIoType
	GetPingTime() uint64
	GetProtocol() string
	GetObjectId() string
	FinallyInit() error
	Write([]byte) error
	IsConnected() bool
	CloseByKernel()
}

// Stellt die Module Funktionen bereit
type ExternalModule interface {
	GetName() string
	GetVersion() uint64
}

// Gibt die Registrierte Paketfunktion an
type PackageTypeFunction interface {
	EnterRecivedPackage(*AddressLayerPackage, RelayConnection) error
	EnterWritableBytesToReciver([]byte, *btcec.PublicKey) error
	EnterCommandData(data []byte) ([]byte, error)
	RegisterKernel(kernel *Kernel) error
	GetProtocolName() string
	GetObjectId() string
}

// Stellt die MetaDaten einer einzelnen Verbindung dar
type RelayConnectionMetaData struct {
	SessionPKey     string
	Id              string
	IsConnected     bool
	Protocol        string
	InboundOutbound uint8
	TxBytes         uint64
	RxBytes         uint64
	Ping            uint64
}

// Stellt die MetaDaten dar
type RelayMetaData struct {
	Connections      []RelayConnectionMetaData
	PublicKey        string
	TotalConnections uint64
	IsConnected      bool
	TotalWrited      uint64
	TotalReaded      uint64
	PingMS           uint64
	BandwithKBs      uint64
	IsTrusted        bool
}

// Gibt an ob es sich um eine Eingehende oder um eine Ausgehende Verbindung handelt
type ConnectionIoType uint8

// Definiert ein oder ausgehende Verbindungestypen
const (
	INBOUND  = ConnectionIoType(1)
	OUTBOUND = ConnectionIoType(2)
)
