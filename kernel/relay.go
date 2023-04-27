package kernel

type Relay struct {
	_db_id      int64
	_hexed_id   string
	_public_key string
	_last_used  uint64
	_end_point  string
	_active     bool
	_type       string
	_trusted    bool
}

func (obj *Relay) GetProtocol() string {
	return obj._type
}

func (obj *Relay) GetEndpoint() string {
	return obj._end_point
}

func (obj *Relay) GetPublicKey() string {
	return obj._public_key
}

type RelayOutboundPair struct {
	_relay     *Relay
	_cl_module *ClientModule
}

func (obj *RelayOutboundPair) GetRelay() *Relay {
	return obj._relay
}

func (obj *RelayOutboundPair) GetClientConnModule() *ClientModule {
	return obj._cl_module
}
