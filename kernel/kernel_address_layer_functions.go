package kernel

import (
	"bytes"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	addresspackages "github.com/fluffelpuff/RoueX/address_packages"
	"github.com/fluffelpuff/RoueX/kernel/extra"
	"github.com/fluffelpuff/RoueX/utils"
)

// Stellt den Eintrag für einen Funktionstypen Hadnler dar
type KernelPackageProtocolEntry struct {
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

	// Es wird geprüft ob es bereits eine Eintrag für das Protokoll gibt
	_, hfound := obj._protocols[int(tpe)]
	if hfound {
		obj._lock.Unlock()
		return fmt.Errorf("RegisterNewKernelTypeProtocol: 2: type always registred")
	}

	// Das Objekt wird registriert
	nwo := &KernelPackageProtocolEntry{Tpe: tpe, Ptf: pckgtf}
	obj._protocols[int(tpe)] = nwo

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
	defer obj._lock.Unlock()

	// Es wird geprüft ob der Eintrag bereits hinzugefügt wurde
	re, has_found := obj._protocols[int(tpe)]
	if !has_found {
		return nil, fmt.Errorf("GetRegisteredKernelTypeProtocol: 2: type always registred")
	}

	// Die Daten werden zurückgegeben
	return re.Ptf, nil
}

