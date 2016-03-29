package utils

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/providers/dns/cloudflare"
	"github.com/xenolf/lego/providers/dns/digitalocean"
	"github.com/xenolf/lego/providers/dns/dnsimple"
	"github.com/xenolf/lego/providers/dns/dyn"
	"github.com/xenolf/lego/providers/dns/gandi"
	"github.com/xenolf/lego/providers/dns/googlecloud"
	"github.com/xenolf/lego/providers/dns/namecheap"
	"github.com/xenolf/lego/providers/dns/rfc2136"
	"github.com/xenolf/lego/providers/dns/route53"
	"github.com/xenolf/lego/providers/dns/vultr"
	"github.com/xenolf/lego/providers/http/webroot"
)

var Logger *log.Logger

func logger() *log.Logger {
	if Logger == nil {
		Logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	return Logger
}

func CheckFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0700)
	}
	return nil
}

func SaveCertRes(certRes acme.CertificateResource, conf *Configuration) {
	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	certOut := path.Join(conf.CertPath(), certRes.Domain+".crt")
	privOut := path.Join(conf.CertPath(), certRes.Domain+".key")
	metaOut := path.Join(conf.CertPath(), certRes.Domain+".json")

	err := ioutil.WriteFile(certOut, certRes.Certificate, 0600)
	if err != nil {
		logger().Fatalf("Unable to save Certificate for domain %s\n\t%s", certRes.Domain, err.Error())
	}

	err = ioutil.WriteFile(privOut, certRes.PrivateKey, 0600)
	if err != nil {
		logger().Fatalf("Unable to save PrivateKey for domain %s\n\t%s", certRes.Domain, err.Error())
	}

	jsonBytes, err := json.MarshalIndent(certRes, "", "\t")
	if err != nil {
		logger().Fatalf("Unable to marshal CertResource for domain %s\n\t%s", certRes.Domain, err.Error())
	}

	err = ioutil.WriteFile(metaOut, jsonBytes, 0600)
	if err != nil {
		logger().Fatalf("Unable to save CertResource for domain %s\n\t%s", certRes.Domain, err.Error())
	}
}

func Setup(c *cobra.Command) (*Configuration, *Account, *acme.Client) {
	pathS, err := c.PersistentFlags().GetString("path")
	if err != nil {
		logger().Fatalf(err.Error())
	}
	err = CheckFolder(pathS)
	if err != nil {
		logger().Fatalf("Could not check/create path: %s", err.Error())
	}

	conf := NewConfiguration(c)

	email, err := c.PersistentFlags().GetString("email")
	if err != nil {
		logger().Fatalln(err.Error())
	}

	if len(email) == 0 {
		logger().Fatal("You have to pass an account (email address) to the program using --email or -m")
	}

	//TODO: move to account struct? Currently MUST pass email.
	acc := NewAccount(email, conf)

	keyType, err := conf.KeyType()
	if err != nil {
		logger().Fatal(err.Error())
	}

	server, err := c.PersistentFlags().GetString("server")
	if err != nil {
		logger().Fatal(err.Error())
	}
	client, err := acme.NewClient(server, acc, keyType)
	if err != nil {
		logger().Fatalf("Could not create client: %s", err.Error())
	}

	excludeS, err := c.PersistentFlags().GetStringSlice("exclude")
	if err != nil {
		logger().Fatal(err.Error())
	}
	if len(excludeS) > 0 {
		client.ExcludeChallenges(conf.ExcludedSolvers())
	}

	webrootS, err := c.PersistentFlags().GetString("webroot")
	if err != nil {
		logger().Fatal(err.Error())
	}

	if len(webrootS) > 0 {
		provider, err := webroot.NewHTTPProvider(webrootS)
		if err != nil {
			logger().Fatal(err)
		}

		client.SetChallengeProvider(acme.HTTP01, provider)

		// --webroot=foo indicates that the user specifically want to do a HTTP challenge
		// infer that the user also wants to exclude all other challenges
		client.ExcludeChallenges([]acme.Challenge{acme.DNS01, acme.TLSSNI01})
	}

	httpS, err := c.PersistentFlags().GetString("http")
	if err != nil {
		logger().Fatal(err.Error())
	}
	if len(httpS) > 0 {
		if strings.Index(httpS, ":") == -1 {
			logger().Fatalf("The --http switch only accepts interface:port or :port for its argument.")
		}
		client.SetHTTPAddress(httpS)
	}

	tls, err := c.PersistentFlags().GetString("tls")
	if err != nil {
		logger().Fatal(err.Error())
	}

	if len(tls) > 0 {
		if strings.Index(tls, ":") == -1 {
			logger().Fatalf("The --tls switch only accepts interface:port or :port for its argument.")
		}
		client.SetTLSAddress(tls)
	}

	dns, err := c.PersistentFlags().GetString("dns")
	if err != nil {
		logger().Fatal(err.Error())
	}

	if len(dns) > 0 {
		var err error
		var provider acme.ChallengeProvider
		switch dns {
		case "cloudflare":
			provider, err = cloudflare.NewDNSProvider()
		case "digitalocean":
			provider, err = digitalocean.NewDNSProvider()
		case "dnsimple":
			provider, err = dnsimple.NewDNSProvider()
		case "dyn":
			provider, err = dyn.NewDNSProvider()
		case "gandi":
			provider, err = gandi.NewDNSProvider()
		case "gcloud":
			provider, err = googlecloud.NewDNSProvider()
		case "manual":
			provider, err = acme.NewDNSProviderManual()
		case "namecheap":
			provider, err = namecheap.NewDNSProvider()
		case "route53":
			provider, err = route53.NewDNSProvider()
		case "rfc2136":
			provider, err = rfc2136.NewDNSProvider()
		case "vultr":
			provider, err = vultr.NewDNSProvider()
		}

		if err != nil {
			logger().Fatal(err)
		}

		client.SetChallengeProvider(acme.DNS01, provider)

		// --dns=foo indicates that the user specifically want to do a DNS challenge
		// infer that the user also wants to exclude all other challenges
		client.ExcludeChallenges([]acme.Challenge{acme.HTTP01, acme.TLSSNI01})
	}

	return conf, acc, client
}

func HandleTOS(c *cobra.Command, client *acme.Client, acc *Account) {
	// Check for a global accept override
	accepttos, err := c.PersistentFlags().GetBool("accept-tos")
	if err != nil {
		logger().Fatalf(err.Error())
	}

	if accepttos {
		err := client.AgreeToTOS()
		if err != nil {
			logger().Fatalf("Could not agree to TOS: %s", err.Error())
		}

		acc.Save()
		return
	}

	reader := bufio.NewReader(os.Stdin)
	logger().Printf("Please review the TOS at %s", acc.Registration.TosURL)

	for {
		logger().Println("Do you accept the TOS? Y/n")
		text, err := reader.ReadString('\n')
		if err != nil {
			logger().Fatalf("Could not read from console: %s", err.Error())
		}

		text = strings.Trim(text, "\r\n")

		if text == "n" {
			logger().Fatal("You did not accept the TOS. Unable to proceed.")
		}

		if text == "Y" || text == "y" || text == "" {
			err = client.AgreeToTOS()
			if err != nil {
				logger().Fatalf("Could not agree to TOS: %s", err.Error())
			}
			acc.Save()
			break
		}

		logger().Println("Your input was invalid. Please answer with one of Y/y, n or by pressing enter.")
	}
}
