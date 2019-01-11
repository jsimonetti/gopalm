package transport

import "io"

type Transport interface {
	io.ReadWriteCloser
	Start() error
	Pull(source, destination string) error
	Push(source, destination string) error
}