// Nimmt Lokale Pakete entgegen und verarbeitet sie
func (obj *Kernel) EnterLocallyPackage(pckge *addresspackages.AddressLayerPackage) error {
	// Es wird geprüft ob der Data mindestens 1 Byte groß ist
	if len(pckge.Data) < 1 {
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

// Nimmt ein Plain PCI Paket entgegen
func (obj *Kernel) EnterPlainPCIPackage(pckge *addresspackages.SendableAddressLayerPackage, conn RelayConnection) {
	// Es wird geprüft ob der Kernel noch ausgeführt wird
	if !obj.IsRunning() {
		return
	}

	// Es wird versucht die Inneren Daten einzulesen
	inner_data, errr := addresspackages.ReadInnerFrameFromBytes(pckge.Data)
	if errr != nil {
		log.Println("")
		return
	}

	// Es wird geprüft um welches Protokoll es sich handelt
	fmt.Println(inner_data)
}

// Nimmt ein nicht verschlüsseltes Lokales Paket entgegen
func (obj *Kernel) PlainLocallyPackageToBuffer(pckge *addresspackages.SendableAddressLayerPackage) error {
	// Es wird geprüft ob das Paket für diesen Node bestimmt ist
	if !bytes.Equal(pckge.Reciver.SerializeCompressed(), obj.GetPublicKey().SerializeCompressed()) {
		return fmt.Errorf("DecryptLocallyPackageToBuffer: unkown reciver address, is not locally address")
	}

	// Die Daten werden eingelesen
	readed_inner, err := addresspackages.ReadInnerFrameFromBytes(pckge.Data)
	if err != nil {
		return fmt.Errorf("DecryptLocallyPackageToBuffer: " + err.Error())
	}

	// Das Paket zur Internen Verarbeitung wird erstellt
	internal_package := &addresspackages.AddressLayerPackage{
		Reciver:  pckge.Reciver,
		Sender:   pckge.Sender,
		Protocol: readed_inner.Protocol,
		Version:  readed_inner.Version,
		Data:     readed_inner.Data,
	}

	// Das Paket wird für Lokale Weiterverabeitung weitergereicht
	if err := obj.EnterLocallyPackage(internal_package); err != nil {
		return fmt.Errorf("DecryptLocallyPackageToBuffer: " + err.Error())
	}

	// Das Paket wurde erfolgreich Lokal ausgewertet
	return nil
}

// Entschlüsselt ein Lokales Paket und Speichert es im Puffer
func (obj *Kernel) DecryptedLocallyPackageToBuffer(pckge *addresspackages.SendableAddressLayerPackage) error {
	// Es wird geprüft ob das Paket für diesen Node bestimmt ist
	if !bytes.Equal(pckge.Reciver.SerializeCompressed(), obj.GetPublicKey().SerializeCompressed()) {
		return fmt.Errorf("DecryptLocallyPackageToBuffer: unkown reciver address, is not locally address")
	}

	// Die Daten werden versucht zu entschlüsseöm
	dcrypted, err := utils.DecryptDataWithPrivateKey(obj._private_key, pckge.Data)
	if err != nil {
		return fmt.Errorf("DecryptLocallyPackageToBuffer: " + err.Error())
	}

	// Die Daten werden eingelesen
	readed_inner, err := addresspackages.ReadInnerFrameFromBytes(dcrypted)
	if err != nil {
		return fmt.Errorf("DecryptLocallyPackageToBuffer: " + err.Error())
	}

	// Das Paket zur Internen Verarbeitung wird erstellt
	internal_package := &addresspackages.AddressLayerPackage{
		Reciver:  pckge.Reciver,
		Sender:   pckge.Sender,
		Protocol: readed_inner.Protocol,
		Version:  readed_inner.Version,
		Data:     readed_inner.Data,
	}

	// Das Paket wird für Lokale Weiterverabeitung weitergereicht
	if err := obj.EnterLocallyPackage(internal_package); err != nil {
		return fmt.Errorf("DecryptLocallyPackageToBuffer: " + err.Error())
	}

	// Das Paket wurde erfolgreich Lokal ausgewertet
	return nil
}

// Nimmt eintreffende Layer 2 Pakete entgegen
func (obj *Kernel) EnterL2Package(pckge *addresspackages.SendableAddressLayerPackage, conn RelayConnection) error {
	// Es wird geprüft ob der Kernel noch ausgeführt wird
	if !obj.IsRunning() {
		return fmt.Errorf("EnterL2Package: kernel is not running")
	}

	// Es wird geprüft ob es sich um eine Lokale Adresse handelt, wenn ja wird sie Lokal weiterverabeitet
	if obj.IsLocallyAddress(pckge.Reciver) {
		// Wenn es sich um ein Verschlüsseltes Paket handelt, wird versuch dieses zu Entschlüsseln,
		// bei einem Unverschlüsseltes Paket wird es direkt Verarbeitet
		if !pckge.Plain {
			if err := obj.DecryptedLocallyPackageToBuffer(pckge); err != nil {
				return fmt.Errorf("EnterL2Package: 1: " + err.Error())
			}
		} else {
			if err := obj.PlainLocallyPackageToBuffer(pckge); err != nil {
				return fmt.Errorf("EnterL2Package: 1: " + err.Error())
			}
		}

		// Der Vorgang wurde ohne Fehler durchgeführt
		return nil
	}

	// Es wird geprüft ob der Kernel noch ausgeführt wird
	if !obj.IsRunning() {
		return fmt.Errorf("EnterL2Package: kernel is not running")
	}

	// Es wird geprüft ob es sich um Nicht verschlüsseltes PCI Paket handelt
	if pckge.Plain && pckge.PCI {
		obj.EnterPlainPCIPackage(pckge, conn)
	}

	// Es wird geprüft ob der Kernel noch ausgeführt wird
	if !obj.IsRunning() {
		return fmt.Errorf("EnterL2Package: kernel is not running")
	}

	// Das Paket wird an das Netzwerk gesendet, sofern eine Route vorhanden ist, ansonsten wird das Paket verworfen
	_, err := obj.WriteL2PackageByNetworkRoute(pckge)
	if err != nil {
		return fmt.Errorf("EnterL2Package: 2: " + err.Error())
	}

	// Der Vorgang wurde ohne Fehler beendet
	return nil
}

// Wird verwendet um Pakete an das Netzwerk zu senden
func (obj *Kernel) WriteL2PackageByNetworkRoute(pckge *addresspackages.SendableAddressLayerPackage) (*extra.PackageSendState, error) {
	// Das Paket wird an den Routing Manager übergeben
	sstate, err := obj._connection_manager.EnterPackageToRoutingManger(pckge)
	if err != nil {
		return nil, fmt.Errorf("WriteL2PackageByNetworkRoute: 1: " + err.Error())
	}

	// Das Paket wurde erfolgreich an den Routing Manager übergeben
	return sstate, nil
}

// Verschlüsselt ein nicht Verschlüsseltes Layer 2 Paket, Signiert es und Sendet es ins Netzwerk
func (obj *Kernel) EncryptPlainL2PackageAndWriteByNetworkRoute(pckge *addresspackages.AddressLayerPackage) (*extra.PackageSendState, error) {
	// Die Inneren Verschlüsselten Daten werden übertragen
	internal_data := addresspackages.InnerFrame{
		Protocol: pckge.Protocol,
		Data:     pckge.Data,
		Version:  pckge.Version,
	}

	// Die Inneren Daten werden in Bytes umgewandelt
	byted_inner_data, err := internal_data.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("EncryptPlainL2PackageAndWriteByNetworkRoute: 1: " + err.Error())
	}

	// Die Daten werden für den Empfänger verschlüsselt
	encrypted_data, err := utils.EncryptECIESPublicKey(&pckge.Reciver, byted_inner_data)
	if err != nil {
		return nil, fmt.Errorf("EncryptPlainL2PackageAndWriteByNetworkRoute: 2: " + err.Error())
	}

	// Es wird ein Hash aus den Daten erstellt
	sign_hash := utils.ComputeSha3256Hash(
		pckge.Sender.SerializeCompressed(),
		pckge.Reciver.SerializeCompressed(),
		[]byte("cipher"),
		encrypted_data,
	)

	// Der Paket Hash wird Signiert
	package_signature, err := utils.Sign(obj._private_key, sign_hash)
	if err != nil {
		return nil, fmt.Errorf("EncryptPlainL2PackageAndWriteByNetworkRoute: 3: " + err.Error())
	}

	// Das Verschlüsselte Paket wird erstellt
	builded_encrypted_package := addresspackages.SendableAddressLayerPackage{
		Sender:  pckge.Sender,
		Reciver: pckge.Reciver,
		Data:    encrypted_data,
		Sig:     package_signature,
		Plain:   false,
		PCI:     false,
	}

	// Das Paket wird an den Routing Manager übergebene
	sstate, err := obj.WriteL2PackageByNetworkRoute(&builded_encrypted_package)
	if err != nil {
		return nil, fmt.Errorf("EncryptPlainL2PackageAndWriteByNetworkRoute: 4: " + err.Error())
	}

	// Der Vorgang wurde ohne Fehler durchgeführt
	return sstate, nil
}

// Signiert ein Layer 2 Paket und sendet es unverschlüsselt an das Netzwerk
func (obj *Kernel) PlainL2PackageAndWriteByNetworkRoute(pckge *addresspackages.AddressLayerPackage, please_check_instructions bool) (*extra.PackageSendState, error) {
	// Die Inneren Verschlüsselten Daten werden übertragen
	internal_data := addresspackages.InnerFrame{
		Protocol: pckge.Protocol,
		Data:     pckge.Data,
		Version:  pckge.Version,
	}

	// Die Inneren Daten werden in Bytes umgewandelt
	byted_inner_data, err := internal_data.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("PlainL2PackageAndWriteByNetworkRoute: 1: " + err.Error())
	}

	// Es wird ein Hash aus den Daten erstellt
	sign_hash := utils.ComputeSha3256Hash(
		pckge.Sender.SerializeCompressed(),
		pckge.Reciver.SerializeCompressed(),
		[]byte("unciphered"),
		byted_inner_data,
	)

	// Der Paket Hash wird Signiert
	package_signature, err := utils.Sign(obj._private_key, sign_hash)
	if err != nil {
		return nil, fmt.Errorf("PlainL2PackageAndWriteByNetworkRoute: 3: " + err.Error())
	}

	// Das Verschlüsselte Paket wird erstellt
	builded_encrypted_package := addresspackages.SendableAddressLayerPackage{
		Sender:  pckge.Sender,
		Reciver: pckge.Reciver,
		Data:    byted_inner_data,
		Sig:     package_signature,
		Plain:   true,
		PCI:     please_check_instructions,
	}

	// Das Paket wird an den Routing Manager übergebene
	sstate, err := obj.WriteL2PackageByNetworkRoute(&builded_encrypted_package)
	if err != nil {
		return nil, fmt.Errorf("PlainL2PackageAndWriteByNetworkRoute: 4: " + err.Error())
	}

	// Der Vorgang wurde ohne Fehler durchgeführt
	return sstate, nil
}

