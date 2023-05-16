package kernel

import (
	"log"
	"sync"
	"time"
)

// Diese Funktion prüft ob ein Paket im Buffer vorhanden ist, wenn ja wird das Paket weiterverabeitet
func check_and_handle_package(kernel *Kernel) {
	for kernel._memory.HasPackages() {
		// Das Paket wird abgerufen
		pack := kernel._memory.GetNextPackage()
		if pack == nil {
			return
		}

		// Das Paket wird an den Kernel übergeben
		if err := kernel.EnterLocallyPackage(pack); err != nil {
			log.Println("Kernel: error by reading package. id = "+kernel.GetKernelID(), "error = "+err.Error())
		}

		// Es wird 100 Nanosekunden gewartet
		time.Sleep(100 * time.Nanosecond)
	}
}

// Wird verwendet um Ein und Ausgehende Pakete zu verwalten
func buffer_io_routine(kernel *Kernel) {
	// Dieser Thread wird verwendet um den Aktuellen Status des Readers anzugeben
	thr := new(sync.Mutex)

	// Diese Variable gibt an ob der Thread ausgeführt wird
	thr_running := false

	// Diese Funktion setzt den Aktuellen Status
	sstate := func(sbool *bool, state bool) {
		thr.Lock()
		*sbool = state
		thr.Unlock()
	}

	// Diese Funktion gibt den Aktuellen Status zurück
	gstate := func(sbool *bool) bool {
		return *sbool
	}

	// Diese Funktion wird als Thread ausgeführt
	go func() {
		// Log
		log.Printf("Kernel: buffer io thread started. id = %s\n", kernel._kernel_id)

		// Signalisiert das Thread ausgeführt wurd
		sstate(&thr_running, true)

		// Der Kernel wird ausgeführt
		for kernel.IsRunning() {
			// Es wird geprüf ob ein Paket vorhanden ist
			check_and_handle_package(kernel)

			// Es wird 1 MS gewartet bevor erneut auf verfügbare Paket geprüft wird
			time.Sleep(1 * time.Millisecond)
		}

		// Signalisiert das der Thread nicht mehr ausgeführt wird
		sstate(&thr_running, false)

		// Log
		log.Printf("Kernel: buffer io thread closed. id = %s\n", kernel._kernel_id)
	}()

	// Diese Schleife wird für 30MS ausgeführt
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Millisecond)
	}

	// Es wird geprüft ob der Thread gestartet wurde
	if !gstate(&thr_running) {
		panic("buffer handler starting error, unkown error")
	}
}
