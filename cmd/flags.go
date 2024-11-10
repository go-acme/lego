package cmd

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/urfave/cli/v2"
	"software.sslmate.com/src/go-pkcs12"
)

// Flag names.
const (
	flgDomains                  = "domains"
	flgServer                   = "server"
	flgAcceptTOS                = "accept-tos"
	flgEmail                    = "email"
	flgCSR                      = "csr"
	flgEAB                      = "eab"
	flgKID                      = "kid"
	flgHMAC                     = "hmac"
	flgKeyType                  = "key-type"
	flgFilename                 = "filename"
	flgPath                     = "path"
	flgHTTP                     = "http"
	flgHTTPPort                 = "http.port"
	flgHTTPProxyHeader          = "http.proxy-header"
	flgHTTPWebroot              = "http.webroot"
	flgHTTPMemcachedHost        = "http.memcached-host"
	flgHTTPS3Bucket             = "http.s3-bucket"
	flgTLS                      = "tls"
	flgTLSPort                  = "tls.port"
	flgDNS                      = "dns"
	flgDNSDisableCP             = "dns.disable-cp"
	flgDNSPropagationWait       = "dns.propagation-wait"
	flgDNSPropagationDisableANS = "dns.propagation-disable-ans"
	flgDNSPropagationRNS        = "dns.propagation-rns"
	flgDNSResolvers             = "dns.resolvers"
	flgHTTPTimeout              = "http-timeout"
	flgTLSSkipVerify            = "tls-skip-verify"
	flgDNSTimeout               = "dns-timeout"
	flgPEM                      = "pem"
	flgPFX                      = "pfx"
	flgPFXPass                  = "pfx.pass"
	flgPFXFormat                = "pfx.format"
	flgCertTimeout              = "cert.timeout"
	flgOverallRequestLimit      = "overall-request-limit"
	flgUserAgent                = "user-agent"
)

