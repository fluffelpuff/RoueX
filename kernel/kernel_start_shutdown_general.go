//go:build !wasm

package kernel

import (
	"log"
	"time"
)

// Gibt an ob der Kernel erfolgreich heruntergefahren wurde
func (obj *Kernel) _is_base_closed() bool {
	// Es wird geprüft ob der Kernel noch ausgeführt wird
	if obj.IsRunning() {
		return false
	}

	// Der Threadlock wird verwendet
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Gibt an wieviele API Interfaces ausgeführt wird
	total_interface := 0

	// Der Status der Interfaces wird ermittelt
	for i := range obj._api_interfaces {
		if obj._api_interfaces[i]._irn() {
			total_interface++
		}
	}

	// Gibt an wieviele Server Module ausgeführt wird
	total_server_modules := 0

	// Die Server Module werden gestartet
	for i := range obj._server_modules {
		if obj._server_modules[i].IsRunning() {
			total_server_modules++
		}
	}

	// Es wird geprüft ob alle Interfaces geschlossen wurden
	if total_interface != 0 || total_server_modules != 0 {
		return false
	}

	// Die Daten werden zurückgegeben
	return true
}

// Gibt an ob der Kernel erfolgreich heruntergefahren wurde
func (obj *Kernel) _is_full_closed() bool {
	// Es wird geprüft ob die basis beendet wurde
	if !obj._is_base_closed() {
		return false
	}

	// Der Threadlock wird verwendet
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob der Shutdown vorgang erfolgreich druchgeführt wurde
	if !obj._shutdown_complete {
		return false
	}

	// Die Daten werden zurückgegeben
	return true
}

// Signalisiert das der Server vollständig heruntergefahren wurde
func (obj *Kernel) _signal_shutdown_complete() {
	// Wird verwendet um den Threadlock zu Sperren
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird Signalisiert dass der Shutdown vollständig war
	obj._shutdown_complete = true
}

// Wird ausgeführt wenn das Programm als Dienst ausgeführt wird
func (obj *Kernel) Start() error {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Der Buffer IO Thread wird gestartet
	buffer_io_routine(obj)

	// Diese Schleife startet alle API Interfaces
	for i := range obj._api_interfaces {
		if err := obj._api_interfaces[i]._start_by_kernel(); err != nil {
			obj._lock.Unlock()
			panic(err)
		}
	}

	// Die Server Module werden gestartet
	for i := range obj._server_modules {
		if err := obj._server_modules[i].Start(); err != nil {
			obj._lock.Unlock()
			panic(err)
		}
	}

	// Signalisiert das der Kernel ausgeführt wird
	obj._is_running = true

	// Der Threadlock wird freigegeben
	obj._lock.Unlock()

	// Der Vorgang wurde ohne Fehler durchgeführt
	return nil
}

// Wird ausgeführt um den Kernel zu beenden
func (obj *Kernel) Shutdown() {
	// Der Threadlock wird verwendet
	obj._lock.Lock()

	// Gibt an ob der Server ausgeführt wurde
	server_was_runed := false

	// Es wird geprüft ob der Kernelausgeführt wird
	if obj._is_running {
		// Log
		log.Printf("Kernel: closing server modules. id = %s\n", obj.GetKernelID())

		// Die Internen Dienste werden beendet
		var vat ServerModule
		for _, item := range obj._server_modules {
			// Es wird geprüft ob der Eintrag nicht null
			if item == nil {
				continue
			}

			// Das Element wird zwischengespeichert
			vat = item

			// Das Modul wird beendet
			vat.Shutdown()
		}

		// Log
		log.Printf("Kernel: server modules closed. id = %s\n", obj.GetKernelID())

		// Der Threadlock wird entsperrt
		obj._lock.Unlock()

		// Es werden alle Verbindungen geschlossen
		obj._connection_manager.ShutdownByKernel()

		// Der Threadlock wird verwendet
		obj._lock.Lock()

		// Log
		log.Printf("Kernel: closing api interfaces. id = %s\n", obj.GetKernelID())

		// Die API Schnitstellen werden geschlossen
		for i := range obj._api_interfaces {
			// Das Interface wird beendet
			obj._api_interfaces[i]._close_by_kernel()
		}

		// Log
		log.Printf("Kernel: api interfaces closed. id = %s\n", obj.GetKernelID())

		// Es wird Signalisiert dass der Kernel nicht mehr läuft
		obj._is_running = false

		// Gibt an das der Server ausgeführt wird
		server_was_runed = true
	}

	// Der Threadlock wird entsperrt
	obj._lock.Unlock()

	// Sollte der Server ausgeführt werden wird gewartet ob alle Dienste beendet wurden
	if server_was_runed {
		// Log
		log.Printf("Kernel: wait of closing. id = %s\n", obj.GetKernelID())

		// Es wird gewartet bis alles geschlossen wurde
		for range time.Tick(1 * time.Millisecond) {
			if obj._is_base_closed() {
				break
			}
		}

		// Log
		log.Printf("Kernel: finally closed. id = %s\n", obj.GetKernelID())

		// Die Datenbanken werden geschlossen
		obj._routing_table.Shutdown()
		obj._trusted_relays.Shutdown()

		// Es wird Signalisiert dass der Kernel vollständig beendet wurde
		obj._signal_shutdown_complete()
	}
}
