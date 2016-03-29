package utils

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xenolf/lego/acme"
)

// Configuration type from CLI and config files.
type Configuration struct {
	context *cobra.Command
}

// NewConfiguration creates a new configuration from CLI data.
func NewConfiguration(c *cobra.Command) *Configuration {
	return &Configuration{context: c}
}

// KeyType the type from which private keys should be generated
func (c *Configuration) KeyType() (acme.KeyType, error) {
	keytype, err := c.context.PersistentFlags().GetString("key-type")
	if err != nil {
		return "", err
	}
	switch strings.ToUpper(keytype) {
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

	return "", fmt.Errorf("Unsupported KeyType: %s", keytype)
}

// ExcludedSolvers is a list of solvers that are to be excluded.
func (c *Configuration) ExcludedSolvers() (cc []acme.Challenge) {
	exclude, err := c.context.PersistentFlags().GetStringSlice("exclude")
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
    server, err := c.context.PersistentFlags().GetString("server")
    if err != nil {
        log.Fatalln(err.Error())
    }
	srv, _ := url.Parse(server)
	srvStr := strings.Replace(srv.Host, ":", "_", -1)
	return strings.Replace(srvStr, "/", string(os.PathSeparator), -1)
}

// CertPath gets the path for certificates.
func (c *Configuration) CertPath() string {
    pathS, err := c.context.PersistentFlags().GetString("path")
    if err != nil {
        log.Fatalln(err.Error())
    }
	return path.Join(pathS, "certificates")
}

// AccountsPath returns the OS dependent path to the
// local accounts for a specific CA
func (c *Configuration) AccountsPath() string {
    pathS, err := c.context.PersistentFlags().GetString("path")
    if err != nil {
        log.Fatalln(err.Error())
    }
	return path.Join(pathS, "accounts", c.ServerPath())
}

// AccountPath returns the OS dependent path to a particular account
func (c *Configuration) AccountPath(acc string) string {
	return path.Join(c.AccountsPath(), acc)
}

// AccountKeysPath returns the OS dependent path to the keys of a particular account
func (c *Configuration) AccountKeysPath(acc string) string {
	return path.Join(c.AccountPath(acc), "keys")
}
