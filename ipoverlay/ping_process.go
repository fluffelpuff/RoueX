package ipoverlay

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fluffelpuff/RoueX/kernel"
)

type PingResult struct {
	State     uint8
	TotalTime float64
}

type PingProcess struct {
	ObjectId      string
	ProcessId     []byte
	CreatedAt     time.Time
	ResultChannel chan PingResult
	_is_closed    bool
	_lock         *sync.Mutex
}

func (obj *PingProcess) untilWaitOfPong() (uint64, error) {
	result := <-obj.ResultChannel
	if result.State != 1 {
		return 0, fmt.Errorf("PingProcess:untilWaitOfPong: invalid ping result")
	}
	return uint64(result.TotalTime), nil
}

func (obj *PingProcess) GetPingPackage() (*PingPackage, error) {
	return &PingPackage{PingId: obj.ProcessId}, nil
}

func (obj *PingProcess) _signal_recived_pong() {
	obj._lock.Lock()
	if !obj._is_closed {
		obj._is_closed = true
		log.Println("PingProcess: close ping process by pong. pingid =", hex.EncodeToString(obj.ProcessId))
		ntime := time.Until(obj.CreatedAt).Seconds()
		if ntime < 1 {
			ntime = 1
		}
		obj.ResultChannel <- PingResult{State: 1, TotalTime: ntime}
	} else {
		log.Println("PingProcess: alwasy closed. pingid =", hex.EncodeToString(obj.ProcessId))
	}
	obj._lock.Unlock()
}

func (obj *PingProcess) _signal_abort() {
	obj._lock.Lock()
	if !obj._is_closed {
		obj._is_closed = true
		log.Println("PingProcess: aborted. pingid =", hex.EncodeToString(obj.ProcessId))
		ntime := time.Until(obj.CreatedAt).Seconds()
		if ntime < 1 {
			ntime = 1
		}
		obj.ResultChannel <- PingResult{State: 0, TotalTime: ntime}
	} else {
		log.Println("PingProcess: aborted. pingid =", hex.EncodeToString(obj.ProcessId))
	}
	obj._lock.Unlock()
}

func newPingProcess() (*PingProcess, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	r := &PingProcess{ProcessId: randomBytes, ResultChannel: make(chan PingResult), CreatedAt: time.Now(), ObjectId: kernel.RandStringRunes(16), _lock: new(sync.Mutex)}
	log.Println("PingProcess: created new process. pingid =", hex.EncodeToString(randomBytes))
	return r, nil
}
