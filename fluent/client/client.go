package client

import (
	// "fmt"
	"net"
	"time"

	"github.com/tinylib/msgp/msgp"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const (
	DEFAULT_CONNECTION_TIMEOUT time.Duration = 60 * time.Second
)

type Client struct {
	ConnectionFactory
	Timeout time.Duration
	Session *Session
}

type ServerAddress struct {
	Hostname string
	Port     int
}

type AuthInfo struct {
	SharedKey []byte
	Username  string
	Password  string
}

type Session struct {
	ServerAddress
	SharedKey  []byte
	Connection net.Conn
	AuthInfo   AuthInfo
}

// ConnectionFactory implementations create new connections
//counterfeiter:generate . ConnectionFactory
type ConnectionFactory interface {
	New() (net.Conn, error)
}

// Connect initializes the Session and Connection objects by opening
// a client connect to the target configured in the ConnectionFactory
func (c *Client) Connect() error {
	conn, err := c.New()
	if err != nil {
		return err
	}

	c.Session = &Session{
		Connection: conn,
	}
	return nil
}

func (c *Client) Reconnect() error {
	// var t time.Duration
	// if c.Timeout != 0 {
	// 	t = c.Timeout
	// } else {
	// 	t = DEFAULT_CONNECTION_TIMEOUT
	// }
	//
	// if c.Session != nil {
	// 	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.Session.Hostname,
	// 		c.Session.Port), t)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	c.Session.Connection = conn
	// }

	return nil
}

// Disconnect terminates a client connection
func (c *Client) Disconnect() {
	if c.Session != nil {
		if c.Session.Connection != nil {
			c.Session.Connection.Close()
		}
	}
}

// SendMessage sends a single msgp.Encodable across the wire
func (c *Client) SendMessage(e msgp.Encodable) error {
	w := msgp.NewWriter(c.Session.Connection)
	e.EncodeMsg(w)
	w.Flush()
	return nil
}

// func (c *Client) Handshake() error {
//
// 	return nil
// }
