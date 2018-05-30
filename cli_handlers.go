package main

import (
	"bufio"
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/urfave/cli"
	"github.com/xenolf/lego/acme"
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

func setup(c *cli.Context) (*Configuration, *Account, *acme.Client) {

	if c.GlobalIsSet("http-timeout") {
		acme.HTTPClient = http.Client{Timeout: time.Duration(c.GlobalInt("http-timeout")) * time.Second}
	}

	if c.GlobalIsSet("dns-timeout") {
		acme.DNSTimeout = time.Duration(c.GlobalInt("dns-timeout")) * time.Second
	}

	if len(c.GlobalStringSlice("dns-resolvers")) > 0 {
		resolvers := []string{}
		for _, resolver := range c.GlobalStringSlice("dns-resolvers") {
			if !strings.Contains(resolver, ":") {
				resolver += ":53"
			}
			resolvers = append(resolvers, resolver)
		}
		acme.RecursiveNameservers = resolvers
	}

	err := checkFolder(c.GlobalString("path"))
	if err != nil {
		log.Fatalf("Could not check/create path: %v", err)
	}

	conf := NewConfiguration(c)
	if len(c.GlobalString("email")) == 0 {
		log.Fatal("You have to pass an account (email address) to the program using --email or -m")
	}

	//TODO: move to account struct? Currently MUST pass email.
	acc := NewAccount(c.GlobalString("email"), conf)

	keyType, err := conf.KeyType()
	if err != nil {
		log.Fatal(err)
	}

	acme.UserAgent = fmt.Sprintf("le-go/cli %s", c.App.Version)

	client, err := acme.NewClient(c.GlobalString("server"), acc, keyType)
	if err != nil {
		log.Fatalf("Could not create client: %v", err)
	}

	if len(c.GlobalStringSlice("exclude")) > 0 {
		client.ExcludeChallenges(conf.ExcludedSolvers())
	}

	if c.GlobalIsSet("webroot") {
		provider, err := webroot.NewHTTPProvider(c.GlobalString("webroot"))
		if err != nil {
			log.Fatal(err)
		}

		err = client.SetChallengeProvider(acme.HTTP01, provider)
		if err != nil {
			log.Fatal(err)
		}

		// --webroot=foo indicates that the user specifically want to do a HTTP challenge
		// infer that the user also wants to exclude all other challenges
		client.ExcludeChallenges([]acme.Challenge{acme.DNS01})
	}
	if c.GlobalIsSet("memcached-host") {
		provider, err := memcached.NewMemcachedProvider(c.GlobalStringSlice("memcached-host"))
		if err != nil {
			log.Fatal(err)
		}

		err = client.SetChallengeProvider(acme.HTTP01, provider)
		if err != nil {
			log.Fatal(err)
		}

		// --memcached-host=foo:11211 indicates that the user specifically want to do a HTTP challenge
		// infer that the user also wants to exclude all other challenges
		client.ExcludeChallenges([]acme.Challenge{acme.DNS01})
	}
	if c.GlobalIsSet("http") {
		if !strings.Contains(c.GlobalString("http"), ":") {
			log.Fatalf("The --http switch only accepts interface:port or :port for its argument.")
		}

		err = client.SetHTTPAddress(c.GlobalString("http"))
		if err != nil {
			log.Fatal(err)
		}
	}

	if c.GlobalIsSet("dns") {
		provider, err := dns.NewDNSChallengeProviderByName(c.GlobalString("dns"))
		if err != nil {
			log.Fatal(err)
		}

		err = client.SetChallengeProvider(acme.DNS01, provider)
		if err != nil {
			log.Fatal(err)
		}

		// --dns=foo indicates that the user specifically want to do a DNS challenge
		// infer that the user also wants to exclude all other challenges
		client.ExcludeChallenges([]acme.Challenge{acme.HTTP01})
	}

	if client.GetExternalAccountRequired() && !c.GlobalIsSet("eab") {
		log.Fatal("Server requires External Account Binding. Use --eab with --kid and --hmac.")
	}

	return conf, acc, client
}

func saveCertRes(certRes *acme.CertificateResource, conf *Configuration) {
	// make sure no funny chars are in the cert names (like wildcards ;))
	domainName := strings.Replace(certRes.Domain, "*", "_", -1)

	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	certOut := path.Join(conf.CertPath(), domainName+".crt")
	privOut := path.Join(conf.CertPath(), domainName+".key")
	pemOut := path.Join(conf.CertPath(), domainName+".pem")
	metaOut := path.Join(conf.CertPath(), domainName+".json")
	issuerOut := path.Join(conf.CertPath(), domainName+".issuer.crt")

	err := ioutil.WriteFile(certOut, certRes.Certificate, 0600)
	if err != nil {
		log.Fatalf("Unable to save Certificate for domain %s\n\t%v", certRes.Domain, err)
	}

	if certRes.IssuerCertificate != nil {
		err = ioutil.WriteFile(issuerOut, certRes.IssuerCertificate, 0600)
		if err != nil {
			log.Fatalf("Unable to save IssuerCertificate for domain %s\n\t%v", certRes.Domain, err)
		}
	}

	if certRes.PrivateKey != nil {
		// if we were given a CSR, we don't know the private key
		err = ioutil.WriteFile(privOut, certRes.PrivateKey, 0600)
		if err != nil {
			log.Fatalf("Unable to save PrivateKey for domain %s\n\t%v", certRes.Domain, err)
		}

		if conf.context.GlobalBool("pem") {
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

	err = ioutil.WriteFile(metaOut, jsonBytes, 0600)
	if err != nil {
		log.Fatalf("Unable to save CertResource for domain %s\n\t%v", certRes.Domain, err)
	}
}

func handleTOS(c *cli.Context, client *acme.Client) bool {
	// Check for a global accept override
	if c.GlobalBool("accept-tos") {
		return true
	}

	reader := bufio.NewReader(os.Stdin)
	log.Printf("Please review the TOS at %s", client.GetToSURL())

	for {
		log.Println("Do you accept the TOS? Y/n")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Could not read from console: %v", err)
		}

		text = strings.Trim(text, "\r\n")

		if text == "n" {
			log.Fatal("You did not accept the TOS. Unable to proceed.")
		}

		if text == "Y" || text == "y" || text == "" {
			return true
		}

		log.Println("Your input was invalid. Please answer with one of Y/y, n or by pressing enter.")
	}
}

func readCSRFile(filename string) (*x509.CertificateRequest, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	raw := bytes

	// see if we can find a PEM-encoded CSR
	var p *pem.Block
	rest := bytes
	for {
		// decode a PEM block
		p, rest = pem.Decode(rest)

		// did we fail?
		if p == nil {
			break
		}

		// did we get a CSR?
		if p.Type == "CERTIFICATE REQUEST" {
			raw = p.Bytes
		}
	}

	// no PEM-encoded CSR
	// assume we were given a DER-encoded ASN.1 CSR
	// (if this assumption is wrong, parsing these bytes will fail)
	return x509.ParseCertificateRequest(raw)
}

func run(c *cli.Context) error {
	var err error

	conf, acc, client := setup(c)
	if acc.Registration == nil {
		accepted := handleTOS(c, client)
		if !accepted {
			log.Fatal("You did not accept the TOS. Unable to proceed.")
		}

		var reg *acme.RegistrationResource

		if c.GlobalBool("eab") {
			kid := c.GlobalString("kid")
			hmacEncoded := c.GlobalString("hmac")

			if kid == "" || hmacEncoded == "" {
				log.Fatalf("Requires arguments --kid and --hmac.")
			}

			reg, err = client.RegisterWithExternalAccountBinding(
				accepted,
				kid,
				hmacEncoded,
			)
		} else {
			reg, err = client.Register(accepted)
		}

		if err != nil {
			log.Fatalf("Could not complete registration\n\t%v", err)
		}

		acc.Registration = reg
		acc.Save()

		log.Print("!!!! HEADS UP !!!!")
		log.Printf(`
		Your account credentials have been saved in your Let's Encrypt
		configuration directory at "%s".
		You should make a secure backup	of this folder now. This
		configuration directory will also contain certificates and
		private keys obtained from Let's Encrypt so making regular
		backups of this folder is ideal.`, conf.AccountPath(c.GlobalString("email")))

	}

	// we require either domains or csr, but not both
	hasDomains := len(c.GlobalStringSlice("domains")) > 0
	hasCsr := len(c.GlobalString("csr")) > 0
	if hasDomains && hasCsr {
		log.Fatal("Please specify either --domains/-d or --csr/-c, but not both")
	}
	if !hasDomains && !hasCsr {
		log.Fatal("Please specify --domains/-d (or --csr/-c if you already have a CSR)")
	}

	var cert *acme.CertificateResource

	if hasDomains {
		// obtain a certificate, generating a new private key
		cert, err = client.ObtainCertificate(c.GlobalStringSlice("domains"), !c.Bool("no-bundle"), nil, c.Bool("must-staple"))
	} else {
		// read the CSR
		csr, err := readCSRFile(c.GlobalString("csr"))
		if err == nil {
			// obtain a certificate for this CSR
			cert, err = client.ObtainCertificateForCSR(*csr, !c.Bool("no-bundle"))
		}
	}

	if err != nil {
		log.Printf("Could not obtain certificates\n\t%v", err)

		// Make sure to return a non-zero exit code if ObtainSANCertificate
		// returned at least one error. Due to us not returning partial
		// certificate we can just exit here instead of at the end.
		os.Exit(1)
	}

	if err = checkFolder(conf.CertPath()); err != nil {
		log.Fatalf("Could not check/create path: %v", err)
	}

	saveCertRes(cert, conf)

	return nil
}

func revoke(c *cli.Context) error {
	conf, acc, client := setup(c)
	if acc.Registration == nil {
		log.Fatalf("Account %s is not registered. Use 'run' to register a new account.\n", acc.Email)
	}

	if err := checkFolder(conf.CertPath()); err != nil {
		log.Fatalf("Could not check/create path: %v", err)
	}

	for _, domain := range c.GlobalStringSlice("domains") {
		log.Printf("Trying to revoke certificate for domain %s", domain)

		certPath := path.Join(conf.CertPath(), domain+".crt")
		certBytes, err := ioutil.ReadFile(certPath)
		if err != nil {
			log.Println(err)
		}

		err = client.RevokeCertificate(certBytes)
		if err != nil {
			log.Fatalf("Error while revoking the certificate for domain %s\n\t%v", domain, err)
		} else {
			log.Println("Certificate was revoked.")
		}
	}

	return nil
}

func renew(c *cli.Context) error {
	conf, acc, client := setup(c)
	if acc.Registration == nil {
		log.Fatalf("Account %s is not registered. Use 'run' to register a new account.\n", acc.Email)
	}

	if len(c.GlobalStringSlice("domains")) <= 0 {
		log.Fatal("Please specify at least one domain.")
	}

	domain := c.GlobalStringSlice("domains")[0]
	domain = strings.Replace(domain, "*", "_", -1)

	// load the cert resource from files.
	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	certPath := path.Join(conf.CertPath(), domain+".crt")
	privPath := path.Join(conf.CertPath(), domain+".key")
	metaPath := path.Join(conf.CertPath(), domain+".json")

	certBytes, err := ioutil.ReadFile(certPath)
	if err != nil {
		log.Fatalf("Error while loading the certificate for domain %s\n\t%v", domain, err)
	}

	if c.IsSet("days") {
		expTime, err := acme.GetPEMCertExpiration(certBytes)
		if err != nil {
			log.Printf("Could not get Certification expiration for domain %s", domain)
		}

		if int(expTime.Sub(time.Now()).Hours()/24.0) > c.Int("days") {
			return nil
		}
	}

	metaBytes, err := ioutil.ReadFile(metaPath)
	if err != nil {
		log.Fatalf("Error while loading the meta data for domain %s\n\t%v", domain, err)
	}

	var certRes acme.CertificateResource
	if err := json.Unmarshal(metaBytes, &certRes); err != nil {
		log.Fatalf("Error while marshalling the meta data for domain %s\n\t%v", domain, err)
	}

	if c.Bool("reuse-key") {
		keyBytes, err := ioutil.ReadFile(privPath)
		if err != nil {
			log.Fatalf("Error while loading the private key for domain %s\n\t%v", domain, err)
		}
		certRes.PrivateKey = keyBytes
	}

	certRes.Certificate = certBytes

	newCert, err := client.RenewCertificate(certRes, !c.Bool("no-bundle"), c.Bool("must-staple"))
	if err != nil {
		log.Fatal(err)
	}

	saveCertRes(newCert, conf)

	return nil
}
