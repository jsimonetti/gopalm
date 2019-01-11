package ssh

import (
	"gopalm/transport/tests/mockDevice"
	"gopalm/transport/tests/mock_nxos_cli"
	"os"
	"testing"
	"time"

	"gopalm/expect"
)

func TestShellExpect(t *testing.T) {
	if testing.Verbose() {
		mock_nxos_cli.Verbose = true
	}
	host, port, d := mockDevice.New(t, mockDevice.SSH, mock_nxos_cli.Handle)
	defer d()

	c := New(host, port)
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	exp := expect.Create(c, func() {})
	defer exp.Close()

	exp.SetTimeout(5 * time.Second)
	if testing.Verbose() {
		exp.SetLogger(expect.StderrLogger())
	}

	if err := expect.WaitFor(exp, "prompt>", "enable\r\n", 0); err != nil {
		t.Fatal(err)
	}
	if err := expect.WaitFor(exp, "prompt#", "configure terminal\r\n", 0); err != nil {
		t.Fatal(err)
	}
	if err := expect.WaitFor(exp, "prompt\\(config\\)#", "exit\r\n", 0); err != nil {
		t.Fatal(err)
	}
	if err := expect.WaitFor(exp, "prompt#", "exit\r\n", 0); err != nil {
		t.Fatal(err)
	}
}

func TestSFTPPull(t *testing.T) {
	host, port, d := mockDevice.New(t, mockDevice.SFTP, nil)
	defer d()

	c := New(host, port, User("testuser"), Password("tiger"))
	if err := c.Connect(); err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	err := c.Pull("client_test.go", "tmp-pull_dst")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("tmp-pull_dst")
}

func TestSFTPPush(t *testing.T) {
	host, port, d := mockDevice.New(t, mockDevice.SFTP, nil)
	defer d()

	c := New(host, port, User("testuser"), Password("tiger"))
	if err := c.Connect(); err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	err := c.Push("client_test.go", "tmp-push_dst")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("tmp-push_dst")
}
