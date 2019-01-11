package mock_nxos_cli

import (
	"os"
	"testing"
)

func TestManual(t *testing.T) {
	t.Skip()
	// TODO: actually make this work
	Verbose = true
	code, err := Handle(stdinRWC{})
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Errorf("wanted code 0, got %d", code)
	}
}

type stdinRWC struct {
}

func (rwc stdinRWC) Read(p []byte) (n int, err error) {
	return os.Stdin.Read(p)
}

func (rwc stdinRWC) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

func (stdinRWC) Close() error {
	return nil
}
