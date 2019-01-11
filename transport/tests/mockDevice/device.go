package mockDevice

import (
	"io"
	"net"
	"testing"
)

type Server interface {
	Close() error
	Serve(l net.Listener) error
}

type Handler func(io.ReadWriteCloser) (exitCode int, err error)
type Type int

const (
	_ Type = iota
	SSH
	SFTP
)

func (t Type) Server(h Handler) Server {
	switch t {
	case SSH:
		return newSSH(h)
	case SFTP:
		return newSFTP(h)
	}
	return nil
}

func New(t *testing.T, sv Type, h Handler) (hostname string, port string, deferfn func() error) {
	lst, _ := net.Listen("tcp", "localhost:0")
	_, port, _ = net.SplitHostPort(lst.Addr().String())
	srv := sv.Server(h)

	if srv == nil {
		t.Fatalf("unknown type %v", sv)
	}

	go func() {
		srv.Serve(lst)
	}()

	return "localhost", port, srv.Close
}
