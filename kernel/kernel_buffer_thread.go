package kernel

import (
	"log"
	"sync"
	"time"
)

// Wird verwendet um Ein und Ausgehende Pakete zu verwalten
func buffer_io_routine(kernel *Kernel) {
	// Dieser Thread wird verwendet um den Aktuellen Status des Readers anzugeben
	thr := new(sync.Mutex)

	// Diese Variable gibt an ob der Thread ausgeführt wird
	thr_running := uint(0)

	// Diese Funktion setzt den Aktuellen Status
	add_thr := func(sbool *uint) {
		thr.Lock()
		*sbool++
		thr.Unlock()
	}

	// Diese Funktion gibt an das ein Thread weniger ausgeführt wird
	rem_thr := func(sbool *uint) {
		thr.Lock()
		*sbool--
		thr.Unlock()
	}

	// Diese Funktion gibt den Aktuellen Status zurück
	gstate := func(sbool *uint) bool {
		return *sbool > 0
	}

	// Diese Funktion wird als Thread ausgeführt
	thr_func := func() {
		// Log
		log.Printf("Kernel: buffer io thread started. id = %s\n", kernel._kernel_id)

		// Signalisiert das Thread ausgeführt wurd
		add_thr(&thr_running)

		// Der Kernel wird ausgeführt
		for kernel.IsRunning() {
			// Das Paket wird abgerufen
			pack := kernel._memory.GetNextPackage()

			// Das Paket wird an den Kernel übergeben
			if err := kernel.EnterLocallyPackage(pack); err != nil {
				log.Println("Kernel: error by reading package. id = "+kernel.GetKernelID(), "error = "+err.Error())
			}
		}

		// Signalisiert das der Thread nicht mehr ausgeführt wird
		rem_thr(&thr_running)

		// Log
		log.Printf("Kernel: buffer io thread closed. id = %s\n", kernel._kernel_id)
	}

	// Es werden 4 Writer Threads gestartet
	go thr_func()

	// Diese Schleife wird für 30MS ausgeführt
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Millisecond)
	}

	// Es wird geprüft ob der Thread gestartet wurde
	if !gstate(&thr_running) {
		panic("buffer handler starting error, unkown error")
	}
}
