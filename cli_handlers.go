package main

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
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

func setup(c *cli.Context) (*Configuration, *Account, *acme.Client) {
	err := checkFolder(c.GlobalString("path"))
	if err != nil {
		logger().Fatalf("Cound not check/create path: %v", err)
	}

	conf := NewConfiguration(c)
	if !c.GlobalIsSet("email") {
		logger().Fatal("You have to pass an account (email address) to the program using --email or -m")
	}

	//TODO: move to account struct? Currently MUST pass email.
	acc := NewAccount(c.GlobalString("email"), conf)

	client, err := acme.NewClient(c.GlobalString("server"), acc, conf.RsaBits(), conf.OptPort())
	if err != nil {
		logger().Fatal("Could not create client:", err)
	}

	return conf, acc, client
}

func saveCertRes(certRes acme.CertificateResource, conf *Configuration) {
	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	certOut := path.Join(conf.CertPath(), certRes.Domain+".crt")
	privOut := path.Join(conf.CertPath(), certRes.Domain+".key")
	metaOut := path.Join(conf.CertPath(), certRes.Domain+".json")

	err := ioutil.WriteFile(certOut, certRes.Certificate, 0600)
	if err != nil {
		logger().Printf("Unable to save Certificate for domain %s\n\t%v", certRes.Domain, err)
	}

	err = ioutil.WriteFile(privOut, certRes.PrivateKey, 0600)
	if err != nil {
		logger().Printf("Unable to save PrivateKey for domain %s\n\t%v", certRes.Domain, err)
	}

	jsonBytes, err := json.MarshalIndent(certRes, "", "\t")
	if err != nil {
		logger().Printf("Unable to marshal CertResource for domain %s\n\t%v", certRes.Domain, err)
	}

	err = ioutil.WriteFile(metaOut, jsonBytes, 0600)
	if err != nil {
		logger().Printf("Unable to save CertResource for domain %s\n\t%v", certRes.Domain, err)
	}
}

func run(c *cli.Context) {
	conf, acc, client := setup(c)
	if acc.Registration == nil {
		reg, err := client.Register()
		if err != nil {
			logger().Fatalf("Could not complete registration\n\t%v", err)
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
		backups of this folder is ideal.`, conf.AccountPath(c.GlobalString("email")))

	}

	if acc.Registration.Body.Agreement == "" {
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
				err = client.AgreeToTOS()
				if err != nil {
					logger().Fatalf("Could not agree to tos -> %v", err)
				}
				acc.Save()
				break
			}

			logger().Println("Your input was invalid. Please answer with one of Y/y, n or by pressing enter.")
		}
	}

	if !c.GlobalIsSet("domains") {
		logger().Fatal("Please specify --domains")
	}

	certs, failures := client.ObtainCertificates(c.GlobalStringSlice("domains"), true)
	if len(failures) > 0 {
		for k, v := range failures {
			logger().Printf("[%s] Could not obtain certificates\n\t%v", k, v)
		}
	}

	err := checkFolder(conf.CertPath())
	if err != nil {
		logger().Fatalf("Cound not check/create path: %v", err)
	}

	for _, certRes := range certs {
		saveCertRes(certRes, conf)
	}
}

func revoke(c *cli.Context) {

	conf, _, client := setup(c)

	err := checkFolder(conf.CertPath())
	if err != nil {
		logger().Fatalf("Cound not check/create path: %v", err)
	}

	for _, domain := range c.GlobalStringSlice("domains") {
		logger().Printf("Trying to revoke certificate for domain %s", domain)

		certPath := path.Join(conf.CertPath(), domain+".crt")
		certBytes, err := ioutil.ReadFile(certPath)

		err = client.RevokeCertificate(certBytes)
		if err != nil {
			logger().Printf("Error while revoking the certificate for domain %s\n\t%v", domain, err)
		} else {
			logger().Print("Certificate was revoked.")
		}
	}
}

func renew(c *cli.Context) {
	conf, _, client := setup(c)

	for _, domain := range c.GlobalStringSlice("domains") {
		// load the cert resource from files.
		// We store the certificate, private key and metadata in different files
		// as web servers would not be able to work with a combined file.
		certPath := path.Join(conf.CertPath(), domain+".crt")
		privPath := path.Join(conf.CertPath(), domain+".key")
		metaPath := path.Join(conf.CertPath(), domain+".json")

		certBytes, err := ioutil.ReadFile(certPath)
		if err != nil {
			logger().Printf("Error while loading the certificate for domain %s\n\t%v", domain, err)
			return
		}

		keyBytes, err := ioutil.ReadFile(privPath)
		if err != nil {
			logger().Printf("Error while loading the private key for domain %s\n\t%v", domain, err)
			return
		}

		metaBytes, err := ioutil.ReadFile(metaPath)
		if err != nil {
			logger().Printf("Error while loading the meta data for domain %s\n\t%v", domain, err)
			return
		}

		var certRes acme.CertificateResource
		err = json.Unmarshal(metaBytes, &certRes)
		if err != nil {
			logger().Printf("Error while marshalling the meta data for domain %s\n\t%v", domain, err)
			return
		}

		certRes.PrivateKey = keyBytes
		certRes.Certificate = certBytes

		newCert, err := client.RenewCertificate(certRes, true, true)
		if err != nil {
			logger().Printf("%v", err)
			return
		}

		saveCertRes(newCert, conf)
	}
}
