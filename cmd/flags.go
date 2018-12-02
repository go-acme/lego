package cmd

import (
	"github.com/urfave/cli"
	"github.com/xenolf/lego/lego"
)

func CreateFlags(defaultPath string) []cli.Flag {
	return []cli.Flag{
		cli.StringSliceFlag{
			Name:  "domains, d",
			Usage: "Add a domain to the process. Can be specified multiple times.",
		},
		cli.StringFlag{
			Name:  "server, s",
			Value: lego.LEDirectoryProduction,
			Usage: "CA hostname (and optionally :port). The server certificate must be trusted in order to avoid further modifications to the client.",
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
			Usage: "Certificate signing request filename, if an external CSR is to be used",
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
			Value: "rsa2048",
			Usage: "Key type to use for private keys. Supported: rsa2048, rsa4096, rsa8192, ec256, ec384",
		},
		cli.StringFlag{
			Name:  "filename",
			Usage: "Filename of the generated certificate",
		},
		cli.StringFlag{
			Name:  "path",
			Usage: "Directory to use for storing the data",
			Value: defaultPath,
		},
		cli.StringSliceFlag{
			Name:  "exclude, x",
			Usage: "Explicitly disallow solvers by name from being used. Solvers: \"http-01\", \"dns-01\", \"tls-alpn-01\".",
		},
		cli.IntFlag{
			Name:  "http-timeout",
			Usage: "Set the HTTP timeout value to a specific value in seconds. The default is 10 seconds.",
		},
		cli.StringFlag{
			Name:  "webroot",
			Usage: "Set the webroot folder to use for HTTP based challenges to write directly in a file in .well-known/acme-challenge",
		},
		cli.StringSliceFlag{
			Name:  "memcached-host",
			Usage: "Set the memcached host(s) to use for HTTP based challenges. Challenges will be written to all specified hosts.",
		},
		cli.StringFlag{
			Name:  "http",
			Usage: "Set the port and interface to use for HTTP based challenges to listen on. Supported: interface:port or :port",
		},
		cli.StringFlag{
			Name:  "tls",
			Usage: "Set the port and interface to use for TLS based challenges to listen on. Supported: interface:port or :port",
		},
		cli.StringFlag{
			Name:  "dns",
			Usage: "Solve a DNS challenge using the specified provider. Disables all other challenges. Run 'lego dnshelp' for help on usage.",
		},
		cli.BoolFlag{
			Name:  "dns-disable-cp",
			Usage: "By setting this flag to true, disables the need to wait the propagation of the TXT record to all authoritative name servers.",
		},
		cli.StringSliceFlag{
			Name:  "dns-resolvers",
			Usage: "Set the resolvers to use for performing recursive DNS queries. Supported: host:port. The default is to use the system resolvers, or Google's DNS resolvers if the system's cannot be determined.",
		},
		cli.IntFlag{
			Name:  "dns-timeout",
			Usage: "Set the DNS timeout value to a specific value in seconds. Used only when performing authoritative name servers queries. The default is 10 seconds.",
		},
		cli.BoolFlag{
			Name:  "pem",
			Usage: "Generate a .pem file by concatenating the .key and .crt files together.",
		},
	}
}
