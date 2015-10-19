# lego
Let's Encrypt client and ACME library written in Go

[![GoDoc](https://godoc.org/github.com/xenolf/lego/acme?status.svg)](https://godoc.org/github.com/xenolf/lego/acme)
[![Build Status](https://travis-ci.org/xenolf/lego.svg?branch=master)](https://travis-ci.org/xenolf/lego)

This is a work in progress. Please do *NOT* run this on a production server.

#### Current Status
The code in this repository is under development.

Current features:
- [x] Registering with a CA
- [x] Requesting Certificates
- [x] Renewing Certificates
- [x] Revoking Certificates
- [ ] Initiating account recovery
- Identifier validation challenges
  - [x] SimpleHTTP Challenge
  - [ ] DVSNI Challenge
  - [ ] Proof of Possession of a Prior Key
  - [ ] DNS Challenge

Please keep in mind that CLI switches and APIs are still subject to change.

When using the standard --path option, all certificates and account configurations are saved to a folder *.lego* in the current working directory.

#### Sudo
I tried to not need sudo apart from challenges where binding to a privileged port is necessary.
To run the CLI without sudo, you have two options:
- Use ```setcap 'cap_net_bind_service=+ep' /path/to/program```
- Pass the --port option and specify a custom port to bind to. In this case you have to forward port 443 to this custom port.

#### Usage

```
NAME:
   lego - Let's encrypt client to go!

USAGE:
   ./lego [global options] command [command options] [arguments...]
   
VERSION:
   0.0.2
   
COMMANDS:
   run		Register an account, then create and install a certificate
   auth		Create a certificate - must already have an account
   revoke	Revoke a certificate
   renew	Renew a certificate
   help, h	Shows a list of commands or help for one command
   
GLOBAL OPTIONS:
   --domains, -d [--domains option --domains option]			Add domains to the process
   --server, -s "https://acme-staging.api.letsencrypt.org/"		CA hostname (and optionally :port). The server certificate must be trusted in order to avoid further modifications to the client.
   --email, -m 								Email used for registration and recovery contact.
   --rsa-key-size, -B "2048"						Size of the RSA key.
   --path "/home/azhwkd/Projects/go/src/github.com/xenolf/lego/.lego"	Directory to use for storing the data
   --port 								Challenges will use this port to listen on. Please make sure to forward port 443 to this port on your machine. Otherwise use setcap on the binary
   --devMode								If set to true, all client side challenge pre-tests are skipped.
   --help, -h								show help
   --version, -v							print the version
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
// bind a listener to port 443 (used later when we attempt to pass challenge).
client := acme.NewClient("http://192.168.99.100:4000", &myUser, rsaKeySize, "5001")

// New users will need to register; be sure to save it
reg, err := client.Register()
if err != nil {
	log.Fatal(err)
}
myUser.Registration = reg

// The client has a URL to the current Let's Encrypt Subscriber
// Agreement. The user will need to agree to it.
err = client.AgreeToTos()
if err != nil {
	log.Fatal(err)
}

// The acme library takes care of completing the challenges to obtain the certificate(s).
// Of course, the hostnames must resolve to this machine or it will fail.
certificates, err := client.ObtainCertificates([]string{"mydomain.com"})
if err != nil {
	log.Fatal(err)
}

// Each certificate comes back with the cert bytes, the bytes of the client's
// private key, and a certificate URL. This is where you should save them to files!
fmt.Printf("%#v\n", certificates)

// ... all done.
```
