package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/xenolf/lego/acme"
)

func checkFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0700)
	}
	return nil
}

func run(c *cli.Context) {
	err := checkFolder(c.GlobalString("path"))
	if err != nil {
		logger().Fatalf("Cound not check/create path: %v", err)
	}

	conf := NewConfiguration(c)

	//TODO: move to account struct? Currently MUST pass email.
	if !c.GlobalIsSet("email") {
		logger().Fatal("You have to pass an account (email address) to the program using --email or -m")
	}

	acc := NewAccount(c.GlobalString("email"), conf)
	client := acme.NewClient(c.GlobalString("server"), acc)
	if acc.Registration == nil {
		reg, err := client.Register()
		if err != nil {
			logger().Fatalf("Could not complete registration -> %v", err)
		}

		acc.Registration = reg
		acc.Save()

		logger().Print("!!!! HEADS UP !!!!")
		logger().Printf(`
			Your account credentials have been saved in your Let's Encrypt
			configuration directory at "%s".
			You should make a secure backup	of this folder now. This
			configuration directory will also contain certificates and
			private keys obtained from Let's Encrypt so making regular
			backups of this folder is ideal.

			If you lose your account credentials, you can recover
			them using the token
			"%s".
			You must write that down and put it in a safe place.`, c.GlobalString("config-dir"), reg.Body.Recoverytoken)

	}

	if acc.Registration.Body.Agreement == "" {
		if !c.GlobalBool("agree-tos") {
			reader := bufio.NewReader(os.Stdin)
			logger().Printf("Please review the TOS at %s", acc.Registration.TosURL)

			for {
				logger().Println("Do you accept the TOS? Y/n")
				text, err := reader.ReadString('\n')
				if err != nil {
					logger().Fatalf("Could not read from console -> %v", err)
				}

				text = strings.Trim(text, "\r\n")

				if text == "n" {
					logger().Fatal("You did not accept the TOS. Unable to proceed.")
				}

				if text == "Y" || text == "y" || text == "" {
					err = client.AgreeToTos()
					if err != nil {
						logger().Fatalf("Could not agree to tos -> %v", err)
					}
					acc.Save()
					break
				}

				logger().Println("Your input was invalid. Please answer with one of Y/y, n or by pressing enter.")
			}
		}
	}

	if !c.GlobalIsSet("domains") {
		logger().Fatal("Please specify --domains")
	}

	client.ObtainCertificates(c.GlobalStringSlice("domains"))
}
