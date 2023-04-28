package kernel

import "github.com/btcsuite/btcd/btcec/v2"

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

// Stellt ein Client Modul dar
type ClientModule interface {
	GetObjectId() string
	RegisterKernel(kernel *Kernel) error
	GetMetaDataInfo() ClientModuleMetaData
	ConnectTo(string, *btcec.PublicKey) error
	GetProtocol() string
	Shutdown()
}

// Stellt eine Verbindung dar
type RelayConnection interface {
	IsConnected() bool
	RegisterKernel(kernel *Kernel) error
	Read() ([]byte, error)
	Write([]byte) error
}

// Stellt die Module Funktionen bereit
type ExternalModule interface {
	Info() error
}
