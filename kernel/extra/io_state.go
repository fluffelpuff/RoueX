package extra

import "sync"

type SendState uint8

const (
	WAIT   SendState = 0
	SEND   SendState = 1
	DROPED SendState = 2
)

// Stellt den Aktuellen Sendestatus dar
type PackageSendState struct {
	_lock  *sync.Mutex
	_state SendState
}

// Gibt den Aktuellen Status zurück
func (obj *PackageSendState) GetState() SendState {
	obj._lock.Lock()
	defer obj._lock.Unlock()
	return obj._state
}

// Setzt den neuen Status
func (obj *PackageSendState) SetFinallyState(fstate SendState) {
	obj._lock.Lock()
	defer obj._lock.Unlock()
	if obj._state != WAIT {
		return
	}
	obj._state = fstate
}

// Erzeugt ein neues Package Sendstate
func NewPackageSendState() *PackageSendState {
	return &PackageSendState{_state: WAIT, _lock: new(sync.Mutex)}
}
