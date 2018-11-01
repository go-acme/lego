package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli"
	"github.com/xenolf/lego/emca"
	"github.com/xenolf/lego/emca/certificate"
	"github.com/xenolf/lego/emca/challenge"
	"github.com/xenolf/lego/emca/challenge/dns01"
	"github.com/xenolf/lego/log"
	"github.com/xenolf/lego/providers/dns"
	"github.com/xenolf/lego/providers/http/memcached"
	"github.com/xenolf/lego/providers/http/webroot"
)

func checkFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0700)
	}
	return nil
}

func setup(c *cli.Context) (*Configuration, *Account, *emca.Client) {
	if c.GlobalIsSet("dns-timeout") {
		// FIXME move to Config?
		dns01.DNSTimeout = time.Duration(c.GlobalInt("dns-timeout")) * time.Second
	}

	if len(c.GlobalStringSlice("dns-resolvers")) > 0 {
		var resolvers []string
		for _, resolver := range c.GlobalStringSlice("dns-resolvers") {
			if !strings.Contains(resolver, ":") {
				resolver += ":53"
			}
			resolvers = append(resolvers, resolver)
		}
		// FIXME move to Config?
		dns01.RecursiveNameservers = resolvers
	}

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

	config := emca.NewDefaultConfig(acc).
		WithKeyType(keyType).
		WithCADirURL(c.GlobalString("server")).
		WithUserAgent(fmt.Sprintf("lego-cli/%s", c.App.Version))

	if c.GlobalIsSet("http-timeout") {
		config.HTTPClient.Timeout = time.Duration(c.GlobalInt("http-timeout")) * time.Second
	}

	client, err := emca.NewClient(config)
	if err != nil {
		log.Fatalf("Could not create client: %v", err)
	}

	if len(c.GlobalStringSlice("exclude")) > 0 {
		client.Challenge.Exclude(conf.ExcludedSolvers())
	}

	if c.GlobalIsSet("webroot") {
		provider, errO := webroot.NewHTTPProvider(c.GlobalString("webroot"))
		if errO != nil {
			log.Fatal(errO)
		}

		errO = client.Challenge.SetProvider(challenge.HTTP01, provider)
		if errO != nil {
			log.Fatal(errO)
		}

		// --webroot=foo indicates that the user specifically want to do a HTTP challenge
		// infer that the user also wants to exclude all other challenges
		client.Challenge.Exclude([]challenge.Type{challenge.DNS01, challenge.TLSALPN01})
	}

	if c.GlobalIsSet("memcached-host") {
		provider, errO := memcached.NewMemcachedProvider(c.GlobalStringSlice("memcached-host"))
		if errO != nil {
			log.Fatal(errO)
		}

		errO = client.Challenge.SetProvider(challenge.HTTP01, provider)
		if errO != nil {
			log.Fatal(errO)
		}

		// --memcached-host=foo:11211 indicates that the user specifically want to do a HTTP challenge
		// infer that the user also wants to exclude all other challenges
		client.Challenge.Exclude([]challenge.Type{challenge.DNS01, challenge.TLSALPN01})
	}

	if c.GlobalIsSet("http") {
		if !strings.Contains(c.GlobalString("http"), ":") {
			log.Fatalf("The --http switch only accepts interface:port or :port for its argument.")
		}

		err = client.Challenge.SetHTTPAddress(c.GlobalString("http"))
		if err != nil {
			log.Fatal(err)
		}
	}

	if c.GlobalIsSet("tls") {
		if !strings.Contains(c.GlobalString("tls"), ":") {
			log.Fatalf("The --tls switch only accepts interface:port or :port for its argument.")
		}

		err = client.Challenge.SetTLSAddress(c.GlobalString("tls"))
		if err != nil {
			log.Fatal(err)
		}
	}

	if c.GlobalIsSet("dns") {
		provider, errO := dns.NewDNSChallengeProviderByName(c.GlobalString("dns"))
		if errO != nil {
			log.Fatal(errO)
		}

		errO = client.Challenge.SetProvider(challenge.DNS01, provider)
		if errO != nil {
			log.Fatal(errO)
		}

		// --dns=foo indicates that the user specifically want to do a DNS challenge
		// infer that the user also wants to exclude all other challenges
		client.Challenge.Exclude([]challenge.Type{challenge.HTTP01, challenge.TLSALPN01})
	}

	if client.GetExternalAccountRequired() && !c.GlobalIsSet("eab") {
		log.Fatal("Server requires External Account Binding. Use --eab with --kid and --hmac.")
	}

	return conf, acc, client
}

func saveCertRes(certRes *certificate.Resource, conf *Configuration) {
	var domainName string

	// Check filename cli parameter
	if conf.context.GlobalString("filename") == "" {
		// Make sure no funny chars are in the cert names (like wildcards ;))
		domainName = strings.Replace(certRes.Domain, "*", "_", -1)
	} else {
		domainName = conf.context.GlobalString("filename")
	}

	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	certOut := filepath.Join(conf.CertPath(), domainName+".crt")

	err := checkFolder(filepath.Dir(certOut))
	if err != nil {
		log.Fatalf("Could not check/create path: %v", err)
	}

	err = ioutil.WriteFile(certOut, certRes.Certificate, 0600)
	if err != nil {
		log.Fatalf("Unable to save Certificate for domain %s\n\t%v", certRes.Domain, err)
	}

	issuerOut := filepath.Join(conf.CertPath(), domainName+".issuer.crt")

	if certRes.IssuerCertificate != nil {
		err = ioutil.WriteFile(issuerOut, certRes.IssuerCertificate, 0600)
		if err != nil {
			log.Fatalf("Unable to save IssuerCertificate for domain %s\n\t%v", certRes.Domain, err)
		}
	}

	if certRes.PrivateKey != nil {
		privOut := filepath.Join(conf.CertPath(), domainName+".key")

		// if we were given a CSR, we don't know the private key
		err = ioutil.WriteFile(privOut, certRes.PrivateKey, 0600)
		if err != nil {
			log.Fatalf("Unable to save PrivateKey for domain %s\n\t%v", certRes.Domain, err)
		}

		if conf.context.GlobalBool("pem") {
			pemOut := filepath.Join(conf.CertPath(), domainName+".pem")
			err = ioutil.WriteFile(pemOut, bytes.Join([][]byte{certRes.Certificate, certRes.PrivateKey}, nil), 0600)
			if err != nil {
				log.Fatalf("Unable to save Certificate and PrivateKey in .pem for domain %s\n\t%v", certRes.Domain, err)
			}
		}

	} else if conf.context.GlobalBool("pem") {
		// we don't have the private key; can't write the .pem file
		log.Fatalf("Unable to save pem without private key for domain %s\n\t%v; are you using a CSR?", certRes.Domain, err)
	}

	jsonBytes, err := json.MarshalIndent(certRes, "", "\t")
	if err != nil {
		log.Fatalf("Unable to marshal CertResource for domain %s\n\t%v", certRes.Domain, err)
	}

	metaOut := filepath.Join(conf.CertPath(), domainName+".json")
	err = ioutil.WriteFile(metaOut, jsonBytes, 0600)
	if err != nil {
		log.Fatalf("Unable to save CertResource for domain %s\n\t%v", certRes.Domain, err)
	}
}
