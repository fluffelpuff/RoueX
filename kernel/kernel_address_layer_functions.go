package kernel

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	addresspackages "github.com/fluffelpuff/RoueX/address_packages"
	"github.com/fluffelpuff/RoueX/kernel/extra"
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

// Nimmt Lokale Pakete entgegen und verarbeitet sie
func (obj *Kernel) EnterLocallyPackage(pckge *addresspackages.PreAddressLayerPackage) error {
	// Es wird geprüft ob der Body mindestens 1 Byte groß ist
	if len(pckge.Body) < 1 {
		return fmt.Errorf("EnterLocallyPackage: 1: Invalid layer two package recived")
	}

	// Es wird geprüft ob es sich um eine Registrierte Paket Funktion handelt
	register_package_type_handler, err := obj.GetRegisteredKernelTypeProtocol(pckge.Protocol)
	if err != nil {
		return fmt.Errorf("EnterLocallyPackage: 4: " + err.Error())
	}

	// Sollte kein Paket Type handler vorhanden sein, wird das Paket verworfen
	if register_package_type_handler == nil {
		log.Println("Kernel: unkown package type, package droped. sender =", pckge.Sender, "reciver =", pckge.Reciver)
		return nil
	}

	// Es wird geprüft ob der Kernel noch ausgeführt wird
	if !obj.IsRunning() {
		return fmt.Errorf("EnterLocallyPackage: kernel is not running")
	}

	// Das Paket wird an den Handler übergeben
	err = register_package_type_handler.EnterRecivedPackage(pckge)
	if err != nil {
		return fmt.Errorf("EnterLocallyPackage: 5: " + err.Error())
	}

	// Der Vorgang wurde ohne Fehler durchgeführt
	return nil
}

// Nimmt eintreffende Layer 2 Pakete entgegen
func (obj *Kernel) EnterL2Package(pckge *addresspackages.FinalAddressLayerPackage, conn RelayConnection) error {
	// Es wird geprüft ob die Signatur korrekt ist
	if !pckge.ValidateSignature() {
		return fmt.Errorf("EnterL2Package: 1: invalid package signature")
	}

	// Es wird geprüft ob der Kernel noch ausgeführt wird
	if !obj.IsRunning() {
		return fmt.Errorf("EnterL2Package: kernel is not running")
	}

	// Es wird geprüft ob es sich um eine Lokale Adresse handelt, wenn ja wird sie Lokal weiterverabeitet
	if obj.IsLocallyAddress(pckge.Reciver) {
		// Das Paket wird Entschlüsselt und dann an den Lokalen Buffer übergeben
		if err := obj.DecryptLocallyPackageToBuffer(pckge); err != nil {
			return fmt.Errorf("EnterL2Package: 1: " + err.Error())
		}

		// Der Vorgang wurde ohne Fehler durchgeführt
		return nil
	}

	// Es wird geprüft ob der Kernel noch ausgeführt wird
	if !obj.IsRunning() {
		return fmt.Errorf("EnterL2Package: kernel is not running")
	}

	// Das Paket wird an das Netzwerk gesendet, sofern eine Route vorhanden ist, ansonsten wird das Paket verworfen
	_, _, err := obj.WriteL2PackageByNetworkRoute(pckge)
	if err != nil {
		return fmt.Errorf("EnterL2Package: 2: " + err.Error())
	}

	// Es wird geprüft ob der Kernel noch ausgeführt wird
	if !obj.IsRunning() {
		return fmt.Errorf("EnterL2Package: kernel is not running")
	}

	// Der Vorgang wurde ohne Fehler beendet
	return nil
}

// Wird verwendet um Pakete an das Netzwerk zu senden
func (obj *Kernel) WriteL2PackageByNetworkRoute(pckge *addresspackages.FinalAddressLayerPackage) (*extra.PackageSendState, bool, error) {
	// Das Paket wird an den Routing Manager übergeben
	sstate, has_found, err := obj._connection_manager.EnterPackageToRoutingManger(pckge)
	if err != nil {
		return nil, true, fmt.Errorf("WriteL2PackageByNetworkRoute: 1: " + err.Error())
	}

	// Sollte keine Route vorhanden sein, wird das Paket verworfen
	if !has_found {
		log.Println("Kernel: package droped, no route for host. host =", hex.EncodeToString(pckge.Reciver.SerializeCompressed()))
		return nil, false, nil
	}

	// Das Paket wurde erfolgreich an den Routing Manager übergeben
	return sstate, true, nil
}

