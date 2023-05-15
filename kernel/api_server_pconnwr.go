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
	id           string
	conn         net.Conn
	lock         *sync.Mutex
	isconn       bool
	live_service []APIConnectionLiveService
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
func (c *APIProcessConnectionWrapper) AddProcessInvigoratingService(new_lservice APIConnectionLiveService) error {
	c.lock.Lock()
	c.live_service = append(c.live_service, new_lservice)
	c.lock.Unlock()
	log.Printf("APIProcessConnectionWrapper: add service. sid = %s, process = %s\n", new_lservice.GetId(), c.GetObjectId())
	return nil
}

// Entfernt einen "Service"
func (c *APIProcessConnectionWrapper) RemoveProcessInvigoratingService(new_lservice APIConnectionLiveService) error {
	found_i := -1
	c.lock.Lock()
	for i := range c.live_service {
		if c.live_service[i].GetId() == new_lservice.GetId() {
			found_i = i
			break
		}
	}
	if found_i > -1 {
		c.live_service = append(c.live_service[:found_i], c.live_service[found_i+1:]...)
		log.Printf("APIProcessConnectionWrapper: remove service. sid = %s, process = %s\n", new_lservice.GetId(), c.GetObjectId())
	}
	c.lock.Unlock()
	return nil
}

// Gibt die Objekt ID aus
func (c *APIProcessConnectionWrapper) GetObjectId() string {
	return c.id
}

// Signalisiert dass alle Prozesse geschlossen werden sollen
func (c *APIProcessConnectionWrapper) Kill() {
	c.lock.Lock()
	total := len(c.live_service)
	for i := range c.live_service {
		c.live_service[i].Close()
	}
	c.lock.Unlock()
	log.Println("APIProcessConnectionWrapper: process api connection closed. id = "+c.id, "total = "+strconv.Itoa(total))
}
