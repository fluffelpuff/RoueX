package kernel

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
)

type Relay struct {
	_db_id      int64
	_hexed_id   string
	_public_key *btcec.PublicKey
	_last_used  uint64
	_end_point  string
	_active     bool
	_type       string
	_trusted    bool
}

// Gibt das verwendete Protkoll aus
func (obj *Relay) GetProtocol() string {
	return obj._type
}

// Gibt den Endpunkt der Gegenseite aus
func (obj *Relay) GetEndpoint() string {
	return obj._end_point
}

// Gibt den Öffentlichen Schlüssel aus
func (obj *Relay) GetPublicKey() *btcec.PublicKey {
	return obj._public_key
}

// Gibt den Öffentlichen Schlüssel als hexstring aus
func (obj *Relay) GetPublicKeyHexString() string {
	return hex.EncodeToString(obj._public_key.SerializeCompressed())
}

// Gibt an ob dem Relay vertraut wird
func (obj *Relay) IsTrusted() bool {
	return obj._trusted
}

// Gibt an, ob es sich um die gleiche Verbindung handelt
func (obj *Relay) Equal(p2 *Relay) bool {
	return bytes.Equal(obj.GetPublicKey().SerializeCompressed(), p2.GetPublicKey().SerializeCompressed())
}

// Gibt den Hashwert des Objekts zurück
func (obj *Relay) Hash() uint32 {
	var hash uint32
	for _, c := range obj.GetPublicKey().SerializeCompressed() {
		hash = 31*hash + uint32(c)
	}
	return hash
}

// Stellt ein mögliches Relay dar mit welchen nocht keien Verbindung besteht
type RelayOutboundPair struct {
	_relay     *Relay
	_cl_module *ClientModule
}

// Gibt das Relay Objekt zurück
func (obj *RelayOutboundPair) GetRelay() *Relay {
	return obj._relay
}

// Gibt das Client Protkoll zurück
func (obj *RelayOutboundPair) GetClientConnModule() *ClientModule {
	return obj._cl_module
}

// Erstellt ein nicht Vertrauenswürdiges Relay
func NewUntrustedRelay(public_key *btcec.PublicKey, last_useed int64, end_point string, tpe string) *Relay {
	log.Println("New temporary untrusted relay created", hex.EncodeToString(public_key.SerializeCompressed()))
	return &Relay{_public_key: public_key, _last_used: uint64(last_useed), _type: tpe, _trusted: false, _end_point: end_point, _active: true, _hexed_id: "", _db_id: -1}
}