// Entschlüsselt ein Lokales Paket und Speichert es im Puffer
func (obj *Kernel) DecryptLocallyPackageToBuffer(pckge *addresspackages.FinalAddressLayerPackage) error {
	// Es wird geprüft ob das Paket für diesen Node bestimmt ist
	return nil
}

// Verschlüsselt ein nicht Verschlüsseltes Layer 2 Paket, Signiert es und Sendet es ins Netzwerk
func (obj *Kernel) EncryptPlainL2PackageAndWriteByNetworkRoute(pckge *addresspackages.PreAddressLayerPackage) (*extra.PackageSendState, bool, error) {
	// Die Inneren Verschlüsselten Daten werden übertragen
	internal_data := addresspackages.EncryptedInnerData{
		Protocol: pckge.Protocol,
		Body:     pckge.Body,
		Version:  pckge.Version,
	}

	// Die Inneren Daten werden in Bytes umgewandelt
	byted_inner_data, err := internal_data.ToBytes()
	if err != nil {
		return nil, false, fmt.Errorf("EncryptPlainL2PackageAndWriteByNetworkRoute: 1: " + err.Error())
	}

	// Die Daten werden für den Empfänger verschlüsselt
	encrypted_data, err := utils.EncryptECIESPublicKey(&pckge.Reciver, byted_inner_data)
	if err != nil {
		return nil, false, fmt.Errorf("EncryptPlainL2PackageAndWriteByNetworkRoute: 2: " + err.Error())
	}

	// Es wird ein Hash aus den Daten erstellt
	sign_hash := utils.ComputeSha3256Hash(
		pckge.Sender.SerializeCompressed(),
		pckge.Reciver.SerializeCompressed(),
		encrypted_data,
	)

	// Der Paket Hash wird Signiert
	package_signature, err := utils.Sign(obj._private_key, sign_hash)
	if err != nil {
		return nil, false, fmt.Errorf("EncryptPlainL2PackageAndWriteByNetworkRoute: 3: " + err.Error())
	}

	// Das Verschlüsselte Paket wird erstellt
	builded_encrypted_package := addresspackages.FinalAddressLayerPackage{
		Sender:           pckge.Sender,
		Reciver:          pckge.Reciver,
		InnerData:        encrypted_data,
		SSig:             package_signature,
		IsLocallyCreated: pckge.IsLocallyCreated,
	}

	// Das Paket wird an den Routing Manager übergebene
	sstate, has_route, err := obj.WriteL2PackageByNetworkRoute(&builded_encrypted_package)
	if err != nil {
		return nil, false, fmt.Errorf("EncryptPlainL2PackageAndWriteByNetworkRoute: 4: " + err.Error())
	}

	// Der Vorgang wurde ohne Fehler durchgeführt
	return sstate, has_route, nil
}

// Signiert ein Layer 2 Paket und sendet es unverschlüsselt an das Netzwerk
func (obj *Kernel) PlainL2PackageAndWriteByNetworkRoute(pckge *addresspackages.PreAddressLayerPackage) (*extra.PackageSendState, bool, error) {
	// Die Inneren Verschlüsselten Daten werden übertragen
	internal_data := addresspackages.EncryptedInnerData{
		Protocol: pckge.Protocol,
		Body:     pckge.Body,
		Version:  pckge.Version,
	}

	// Die Inneren Daten werden in Bytes umgewandelt
	byted_inner_data, err := internal_data.ToBytes()
	if err != nil {
		return nil, false, fmt.Errorf("PlainL2PackageAndWriteByNetworkRoute: 1: " + err.Error())
	}

	// Die Daten werden für den Empfänger verschlüsselt
	encrypted_data, err := utils.EncryptECIESPublicKey(&pckge.Reciver, byted_inner_data)
	if err != nil {
		return nil, false, fmt.Errorf("PlainL2PackageAndWriteByNetworkRoute: 2: " + err.Error())
	}

	// Es wird ein Hash aus den Daten erstellt
	sign_hash := utils.ComputeSha3256Hash(
		pckge.Sender.SerializeCompressed(),
		pckge.Reciver.SerializeCompressed(),
		encrypted_data,
	)

	// Der Paket Hash wird Signiert
	package_signature, err := utils.Sign(obj._private_key, sign_hash)
	if err != nil {
		return nil, false, fmt.Errorf("PlainL2PackageAndWriteByNetworkRoute: 3: " + err.Error())
	}

	// Das Verschlüsselte Paket wird erstellt
	builded_encrypted_package := addresspackages.FinalAddressLayerPackage{
		Sender:           pckge.Sender,
		Reciver:          pckge.Reciver,
		InnerData:        encrypted_data,
		SSig:             package_signature,
		IsLocallyCreated: pckge.IsLocallyCreated,
	}

	// Das Paket wird an den Routing Manager übergebene
	sstate, has_route, err := obj.WriteL2PackageByNetworkRoute(&builded_encrypted_package)
	if err != nil {
		return nil, false, fmt.Errorf("PlainL2PackageAndWriteByNetworkRoute: 4: " + err.Error())
	}

	// Der Vorgang wurde ohne Fehler durchgeführt
	return sstate, has_route, nil
}

