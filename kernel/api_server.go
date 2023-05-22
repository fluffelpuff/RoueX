package kernel

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"

	"github.com/fluffelpuff/RoueX/static"
	"github.com/fluffelpuff/RoueX/utils"
)

// Stellt die KernelAPI dar
type KernelAPI struct {
	_process_connections []*APIProcessConnectionWrapper
	_socket              net.Listener
	_lock                sync.Mutex
	_socket_unix_path    string
	_kernel              *Kernel
	_is_running          bool
	_signal_shutdown     bool
	_object_id           string
}

// Registriert den Kernel in der API
func (obj *KernelAPI) _register_kernel(kernel *Kernel) error {
	obj._lock.Lock()
	obj._kernel = kernel
	obj._lock.Unlock()

	log.Println("KernelAPI: registered in kernel. id =", obj._object_id, ", kernel =", kernel.GetKernelID(), ", path =", obj._socket_unix_path)
	return nil
}

// Gibt an ob die Acceptor Schleife ausgeführt werden soll
func (obj *KernelAPI) _ral() bool {
	obj._lock.Lock()
	r := obj._signal_shutdown
	obj._lock.Unlock()
	return !r
}

// Gibt an ob die API noch ausgeführt wird
func (obj *KernelAPI) _irn() bool {
	// Es wird geprüft
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob ein Shutdown Signal eingegangen ist
	if !obj._signal_shutdown {
		return true
	}

	// Es wird geprüft ob die Verbindung vorhanden ist
	for i := range obj._process_connections {
		if obj._process_connections[i].IsConnected() {
			return true
		}
	}

	// Der Endgültüge Status wird zurückgegeben
	return obj._is_running
}

// Handelt eine neue Verbindung
func (obj *KernelAPI) _handle_conn(conn net.Conn) {
	// Der RPC Server wird erstellt
	rpcServer := rpc.NewServer()

	// Die Prozess ID wird erstellt
	process_id := utils.RandStringRunes(16)

	// Das Wrapper Objekt wird erzeugt
	wrapper_obj := &APIProcessConnectionWrapper{conn: conn, lock: new(sync.Mutex), isconn: true, id: process_id}

	// Die Funktionen werden bereitgestellt
	pkf := &Kf{_kernel: obj._kernel, _process_id: process_id, _connection: wrapper_obj}

	// Die RPC Funktionen werden registriert
	err := rpcServer.Register(pkf)
	if err != nil {
		panic(err)
	}

	// Log
	log.Printf("KernelAPI: new api connection established. connection = %s\n", process_id)

	// Der JSON RPC wird ausgeführt
	go rpcServer.ServeConn(wrapper_obj)

	// Die Schleife wird solange ausgeführt, solange die Verbindung verbunden ist
	for wrapper_obj.IsConnected() {
		time.Sleep(1 * time.Millisecond)
	}

	// Log
	log.Printf("KernelAPI: api connection closed. connection = %s\n", process_id)

	// Dem API Prozess Handler wird Signalisiert dass alle Vorgänge verworfen werden sollen
	wrapper_obj.Kill()
}

// Startet das API Interface
func (obj *KernelAPI) _start_by_kernel() error {
	// Der Lokale nicht Root CLI Socket wird erstellt
	t := make(chan error)
	go func(err chan error) {
		// Es wird Signalisiert dass der Server ausgeführt wird
		obj._lock.Lock()
		obj._is_running = true
		obj._lock.Unlock()

		// Diese Funktion prüft 10 MS ob der CLI Server gestartet wurde
		go func(err chan error) {
			for i := 1; i <= 10; i++ {
				obj._lock.Lock()
				t := obj._is_running
				obj._lock.Unlock()
				if !t {
					err <- fmt.Errorf("rpc server not started")
					break
				}
			}
			err <- nil
		}(err)

		// Diese Schleife wird solange ausgeführt bis das Objekt geschlossen wurde
		for obj._ral() {
			// Nimmt neue Verbindungen entgegen
			conn, err := obj._socket.Accept()
			if err != nil {
				obj._lock.Lock()
				if obj._signal_shutdown {
					obj._lock.Unlock()
					break
				}
				obj._lock.Unlock()
				fmt.Println("KernelAPI: error by accepting new api-connection. id =", obj._object_id)
			}

			// Startet die Serverseitige verwaltung der Verbindung
			go obj._handle_conn(conn)
		}

		// Es wird Signalisiert dass der Server nicht mehr ausgeführt wird
		obj._lock.Lock()
		obj._is_running = false
		obj._lock.Unlock()

	}(t)

	// Es wird geprüft ob beim Starten des RPC Servers aufgetreten ist
	if err := <-t; err != nil {
		return fmt.Errorf("KernelAPI: internal error: " + err.Error())
	}

	// Log
	log.Println("KernelAPI: started by kernel. id =", obj._object_id, ", path =", obj._socket_unix_path)

	// Der Vorgang wurde ohne fehler beendet
	return nil
}

// Schließt die API durch den Kernel
func (obj *KernelAPI) _close_by_kernel() {
	// Der Threadlock wird verwendet
	obj._lock.Lock()

	// Es wird Signalisiert dass die Verbindung getrennt wurde
	obj._signal_shutdown = true

	// Der Socket wird geschlossen
	obj._socket.Close()

	// Die einzelnen API Verbindungen werden geschlossen
	obj._lock.Unlock()

	// Wartet bis alle Verbindungen geschlossen wurden
	for range time.Tick(1 * time.Millisecond) {
		if !obj._irn() {
			break
		}
	}

	// Log
	log.Println("KernelAPI: closed by kernel. id =", obj._object_id, ", rpc path =", obj._socket_unix_path)
}

// Erstellt eine neue Kernel API
func newKernelAPI() (*KernelAPI, error) {
	// Gibt die
	_rpc_socket_path := static.GetFilePathFor(static.API_SOCKET)

	// Überprüfen, ob die Datei vorhanden ist
	delete := true
	if _, err := os.Stat(_rpc_socket_path); os.IsNotExist(err) {
		delete = false
	}
	if delete {
		err := os.Remove(_rpc_socket_path)
		if err != nil {
			return nil, err
		}
	}

	// Der Unix Socket wird vorbereitet
	l, err := net.Listen("unix", _rpc_socket_path)
	if err != nil {
		return nil, fmt.Errorf("newKernelAPI: " + err.Error())
	}

	// Es wird eine ObjektId erstellt
	obj_id := utils.RandStringRunes(12)

	// Log
	log.Println("KernelAPI: new api created. id =", obj_id, ", rpc path =", _rpc_socket_path)

	// Das Onjekt wird zurückgegeben
	rewa := KernelAPI{_socket: l, _socket_unix_path: _rpc_socket_path, _lock: sync.Mutex{}, _object_id: obj_id}
	return &rewa, nil
}
