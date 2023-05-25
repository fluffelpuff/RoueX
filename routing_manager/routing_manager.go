package routingmanager

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	"github.com/btcsuite/btcd/btcec/v2"
)

type RelayInterface interface {
}

// Stellt einen Routing Eintrag dar
type RoutingManagerEntry struct {
	_db_id        int64
	_route_hex_id string
	_relay_hex_id string
	_last_used    int64
	_active       bool
	_public_key   *btcec.PublicKey
}

// Stellt einen Routen Link dar
type RelayRoutesList struct {
}

// Gibt die Anzahl alle Verbindungen aus
func (o *RelayRoutesList) GetTotalConnections() uint64 {
	return 0
}

// Stellt einen Routing Manager dar
type RoutingManager struct {
	_routes []*RoutingManagerEntry
	_lock   *sync.Mutex
	_db     *sql.DB
}

// Wird verwendet um den Routing Manager herunterzufahren
func (obj *RoutingManager) Shutdown() {
	// Log
	log.Println("RoutingManager: shutingdown routing table manager...")

	// Der Threadlock wird verwendet
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Die Datenbank wird geschlossen
	obj._db.Close()
}

// Gibt die Routen für einen Spiziellen Relay aus
func (obj *RoutingManager) FetchRoutesByRelay(relay RelayInterface) (*RelayRoutesList, error) {
	return &RelayRoutesList{}, nil
}

// Wird verwendet um den Routing Manager zu laden
func LoadRoutingManager(path string) (RoutingManager, error) {
	// Es wird versucht die SQLite Datei zu laden
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return RoutingManager{}, err
	}

	// Log
	fmt.Printf("Loading Routing database from %s...\n", path)

	// Die Anzahl der Tabellen mit dem Namen relays wird abgerufen
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master  WHERE type='table' AND name='routes'").Scan(&count)
	if err != nil {
		return RoutingManager{}, err
	}

	// Sollten mehr als eine Tabelle mit diesem Namen vorhanden sein, wird ein Fheler ausgelöst
	if count != 1 && count != 0 {
		return RoutingManager{}, fmt.Errorf("loadRoutingManager: invalid routing table database, file error, panic")
	}

	// Sollte die Tabelle nicht vorhanden sein, wird sie hinzugefügt
	// sollte die Tabelle jedoch bereits vorhanden sein, so werden alle Routen abgerufen
	route_entrys := make([]*RoutingManagerEntry, 0)
	if count == 0 {
		_, err = db.Exec(`CREATE TABLE "routes" (
			"route_id"	INTEGER UNIQUE,
			"route_hex_id" TEXT,
			"relay_hex_id"	TEXT,
			"last_used"	INTEGER DEFAULT -1,
			"active"	INTEGER DEFAULT 1,
			"public_key"	TEXT,
			PRIMARY KEY("route_id" AUTOINCREMENT)
		);`)
		if err != nil {
			return RoutingManager{}, err
		}
		fmt.Printf("New routing table database created %s\n", path)
	} else {
		// Es werden alle Verfügabren Relays abgerufen
		rows, err := db.Query("SELECT * FROM routes")
		if err != nil {
			return RoutingManager{}, err
		}

		// Die Einzelnen Relays werden abgerbeietet
		for rows.Next() {
			// Die Daten des Aktuellen Relays werden ausgelesen
			var pre_result RoutingManagerEntry
			var is_active_res int64
			err := rows.Scan(&pre_result._db_id, &pre_result._route_hex_id, &pre_result._relay_hex_id, &pre_result._last_used, &is_active_res, &pre_result._public_key)
			if err != nil {
				return RoutingManager{}, err
			}

			// Es wird geprüft ob die Verbindung aktiv ist
			if is_active_res == 1 {
				pre_result._active = true
			} else {
				pre_result._active = false
			}

			// Die Verbindung wird zwischen gespeichert
			route_entrys = append(route_entrys, &pre_result)
		}

		// Log
		fmt.Printf("Routes from routing table loaded, total = %d, path = %s\n", len(route_entrys), path)
	}

	// Es wird ein neues Routing Table objekt erstetllt
	re_result := RoutingManager{_routes: route_entrys, _db: db, _lock: new(sync.Mutex)}

	// Die Daten werden ohne Fehler zurückgeben
	return re_result, nil
}
