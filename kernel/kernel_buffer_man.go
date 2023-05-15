package kernel

import (
	"log"
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
	log.Printf("Kernel: buffer io thread started. id = %s\n", kernel._kernel_id)
	for kernel.IsRunning() {
		check_and_handle_package(kernel)
		time.Sleep(1 * time.Millisecond)
	}
	log.Printf("Kernel: buffer io thread closed. id = %s\n", kernel._kernel_id)
}
