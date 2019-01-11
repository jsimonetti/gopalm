package ssh

import (
	"gopalm/transport"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/user"

	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Client is a SSH conn transport implementation
type Client struct {
	User     string
	Password string
	KeyFile  string
	Server   string
	Port     string
	config   ssh.ClientConfig
	conn     *ssh.Client
	stdout   io.Reader
	stdin    io.WriteCloser
}

var _ transport.Transport = &Client{}

// New returns a new Client to connect to `server:port`
// with the `opts` options applied.
func New(server string, port string, opts ...Option) *Client {
	c := &Client{Server: server, Port: port}

	for _, opt := range opts {
		c.SetOption(opt)
	}

	return c
}

// Start starts the transport Client
func (c *Client) Start() error {
	if err := c.Connect(); err != nil {
		return err
	}
	if err := c.Shell(); err != nil {
		return err
	}
	return nil
}

// Write implements io.Writer
func (c *Client) Write(p []byte) (n int, err error) {
	return c.stdin.Write(p)
}

// Read implements io.Reader
func (c *Client) Read(p []byte) (n int, err error) {
	return c.stdout.Read(p)
}

// Close implements io.Closer
func (c *Client) Close() error {
	return c.conn.Close()
}

// Connect builds an ssh config and connects to the server
func (c *Client) Connect() error {
	c.config = ssh.ClientConfig{
		User:            c.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	defer authMethods(c)()

	client, err := ssh.Dial("tcp", c.Server+":"+string(c.Port), &c.config)
	if err != nil {
		return err
	}

	c.conn = client

	return nil
}

// Pull will pull source file from the device and save to destination
func (c *Client) Pull(source, destination string) (err error) {
	client, err := sftp.NewClient(c.conn)
	if err != nil {
		return err
	}

	src, err := client.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer func() {
		cerr := dst.Sync()
		if err == nil {
			err = cerr
		} else {
			cerr = dst.Close()
			if err == nil {
				err = cerr
			}
		}
	}()

	_, err = io.Copy(dst, src)
	return
}

// Push will push source file to the device and save at destination
func (c *Client) Push(source, destination string) (err error) {
	client, err := sftp.NewClient(c.conn)
	if err != nil {
		return err
	}

	err = client.Remove(destination)
	if err != nil && err.Error() != "file does not exist" {
		return err
	}

	dst, err := client.Create(destination)
	if err != nil {
		return err
	}
	defer dst.Close()

	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() {
		cerr := src.Close()
		if err == nil {
			err = cerr

		}
	}()

	_, err = io.Copy(dst, src)
	return
}

// Shell requests a login shell on the remote device
// It also requests a new session over the existing
// connection and a Pty for a 'dumb' terminal
func (c *Client) Shell() error {
	var (
		termWidth, termHeight = 80, 24 // default terminal sizes
	)

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// request a new session on the existing connection
	session, err := c.conn.NewSession()
	if err != nil {
		return err
	}

	var stdout io.Reader
	var stderr io.Reader

	if stdout, err = session.StdoutPipe(); err != nil {
		return err
	}
	if stderr, err = session.StderrPipe(); err != nil {
		return err
	}
	// merge stdout and stderr into a single io.Reader
	c.stdout = io.MultiReader(stdout, stderr)

	if c.stdin, err = session.StdinPipe(); err != nil {
		return err
	}

	// request a dumb terminal
	// this prevents ansi colors and other smart things
	if err := session.RequestPty("dumb", termHeight, termWidth, modes); err != nil {
		return err
	}

	if err := session.Shell(); err != nil {
		return err
	}

	return nil
}

// returns ssh.Signer from user you running app home path + cutted key path.
// (ex. pubkey,err := getKeyFile("/.ssh/id_rsa") )
func getKeyFile(keypath string) (ssh.Signer, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	file := usr.HomeDir + keypath
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	pubkey, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		return nil, err
	}

	return pubkey, nil
}

// authMethods will build the correct `ssh.Config.Auth` settings
// applicable for this Client
func authMethods(c *Client) (f func()) {
	if c.Password != "" {
		c.config.Auth = append(c.config.Auth, ssh.Password(c.Password))
	}

	// use existing SSH agent if available
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		c.config.Auth = append(c.config.Auth, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
		f = func() { sshAgent.Close() }
	}

	if publicKey, err := getKeyFile(c.KeyFile); err == nil {
		c.config.Auth = append(c.config.Auth, ssh.PublicKeys(publicKey))
	}

	return
}
