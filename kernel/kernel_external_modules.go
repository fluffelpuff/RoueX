package kernel

import (
	"fmt"
	"log"
	"os"
	"plugin"
	"strings"
)

// Gibt an ob es sich um ein zulässiges Dateiformat handelt
func _is_validate_file_type(tpe string) bool {
	rw := strings.Split(tpe, ".")
	return rw[len(rw)-1] == "so"
}

// Wird verwendet um Third Party oder Externe Kernel Module zu laden
func (obj *Kernel) LoadExternalKernelModules() error {
	// Es wird versucht das Verzeichniss einzulesen
	files, err := os.ReadDir(obj._external_modules_path)
	if err != nil {
		return fmt.Errorf("LoadExternalKernelModules: 1: " + err.Error())
	}

	// Log
	log.Printf("Kernel: loading external Kernel Modules. id = %s path = %s\n", obj._kernel_id, obj._external_modules_path)

	// Die Externen Module werden geladen
	loaded_modules := []*ExternalModule{}
	for _, file := range files {
		// Es wird geprüft ob es sich um ein Verzeichniss handelt
		if file.IsDir() {
			continue
		}

		// Es wird geprüft ob es sich um eine zulässiges Dateiformat handelt
		if !_is_validate_file_type(file.Name()) {
			continue
		}

		fmt.Println("SO FILE")

		// Es wird versucht das Module einzulesen
		plug, err := plugin.Open(obj._external_modules_path + obj._os_path_trimmer + file.Name())
		if err != nil {
			panic(err)
		}

		// Es wird geprüft ob es die Schnittstelle (Variable) "Module" gibt
		lamda_kernel_mod, err := plug.Lookup("Module")
		if err != nil {
			panic(err)
		}

		// Es wird geprüft ob die Eigentliche Module Information vorhanden ist
		extr_module, ok := lamda_kernel_mod.(ExternalModule)
		if !ok {
			panic(fmt.Sprint("Kernel: invalid kernel module cant load. id =", obj._kernel_id))
		}

		fmt.Println("ASOPA")

		// Log
		log.Printf("Kernel: module loaded. id = %s, module = %s, name = %s\n", obj._kernel_id, obj._external_modules_path+obj._os_path_trimmer+file.Name(), extr_module.GetName())

		// Das Module wird der Modules liste hinzugefügt
		loaded_modules = append(loaded_modules, &extr_module)
	}

	// Log
	log.Printf("Kernel: total modules loaded. id = %s, total = %d\n", obj._kernel_id, len(loaded_modules))

	// Der Vorgang wurde ohne Fehler abgeschlossen
	return nil
}

// Wird verwendet um die Externen Kernel Module zu starten
func (obj *Kernel) StartExternalKernelModules() error {
	log.Println("Kernel: starting external kernel modules...")
	return nil
}
