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
	pckge  *addresspackages.AddressLayerPackage
}

// Stellt ein Kernel Speicher dar
type kernel_package_buffer struct {
	thrLock *sync.Mutex
	buffer  chan *kernel_package_buffer_entry
}

// Nimmt ein eintreffendes Paket entgegen
func (obj *kernel_package_buffer) AddL2Package(pckge *addresspackages.AddressLayerPackage) (*extra.PackageSendState, error) {
	// Es wird ein neuer Status erstellt
	new_sstate := extra.NewPackageSendState()

	// Das Paket wird zwischengespeichert
	entry := &kernel_package_buffer_entry{sstate: new_sstate, pckge: pckge}
	obj.buffer <- entry

	// Log
	log.Println("kernel_package_buffer: add package to buffer. phash = " + hex.EncodeToString(pckge.GetPackageHash()))

	// Der Vorgang wurde ohne Fehler durchgeführt
	return new_sstate, nil
}

// Gibt das nächste Paket aus dem Buffer zurück
func (obj *kernel_package_buffer) GetNextPackage() *addresspackages.AddressLayerPackage {
	// Das nächste Paket aus dem Buffer wird abgerufen
	retiv := <-obj.buffer

	// Log
	log.Println("kernel_package_buffer: retrive package from buffer. phash = " + hex.EncodeToString(retiv.pckge.GetPackageHash()))

	// Die Daten werden zurückgegeben
	return retiv.pckge
}

// Erzeugt einen neuen Kernel
func new_kernel_package_buffer() (*kernel_package_buffer, error) {
	// Log:
	fmt.Printf("New kernel package buffer created.\n")

	// Der kernel_package_buffer wird zurückgegeben
	return &kernel_package_buffer{new(sync.Mutex), make(chan *kernel_package_buffer_entry, local_package_buffer_max)}, nil
}
