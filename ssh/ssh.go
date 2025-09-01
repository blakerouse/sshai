package ssh

import (
	"errors"
	"fmt"
	"net/url"

	"golang.org/x/crypto/ssh"
)

// ErrNotConnected returned when the client is not connected.
var ErrNotConnected = errors.New("not connected")

// Client is an SSH client.
type Client struct {
	hostPort string
	config   *ssh.ClientConfig

	client *ssh.Client
}

// NewClient creates the client with the hostPort and configuration.
func NewClient(hostPort string, config *ssh.ClientConfig) *Client {
	return &Client{
		hostPort: hostPort,
		config:   config,
	}
}

// NewClientFromString creates a client from a connection striong.
//
// Currently this requires the connection string to include both the username and password
// and it ignores the SSH host key.
func NewClientFromString(connStr string) (*Client, error) {
	sshURL, err := url.Parse(connStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SSH connection string: %w", err)
	}
	if sshURL.Scheme != "ssh" {
		return nil, errors.New("invalid SSH connection string: not ssh scheme")
	}
	if sshURL.User == nil {
		return nil, errors.New("invalid SSH connection string: missing user info")
	}

	user := sshURL.User.Username()
	if user == "" {
		return nil, errors.New("invalid SSH connection string: missing username")
	}
	pass, _ := sshURL.User.Password()
	if pass == "" {
		return nil, errors.New("invalid SSH connection string: missing password")
	}
	host := sshURL.Hostname()
	if host == "" {
		return nil, errors.New("invalid SSH connection string: missing host")
	}

	port := sshURL.Port()
	if port == "" {
		port = "22" // default SSH port
	}

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			// TODO: Use SSH key instead.
			ssh.Password(pass),
		},
		// TODO: This should be improved to not ignore the host keys.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return NewClient(fmt.Sprintf("%s:%s", host, port), sshConfig), nil
}

// Connect connects to the SSH server.
func (c *Client) Connect() error {
	var err error
	c.client, err = ssh.Dial("tcp", c.hostPort, c.config)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH server: %w", err)
	}
	return nil
}

// Close closes the connection to the SSH server.
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Exec runs a command on the remote SSH server.
func (c *Client) Exec(cmd string) ([]byte, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return nil, err
	}
	return output, nil
}
