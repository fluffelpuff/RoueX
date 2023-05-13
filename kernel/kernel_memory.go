package kernel

import (
	"bytes"
	"fmt"

	"github.com/fluffelpuff/RoueX/utils"
	mem "github.com/pbnjay/memory"
)

// Stellt einen Memory eintrag dar
type mem_entry struct {
}

// Stellt ein Kernel Speicher dar
type kernel_memory struct {
	totalMem       uint64
	reservedBuffer []byte
	data           []mem_entry
}

// Gibt die Gesamtgröße des Speichers an
func (obj *kernel_memory) GetTotalSize() uint64 {
	return obj.totalMem
}

// Gibt die Verwendete Größe an
func (obj *kernel_memory) GetUsedSize() uint64 {
	return 0
}

// Gibt an wieviel Spiecher noch frei ist
func (obj *kernel_memory) GetAvailableSize() uint64 {
	return obj.totalMem - obj.GetUsedSize()
}

// Erzeugt einen neuen Kernel
func new_memory() (*kernel_memory, error) {
	// Aktuelle Größe des Arbeitsspeichers ermitteln
	totalMem := mem.TotalMemory()

	// 10% des Arbeitsspeichers als reservierten Speicherplatz reservieren
	reservedMem := uint64(float64(totalMem) * 0.01)

	// Speicherblock für den Puffer reservieren
	reservedBuffer := bytes.Repeat([]byte{0xff}, int(reservedMem))

	// Ihr eigentlicher Datenbereich initialisieren
	data := make([]mem_entry, 0)

	// Log:
	fmt.Printf("New kernel memory created. total-size = %s \n", utils.FormatSize(uint64(len(reservedBuffer))))

	// Der kernel_memory wird zurückgegeben
	return &kernel_memory{totalMem, reservedBuffer, data}, nil
}
