# lego
Let's Encrypt client and ACME library written in Go

[![GoDoc](https://godoc.org/github.com/xenolf/lego/acme?status.svg)](https://godoc.org/github.com/xenolf/lego/acme)
[![Build Status](https://travis-ci.org/xenolf/lego.svg?branch=master)](https://travis-ci.org/xenolf/lego)

#### General
This is a work in progress. Please do *NOT* run this on a production server and please report any bugs you find!

#### Installation
lego supports both binary installs and install from source.

To get the binary just download the latest release for your OS/Arch from [the release page](https://github.com/xenolf/lego/releases)
and put the binary somewhere convenient. lego does not assume anything about the location you run it from.

To install from source, just run 
```
go get -u github.com/xenolf/lego
```

#### Current Status
The code in this repository is under development.

Current features:
- [x] Registering with a CA
- [x] Requesting Certificates
- [x] Renewing Certificates
- [x] Revoking Certificates
- [ ] Initiating account recovery
- Identifier validation challenges
  - [x] HTTP (http-01)
  - [x] TLS with Server Name Indication (tls-sni-01)
  - [ ] Proof of Possession of a Prior Key (proofOfPossession-01)
  - [ ] DNS (dns-01) - Implemented in branch, blocked by upstream.
- [x] Certificate bundling
- [x] Library support for OCSP

Please keep in mind that CLI switches and APIs are still subject to change.

When using the standard --path option, all certificates and account configurations are saved to a folder *.lego* in the current working directory.

#### Sudo
The CLI does not require root permissions but needs to bind to port 80 and 443 for certain challenges. 
To run the CLI without sudo, you have two options:

- Use setcap 'cap_net_bind_service=+ep' /path/to/program
- Pass the `--port` option and specify a custom port to bind to. In this case you have to forward port 443 to this custom port.

#### Port Usage
By default lego assumes it is able to bind to ports 80 and 443 to solve challenges.
If this is not possible in your environment, you can use the `--port` option to instruct
lego to listen on that port for any incoming challenges.

If you are using this option, make sure you proxy all of the following traffic to that port:
- All plaintext HTTP requests to port 80 which begin with a request path of `/.well-known/acme-challenge/` for the HTTP-01 challenge.
- All TLS handshakes on port 443 for TLS-SNI-01.

This traffic redirection is only needed as long as lego solves challenges. As soon as you have received your certificates you can deactivate the forwarding.

#### Usage

```
NAME:
   lego - Let's encrypt client to go!

USAGE:
   ./lego [global options] command [command options] [arguments...]
   
VERSION:
   0.1.0
   
COMMANDS:
   run		Register an account, then create and install a certificate
   revoke	Revoke a certificate
   renew	Renew a certificate
   help, h	Shows a list of commands or help for one command
   
GLOBAL OPTIONS:
   --domains, -d [--domains option --domains option]			Add domains to the process
   --server, -s "https://acme-v01.api.letsencrypt.org/directory"	CA hostname (and optionally :port). The server certificate must be trusted in order to avoid further modifications to the client.
   --email, -m 								Email used for registration and recovery contact.
   --rsa-key-size, -B "2048"						Size of the RSA key.
   --path "${CWD}"							Directory to use for storing the data
   --port 								Challenges will use this port to listen on. Please make sure to forward port 443 to this port on your machine. Otherwise use setcap on the binary
   --help, -h								show help
   --version, -v							print the version

```

##### CLI Example

Assumes the `lego` binary has permission to bind to ports 80 and 443. You can get a pre-built binary from the [releases](https://github.com/xenolf/lego/releases) page.
If your environment does not allow you to bind to these ports, please read [Port Usage](#port-usage).

Obtain a certificate:

```bash
$ lego --email="foo@bar.com" --domains="example.com" run
```

(Find your certificate in the `.lego` folder of current working directory.)

To renew the certificate:

```bash
$ lego --email="foo@bar.com" --domains="example.com" renew
```

#### ACME Library Usage

A valid, but bare-bones example use of the acme package:

```go
// You'll need a user or account type that implements acme.User
type MyUser struct {
	Email        string
	Registration *acme.RegistrationResource
	key          *rsa.PrivateKey
}
func (u MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *acme.RegistrationResource {
	return u.Registration
}
func (u MyUser) GetPrivateKey() *rsa.PrivateKey {
	return u.key
}

// Create a user. New accounts need an email and private key to start.
const rsaKeySize = 2048
privateKey, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
if err != nil {
	log.Fatal(err)
}
myUser := MyUser{
	Email: "you@yours.com",
	key: privateKey,
}

// A client facilitates communication with the CA server. This CA URL is
// configured for a local dev instance of Boulder running in Docker in a VM.
// We specify an optPort of 5001 because we aren't running as root and can't
// bind a listener to port 80 or 443 (used later when we attempt to pass challenges).
// Keep in mind that we still need to proxy challenge traffic to port 5001.
client, err := acme.NewClient("http://192.168.99.100:4000", &myUser, rsaKeySize, "5001")
if err != nil {
  log.Fatal(err)
}

// New users will need to register; be sure to save it
reg, err := client.Register()
if err != nil {
	log.Fatal(err)
}
myUser.Registration = reg

// The client has a URL to the current Let's Encrypt Subscriber
// Agreement. The user will need to agree to it.
err = client.AgreeToTOS()
if err != nil {
	log.Fatal(err)
}

// The acme library takes care of completing the challenges to obtain the certificate(s).
// Of course, the hostnames must resolve to this machine or it will fail.
bundle := false
certificates, err := client.ObtainCertificates([]string{"mydomain.com"}, bundle)
if err != nil {
	log.Fatal(err)
}

// Each certificate comes back with the cert bytes, the bytes of the client's
// private key, and a certificate URL. This is where you should save them to files!
fmt.Printf("%#v\n", certificates)

// ... all done.
```
