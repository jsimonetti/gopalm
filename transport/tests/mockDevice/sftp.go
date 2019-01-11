package mockDevice

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
)

func newSFTP(_ Handler) Server {

	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	srv := &sftp_server{
		config: &ssh.ServerConfig{PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// Should use constant-time compare (or better, salt+hash) in
			// a production setting.
			if c.User() == "testuser" && string(pass) == "tiger" {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
		},
	}

	return srv
}

type sftp_server struct {
	config *ssh.ServerConfig
	nConn  net.Conn
}

func (s *sftp_server) Serve(l net.Listener) (err error) {
	private, err := ssh.ParsePrivateKey(privKey)
	if err != nil {
		log.Fatal("Failed to parse private key", err)
	}

	s.config.AddHostKey(private)

	s.nConn, err = l.Accept()
	if err != nil {
		return err
	}

	// Before use, a handshake must be performed on the incoming
	// net.Conn.
	_, chans, reqs, err := ssh.NewServerConn(s.nConn, s.config)
	if err != nil {
		return err
	}

	// The incoming Request channel must be serviced.
	go ssh.DiscardRequests(reqs)

	// Service the incoming Channel channel.
	for newChannel := range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of an SFTP session, this is "subsystem"
		// with a payload string of "<length=4>sftp"
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Fatal("could not accept channel.", err)
		}

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".  Here we handle only the
		// "subsystem" request.
		go func(in <-chan *ssh.Request) {
			for req := range in {
				ok := false
				switch req.Type {
				case "subsystem":
					if string(req.Payload[4:]) == "sftp" {
						ok = true
					}
				}
				req.Reply(ok, nil)
			}
		}(requests)

		server, err := sftp.NewServer(
			channel,
		)
		if err != nil {
			log.Fatal(err)
		}
		if err := server.Serve(); err == io.EOF {
			server.Close()
		} else if err != nil {
			log.Fatal("sftp server completed with error:", err)
		}
	}

	return nil
}

func (s *sftp_server) Close() error {
	return nil
}

var privKey = []byte(`-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABFwAAAAdzc2gtcn
NhAAAAAwEAAQAAAQEAuI8PoR/XRVcd8wCXKxgZPNLvE03W/BwoR8/Sdn2cs3/VMUmpM+rA
zSLUKW/+4fWLIPEzlCwe0AMs4MsD3QGf121pxpgFfC09FZiN2VNK2+FGj//g3DMLMe5GOK
z5ZxLH2wqI2YtJqaTl4qUOQBgSg+NGSzH7N0JH3aYPHRDb3S1CUy41y+v8arnQt7vh/+9e
UFeLBJaVGoDukVutbzcm1ElzKFeB6SUOauZnaRSRSR1FD6it+0T18FBPZa/Rbr8WT1J7+d
8RRr4xEKbWAmJGZMX1uTQweZ36bqSLYRfDgnTEmDOKmSQfGwc+evGAiabMp9AYzXvoRNTn
SKWpX2rrkwAAA9DPHXxmzx18ZgAAAAdzc2gtcnNhAAABAQC4jw+hH9dFVx3zAJcrGBk80u
8TTdb8HChHz9J2fZyzf9UxSakz6sDNItQpb/7h9Ysg8TOULB7QAyzgywPdAZ/XbWnGmAV8
LT0VmI3ZU0rb4UaP/+DcMwsx7kY4rPlnEsfbCojZi0mppOXipQ5AGBKD40ZLMfs3Qkfdpg
8dENvdLUJTLjXL6/xqudC3u+H/715QV4sElpUagO6RW61vNybUSXMoV4HpJQ5q5mdpFJFJ
HUUPqK37RPXwUE9lr9FuvxZPUnv53xFGvjEQptYCYkZkxfW5NDB5nfpupIthF8OCdMSYM4
qZJB8bBz568YCJpsyn0BjNe+hE1OdIpalfauuTAAAAAwEAAQAAAQBF+ukAPWSRBFF03Np1
GrQnHgxNE4zbF4omgKTbDRIn9ebOw5GHABKPNg+gjrjk0QgqO4tFOd2NHkccDZ6vZHhJZV
FgXjBmP3kUAT54E18lNKxe2bVXiXtLOYAi6WPAM5zYb4wogOozizUn1VIr93S90aXLyW3q
LBW388lzSfs0R9ow/BGljWCjQ7cNnA0aaZVBzPMheZMrc875ScFcEy6JQe4IFzUCPSGttM
Gm0/90v5vcF7nuBKt5WH8PVx7GgxZTupZZXfmSmO8xSOv8Exz/aOH/vR8UjaGVYyIEIgQ8
TL9ovrWH3IsSQvQ1VmPMJRgpWzttl9LkZ6/sohRMsEUxAAAAgQDA1XtCuRxEveY5nH5WGS
PEGwSVPz9qGUDTjIfkLyaQFJR8h0Poety085a4Iqag49xiyEFMEvC6CKd49eq9LjziGTM+
GY7BRdHp/4AR9AgcSdw07fWQXMuOdghnU+oY8RFtTYoDp5Z9nJyrGn4TmHO+vJVo/Ape7P
mc0WTD22ABEAAAAIEA2gENTgkzyj8/2ND5VFY37PTXagnE89YvXdqTtthK3t3rFUAmb3uG
3mgrEsNwy9Fu0FAP2SE1ZlSghRAt/CZ748w3pr0chAgIZNnZL6UKWmEiC7i6acbIG/M0cx
qewiKjkiFnb1fwrHFqNRdjtgXMrY77uCEdEa03exMxIM3VyIUAAACBANi5vdDk/k7ieWNg
eYwiLcsc92TC5LPHaM1BuXOdM5V9QXeev2P2YiD7ZHOo5hIlgNVXXhcdnSd/dcQyIP6zCC
jHYDRdX0YYQjFlbZpyCe0Zz57VXIK1wPnAOdoEyboABVPJC7T+N8PaZwzjdHpnFIeSxKDd
nu5d0NiPWUkvHKs3AAAAFWpzaW1vbmV0dGlAanNpbW9uZXR0aQECAwQF
-----END OPENSSH PRIVATE KEY-----`)
