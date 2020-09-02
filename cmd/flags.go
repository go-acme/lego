package cmd

import (
	"github.com/go-acme/lego/v4/lego"
	"github.com/urfave/cli"
)

func CreateFlags(defaultPath string) []cli.Flag {
	return []cli.Flag{
		cli.StringSliceFlag{
			Name:  "domains, d",
			Usage: "Add a domain to the process. Can be specified multiple times.",
		},
		cli.StringFlag{
			Name:  "server, s",
			Usage: "CA hostname (and optionally :port). The server certificate must be trusted in order to avoid further modifications to the client.",
			Value: lego.LEDirectoryProduction,
		},
		cli.BoolFlag{
			Name:  "accept-tos, a",
			Usage: "By setting this flag to true you indicate that you accept the current Let's Encrypt terms of service.",
		},
		cli.StringFlag{
			Name:  "email, m",
			Usage: "Email used for registration and recovery contact.",
		},
		cli.StringFlag{
			Name:  "csr, c",
			Usage: "Certificate signing request filename, if an external CSR is to be used.",
		},
		cli.BoolFlag{
			Name:  "eab",
			Usage: "Use External Account Binding for account registration. Requires --kid and --hmac.",
		},
		cli.StringFlag{
			Name:  "kid",
			Usage: "Key identifier from External CA. Used for External Account Binding.",
		},
		cli.StringFlag{
			Name:  "hmac",
			Usage: "MAC key from External CA. Should be in Base64 URL Encoding without padding format. Used for External Account Binding.",
		},
		cli.StringFlag{
			Name:  "key-type, k",
			Value: "ec256",
			Usage: "Key type to use for private keys. Supported: rsa2048, rsa4096, rsa8192, ec256, ec384.",
		},
		cli.StringFlag{
			Name:  "filename",
			Usage: "(deprecated) Filename of the generated certificate.",
		},
		cli.StringFlag{
			Name:   "path",
			EnvVar: "LEGO_PATH",
			Usage:  "Directory to use for storing the data.",
			Value:  defaultPath,
		},
		cli.BoolFlag{
			Name:  "http",
			Usage: "Use the HTTP challenge to solve challenges. Can be mixed with other types of challenges.",
		},
		cli.StringFlag{
			Name:  "http.port",
			Usage: "Set the port and interface to use for HTTP based challenges to listen on.Supported: interface:port or :port.",
			Value: ":80",
		},
		cli.StringFlag{
			Name:  "http.proxy-header",
			Usage: "Validate against this HTTP header when solving HTTP based challenges behind a reverse proxy.",
			Value: "Host",
		},
		cli.StringFlag{
			Name:  "http.webroot",
			Usage: "Set the webroot folder to use for HTTP based challenges to write directly in a file in .well-known/acme-challenge. This disables the built-in server and expects the given directory to be publicly served with access to .well-known/acme-challenge",
		},
		cli.StringSliceFlag{
			Name:  "http.memcached-host",
			Usage: "Set the memcached host(s) to use for HTTP based challenges. Challenges will be written to all specified hosts.",
		},
		cli.BoolFlag{
			Name:  "tls",
			Usage: "Use the TLS challenge to solve challenges. Can be mixed with other types of challenges.",
		},
		cli.StringFlag{
			Name:  "tls.port",
			Usage: "Set the port and interface to use for TLS based challenges to listen on. Supported: interface:port or :port.",
			Value: ":443",
		},
		cli.StringFlag{
			Name:  "dns",
			Usage: "Solve a DNS challenge using the specified provider. Can be mixed with other types of challenges. Run 'lego dnshelp' for help on usage.",
		},
		cli.BoolFlag{
			Name:  "dns.disable-cp",
			Usage: "By setting this flag to true, disables the need to wait the propagation of the TXT record to all authoritative name servers.",
		},
		cli.StringSliceFlag{
			Name:  "dns.resolvers",
			Usage: "Set the resolvers to use for performing recursive DNS queries. Supported: host:port. The default is to use the system resolvers, or Google's DNS resolvers if the system's cannot be determined.",
		},
		cli.IntFlag{
			Name:  "http-timeout",
			Usage: "Set the HTTP timeout value to a specific value in seconds.",
		},
		cli.IntFlag{
			Name:  "dns-timeout",
			Usage: "Set the DNS timeout value to a specific value in seconds. Used only when performing authoritative name servers queries.",
			Value: 10,
		},
		cli.BoolFlag{
			Name:  "pem",
			Usage: "Generate a .pem file by concatenating the .key and .crt files together.",
		},
		cli.IntFlag{
			Name:  "cert.timeout",
			Usage: "Set the certificate timeout value to a specific value in seconds. Only used when obtaining certificates.",
			Value: 30,
		},
	}
}
