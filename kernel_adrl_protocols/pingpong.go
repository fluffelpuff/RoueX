package kernelprotocols

import (
	"fmt"
	"log"
	"sync"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/fluffelpuff/RoueX/kernel"
	"github.com/fluffelpuff/RoueX/utils"
)

type ROUEX_PING_PONG_PROTOCOL struct {
	_objid  string
	_kernel *kernel.Kernel
	_lock   *sync.Mutex
}

// Nimmt eingetroffene Pakete aus dem Netzwerk Entgegen
func (obj *ROUEX_PING_PONG_PROTOCOL) EnterRecivedPackage(pckage *kernel.AddressLayerPackage, conn kernel.RelayConnection) error {
	return nil
}

// Nimmt Datensätze entgegen und übergibt diese an den Kernel um das Paket entgültig abzusenden
func (obj *ROUEX_PING_PONG_PROTOCOL) EnterWritableBytesToReciver(data []byte, reciver *btcec.PublicKey) error {
	return nil
}

// Nimmt eintreffende Steuer Befehele entgegen
func (obj *ROUEX_PING_PONG_PROTOCOL) EnterCommandData(data []byte) ([]byte, error) {
	return nil, nil
}

// Registriert den Kernel im Protokoll
func (obj *ROUEX_PING_PONG_PROTOCOL) RegisterKernel(kernel *kernel.Kernel) error {
	obj._lock.Lock()
	if obj._kernel != nil {
		obj._lock.Unlock()
		return fmt.Errorf("kernel always registered")
	}
	obj._kernel = kernel
	obj._lock.Unlock()
	log.Println("ROUEX_PING_PONG_PROTOCOL: kernel registrated. id =", kernel.GetKernelID(), "object-id =", obj._objid)
	return nil
}

// Gibt den Namen des Protokolles zurück
func (obj *ROUEX_PING_PONG_PROTOCOL) GetProtocolName() string {
	return "ROUEX_PING_PONG_PROTOCOL"
}

// Gibt die ObjektID des Protokolls zurück
func (obj *ROUEX_PING_PONG_PROTOCOL) GetObjectId() string {
	return obj._objid
}

// Erzeugt ein neues PING PONG Protokoll
func NEW_ROUEX_PING_PONG_PROTOCOL_HANDLER() *ROUEX_PING_PONG_PROTOCOL {
	return &ROUEX_PING_PONG_PROTOCOL{_lock: &sync.Mutex{}, _objid: utils.RandStringRunes(12)}
}
