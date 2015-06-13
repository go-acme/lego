# lego
Let's Encrypt client and library in go!

[![Build Status](https://travis-ci.org/xenolf/lego.svg?branch=master)](https://travis-ci.org/xenolf/lego)

This is a work in progress. Please do *NOT* run this on a production server. 

####Current Status
The code in this repository is currently quite raw.
You are currently able to register an account with the ACME server as well as request certificates through the CLI.

Please keep in mind that CLI switches and APIs are still subject to change.

When using the standard --path option, all certificates and account configurations are saved to a folder *.lego* in the current working directory.

####Sudo
I tried to not need sudo apart from challenges where binding to a privileged port is necessary.
To run the CLI without sudo, you have two options:
- Use ```setcap 'cap_net_bind_service=+ep' /path/to/program```
- Pass the --port option and specify a custom port to bind to. In this case you have to forward port 443 to this custom port.

#### Usage

```
NAME:
   lego - Let's encrypt client to go!

USAGE:
   lego [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
   run		Create and install a certificate
   auth		Create a certificate
   install	Install a certificate
   revoke	Revoke a certificate
   rollback	Rollback a certificate
   help, h	Shows a list of commands or help for one command
   
GLOBAL OPTIONS:
   --domains, -d [--domains option --domains option]					Add domains to the process
   --server, -s "https://www.letsencrypt-demo.org/acme/new-reg"				CA hostname (and optionally :port). The server certificate must be trusted in order to avoid further modifications to the client.
   --email, -m 										Email used for registration and recovery contact.
   --rsa-key-size, -B "2048"								Size of the RSA key.
   --no-confirm										Turn off confirmation screens.
   --agree-tos, -e									Skip the end user license agreement screen.
   --path "/Volumes/Data/Users/azhwkd/Projects/go/src/github.com/xenolf/lego/.lego"	Directory to use for storing the data
   --port 										Challenges will use this port to listen on. Please make sure to forward port 443 to this port on your machine. Otherwise use setcap on the binary
   --help, -h										show help
   --version, -v									print the version
```
