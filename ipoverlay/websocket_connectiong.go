package ipoverlay

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/gorilla/websocket"
)

type WebsocketKernelConnection struct {
}

func (obj *WebsocketKernelConnection) IsConnected() bool {
	return true
}

func createFinallyKernelConnection(conn *websocket.Conn, local_otk_key_pair_id string, relay_public_key *btcec.PublicKey, relay_otk_public_key *btcec.PublicKey) (*WebsocketKernelConnection, error) {
	return nil, nil
}
