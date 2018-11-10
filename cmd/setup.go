package cmd

import (
	"fmt"
	"time"

	"github.com/urfave/cli"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/log"
)

func setup(c *cli.Context) (*Configuration, *Account, *acme.Client, error) {
	err := checkFolder(c.GlobalString("path"))
	if err != nil {
		log.Fatalf("Could not check/create path: %v", err)
	}

	conf := NewConfiguration(c)
	if len(c.GlobalString("email")) == 0 {
		log.Fatal("You have to pass an account (email address) to the program using --email or -m")
	}

	// TODO: move to account struct? Currently MUST pass email.
	acc := NewAccount(c.GlobalString("email"), conf)

	keyType, err := conf.KeyType()
	if err != nil {
		log.Fatal(err)
	}

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

	setupChallenges(client, c, conf)

	if client.GetExternalAccountRequired() && !c.GlobalIsSet("eab") {
		log.Fatal("Server requires External Account Binding. Use --eab with --kid and --hmac.")
	}

	return conf, acc, client
}
