package ipoverlay

import (
	"log"

	"github.com/fluffelpuff/RoueX/kernel"
)

type TcpKernelServerEP struct {
}

func (obj *TcpKernelServerEP) RegisterKernel(kernel *kernel.Kernel) error {
	log.Printf("TCP Server EndPoint registrated on kernel %s\n", kernel.GetKernelID())
	return nil
}

func (obj *TcpKernelServerEP) Shutdown() {
	log.Printf("TCP Server EndPoint shutingdown...\n")
}

func (obj *TcpKernelServerEP) Start() error {
	log.Printf("New TCP Server EndPoint started\n")
	return nil
}

func (obj *TcpKernelServerEP) GetProtocol() string {
	return "ws"
}

func (obj *TcpKernelServerEP) GetObjectId() string {
	return ""
}

func CreateNewTCPKernelServerEP(ip_adr string, port uint64) (*TcpKernelServerEP, error) {
	log.Printf("New TCP Server EndPoint on %s and port %d created\n", ip_adr, port)
	return nil, nil
}
