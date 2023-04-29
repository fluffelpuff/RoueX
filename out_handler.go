package main

import (
	"fmt"

	"github.com/fluffelpuff/RoueX/kernel"
)

func outboundHandler(core *kernel.Kernel) {
	rt, err := core.ListOutboundTrustedAvaileRelays()
	if err != nil {
		panic(err)
	}

	for _, o := range rt {
		client_conn := *o.GetClientConnModule()
		err := client_conn.ConnectTo(o.GetRelay().GetEndpoint(), o.GetRelay().GetPublicKey())
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(o.GetRelay().GetProtocol(), o.GetRelay().GetEndpoint())
	}
}
