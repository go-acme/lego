package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/certcrypto"
	"github.com/xenolf/lego/log"
)

func setup(c *cli.Context) (*Account, *acme.Client) {
	email := c.GlobalString("email")
	if len(email) == 0 {
		log.Fatal("You have to pass an account (email address) to the program using --email or -m")
	}

	keyType := getKeyType(c)

	// TODO: move to account struct? Currently MUST pass email.
	acc := NewAccount(c, email)

	config := acme.NewDefaultConfig(acc).
		WithKeyType(keyType).
		WithCADirURL(c.GlobalString("server")).
		WithUserAgent(fmt.Sprintf("lego-cli/%s", c.App.Version))

	if c.GlobalIsSet("http-timeout") {
		config.HTTPClient.Timeout = time.Duration(c.GlobalInt("http-timeout")) * time.Second
	}

	client, err := acme.NewClient(config)
	if err != nil {
		log.Fatalf("Could not create client: %v", err)
	}

	setupChallenges(c, client)

	if client.GetExternalAccountRequired() && !c.GlobalIsSet("eab") {
		log.Fatal("Server requires External Account Binding. Use --eab with --kid and --hmac.")
	}

	return acc, client
}

// getKeyType the type from which private keys should be generated
func getKeyType(c *cli.Context) certcrypto.KeyType {
	keyType := c.GlobalString("key-type")
	switch strings.ToUpper(keyType) {
	case "RSA2048":
		return certcrypto.RSA2048
	case "RSA4096":
		return certcrypto.RSA4096
	case "RSA8192":
		return certcrypto.RSA8192
	case "EC256":
		return certcrypto.EC256
	case "EC384":
		return certcrypto.EC384
	}

	log.Fatalf("Unsupported KeyType: %s", keyType)
	return ""
}
