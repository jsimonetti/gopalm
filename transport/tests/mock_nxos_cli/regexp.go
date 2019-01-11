package mock_nxos_cli

import (
	"regexp"

	"github.com/google/goexpect"
)

var (
	stateUnknownCommand = expect.NewStatus(999, "unknown command")
	stateEnable         = expect.NewStatus(1, "enable")
	stateConfigure      = expect.NewStatus(2, "configure terminal")
	stateExit           = expect.NewStatus(3, "exit")
)

var (
	caserEnable = []expect.Caser{
		&expect.Case{
			R: regexp.MustCompile(`exit` + lfc),
			T: expect.Continue(stateExit)},
		&expect.Case{
			R: regexp.MustCompile(`enable` + lfc),
			S: "\nentering enable\n",
			T: expect.Continue(stateEnable)},
		&expect.Case{
			R: regexp.MustCompile(lfc),
			S: "\nunknown command\n",
			T: expect.Continue(stateUnknownCommand)},
	}
	caserOper = []expect.Caser{
		&expect.Case{
			R: regexp.MustCompile(`exit` + lfc),
			T: expect.Continue(stateExit)},
		&expect.Case{
			R: regexp.MustCompile(`configure terminal` + lfc),
			S: "\nentering configure\n",
			T: expect.Continue(stateConfigure)},
		&expect.Case{
			R: regexp.MustCompile(lfc),
			S: "\nunknown command\n",
			T: expect.Continue(stateUnknownCommand)},
	}
	caserConfigure = []expect.Caser{
		&expect.Case{
			R: regexp.MustCompile(`exit` + lfc),
			S: "\n",
			T: expect.Continue(stateEnable)},
		&expect.Case{
			R: regexp.MustCompile(lfc),
			S: "\nunknown command\n",
			T: expect.Continue(stateUnknownCommand)},
	}
)

const (
	lfc = `\r\n??$`
)
