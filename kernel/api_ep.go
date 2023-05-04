package kernel

type Kf struct {
}

func (s *Kf) Hello(name string, reply *string) error {
	*reply = "Hello, " + name
	return nil
}
