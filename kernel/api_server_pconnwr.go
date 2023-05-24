package kernel

import (
	"log"
	"net"
	"strconv"
	"sync"
)

// Stellt einen Dienst dar
type APIConnectionLiveService interface {
	GetId() string
	Close()
}

// Stellt den Verbindungs Wrapper dar
type APIProcessConnectionWrapper struct {
	id          string
	conn        net.Conn
	lock        *sync.Mutex
	isconn      bool
	service_map map[string]APIConnectionLiveService
}

// Ließt Daten aus der Verbindung
func (c *APIProcessConnectionWrapper) Read(b []byte) (int, error) {
	r, e := c.conn.Read(b)
	if e != nil {
		c.lock.Lock()
		c.isconn = false
		c.lock.Unlock()
		return r, e
	}
	return r, nil
}

// Schreibt Daten in die Verbindung
func (c *APIProcessConnectionWrapper) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

// Schließt die Verbindung
func (c *APIProcessConnectionWrapper) Close() error {
	return c.conn.Close()
}

// Gibt an ob die Verbindung aufgebaut ist
func (c *APIProcessConnectionWrapper) IsConnected() bool {
	c.lock.Lock()
	t := c.isconn
	c.lock.Unlock()
	return t
}

// Registriert einen "Service" welcher geschlossen wird sobald die Verbindung zum Prozess getrennt wird
func (c *APIProcessConnectionWrapper) AddProcessInvigoratingService(new_lservice APIConnectionLiveService) {
	// Der Threadlock wird verwendet
	c.lock.Lock()
	defer c.lock.Unlock()

	// Die Daten werden abgespeichert
	c.service_map[new_lservice.GetId()] = new_lservice

	// Log
	log.Printf("APIProcessConnectionWrapper: add service. sid = %s, process = %s\n", new_lservice.GetId(), c.GetObjectId())
}

// Entfernt einen "Service"
func (c *APIProcessConnectionWrapper) RemoveProcessInvigoratingService(new_lservice APIConnectionLiveService) {
	// Der Threadlock wird verwendet
	c.lock.Lock()
	defer c.lock.Unlock()

	// Es wird ermittet ob es einen passenden Dienst in dieser Verbindung gibt
	_, found := c.service_map[new_lservice.GetId()]
	if !found {
		log.Printf("APIProcessConnectionWrapper: cant remove unkown service. sid = %s, process = %s\n", new_lservice.GetId(), c.GetObjectId())
		return
	}

	// Der Eintrag wird entfernt
	delete(c.service_map, new_lservice.GetId())

	// Log
	log.Printf("APIProcessConnectionWrapper: remove service. sid = %s, process = %s\n", new_lservice.GetId(), c.GetObjectId())
}

// Gibt die Objekt ID aus
func (c *APIProcessConnectionWrapper) GetObjectId() string {
	return c.id
}

// Signalisiert dass alle Prozesse geschlossen werden sollen
func (c *APIProcessConnectionWrapper) Kill() {
	// Der Threadlock wird verwendet
	c.lock.Lock()
	defer c.lock.Unlock()

	// Es werden alle Dienste durch Iterriert
	total := len(c.service_map)
	for i := range c.service_map {
		c.service_map[i].Close()
	}

	// Log
	log.Println("APIProcessConnectionWrapper: process api connection closed. id = "+c.id, "total = "+strconv.Itoa(total))
}
