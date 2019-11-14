---
title: "CLI"
date: 2019-03-03T16:39:46+01:00
draft: false
---

Lego can be use as a CLI.

<!--more-->

## Usage

{{%expand "CLI help" %}}
```slim
NAME:
   lego - Let's Encrypt client written in Go

USAGE:
   lego [global options] command [command options] [arguments...]

COMMANDS:
   run      Register an account, then create and install a certificate
   revoke   Revoke a certificate
   renew    Renew a certificate
   dnshelp  Shows additional help for the '--dns' global option
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
   --key-type value, -k value   Key type to use for private keys. Supported: rsa2048, rsa4096, rsa8192, ec256, ec384. (default: "ec384")
   --filename value             (deprecated) Filename of the generated certificate.
   --path value                 Directory to use for storing the data. (default: "./.lego")
   --http                       Use the HTTP challenge to solve challenges. Can be mixed with other types of challenges.
   --http.port value            Set the port and interface to use for HTTP based challenges to listen on.Supported: interface:port or :port. (default: ":80")
   --http.proxy-header value    Validate against this HTTP header when solving HTTP based challenges behind a reverse proxy. (default: "Host")
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
{{% /expand%}}


When using the standard `--path` option, all certificates and account configurations are saved to a folder `.lego` in the current working directory.


## Let's Encrypt ACME server

lego defaults to communicating with the production Let's Encrypt ACME server.
If you'd like to test something without issuing real certificates, consider using the staging endpoint instead:

```bash
lego --server=https://acme-staging-v02.api.letsencrypt.org/directory …
```

## Sudo

The CLI does not require root permissions but needs to bind to port 80 and 443 for certain challenges.
To run the CLI without sudo, you have four options:

- Use setcap 'cap_net_bind_service=+ep' /path/to/program
- Pass the `--http.port` or/and the `--tls.port` option and specify a custom port to bind to. In this case you have to forward port 80/443 to these custom ports (see [Port Usage](usage/cli#port-usage)).
- Pass the `--http.webroot` option and specify the path to your webroot folder. In this case the challenge will be written in a file in `.well-known/acme-challenge/` inside your webroot.
- Pass the `--dns` option and specify a DNS provider.

## Port Usage

By default lego assumes it is able to bind to ports 80 and 443 to solve challenges.
If this is not possible in your environment, you can use the `--http.port` and `--tls.port` options to instruct
lego to listen on that interface:port for any incoming challenges.

If you are using this option, make sure you proxy all of the following traffic to these ports.

**HTTP Port:** All plaintext HTTP requests to port **80** which begin with a request path of `/.well-known/acme-challenge/` for the HTTP challenge.[^header]

**TLS Port:** All TLS handshakes on port **443** for the TLS-ALPN challenge.

This traffic redirection is only needed as long as lego solves challenges. As soon as you have received your certificates you can deactivate the forwarding.

[^header]: You must ensure that incoming validation requests containt the correct value for the HTTP `Host` header. If you operate lego behind a non-transparent reverse proxy (such as Apache or NGINX), you might need to alter the header field using `--http.proxy-header X-Forwarded-Host`.
