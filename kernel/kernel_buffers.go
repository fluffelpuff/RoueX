package kernel

import (
	"encoding/hex"
	"fmt"
	"log"
	"sync"
)

const (
	local_package_buffer_max uint = 1024 // Gibt an, wieviele Pakete der Kernel Paket Buffer zwischenspeichern darf
)

// Stellt ein Kernel Speicher dar
type kernel_package_buffer struct {
	thrLock *sync.Mutex
	data    []*PlainAddressLayerPackage
}

// Nimmt ein eintreffendes Paket entgegen
func (obj *kernel_package_buffer) AddL2Package(pckge *PlainAddressLayerPackage) error {
	// Der Threadlock wird gesperrt
	obj.thrLock.Lock()

	// Sollten mehr als "local_package_buffer_max" Pakete im Buffer liegen, wird das erste Entfernt
	if len(obj.data) >= int(local_package_buffer_max) {
		obj.data = append(obj.data[:0], obj.data[1:]...)
	}

	// Das Paket wird zwischengespeichert
	obj.data = append(obj.data, pckge)

	// Der Threadlock wird freigegeben
	obj.thrLock.Unlock()

	// Log
	log.Println("kernel_package_buffer: add package to buffer. phash = " + hex.EncodeToString(pckge.GetSignHash()))

	// Der Vorgang wurde ohne Fehler durchgeführt
	return nil
}

// Gibt das nächste Paket aus dem Buffer zurück
func (obj *kernel_package_buffer) GetNextPackage() *PlainAddressLayerPackage {
	// Der Threadlock wird gesperrt
	obj.thrLock.Lock()

	// Es wird geprüft ob es einen Dateneintrag gibt
	if len(obj.data) < 1 {
		obj.thrLock.Unlock()
		return nil
	}

	// Das Objekt wird aus der Warteschlange abgerufen
	retriv := obj.data[0]

	// Das Paket wird aus der Warteschlang enfertn
	obj.data = append(obj.data[:0], obj.data[1:]...)

	// Der Threadlock wird freigegeben
	obj.thrLock.Unlock()

	// Die Daten werden zurückgegeben
	return retriv
}

// Gibt an, ob es derzeit Pakete im Buffer gibt
func (obj *kernel_package_buffer) HasPackages() bool {
	obj.thrLock.Lock()
	r := len(obj.data) > 0
	obj.thrLock.Unlock()
	return r
}

// Löscht alle Verfügbaren Pakete
func (obj *kernel_package_buffer) Flush() {
	obj.thrLock.Lock()
	obj.data = make([]*PlainAddressLayerPackage, 0)
	obj.thrLock.Unlock()
}

// Erzeugt einen neuen Kernel
func new_kernel_package_buffer() (*kernel_package_buffer, error) {
	// Ihr eigentlicher Datenbereich initialisieren
	data := make([]*PlainAddressLayerPackage, 0)

	// Log:
	fmt.Printf("New kernel package buffer created.\n")

	// Der kernel_package_buffer wird zurückgegeben
	return &kernel_package_buffer{new(sync.Mutex), data}, nil
}
