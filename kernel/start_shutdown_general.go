//go:build !wasm

package kernel

import "log"

// Wird ausgeführt wenn das Programm als Dienst ausgeführt wird
func (obj *Kernel) Start() error {
	obj._lock.Lock()
	for i := range obj._api_interfaces {
		if err := obj._api_interfaces[i]._start_by_kernel(); err != nil {
			obj._lock.Unlock()
			return err
		}
	}
	obj._is_running = true
	obj._lock.Unlock()
	return nil
}

// Wird ausgeführt um den Kernel zu beenden
func (obj *Kernel) Shutdown() {
	obj._lock.Lock()
	if obj._is_running {
		// Die Internen Dienste werden beendet
		var vat ServerModule
		for _, item := range obj._server_modules {
			if item == nil {
				continue
			}
			vat = *item
			vat.Shutdown()
		}

		// Es werden alle Verbindungen geschlossen
		obj._connection_manager.ShutdownByKernel()

		// Die API Schnitstellen werden geschlossen
		for i := range obj._api_interfaces {
			obj._api_interfaces[i]._close_by_kernel()
		}

		// Die Datenbanken werden geschlossen
		obj._routing_table.Shutdown()
		obj._trusted_relays.Shutdown()

		// Es wird Signalisiert dass der Kernel nicht mehr läuft
		obj._is_running = false

		// Der Threadlock wird freigegeben
		obj._lock.Unlock()

		// Log
		log.Println("Kernel shutdown...")
		return
	}
	obj._lock.Unlock()
}
