package kernel

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type TrustedRelays struct {
	_lock   *sync.Mutex
	_relays []*Relay
	_db     *sql.DB
}

func (obj *TrustedRelays) Shutdown() {
	log.Println("Shutingdown Trusted Relays database")
	obj._lock.Lock()
	obj._db.Close()
	obj._lock.Unlock()
}

func (obj *TrustedRelays) GetAllRelays() []*Relay {
	return obj._relays
}

func loadTrustedRelaysTable(path string) (TrustedRelays, error) {
	// Es wird versucht die SQLite Datei zu laden
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return TrustedRelays{}, err
	}

	// Log
	fmt.Printf("Loading Trusted Relays database from %s...\n", path)

	// Die Anzahl der Tabellen mit dem Namen relays wird abgerufen
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master  WHERE type='table' AND name='relays'").Scan(&count)
	if err != nil {
		return TrustedRelays{}, err
	}

	// Sollten mehr als eine Tabelle mit diesem Namen vorhanden sein, wird ein Fheler ausgelöst
	if count != 1 && count != 0 {
		return TrustedRelays{}, fmt.Errorf("loadTrustedRelaysTable: invalid trusted relays database file, hard panic")
	}

	// Sollte die Tabelle nicht vorhanden sein, wird sie hinzugefügt
	// sollte die Tabelle jedoch bereits vorhanden sein, so werden alle Relays abgerufen
	re_relays := make([]*Relay, 0)
	if count == 0 {
		_, err = db.Exec(`CREATE TABLE "relays" (
			"rid"	INTEGER UNIQUE,
			"hx_id"	TEXT,
			"type"	TEXT,
			"end_point"	TEXT,
			"last_used"	INTEGER DEFAULT -1,
			"active"	INTEGER DEFAULT 1,
			"public_key"	TEXT,
			PRIMARY KEY("rid" AUTOINCREMENT)
		);`)
		if err != nil {
			return TrustedRelays{}, err
		}
		fmt.Printf("New Trusted Relays Database created %s\n", path)
	} else {
		// Es werden alle Verfügabren Relays abgerufen
		rows, err := db.Query("SELECT * FROM relays")
		if err != nil {
			return TrustedRelays{}, err
		}

		// Die Einzelnen Relays werden abgerbeietet
		for rows.Next() {
			// Die Daten des Aktuellen Relays werden ausgelesen
			var db_uid int64
			var db_hex_id string
			var db_type string
			var end_point string
			var last_used int64
			var active int64
			var public_key string
			err := rows.Scan(&db_uid, &db_hex_id, &db_type, &end_point, &last_used, &active, &public_key)
			if err != nil {
				return TrustedRelays{}, err
			}

			// Es wird geprüft ob die Verbindung aktiv ist
			is_active := false
			if active == 1 {
				is_active = true
			}

			// Es wird versucht den Öffentlichen Schlüssel einzulesn
			decoded_pkey, err := hex.DecodeString(public_key)
			if err != nil {
				return TrustedRelays{}, err
			}
			pkey, err := ReadPublicKeyFromByteSlice(decoded_pkey)
			if err != nil {
				return TrustedRelays{}, err
			}

			// Das Objekt wird wieder hergestellt
			retrived_relay := Relay{_db_id: db_uid, _hexed_id: db_hex_id, _type: db_type, _end_point: end_point, _last_used: uint64(last_used), _active: is_active, _public_key: pkey, _trusted: true}

			// Die Verbindung wird zwischen gespeichert
			re_relays = append(re_relays, &retrived_relay)
		}

		fmt.Printf("Trusted Relays from database loaded, total = %d, path = %s\n", len(re_relays), path)
	}

	// Die Daten werden ohne Fehler zurückgegeben
	return TrustedRelays{_relays: re_relays, _db: db, _lock: new(sync.Mutex)}, nil
}
