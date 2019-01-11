package expect

import (
	"io"
	"regexp"
	"time"

	"github.com/jamesharr/expect"
)

func Create(pty io.ReadWriteCloser, killer func()) *Expect {
	return expect.Create(pty, killer)
}

func StderrLogger() Logger {
	return expect.StderrLogger()
}

func WaitFor(exp *Expect, reg, send string, timeout int) error {
	if timeout == 0 {
		timeout = 100
	}

	exp.SetTimeout(time.Duration(timeout) * time.Millisecond)
	_, err := exp.ExpectRegexp(regexp.MustCompile(reg))
	if err != nil {
		return err
	}

	err = exp.Send(send)
	return err
}

type Expect = expect.Expect
type Logger = expect.Logger
