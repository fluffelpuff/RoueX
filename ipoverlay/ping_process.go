package ipoverlay

import (
	"crypto/rand"
	"time"
)

type PingResult struct {
	State     uint8
	TotalTime uint64
}

type PingProcess struct {
	ProcessId     []byte
	CreatedAt     time.Time
	ResultChannel chan PingResult
}

func (obj *PingProcess) untilWaitOfPong() (uint64, error) {
	return 0, nil
}

func (obj *PingProcess) GetPingPackage() (*PingPackage, error) {
	return nil, nil
}

func newPingProcess() (*PingProcess, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	return &PingProcess{ProcessId: randomBytes, ResultChannel: make(chan PingResult), CreatedAt: time.Now()}, nil
}
