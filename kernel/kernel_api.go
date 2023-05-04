package kernel

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"

	"github.com/fluffelpuff/RoueX/static"
)

type KernelAPI struct {
	_socket           net.Listener
	_lock             sync.Mutex
	_socket_unix_path string
	_kernel           *Kernel
	_is_running       bool
	_signal_shutdown  bool
	_object_id        string
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
	return r
}

// Startet das API Interface
func (obj *KernelAPI) _start_by_kernel() error {
	// Der RPC Server wird erstellt
	rpcServer := rpc.NewServer()
	err := rpcServer.Register(&Kf{})
	if err != nil {
		return fmt.Errorf("KernelAPI:_start_by_kernel: " + err.Error())
	}

	// Der Lokale nicht Root CLI Socket wird erstellt
	go func() {
		obj._lock.Lock()
		obj._is_running = true
		obj._lock.Unlock()

		for obj._ral() {
			conn, err := obj._socket.Accept()
			if err != nil {
				log.Fatal("Accept error: ", err)
			}
			go rpcServer.ServeConn(conn)
		}

		obj._lock.Lock()
		obj._is_running = false
		obj._lock.Unlock()

		log.Println("KernelAPI: stoped. id =", obj._object_id)
	}()

	// Log
	log.Println("KernelAPI: started by kernel. id =", obj._object_id, ", path =", obj._socket_unix_path)

	// Der Vorgang wurde ohne fehler beendet
	return nil
}

// Schließt die API durch den Kernel
func (obj *KernelAPI) _close_by_kernel() {
	obj._lock.Lock()
	obj._signal_shutdown = true
	obj._socket.Close()
	obj._lock.Unlock()

	log.Println("KernelAPI: closed by kernel. id =", obj._object_id, ", path =", obj._socket_unix_path)
}

// Erstellt eine neue Kernel API
func newKernelAPI(_socket_path string) (*KernelAPI, error) {
	// Überprüfen, ob die Datei vorhanden ist
	delete := true
	if _, err := os.Stat(_socket_path); os.IsNotExist(err) {
		delete = false
	}
	if delete {
		err := os.Remove(_socket_path)
		if err != nil {
			return nil, err
		}
	}

	// Der Unix Socket wird vorbereitet
	l, err := net.Listen("unix", static.GetFilePathFor(static.API_SOCKET))
	if err != nil {
		return nil, fmt.Errorf("newKernelAPI: " + err.Error())
	}

	// Es wird eine ObjektId erstellt
	obj_id := RandStringRunes(12)

	// Log
	log.Println("KernelAPI: new api created. id =", obj_id, ", path =", _socket_path)

	// Das Onjekt wird zurückgegeben
	rewa := KernelAPI{_socket: l, _socket_unix_path: _socket_path, _lock: sync.Mutex{}, _object_id: obj_id}
	return &rewa, nil
}
