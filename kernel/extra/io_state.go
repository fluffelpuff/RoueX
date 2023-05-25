package extra

import (
	"sync"
)

type SendState uint8

const (
	WAIT   SendState = 0
	SEND   SendState = 1
	DROPED SendState = 2
)

// Stellt den Aktuellen Sendestatus dar
type PackageSendState struct {
	_lock       *sync.Mutex
	_state      SendState
	_state_chan chan bool
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
	if obj._state != WAIT {
		obj._lock.Unlock()
		return
	}
	obj._state = SendState(fstate)
	obj._lock.Unlock()

	select {
	case obj._state_chan <- true:
	default:
	}
}

// Wird verwendet um zu warten bis sicher der Status geändert hat
func (obj *PackageSendState) WaitOfNewState() {
	// Es wird geprüft ob der Status geändert wurde
	if obj.GetState() != WAIT {
		return
	}

	// Es wird auf die Eintreffende Antwort gewartet
	<-obj._state_chan
}

// Erzeugt ein neues Package Sendstate
func NewPackageSendState() *PackageSendState {
	return &PackageSendState{_state: WAIT, _lock: new(sync.Mutex), _state_chan: make(chan bool, 1)}
}
