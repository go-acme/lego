package cmd

import (
	"time"

	"github.com/go-acme/lego/v4/lego"
	"github.com/urfave/cli/v2"
	"software.sslmate.com/src/go-pkcs12"
)

func CreateFlags(defaultPath string) []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "domains",
			Aliases: []string{"d"},
			Usage:   "Add a domain to the process. Can be specified multiple times.",
		},
		&cli.StringFlag{
			Name:    "server",
			Aliases: []string{"s"},
			Usage:   "CA hostname (and optionally :port). The server certificate must be trusted in order to avoid further modifications to the client.",
			Value:   lego.LEDirectoryProduction,
		},
		&cli.BoolFlag{
			Name:    "accept-tos",
			Aliases: []string{"a"},
			Usage:   "By setting this flag to true you indicate that you accept the current Let's Encrypt terms of service.",
		},
		&cli.StringFlag{
			Name:    "email",
			Aliases: []string{"m"},
			Usage:   "Email used for registration and recovery contact.",
		},
		&cli.StringFlag{
			Name:    "csr",
			Aliases: []string{"c"},
			Usage:   "Certificate signing request filename, if an external CSR is to be used.",
		},
		&cli.BoolFlag{
			Name:    "eab",
			EnvVars: []string{"LEGO_EAB"},
			Usage:   "Use External Account Binding for account registration. Requires --kid and --hmac.",
		},
		&cli.StringFlag{
			Name:    "kid",
			EnvVars: []string{"LEGO_EAB_KID"},
			Usage:   "Key identifier from External CA. Used for External Account Binding.",
		},
		&cli.StringFlag{
			Name:    "hmac",
			EnvVars: []string{"LEGO_EAB_HMAC"},
			Usage:   "MAC key from External CA. Should be in Base64 URL Encoding without padding format. Used for External Account Binding.",
		},
		&cli.StringFlag{
			Name:    "key-type",
			Aliases: []string{"k"},
			Value:   "ec256",
			Usage:   "Key type to use for private keys. Supported: rsa2048, rsa3072, rsa4096, rsa8192, ec256, ec384.",
		},
		&cli.StringFlag{
			Name:  "filename",
			Usage: "(deprecated) Filename of the generated certificate.",
		},
		&cli.StringFlag{
			Name:    "path",
			EnvVars: []string{"LEGO_PATH"},
			Usage:   "Directory to use for storing the data.",
			Value:   defaultPath,
		},
		&cli.BoolFlag{
			Name:  "http",
			Usage: "Use the HTTP-01 challenge to solve challenges. Can be mixed with other types of challenges.",
		},
		&cli.StringFlag{
			Name:  "http.port",
			Usage: "Set the port and interface to use for HTTP-01 based challenges to listen on. Supported: interface:port or :port.",
			Value: ":80",
		},
		&cli.StringFlag{
			Name:  "http.proxy-header",
			Usage: "Validate against this HTTP header when solving HTTP-01 based challenges behind a reverse proxy.",
			Value: "Host",
		},
		&cli.StringFlag{
			Name: "http.webroot",
			Usage: "Set the webroot folder to use for HTTP-01 based challenges to write directly to the .well-known/acme-challenge file." +
				" This disables the built-in server and expects the given directory to be publicly served with access to .well-known/acme-challenge",
		},
		&cli.StringSliceFlag{
			Name:  "http.memcached-host",
			Usage: "Set the memcached host(s) to use for HTTP-01 based challenges. Challenges will be written to all specified hosts.",
		},
		&cli.StringFlag{
			Name:  "http.s3-bucket",
			Usage: "Set the S3 bucket name to use for HTTP-01 based challenges. Challenges will be written to the S3 bucket.",
		},
		&cli.BoolFlag{
			Name:  "tls",
			Usage: "Use the TLS-ALPN-01 challenge to solve challenges. Can be mixed with other types of challenges.",
		},
		&cli.StringFlag{
			Name:  "tls.port",
			Usage: "Set the port and interface to use for TLS-ALPN-01 based challenges to listen on. Supported: interface:port or :port.",
			Value: ":443",
		},
		&cli.StringFlag{
			Name:  "dns",
			Usage: "Solve a DNS-01 challenge using the specified provider. Can be mixed with other types of challenges. Run 'lego dnshelp' for help on usage.",
		},
		&cli.BoolFlag{
			Name:  "dns.disable-cp",
			Usage: "By setting this flag to true, disables the need to await propagation of the TXT record to all authoritative name servers.",
		},
		&cli.StringSliceFlag{
			Name: "dns.resolvers",
			Usage: "Set the resolvers to use for performing (recursive) CNAME resolving and apex domain determination." +
				" For DNS-01 challenge verification, the authoritative DNS server is queried directly." +
				" Supported: host:port." +
				" The default is to use the system resolvers, or Google's DNS resolvers if the system's cannot be determined.",
		},
		&cli.IntFlag{
			Name:  "http-timeout",
			Usage: "Set the HTTP timeout value to a specific value in seconds.",
		},
		&cli.IntFlag{
			Name:  "dns-timeout",
			Usage: "Set the DNS timeout value to a specific value in seconds. Used only when performing authoritative name server queries.",
			Value: 10,
		},
		&cli.BoolFlag{
			Name:  "pem",
			Usage: "Generate an additional .pem (base64) file by concatenating the .key and .crt files together.",
		},
		&cli.BoolFlag{
			Name:  "pfx",
			Usage: "Generate an additional .pfx (PKCS#12) file by concatenating the .key and .crt and issuer .crt files together.",
		},
		&cli.StringFlag{
			Name:  "pfx.pass",
			Usage: "The password used to encrypt the .pfx (PCKS#12) file.",
			Value: pkcs12.DefaultPassword,
		},
		&cli.IntFlag{
			Name:  "cert.timeout",
			Usage: "Set the certificate timeout value to a specific value in seconds. Only used when obtaining certificates.",
			Value: 30,
		},
		&cli.StringFlag{
			Name:  "user-agent",
			Usage: "Add to the user-agent sent to the CA to identify an application embedding lego-cli",
		},
	}
}

func getTime(ctx *cli.Context, name string) time.Time {
	value := ctx.Timestamp(name)
	if value == nil {
		return time.Time{}
	}
	return *value
}
