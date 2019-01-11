package mock_nxos_cli

import (
	"fmt"
	"io"
	"time"

	"github.com/google/goexpect"
)

var (
	TimeOut = 30 * time.Second
	Verbose = false
)

func Handle(rw io.ReadWriteCloser) (int, error) {
	resCh := make(chan error)

	exp, _, err := expect.SpawnGeneric(&expect.GenOptions{
		In:  rw,
		Out: rw,
		Wait: func() error {
			return <-resCh
		},
		Close: func() error {
			close(resCh)
			return rw.Close()
		},
		Check: func() bool { return true },
	}, TimeOut, expect.Verbose(Verbose))

	if err != nil {
		return 1, err
	}

	var state error = stateExit
	var noPrompt = false
	var timeOutErr = fmt.Sprintf("expect: timer expired after %d seconds", time.Duration(TimeOut)/time.Second)

	caser := caserEnable
	prompt := "prompt> "

	for {
		if !noPrompt {
			exp.Send(prompt)
		}
		noPrompt = false

		_, _, _, state = exp.ExpectSwitchCase(
			caser,
			30*time.Second,
		)
		if state.Error() == timeOutErr {
			noPrompt = true
			continue
		}

		switch state {
		case stateEnable: // enter enable mode
			prompt = "prompt# "
			caser = caserOper
		case stateConfigure: // enter configure mode
			prompt = "prompt(config)# "
			caser = caserConfigure
		case stateExit: // exit session
			return 0, nil
		case stateUnknownCommand: // unknown command entered
		}
	}
}
