# lego

Let's Encrypt client and ACME library written in Go

[![GoDoc](https://godoc.org/github.com/xenolf/lego?status.svg)](https://godoc.org/github.com/xenolf/lego/acme)
[![Build Status](https://travis-ci.org/xenolf/lego.svg?branch=master)](https://travis-ci.org/xenolf/lego)
[![Docker Pulls](https://img.shields.io/docker/pulls/xenolf/lego.svg)](https://hub.docker.com/r/xenolf/lego/)
[![Dev Chat](https://img.shields.io/badge/dev%20chat-gitter-blue.svg?label=dev+chat)](https://gitter.im/xenolf/lego)
[![Beerpay](https://beerpay.io/xenolf/lego/badge.svg)](https://beerpay.io/xenolf/lego)

## Installation

### Binaries

To get the binary just download the latest release for your OS/Arch from [the release page](https://github.com/xenolf/lego/releases) and put the binary somewhere convenient.
lego does not assume anything about the location you run it from.

### From Docker

```bash
docker run xenolf/lego -h
```

### From package managers

- [ArchLinux (AUR)](https://aur.archlinux.org/packages/lego):

```bash
yay -S lego
```

**Note**: only the package manager for Arch Linux is officially supported by the lego team.

### From sources

To install from sources, just run:

```bash
go get -u github.com/xenolf/lego/cmd/lego
```

## Features

- Register with CA
- Obtain certificates, both from scratch or with an existing CSR
- Renew certificates
- Revoke certificates
- Robust implementation of all ACME challenges
  - HTTP (http-01)
  - DNS (dns-01)
  - TLS (tls-alpn-01)
- SAN certificate support
- Comes with multiple optional [DNS providers](https://github.com/xenolf/lego/tree/master/providers/dns)
- [Custom challenge solvers](https://github.com/xenolf/lego/wiki/Writing-a-Challenge-Solver)
- Certificate bundling
- OCSP helper function

Please keep in mind that CLI switches and APIs are still subject to change.

When using the standard `--path` option, all certificates and account configurations are saved to a folder `.lego` in the current working directory.

## Usage

```text
NAME:
   lego - Let's Encrypt client written in Go

USAGE:
   lego [global options] command [command options] [arguments...]

COMMANDS:
     run      Register an account, then create and install a certificate
     revoke   Revoke a certificate
     renew    Renew a certificate
     dnshelp  Shows additional help for the --dns global option
     list     Display certificates and accounts information.
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --domains value, -d value    Add a domain to the process. Can be specified multiple times.
   --server value, -s value     CA hostname (and optionally :port). The server certificate must be trusted in order to avoid further modifications to the client. (default: "https://acme-v02.api.letsencrypt.org/directory")
   --accept-tos, -a             By setting this flag to true you indicate that you accept the current Let's Encrypt terms of service.
   --email value, -m value      Email used for registration and recovery contact.
   --csr value, -c value        Certificate signing request filename, if an external CSR is to be used.
   --eab                        Use External Account Binding for account registration. Requires --kid and --hmac.
   --kid value                  Key identifier from External CA. Used for External Account Binding.
   --hmac value                 MAC key from External CA. Should be in Base64 URL Encoding without padding format. Used for External Account Binding.
   --key-type value, -k value   Key type to use for private keys. Supported: rsa2048, rsa4096, rsa8192, ec256, ec384. (default: "rsa2048")
   --filename value             (deprecated) Filename of the generated certificate.
   --path value                 Directory to use for storing the data. (default: "./.lego")
   --http                       Use the HTTP challenge to solve challenges. Can be mixed with other types of challenges.
   --http.port value            Set the port and interface to use for HTTP based challenges to listen on.Supported: interface:port or :port. (default: ":80")
   --http.webroot value         Set the webroot folder to use for HTTP based challenges to write directly in a file in .well-known/acme-challenge.
   --http.memcached-host value  Set the memcached host(s) to use for HTTP based challenges. Challenges will be written to all specified hosts.
   --tls                        Use the TLS challenge to solve challenges. Can be mixed with other types of challenges.
   --tls.port value             Set the port and interface to use for TLS based challenges to listen on. Supported: interface:port or :port. (default: ":443")
   --dns value                  Solve a DNS challenge using the specified provider. Can be mixed with other types of challenges. Run 'lego dnshelp' for help on usage.
   --dns.disable-cp             By setting this flag to true, disables the need to wait the propagation of the TXT record to all authoritative name servers.
   --dns.resolvers value        Set the resolvers to use for performing recursive DNS queries. Supported: host:port. The default is to use the system resolvers, or Google's DNS resolvers if the system's cannot be determined.
   --http-timeout value         Set the HTTP timeout value to a specific value in seconds. (default: 0)
   --dns-timeout value          Set the DNS timeout value to a specific value in seconds. Used only when performing authoritative name servers queries. (default: 10)
   --pem                        Generate a .pem file by concatenating the .key and .crt files together.
   --cert.timeout value         Set the certificate timeout value to a specific value in seconds. Only used when obtaining certificates. (default: 30)
   --help, -h                   show help
   --version, -v                print the version
```

### Sudo

The CLI does not require root permissions but needs to bind to port 80 and 443 for certain challenges.
To run the CLI without sudo, you have four options:

- Use setcap 'cap_net_bind_service=+ep' /path/to/program
- Pass the `--http.port` or/and the `--tls.port` option and specify a custom port to bind to. In this case you have to forward port 80/443 to these custom ports (see [Port Usage](#port-usage)).
- Pass the `--http.webroot` option and specify the path to your webroot folder. In this case the challenge will be written in a file in `.well-known/acme-challenge/` inside your webroot.
- Pass the `--dns` option and specify a DNS provider.

### Port Usage

By default lego assumes it is able to bind to ports 80 and 443 to solve challenges.
If this is not possible in your environment, you can use the `--http.port` and `--tls.port` options to instruct
lego to listen on that interface:port for any incoming challenges.

If you are using this option, make sure you proxy all of the following traffic to these ports.

HTTP Port:

- All plaintext HTTP requests to port 80 which begin with a request path of `/.well-known/acme-challenge/` for the HTTP challenge.

TLS Port:

- All TLS handshakes on port 443 for the TLS-ALPN challenge.

This traffic redirection is only needed as long as lego solves challenges. As soon as you have received your certificates you can deactivate the forwarding.

### CLI Example

Assumes the `lego` binary has permission to bind to ports 80 and 443.
You can get a pre-built binary from the [releases](https://github.com/xenolf/lego/releases) page.
If your environment does not allow you to bind to these ports, please read [Port Usage](#port-usage).

Obtain a certificate:

```bash
lego --email="foo@bar.com" --domains="example.com" --http run
```

(Find your certificate in the `.lego` folder of current working directory.)

To renew the certificate:

```bash
lego --email="foo@bar.com" --domains="example.com" --http renew
```

To renew the certificate only if it expires within 30 days

```bash
lego --email="foo@bar.com" --domains="example.com" --http renew --days 30
```

Obtain a certificate using the DNS challenge and AWS Route 53:

```bash
AWS_REGION=us-east-1 AWS_ACCESS_KEY_ID=my_id AWS_SECRET_ACCESS_KEY=my_key lego --email="foo@bar.com" --domains="example.com" --dns="route53" run
```

Obtain a certificate given a certificate signing request (CSR) generated by something else:

```bash
lego --email="foo@bar.com" --http --csr=/path/to/csr.pem run
```

(lego will infer the domains to be validated based on the contents of the CSR, so make sure the CSR's Common Name and optional SubjectAltNames are set correctly.)

lego defaults to communicating with the production Let's Encrypt ACME server.
If you'd like to test something without issuing real certificates, consider using the staging endpoint instead:

```bash
lego --server=https://acme-staging-v02.api.letsencrypt.org/directory …
```

## ACME Library Usage

A valid, but bare-bones example use of the acme package:

```go
package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"

	"github.com/xenolf/lego/certcrypto"
	"github.com/xenolf/lego/certificate"
	"github.com/xenolf/lego/challenge/http01"
	"github.com/xenolf/lego/challenge/tlsalpn01"
	"github.com/xenolf/lego/lego"
	"github.com/xenolf/lego/registration"
)

// You'll need a user or account type that implements acme.User
type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func main() {

	// Create a user. New accounts need an email and private key to start.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	myUser := MyUser{
		Email: "you@yours.com",
		key:   privateKey,
	}

	config := lego.NewConfig(&myUser)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	config.CADirURL = "http://192.168.99.100:4000/directory"
	config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	// We specify an http port of 5002 and an tls port of 5001 on all interfaces
	// because we aren't running as root and can't bind a listener to port 80 and 443
	// (used later when we attempt to pass challenges). Keep in mind that you still
	// need to proxy challenge traffic to port 5002 and 5001.
	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "5002"))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Challenge.SetTLSALPN01Provider(tlsalpn01.NewProviderServer("", "5001"))
	if err != nil {
		log.Fatal(err)
	}

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
	}
	myUser.Registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{"mydomain.com"},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
	}

	// Each certificate comes back with the cert bytes, the bytes of the client's
	// private key, and a certificate URL. SAVE THESE TO DISK.
	fmt.Printf("%#v\n", certificates)

	// ... all done.
}
```

## DNS Challenge API Details

### AWS Route 53

The following AWS IAM policy document describes the permissions required for lego to complete the DNS challenge.

```json
{
   "Version": "2012-10-17",
   "Statement": [
       {
           "Sid": "",
           "Effect": "Allow",
           "Action": [
               "route53:GetChange",
               "route53:ChangeResourceRecordSets",
               "route53:ListResourceRecordSets"
           ],
           "Resource": [
               "arn:aws:route53:::hostedzone/*",
               "arn:aws:route53:::change/*"
           ]
       },
       {
           "Sid": "",
           "Effect": "Allow",
           "Action": "route53:ListHostedZonesByName",
           "Resource": "*"
       }
   ]
}
```

## ACME v1

lego introduced support for ACME v2 in [v1.0.0](https://github.com/xenolf/lego/releases/tag/v1.0.0), if you still need to utilize ACME v1, you can do so by using the [v0.5.0](https://github.com/xenolf/lego/releases/tag/v0.5.0) version.
