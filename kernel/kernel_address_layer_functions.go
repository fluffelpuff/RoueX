package kernel

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/fluffelpuff/RoueX/utils"
)

// Stellt den Eintrag für einen Funktionstypen Hadnler dar
type kernel_package_type_function_entry struct {
	Ptf KernelTypeProtocol
	Tpe uint8
}

// Registtriert einen neuen Kernel Package Type
func (obj *Kernel) RegisterNewKernelTypeProtocol(tpe uint8, pckgtf KernelTypeProtocol) error {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Es wird geprüft ob die Verbindung besteht
	if obj._is_running {
		obj._lock.Unlock()
		return fmt.Errorf("RegisterNewKernelTypeProtocol: 1: kernel is running, aborted")
	}

	// Es wird geprüft ob der Eintrag bereits hinzugefügt wurde
	var has_found *kernel_package_type_function_entry
	for i := range obj._protocols {
		if obj._protocols[i].Tpe == tpe {
			has_found = obj._protocols[i]
			break
		}
	}

	// Sollte ein passender Wert vorhanden sein, wird der Vorgang abgebrochen
	if has_found != nil {
		obj._lock.Unlock()
		return fmt.Errorf("RegisterNewKernelTypeProtocol: 2: type always registred")
	}

	// Das Objekt wird registriert
	nwo := &kernel_package_type_function_entry{Tpe: tpe, Ptf: pckgtf}
	obj._protocols = append(obj._protocols, nwo)

	// Der Threadlock wird freigegeben
	obj._lock.Unlock()

	// Der Kernel wird im Objekt registriert
	if err := nwo.Ptf.RegisterKernel(obj); err != nil {
		return fmt.Errorf("RegisterNewKernelTypeProtocol: 3: " + err.Error())
	}

	// Log
	log.Println("Kernel: new package type handle function registrated. kernel =", obj.GetKernelID(), "type =", tpe, "name =", pckgtf.GetProtocolName(), "object-id =", pckgtf.GetObjectId())

	// Der Vorgang wurde ohne Fehler durchgeführt
	return nil
}

// Prüft ob es für den Typen einen bekannten Paket handler gibt
func (obj *Kernel) GetRegisteredKernelTypeProtocol(tpe uint8) (KernelTypeProtocol, error) {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Es wird geprüft ob die Verbindung besteht
	if obj._is_running {
		obj._lock.Unlock()
		return nil, fmt.Errorf("GetRegisteredKernelTypeProtocol: 1: kernel is running, aborted")
	}

	// Es wird geprüft ob der Eintrag bereits hinzugefügt wurde
	var has_found *kernel_package_type_function_entry
	for i := range obj._protocols {
		if obj._protocols[i].Tpe == tpe {
			has_found = obj._protocols[i]
			break
		}
	}

	// Sollte kein passender Typ gefunden werden, wird der Vorgang abegbrochen
	if has_found == nil {
		obj._lock.Unlock()
		return nil, fmt.Errorf("GetRegisteredKernelTypeProtocol: 2: type always registred")
	}

	// Der Threadlock wird freigegeben
	obj._lock.Unlock()

	// Die Daten werden zurückgegeben
	return has_found.Ptf, nil
}

// Nimmt Lokale Pakete entgegen
func (obj *Kernel) EnterLocallyPackage(pckge *AddressLayerPackage, conn RelayConnection) error {
	// Es wird geprüft ob der Body mindestens 1 Byte groß ist
	if len(pckge.Body) < 1 {
		return fmt.Errorf("EnterLocallyPackage: 1: Invalid layer two package recived")
	}

	// Es wird geprüft ob es sich um eine Registrierte Paket Funktion handelt
	register_package_type_handler, err := obj.GetRegisteredKernelTypeProtocol(pckge.Type)
	if err != nil {
		return fmt.Errorf("EnterLocallyPackage: 4: " + err.Error())
	}

	// Sollte kein Paket Type handler vorhanden sein, wird das Paket verworfen
	if register_package_type_handler == nil {
		log.Println("Kernel: unkown package type, package droped. connection =", conn.GetObjectId(), "sender =", pckge.Sender, "reciver =", pckge.Reciver)
		return nil
	}

	// Es wird geprüft ob der Kernel noch ausgeführt wird
	if !obj.IsRunning() {
		return fmt.Errorf("EnterLocallyPackage: kernel is not running")
	}

	// Das Paket wird an den Handler übergeben
	err = register_package_type_handler.EnterRecivedPackage(pckge, conn)
	if err != nil {
		return fmt.Errorf("EnterLocallyPackage: 5: " + err.Error())
	}

	// Der Vorgang wurde ohne Fehler durchgeführt
	return nil
}

// Wird verwendet um Pakete an das Netzwerk zu senden
func (obj *Kernel) WriteL2Package(pckge *AddressLayerPackage) error {
	// Das Paket wird an den Routing Manager übergeben
	has_found, err := obj._connection_manager.EnterPackageBufferdAndRoute(pckge)
	if err != nil {
		return fmt.Errorf("EnterL2Package: 2: " + err.Error())
	}

	// Sollte keine Route vorhanden sein, wird das Paket verworfen
	if !has_found {
		finally_reciver_address := utils.ConvertHexStringToAddress(hex.EncodeToString(pckge.Reciver.SerializeCompressed()))
		log.Println("Kernel: package droped, no route for host. host =", finally_reciver_address)
		return nil
	}

	// Das Paket wurde erfolgreich an den Routing Manager übergeben
	return nil
}

// Nimmt eintreffende Layer 2 Pakete entgegen
func (obj *Kernel) EnterL2Package(pckge *AddressLayerPackage, conn RelayConnection) error {
	// Es wird geprüft ob es sich um eine Lokale Adresse handelt, wenn ja wird sie Lokal weiterverabeitet
	if obj.IsLocallyAddress(pckge.Reciver) {
		if err := obj.EnterLocallyPackage(pckge, conn); err != nil {
			return fmt.Errorf("EnterL2Package: 1: " + err.Error())
		} else {
			return nil
		}
	}

	// Es wird geprüft ob der Kernel noch ausgeführt wird
	if !obj.IsRunning() {
		return fmt.Errorf("EnterL2Package: kernel is not running")
	}

	// Das Paket wird an das Netzwerk gesendet, sofern eine Route vorhanden ist, ansonsten wird das Paket verworfen
	if err := obj.WriteL2Package(pckge); err != nil {
		return fmt.Errorf("EnterL2Package: 2: " + err.Error())
	}

	// Der Vorgang wurde ohne Fehler beendet
	return nil
}

// Nimmt einen Datensatz von einem Protokoll entgegen und überträgt es in das Netzwerk
func (obj *Kernel) EnterBytesAndSendL2PackageToNetwork(protocol_type uint8, package_bytes []byte, reciver_pkey *btcec.PublicKey) (bool, error) {
	return false, nil
}
