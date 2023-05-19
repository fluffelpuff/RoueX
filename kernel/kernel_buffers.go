package kernel

import (
	"encoding/hex"
	"fmt"
	"log"
	"sync"

	addresspackages "github.com/fluffelpuff/RoueX/address_packages"
	"github.com/fluffelpuff/RoueX/kernel/extra"
)

const (
	local_package_buffer_max uint = 1024 // Gibt an, wieviele Pakete der Kernel Paket Buffer zwischenspeichern darf
)

// Stellt einen Eintrag dar
type kernel_package_buffer_entry struct {
	sstate *extra.PackageSendState
	pckge  *addresspackages.PreAddressLayerPackage
}

// Stellt ein Kernel Speicher dar
type kernel_package_buffer struct {
	thrLock *sync.Mutex
	data    []*kernel_package_buffer_entry
}

// Nimmt ein eintreffendes Paket entgegen
func (obj *kernel_package_buffer) AddL2Package(pckge *addresspackages.PreAddressLayerPackage) (*extra.PackageSendState, error) {
	// Der Threadlock wird gesperrt
	obj.thrLock.Lock()
	defer obj.thrLock.Unlock()

	// Sollten mehr als "local_package_buffer_max" Pakete im Buffer liegen, wird das erste Entfernt
	if len(obj.data) >= int(local_package_buffer_max) {
		obj.data = append(obj.data[:0], obj.data[1:]...)
	}

	// Es wird ein neuer Status erstellt
	new_sstate := extra.NewPackageSendState()

	// Das Paket wird zwischengespeichert
	entry := &kernel_package_buffer_entry{sstate: new_sstate, pckge: pckge}
	obj.data = append(obj.data, entry)

	// Log
	log.Println("kernel_package_buffer: add package to buffer. phash = " + hex.EncodeToString(pckge.GetPackageHash()))

	// Der Vorgang wurde ohne Fehler durchgeführt
	return new_sstate, nil
}

// Gibt das nächste Paket aus dem Buffer zurück
func (obj *kernel_package_buffer) GetNextPackage() *addresspackages.PreAddressLayerPackage {
	// Der Threadlock wird gesperrt
	obj.thrLock.Lock()
	defer obj.thrLock.Unlock()

	// Es wird geprüft ob es einen Dateneintrag gibt
	if len(obj.data) < 1 {
		return nil
	}

	// Das Objekt wird aus der Warteschlange abgerufen
	retriv := obj.data[0]

	// Der Status wird auf Gesendet gesetzt
	retriv.sstate.SetFinallyState(extra.SEND)

	// Das Paket wird aus der Warteschlang enfertn
	obj.data = append(obj.data[:0], obj.data[1:]...)

	// Log
	log.Println("kernel_package_buffer: retrive package from buffer. phash = " + hex.EncodeToString(retriv.pckge.GetPackageHash()))

	// Die Daten werden zurückgegeben
	return retriv.pckge
}

// Gibt an, ob es derzeit Pakete im Buffer gibt
func (obj *kernel_package_buffer) HasPackages() bool {
	obj.thrLock.Lock()
	defer obj.thrLock.Unlock()
	return len(obj.data) > 0
}

// Löscht alle Verfügbaren Pakete
func (obj *kernel_package_buffer) Flush() {
	obj.thrLock.Lock()
	obj.data = make([]*kernel_package_buffer_entry, 0)
	obj.thrLock.Unlock()
}

// Erzeugt einen neuen Kernel
func new_kernel_package_buffer() (*kernel_package_buffer, error) {
	// Ihr eigentlicher Datenbereich initialisieren
	data := make([]*kernel_package_buffer_entry, 0)

	// Log:
	fmt.Printf("New kernel package buffer created.\n")

	// Der kernel_package_buffer wird zurückgegeben
	return &kernel_package_buffer{new(sync.Mutex), data}, nil
}