// Nimmt einen Datensatz von einem Protokoll entgegen verschlüsselt ihn und überträgt es als Layer 2 Paket (verschlüsselt)
func (obj *Kernel) EnterBytesEncryptAndSendL2PackageToNetwork(protocol_type uint8, package_bytes []byte, reciver_pkey *btcec.PublicKey) (*extra.PackageSendState, bool, error) {
	// Das Paket wird gebaut
	builded_locally_package := addresspackages.PreAddressLayerPackage{Reciver: *reciver_pkey, Sender: *obj.GetPublicKey(), IsLocallyCreated: true, Protocol: protocol_type, Body: package_bytes}

	// Es wird ermittelt ob es sich um eien Lokale Adresse handelt
	is_locally := obj.IsLocallyAddress(*reciver_pkey)

	// Sollte es sich nicht um eine Lokale Adresse handeln, wird das Paket verschlüsselt und signiert
	if !is_locally {
		// Das Paket wird verschlüsselt, Signiert und in das Netzwerk gesendet
		sstate, has_route, err := obj.EncryptPlainL2PackageAndWriteByNetworkRoute(&builded_locally_package)
		if err != nil {
			return nil, false, fmt.Errorf("EnterBytesEncryptAndSendL2PackageToNetwork: 1: " + err.Error())
		}

		// Der Sendestatus wird zurückgegeben
		return sstate, has_route, nil
	}

	// Das Paket wird an den Lokalen Paket Puffer übergeben
	sstate, err := obj._memory.AddL2Package(&builded_locally_package)
	if err != nil {
		return nil, false, fmt.Errorf("EnterBytesEncryptAndSendL2PackageToNetwork: 2: " + err.Error())
	}

	// Der Vorgang wurde ohne Fehler durchgeführt
	return sstate, true, nil
}

// Nimmt einen Datensatz von einem Protokoll entgegen und überträgt es als Layer 2 Paket (unverschlüsselt)
func (obj *Kernel) EnterBytesAndSendL2PackageToNetwork(protocol_type uint8, package_bytes []byte, reciver_pkey *btcec.PublicKey) (*extra.PackageSendState, bool, error) {
	// Das Paket wird gebaut
	builded_locally_package := addresspackages.PreAddressLayerPackage{Reciver: *reciver_pkey, Sender: *obj.GetPublicKey(), IsLocallyCreated: true, Protocol: protocol_type, Body: package_bytes}

	// Es wird ermittelt ob es sich um eien Lokale Adresse handelt
	is_locally := obj.IsLocallyAddress(*reciver_pkey)

	// Sollte es sich nicht um eine Lokale Adresse handeln, wird das Paket verschlüsselt und signiert
	if !is_locally {
		// Das Paket wird an das Netzwerk gesendet
		sstate, has_route, err := obj.PlainL2PackageAndWriteByNetworkRoute(&builded_locally_package)
		if err != nil {
			return nil, false, fmt.Errorf("EnterBytesAndSendL2PackageToNetwork: 1: " + err.Error())
		}

		// Der Sendestatus wird zurückgegeben
		return sstate, has_route, nil
	}

	// Das Paket wird an den Lokalen Paket Puffer übergeben
	sstate, err := obj._memory.AddL2Package(&builded_locally_package)
	if err != nil {
		return nil, false, fmt.Errorf("EnterBytesAndSendL2PackageToNetwork: 2: " + err.Error())
	}

	// Der Vorgang wurde ohne Fehler durchgeführt
	return sstate, true, nil
}
