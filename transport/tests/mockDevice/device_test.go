package mockDevice

import (
	"gopalm/transport/tests/mock_nxos_cli"
	"net"
	"testing"
	"time"
)

func TestSSHManual(t *testing.T) {
	t.Skip()
	lst, _ := net.Listen("tcp", ":2222")
	srv := newSSH(mock_nxos_cli.Handle)

	defer srv.Close()
	go func() {
		srv.Serve(lst)
	}()

	time.Sleep(30 * time.Second)
}

func TestSCPManual(t *testing.T) {
	t.Skip()
	lst, _ := net.Listen("tcp", ":2222")
	srv := newSFTP(nil)

	defer srv.Close()
	go func() {
		srv.Serve(lst)
	}()

	time.Sleep(30 * time.Second)
}
