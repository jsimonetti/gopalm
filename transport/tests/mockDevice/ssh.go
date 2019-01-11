package mockDevice

import (
	"io"
	"time"

	"github.com/gliderlabs/ssh"
)

func newSSH(h Handler) Server {
	handle := func(s ssh.Session) {
		defer s.Close()
		_, _, isPty := s.Pty()
		if isPty {
			code, err := h(s)
			if err != nil {
				io.WriteString(s, "error: "+err.Error())
			}
			s.Exit(code)
		} else {
			io.WriteString(s, "No PTY requested.\n")
			s.Exit(1)
		}
	}

	return &ssh.Server{Handler: handle, MaxTimeout: 5 * time.Minute}
}
