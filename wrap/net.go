package wrap

import (
	"net"
)

type (
	Net interface {
		InterfaceByName(name string) (*net.Interface, error)
		Interfaces() ([]net.Interface, error)
	}

	netImpl struct{}
)

func NewNet() Net {
	return &netImpl{}
}

func (*netImpl) InterfaceByName(name string) (*net.Interface, error) {
	return net.InterfaceByName(name)
}

func (*netImpl) Interfaces() ([]net.Interface, error) {
	return net.Interfaces()
}