// Nimmt einen Datensatz von einem Protokoll entgegen verschlüsselt ihn und überträgt es als Layer 2 Paket (verschlüsselt)
func (obj *Kernel) EnterBytesEncryptAndSendL2PackageToNetwork(protocol_type uint8, package_bytes []byte, reciver_pkey *btcec.PublicKey) (*extra.PackageSendState, error) {
	// Das Paket wird gebaut
	builded_locally_package := addresspackages.AddressLayerPackage{
		Reciver:  *reciver_pkey,
		Sender:   *obj.GetPublicKey(),
		Protocol: protocol_type,
		Data:     package_bytes,
	}

	// Es wird ermittelt ob es sich um eien Lokale Adresse handelt
	is_locally := obj.IsLocallyAddress(*reciver_pkey)

	// Sollte es sich nicht um eine Lokale Adresse handeln, wird das Paket verschlüsselt und signiert
	if !is_locally {
		// Das Paket wird verschlüsselt, Signiert und in das Netzwerk gesendet
		sstate, err := obj.EncryptPlainL2PackageAndWriteByNetworkRoute(&builded_locally_package)
		if err != nil {
			return nil, fmt.Errorf("EnterBytesEncryptAndSendL2PackageToNetwork: 1: " + err.Error())
		}

		// Der Sendestatus wird zurückgegeben
		return sstate, nil
	}

	// Das Paket wird an den Lokalen Paket Puffer übergeben
	sstate, err := obj._memory.AddL2Package(&builded_locally_package)
	if err != nil {
		return nil, fmt.Errorf("EnterBytesEncryptAndSendL2PackageToNetwork: 2: " + err.Error())
	}

	// Der Vorgang wurde ohne Fehler durchgeführt
	return sstate, nil
}

