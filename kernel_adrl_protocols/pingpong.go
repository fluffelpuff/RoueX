package kernelprotocols

import (
	"fmt"
	"log"
	"sync"

	"github.com/fluffelpuff/RoueX/kernel"
	"github.com/fluffelpuff/RoueX/utils"
)

type ROUEX_PING_PONG_PROTOCOL struct {
	_objid  string
	_kernel *kernel.Kernel
	_lock   *sync.Mutex
}

// Nimmt eingetroffene Pakete entegegen
func (obj *ROUEX_PING_PONG_PROTOCOL) EnterRecivedPackage(pckage *kernel.AddressLayerPackage, conn kernel.RelayConnection) error {
	return nil
}

// Nimmt Datensätze entgegen welche

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
