package ipoverlay

import (
	"crypto/rand"
	"encoding/hex"
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
	<-obj.ResultChannel
	return 0, nil
}

func (obj *PingProcess) GetPingPackage() (*PingPackage, error) {
	return &PingPackage{PingId: obj.ProcessId}, nil
}

func (obj *PingProcess) _signal_recived_pong() {
	obj._lock.Lock()
	if !obj._is_closed {
		obj._is_closed = true
		obj._lock.Unlock()
		log.Println("Close ping process by pong. pingid =", hex.EncodeToString(obj.ProcessId))
		ntime := time.Until(obj.CreatedAt).Seconds()
		if ntime < 1 {
			ntime = 1
		}
		obj.ResultChannel <- PingResult{State: 1, TotalTime: ntime}
	} else {
		obj._lock.Unlock()
		log.Println("Ping process alwasy closed. pingid =", hex.EncodeToString(obj.ProcessId))
	}
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