// Nimmt einen Datensatz von einem Protokoll entgegen und überträgt es als Layer 2 Paket (unverschlüsselt)
func (obj *Kernel) EnterBytesAndSendL2PackageToNetwork(protocol_type uint8, package_bytes []byte, reciver_pkey *btcec.PublicKey, please_check_instructions bool) (*extra.PackageSendState, error) {
	// Das Paket wird gebaut
	builded_locally_package := addresspackages.AddressLayerPackage{
		Reciver:  *reciver_pkey,
		Sender:   *obj.GetPublicKey(),
		Protocol: protocol_type,
		Data:     package_bytes,
	}

	// Es wird ermittelt ob es sich um eien Lokale Adresse handelt
	is_locally := obj.IsLocallyAddress(*reciver_pkey)

	// Sollte es sich nicht um eine Lokale Adresse handeln, wird das Paket verschlüsselt und signiert
	if !is_locally {
		// Das Paket wird an das Netzwerk gesendet
		sstate, err := obj.PlainL2PackageAndWriteByNetworkRoute(&builded_locally_package, please_check_instructions)
		if err != nil {
			return nil, fmt.Errorf("EnterBytesAndSendL2PackageToNetwork: 1: " + err.Error())
		}

		// Der Sendestatus wird zurückgegeben
		return sstate, nil
	}

	// Das Paket wird an den Lokalen Paket Puffer übergeben
	sstate, err := obj._memory.AddL2Package(&builded_locally_package)
	if err != nil {
		return nil, fmt.Errorf("EnterBytesAndSendL2PackageToNetwork: 2: " + err.Error())
	}

	// Der Vorgang wurde ohne Fehler durchgeführt
	return sstate, nil
}
