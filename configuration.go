package main

import (
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/codegangsta/cli"
)

// Configuration type from CLI and config files.
type Configuration struct {
	context *cli.Context
}

// NewConfiguration creates a new configuration from CLI data.
func NewConfiguration(c *cli.Context) *Configuration {
	return &Configuration{context: c}
}

// RsaBits returns the current set RSA bit length for private keys
func (c *Configuration) RsaBits() int {
	return c.context.GlobalInt("rsa-key-size")
}

func (c *Configuration) Solvers() []string {
	return c.context.GlobalStringSlice("solvers")
}

// ServerPath returns the OS dependent path to the data for a specific CA
func (c *Configuration) ServerPath() string {
	srv, _ := url.Parse(c.context.GlobalString("server"))
	srvStr := strings.Replace(srv.Host, ":", "_", -1)
	return strings.Replace(srvStr, "/", string(os.PathSeparator), -1)
}

func (c *Configuration) CertPath() string {
	return path.Join(c.context.GlobalString("path"), "certificates")
}

// AccountsPath returns the OS dependent path to the
// local accounts for a specific CA
func (c *Configuration) AccountsPath() string {
	return path.Join(c.context.GlobalString("path"), "accounts", c.ServerPath())
}

// AccountPath returns the OS dependent path to a particular account
func (c *Configuration) AccountPath(acc string) string {
	return path.Join(c.AccountsPath(), acc)
}

// AccountPath returns the OS dependent path to the keys of a particular account
func (c *Configuration) AccountKeysPath(acc string) string {
	return path.Join(c.AccountPath(acc), "keys")
}
