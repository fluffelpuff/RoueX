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
}

// Stellt eine Verbindung dar
type RelayConnection interface {
	RegisterKernel(kernel *Kernel) error
	GetProtocol() string
	GetObjectId() string
	FinallyInit() error
	Write([]byte) error
	IsConnected() bool
	CloseByKernel()
}

// Stellt die Module Funktionen bereit
type ExternalModule interface {
	Info() error
}

// Überträgt Verschlüsselte Daten
type _aes_encrypted_result struct {
	Cipher []byte `cbor:"1,keyasint"`
	Nonce  []byte `cbor:"2,keyasint"`
	Sig    []byte `cbor:"3,keyasint"`
}

// Gibt den Verschlüsselungs Algo an
type EncryptionAlgo uint8

const (
	CHACHA_2020 = EncryptionAlgo(1)
)