func CreateFlags(defaultPath string) []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:    flgDomains,
			Aliases: []string{"d"},
			Usage:   "Add a domain to the process. Can be specified multiple times.",
		},
		&cli.StringFlag{
			Name:    flgServer,
			Aliases: []string{"s"},
			EnvVars: []string{"LEGO_SERVER"},
			Usage:   "CA hostname (and optionally :port). The server certificate must be trusted in order to avoid further modifications to the client.",
			Value:   lego.LEDirectoryProduction,
		},
		&cli.BoolFlag{
			Name:    flgAcceptTOS,
			Aliases: []string{"a"},
			Usage:   "By setting this flag to true you indicate that you accept the current Let's Encrypt terms of service.",
		},
		&cli.StringFlag{
			Name:    flgEmail,
			Aliases: []string{"m"},
			Usage:   "Email used for registration and recovery contact.",
		},
		&cli.StringFlag{
			Name:    flgCSR,
			Aliases: []string{"c"},
			Usage:   "Certificate signing request filename, if an external CSR is to be used.",
		},
		&cli.BoolFlag{
			Name:    flgEAB,
			EnvVars: []string{"LEGO_EAB"},
			Usage:   "Use External Account Binding for account registration. Requires --kid and --hmac.",
		},
		&cli.StringFlag{
			Name:    flgKID,
			EnvVars: []string{"LEGO_EAB_KID"},
			Usage:   "Key identifier from External CA. Used for External Account Binding.",
		},
		&cli.StringFlag{
			Name:    flgHMAC,
			EnvVars: []string{"LEGO_EAB_HMAC"},
			Usage:   "MAC key from External CA. Should be in Base64 URL Encoding without padding format. Used for External Account Binding.",
		},
		&cli.StringFlag{
			Name:    flgKeyType,
			Aliases: []string{"k"},
			Value:   "ec256",
			Usage:   "Key type to use for private keys. Supported: rsa2048, rsa3072, rsa4096, rsa8192, ec256, ec384.",
		},
		&cli.StringFlag{
			Name:  flgFilename,
			Usage: "(deprecated) Filename of the generated certificate.",
		},
		&cli.StringFlag{
			Name:    flgPath,
			EnvVars: []string{"LEGO_PATH"},
			Usage:   "Directory to use for storing the data.",
			Value:   defaultPath,
		},
		&cli.BoolFlag{
			Name:  flgHTTP,
			Usage: "Use the HTTP-01 challenge to solve challenges. Can be mixed with other types of challenges.",
		},
		&cli.StringFlag{
			Name:  flgHTTPPort,
			Usage: "Set the port and interface to use for HTTP-01 based challenges to listen on. Supported: interface:port or :port.",
			Value: ":80",
		},
		&cli.StringFlag{
			Name:  flgHTTPProxyHeader,
			Usage: "Validate against this HTTP header when solving HTTP-01 based challenges behind a reverse proxy.",
			Value: "Host",
		},
		&cli.StringFlag{
			Name: flgHTTPWebroot,
			Usage: "Set the webroot folder to use for HTTP-01 based challenges to write directly to the .well-known/acme-challenge file." +
				" This disables the built-in server and expects the given directory to be publicly served with access to .well-known/acme-challenge",
		},
		&cli.StringSliceFlag{
			Name:  flgHTTPMemcachedHost,
			Usage: "Set the memcached host(s) to use for HTTP-01 based challenges. Challenges will be written to all specified hosts.",
		},
		&cli.StringFlag{
			Name:  flgHTTPS3Bucket,
			Usage: "Set the S3 bucket name to use for HTTP-01 based challenges. Challenges will be written to the S3 bucket.",
		},
		&cli.BoolFlag{
			Name:  flgTLS,
			Usage: "Use the TLS-ALPN-01 challenge to solve challenges. Can be mixed with other types of challenges.",
		},
		&cli.StringFlag{
			Name:  flgTLSPort,
			Usage: "Set the port and interface to use for TLS-ALPN-01 based challenges to listen on. Supported: interface:port or :port.",
			Value: ":443",
		},
		&cli.StringFlag{
			Name:  flgDNS,
			Usage: "Solve a DNS-01 challenge using the specified provider. Can be mixed with other types of challenges. Run 'lego dnshelp' for help on usage.",
		},
		&cli.BoolFlag{
			Name:  flgDNSDisableCP,
			Usage: fmt.Sprintf("(deprecated) use %s instead.", flgDNSPropagationDisableANS),
		},
		&cli.BoolFlag{
			Name:  flgDNSPropagationDisableANS,
			Usage: "By setting this flag to true, disables the need to await propagation of the TXT record to all authoritative name servers.",
		},
		&cli.BoolFlag{
			Name:  flgDNSPropagationRNS,
			Usage: "By setting this flag to true, use all the recursive nameservers to check the propagation of the TXT record.",
		},
		&cli.DurationFlag{
			Name:  flgDNSPropagationWait,
			Usage: "By setting this flag, disables all the propagation checks of the TXT record and uses a wait duration instead.",
		},
		&cli.StringSliceFlag{
			Name: flgDNSResolvers,
			Usage: "Set the resolvers to use for performing (recursive) CNAME resolving and apex domain determination." +
				" For DNS-01 challenge verification, the authoritative DNS server is queried directly." +
				" Supported: host:port." +
				" The default is to use the system resolvers, or Google's DNS resolvers if the system's cannot be determined.",
		},
		&cli.IntFlag{
			Name:  flgHTTPTimeout,
			Usage: "Set the HTTP timeout value to a specific value in seconds.",
		},
		&cli.BoolFlag{
			Name:  flgTLSSkipVerify,
			Usage: "Skip the TLS verification of the ACME server.",
		},
		&cli.IntFlag{
			Name:  flgDNSTimeout,
			Usage: "Set the DNS timeout value to a specific value in seconds. Used only when performing authoritative name server queries.",
			Value: 10,
		},
		&cli.BoolFlag{
			Name:  flgPEM,
			Usage: "Generate an additional .pem (base64) file by concatenating the .key and .crt files together.",
		},
		&cli.BoolFlag{
			Name:    flgPFX,
			Usage:   "Generate an additional .pfx (PKCS#12) file by concatenating the .key and .crt and issuer .crt files together.",
			EnvVars: []string{"LEGO_PFX"},
		},
		&cli.StringFlag{
			Name:    flgPFXPass,
			Usage:   "The password used to encrypt the .pfx (PCKS#12) file.",
			Value:   pkcs12.DefaultPassword,
			EnvVars: []string{"LEGO_PFX_PASSWORD"},
		},
		&cli.StringFlag{
			Name:    flgPFXFormat,
			Usage:   "The encoding format to use when encrypting the .pfx (PCKS#12) file. Supported: RC2, DES, SHA256.",
			Value:   "RC2",
			EnvVars: []string{"LEGO_PFX_FORMAT"},
		},
		&cli.IntFlag{
			Name:  flgCertTimeout,
			Usage: "Set the certificate timeout value to a specific value in seconds. Only used when obtaining certificates.",
			Value: 30,
		},
		&cli.IntFlag{
			Name:  flgOverallRequestLimit,
			Usage: "ACME overall requests limit.",
			Value: certificate.DefaultOverallRequestLimit,
		},
		&cli.StringFlag{
			Name:  flgUserAgent,
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
