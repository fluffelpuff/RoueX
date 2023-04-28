package ipoverlay

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/fluffelpuff/RoueX/kernel"
	"github.com/gorilla/websocket"
)

type WebsocketKernelConnection struct {
}

func (obj *WebsocketKernelConnection) IsConnected() bool {
	return true
}

func (obj *WebsocketKernelConnection) RegisterKernel(kernel *kernel.Kernel) error {
	return nil
}

func (obj *WebsocketKernelConnection) Write(data []byte) error {
	return nil
}

func (obj *WebsocketKernelConnection) Read() ([]byte, error) {
	return nil, nil
}

func createFinallyKernelConnection(conn *websocket.Conn, local_otk_key_pair_id string, relay_public_key *btcec.PublicKey, relay_otk_public_key *btcec.PublicKey) (*WebsocketKernelConnection, error) {
	return &WebsocketKernelConnection{}, nil
}
