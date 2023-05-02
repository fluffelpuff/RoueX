package kernel

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
)

type RouteEntry struct {
	_db_id        int64
	_route_hex_id string
	_relay_hex_id string
	_last_used    int64
	_active       bool
	_public_key   string
}

type RoutingTable struct {
	_routes []*RouteEntry
	_lock   *sync.Mutex
	_db     *sql.DB
}

func (obj *RoutingTable) Shutdown() {
	log.Println("RoutingTable: shutingdown routing table manager...")
	obj._lock.Lock()
	obj._db.Close()
	obj._lock.Unlock()
}

func (obj *RoutingTable) FetchRoutesByRelay(relay *Relay) ([]*RouteEntry, error) {
	return []*RouteEntry{}, nil
}

func loadRoutingTable(path string) (RoutingTable, error) {
	// Es wird versucht die SQLite Datei zu laden
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return RoutingTable{}, err
	}

	// Log
	fmt.Printf("Loading Routing database from %s...\n", path)

	// Die Anzahl der Tabellen mit dem Namen relays wird abgerufen
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master  WHERE type='table' AND name='routes'").Scan(&count)
	if err != nil {
		return RoutingTable{}, err
	}

	// Sollten mehr als eine Tabelle mit diesem Namen vorhanden sein, wird ein Fheler ausgelöst
	if count != 1 && count != 0 {
		return RoutingTable{}, fmt.Errorf("loadRoutingTable: invalid routing table database, file error, panic")
	}

	// Sollte die Tabelle nicht vorhanden sein, wird sie hinzugefügt
	// sollte die Tabelle jedoch bereits vorhanden sein, so werden alle Routen abgerufen
	route_entrys := make([]*RouteEntry, 0)
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
			return RoutingTable{}, err
		}
		fmt.Printf("New routing table database created %s\n", path)
	} else {
		// Es werden alle Verfügabren Relays abgerufen
		rows, err := db.Query("SELECT * FROM routes")
		if err != nil {
			return RoutingTable{}, err
		}

		// Die Einzelnen Relays werden abgerbeietet
		for rows.Next() {
			// Die Daten des Aktuellen Relays werden ausgelesen
			var pre_result RouteEntry
			var is_active_res int64
			err := rows.Scan(&pre_result._db_id, &pre_result._route_hex_id, &pre_result._relay_hex_id, &pre_result._last_used, &is_active_res, &pre_result._public_key)
			if err != nil {
				return RoutingTable{}, err
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
	re_result := RoutingTable{_routes: route_entrys, _db: db, _lock: new(sync.Mutex)}

	// Die Daten werden ohne Fehler zurückgeben
	return re_result, nil
}
