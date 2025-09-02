package ssh

import (
	"errors"
	"fmt"
	"net/url"

	"golang.org/x/crypto/ssh"
)

// ErrNotConnected returned when the client is not connected.
var ErrNotConnected = errors.New("not connected")

// OSInfo provides the OS information.
type OSInfo struct {
	Name     string `yaml:"name" json:"name" jsonschema_description:"The name of the operating system"`
	Platform string `yaml:"platform" json:"platform" jsonschema_description:"The platform of the operating system"`
	Version  string `yaml:"version" json:"version" jsonschema_description:"The version of the operating system"`
	Arch     string `yaml:"arch" json:"arch" jsonschema_description:"The architecture of the operating system"`
}

// ClientInfo stores the generate client information.
type ClientInfo struct {
	Name string `yaml:"name" json:"name" jsonschema_description:"The name of the client"`
	Host string `yaml:"host" json:"host" jsonschema_description:"The host of the client"`
	Port string `yaml:"port" json:"port" jsonschema_description:"The port of the client"`
	User string `yaml:"user" json:"user" jsonschema_description:"The user of the client"`
	Pass string `yaml:"pass" json:"pass" jsonschema_description:"The password of the client"`

	OS OSInfo `yaml:"os" json:"os" jsonschema_description:"The operating system information"`
}

// NewClientInfo returns client information from the connection string.
func NewClientInfo(name string, connStr string) (*ClientInfo, error) {
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
	if name == "" {
		name = host // default name to host (if not provided)
	}

	return &ClientInfo{
		Name: name,
		Host: host,
		Port: port,
		User: user,
		Pass: pass,
	}, nil
}

// Client is an SSH client.
type Client struct {
	info *ClientInfo

	client *ssh.Client
}

// NewClient creates the client with the hostPort and configuration.
func NewClient(info *ClientInfo) *Client {
	return &Client{
		info: info,
	}
}

// Connect connects to the SSH server.
func (c *Client) Connect() error {
	var err error
	host := fmt.Sprintf("%s:%s", c.info.Host, c.info.Port)
	cfg := &ssh.ClientConfig{
		User: c.info.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.info.Pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	c.client, err = ssh.Dial("tcp", host, cfg)
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
