package utils

import (
<<<<<<< HEAD:configuration.go
	"fmt"
=======
	"log"
>>>>>>> Cobra CLI run command:cmd/utils/configuration.go
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xenolf/lego/acme"
)

// Configuration type from CLI and config files.
type Configuration struct {
	command *cobra.Command
}

func CheckFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0700)
	}
	return nil
}

// NewConfiguration creates a new configuration from CLI data.
func NewConfiguration(c *cobra.Command) *Configuration {
	return &Configuration{command: c}
}

<<<<<<< HEAD:configuration.go
// KeyType the type from which private keys should be generated
func (c *Configuration) KeyType() (acme.KeyType, error) {
	switch strings.ToUpper(c.context.GlobalString("key-type")) {
	case "RSA2048":
		return acme.RSA2048, nil
	case "RSA4096":
		return acme.RSA4096, nil
	case "RSA8192":
		return acme.RSA8192, nil
	case "EC256":
		return acme.EC256, nil
	case "EC384":
		return acme.EC384, nil
	}

	return "", fmt.Errorf("Unsupported KeyType: %s", c.context.GlobalString("key-type"))
=======
// RsaBits returns the current set RSA bit length for private keys
func (c *Configuration) RsaBits() int {
	i, err := c.command.PersistentFlags().GetInt("rsa-key-size")
    if err != nil {
        log.Fatalln(err.Error())
    }
	return i
>>>>>>> Cobra CLI run command:cmd/utils/configuration.go
}

// ExcludedSolvers is a list of solvers that are to be excluded.
func (c *Configuration) ExcludedSolvers() (cc []acme.Challenge) {
    exclude, err := c.command.PersistentFlags().GetStringSlice("exclude")
    if err != nil {
        log.Fatalln(err.Error())
    }
	for _, s := range exclude {
		cc = append(cc, acme.Challenge(s))
	}
	return
}

// ServerPath returns the OS dependent path to the data for a specific CA
func (c *Configuration) ServerPath() string {
    server, err := c.command.PersistentFlags().GetString("server")
    if err != nil {
        log.Fatalln(err.Error())
    }
	srv, _ := url.Parse(server)
	srvStr := strings.Replace(srv.Host, ":", "_", -1)
	return strings.Replace(srvStr, "/", string(os.PathSeparator), -1)
}

// CertPath gets the path for certificates.
func (c *Configuration) CertPath() string {
    p, err := c.command.PersistentFlags().GetString("path")
    if err != nil {
        log.Fatalln(err.Error())
    }
	return path.Join(p, "certificates")
}

// AccountsPath returns the OS dependent path to the
// local accounts for a specific CA
func (c *Configuration) AccountsPath() string {
    p, err := c.command.PersistentFlags().GetString("path")
    if err != nil {
        log.Fatalln(err.Error())
    }
	return path.Join(p, "accounts", c.ServerPath())
}

// AccountPath returns the OS dependent path to a particular account
func (c *Configuration) AccountPath(acc string) string {
	return path.Join(c.AccountsPath(), acc)
}

// AccountKeysPath returns the OS dependent path to the keys of a particular account
func (c *Configuration) AccountKeysPath(acc string) string {
	return path.Join(c.AccountPath(acc), "keys")
}
