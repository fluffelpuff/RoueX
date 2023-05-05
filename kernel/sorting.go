package kernel

type ByRelayConnection []RelayConnection

func (a ByRelayConnection) Len() int { return len(a) }

func (a ByRelayConnection) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a ByRelayConnection) Less(i, j int) bool {
	return a[i].GetPingTime() < a[j].GetPingTime()
}
