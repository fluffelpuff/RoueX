package kernel

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"

	"github.com/fluffelpuff/RoueX/utils"
)

type KernelAPI struct {
	_channel_socket      net.Listener
	_socket              net.Listener
	_lock                sync.Mutex
	_socket_unix_path    string
	_ch_socket_unix_path string
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
	obj._lock.Lock()
	r := obj._is_running
	obj._lock.Unlock()
	return r
}

// Wird als Thread ausgeführt um eingehende Channel verbindungen entgegen zu nehmen
func (obj *KernelAPI) _channel_server_thr(result chan error) {
	result <- nil
}

// Startet das API Interface
func (obj *KernelAPI) _start_by_kernel() error {
	// Der RPC Server wird erstellt
	rpcServer := rpc.NewServer()
	err := rpcServer.Register(&Kf{_kernel: obj._kernel, _kernel_api: obj})
	if err != nil {
		return fmt.Errorf("KernelAPI:_start_by_kernel: " + err.Error())
	}

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
			go func() {
				log.Println("KernelAPI: new api connection established.")
				rpcServer.ServeConn(conn)
			}()
		}

		// Es wird Signalisiert dass der Server nicht mehr ausgeführt wird
		obj._lock.Lock()
		obj._is_running = false
		obj._lock.Unlock()

		// Log
		log.Println("KernelAPI: stoped. id =", obj._object_id)
	}(t)

	// Es wird geprüft ob beim Starten des RPC Servers aufgetreten ist
	if err := <-t; err != nil {
		return fmt.Errorf("KernelAPI: internal error: " + err.Error())
	}

	// Es wird versucht den Channel Server zu starten
	c := make(chan error)
	go obj._channel_server_thr(c)

	// Es wird geprüft ob beim Starten des Channel servers ein Fehler aufgetreten ist
	if err := <-c; err != nil {
		return fmt.Errorf("KernelAPI: internal error: " + err.Error())
	}

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
	obj._channel_socket.Close()
	obj._lock.Unlock()

	for obj._irn() {
		time.Sleep(1 * time.Millisecond)
	}

	log.Println("KernelAPI: closed by kernel. id =", obj._object_id, ", rpc path =", obj._socket_unix_path, "channel path =", obj._ch_socket_unix_path)
}

// Erstellt eine neue Kernel API
func newKernelAPI(_rpc_socket_path string, _app_channel_socket_path string) (*KernelAPI, error) {
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
	if _, err := os.Stat(_app_channel_socket_path); os.IsNotExist(err) {
		delete = false
	}
	if delete {
		err := os.Remove(_app_channel_socket_path)
		if err != nil {
			return nil, err
		}
	}

	// Der Unix Socket wird vorbereitet
	l, err := net.Listen("unix", _rpc_socket_path)
	if err != nil {
		return nil, fmt.Errorf("newKernelAPI: " + err.Error())
	}

	// Der Unix Socket für die Channels wird erstellt
	r, err := net.Listen("unix", _app_channel_socket_path)
	if err != nil {
		return nil, fmt.Errorf("newKernelAPI: " + err.Error())
	}

	// Es wird eine ObjektId erstellt
	obj_id := utils.RandStringRunes(12)

	// Log
	log.Println("KernelAPI: new api created. id =", obj_id, ", rpc path =", _rpc_socket_path, "channel path =", _app_channel_socket_path)

	// Das Onjekt wird zurückgegeben
	rewa := KernelAPI{_socket: l, _socket_unix_path: _rpc_socket_path, _lock: sync.Mutex{}, _object_id: obj_id, _channel_socket: r, _ch_socket_unix_path: _app_channel_socket_path}
	return &rewa, nil
}
