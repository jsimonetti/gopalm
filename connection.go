package gomiko

type Connection interface {
	Connect()
}

type SSHConnection struct {
}
