package kernel

import (
	"github.com/btcsuite/btcd/btcec/v2"
	addresspackages "github.com/fluffelpuff/RoueX/address_packages"
	"github.com/fluffelpuff/RoueX/kernel/extra"
)

// Stellt das Gerüst für ein Server Modul dar
type ServerModule interface {
	RegisterKernel(kernel *Kernel) error
	GetObjectId() string
	GetProtocol() string
	IsRunning() bool
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
	EnterSendableData([]byte) (*extra.PackageSendState, error)
	GetSessionPKey() (*btcec.PublicKey, error)
	RegisterKernel(kernel *Kernel) error
	GetTxRxBytes() (uint64, uint64)
	GetIOType() ConnectionIoType
	CannUseToWrite() bool
	GetPingTime() uint64
	GetProtocol() string
	GetObjectId() string
	FinallyInit() error
	Write([]byte) error
	IsConnected() bool
	IsFinally() bool
	CloseByKernel()
}

// Stellt die Module Funktionen bereit
type ExternalModule interface {
	GetName() string
	GetVersion() uint64
}

// Gibt die Registrierte Paketfunktion an
type KernelTypeProtocol interface {
	EnterRecivedPackage(*addresspackages.PreAddressLayerPackage) error
	EnterCommandData(string, [][]byte, *APIProcessConnectionWrapper) (map[string]interface{}, error)
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
	Connections []RelayConnectionMetaData
	PublicKey   string
	IsConnected bool
	TotalWrited uint64
	TotalReaded uint64
	PingMS      uint64
	IsTrusted   bool
}

// Gibt an ob es sich um eine Eingehende oder um eine Ausgehende Verbindung handelt
type ConnectionIoType uint8

// Definiert ein oder ausgehende Verbindungestypen
const (
	INBOUND  = ConnectionIoType(1)
	OUTBOUND = ConnectionIoType(2)
)
