package ssh

// Option is a functional option handler for Client.
type Option func(*Client) error

// SetOption runs a functional option against Client.
func (c *Client) SetOption(option Option) error {
	return option(c)
}

// User sets the user to use for authentication
func User(user string) Option {
	return func(c *Client) error {
		c.User = user
		return nil
	}
}

// Password uses password login
func Password(password string) Option {
	return func(c *Client) error {
		c.Password = password
		return nil
	}
}

// Keyfile uses key login
func Keyfile(file string) Option {
	return func(c *Client) error {
		c.KeyFile = file
		return nil
	}
}
