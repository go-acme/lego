package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
	"github.com/urfave/cli/v3"
	"software.sslmate.com/src/go-pkcs12"
)

// Flag names related to the account and domains.
const (
	flgDomains   = "domains"
	flgAcceptTOS = "accept-tos"
	flgEmail     = "email"
	flgEAB       = "eab"
	flgKID       = "kid"
	flgHMAC      = "hmac"
)

// Flag names related to Obtain certificates.
const (
	flgCSR                            = "csr"
	flgNoBundle                       = "no-bundle"
	flgMustStaple                     = "must-staple"
	flgNotBefore                      = "not-before"
	flgNotAfter                       = "not-after"
	flgPreferredChain                 = "preferred-chain"
	flgProfile                        = "profile"
	flgAlwaysDeactivateAuthorizations = "always-deactivate-authorizations"
)

// Flag names related to the output.
const (
	flgFilename  = "filename"
	flgPath      = "path"
	flgPEM       = "pem"
	flgPFX       = "pfx"
	flgPFXPass   = "pfx.pass"
	flgPFXFormat = "pfx.format"
)

// Flag names related to the ACME client.
const (
	flgServer              = "server"
	flgDisableCommonName   = "disable-cn"
	flgKeyType             = "key-type"
	flgHTTPTimeout         = "http-timeout"
	flgTLSSkipVerify       = "tls-skip-verify"
	flgCertTimeout         = "cert.timeout"
	flgOverallRequestLimit = "overall-request-limit"
	flgUserAgent           = "user-agent"
)

// Flag names related to HTTP-01 challenge.
const (
	flgHTTP              = "http"
	flgHTTPPort          = "http.port"
	flgHTTPDelay         = "http.delay"
	flgHTTPProxyHeader   = "http.proxy-header"
	flgHTTPWebroot       = "http.webroot"
	flgHTTPMemcachedHost = "http.memcached-host"
	flgHTTPS3Bucket      = "http.s3-bucket"
)

// Flag names related to TLS-ALPN-01 challenge.
const (
	flgTLS      = "tls"
	flgTLSPort  = "tls.port"
	flgTLSDelay = "tls.delay"
)

// Flag names related to DNS-01 challenge.
const (
	flgDNS                      = "dns"
	flgDNSDisableCP             = "dns.disable-cp"
	flgDNSPropagationWait       = "dns.propagation-wait"
	flgDNSPropagationDisableANS = "dns.propagation-disable-ans"
	flgDNSPropagationRNS        = "dns.propagation-rns"
	flgDNSResolvers             = "dns.resolvers"
	flgDNSTimeout               = "dns-timeout"
)

// Environment variable names.
const (
	envEAB         = "LEGO_EAB"
	envEABHMAC     = "LEGO_EAB_HMAC"
	envEABKID      = "LEGO_EAB_KID"
	envEmail       = "LEGO_EMAIL"
	envPath        = "LEGO_PATH"
	envPFX         = "LEGO_PFX"
	envPFXFormat   = "LEGO_PFX_FORMAT"
	envPFXPassword = "LEGO_PFX_PASSWORD"
	envServer      = "LEGO_SERVER"
)

func CreateACMEClientFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     flgServer,
			Aliases:  []string{"s"},
			Sources:  cli.EnvVars(envServer),
			Usage:    "CA hostname (and optionally :port). The server certificate must be trusted in order to avoid further modifications to the client.",
			Value:    lego.LEDirectoryProduction,
			Required: true,
		},
		&cli.BoolFlag{
			Name:  flgDisableCommonName,
			Usage: "Disable the use of the common name in the CSR.",
		},
		&cli.StringFlag{
			Name:    flgKeyType,
			Aliases: []string{"k"},
			Value:   "ec256",
			Usage:   "Key type to use for private keys. Supported: rsa2048, rsa3072, rsa4096, rsa8192, ec256, ec384.",
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

func CreateHTTPChallengeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:  flgHTTP,
			Usage: "Use the HTTP-01 challenge to solve challenges. Can be mixed with other types of challenges.",
		},
		&cli.StringFlag{
			Name:  flgHTTPPort,
			Usage: "Set the port and interface to use for HTTP-01 based challenges to listen on. Supported: interface:port or :port.",
			Value: ":80",
		},
		&cli.DurationFlag{
			Name:  flgHTTPDelay,
			Usage: "Delay between the starts of the HTTP server (use for HTTP-01 based challenges) and the validation of the challenge.",
			Value: 0,
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
	}
}

func CreateTLSChallengeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:  flgTLS,
			Usage: "Use the TLS-ALPN-01 challenge to solve challenges. Can be mixed with other types of challenges.",
		},
		&cli.StringFlag{
			Name:  flgTLSPort,
			Usage: "Set the port and interface to use for TLS-ALPN-01 based challenges to listen on. Supported: interface:port or :port.",
			Value: ":443",
		},
		&cli.DurationFlag{
			Name:  flgTLSDelay,
			Usage: "Delay between the start of the TLS listener (use for TLSALPN-01 based challenges) and the validation of the challenge.",
			Value: 0,
		},
	}
}

func CreateDNSChallengeFlags() []cli.Flag {
	return []cli.Flag{
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
			Name:  flgDNSTimeout,
			Usage: "Set the DNS timeout value to a specific value in seconds. Used only when performing authoritative name server queries.",
			Value: 10,
		},
	}
}

func CreateOutputFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  flgFilename,
			Usage: "(deprecated) Filename of the generated certificate.",
		},
		CreatePathFlag(true),
		&cli.BoolFlag{
			Name:  flgPEM,
			Usage: "Generate an additional .pem (base64) file by concatenating the .key and .crt files together.",
		},
		&cli.BoolFlag{
			Name:    flgPFX,
			Usage:   "Generate an additional .pfx (PKCS#12) file by concatenating the .key and .crt and issuer .crt files together.",
			Sources: cli.EnvVars(envPFX),
		},
		&cli.StringFlag{
			Name:    flgPFXPass,
			Usage:   "The password used to encrypt the .pfx (PCKS#12) file.",
			Value:   pkcs12.DefaultPassword,
			Sources: cli.EnvVars(envPFXPassword),
		},
		&cli.StringFlag{
			Name:    flgPFXFormat,
			Usage:   "The encoding format to use when encrypting the .pfx (PCKS#12) file. Supported: RC2, DES, SHA256.",
			Value:   "RC2",
			Sources: cli.EnvVars(envPFXFormat),
		},
	}
}

func CreateAccountFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    flgAcceptTOS,
			Aliases: []string{"a"},
			Usage:   "By setting this flag to true you indicate that you accept the current Let's Encrypt terms of service.",
		},
		&cli.StringFlag{
			Name:    flgEmail,
			Aliases: []string{"m"},
			Sources: cli.EnvVars(envEmail),
			Usage:   "Email used for registration and recovery contact.",
		},
		&cli.BoolFlag{
			Name:    flgEAB,
			Sources: cli.EnvVars(envEAB),
			Usage:   "Use External Account Binding for account registration. Requires --kid and --hmac.",
		},
		&cli.StringFlag{
			Name:    flgKID,
			Sources: cli.EnvVars(envEABKID),
			Usage:   "Key identifier from External CA. Used for External Account Binding.",
		},
		&cli.StringFlag{
			Name:    flgHMAC,
			Sources: cli.EnvVars(envEABHMAC),
			Usage:   "MAC key from External CA. Should be in Base64 URL Encoding without padding format. Used for External Account Binding.",
		},
	}
}

func CreateObtainFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    flgCSR,
			Aliases: []string{"c"},
			Usage:   "Certificate signing request filename, if an external CSR is to be used.",
		},
		&cli.BoolFlag{
			Name:  flgNoBundle,
			Usage: "Do not create a certificate bundle by adding the issuers certificate to the new certificate.",
		},
		&cli.BoolFlag{
			Name: flgMustStaple,
			Usage: "Include the OCSP must staple TLS extension in the CSR and generated certificate." +
				" Only works if the CSR is generated by lego.",
		},
		&cli.TimestampFlag{
			Name:  flgNotBefore,
			Usage: "Set the notBefore field in the certificate (RFC3339 format)",
			Config: cli.TimestampConfig{
				Layouts: []string{time.RFC3339},
			},
		},
		&cli.TimestampFlag{
			Name:  flgNotAfter,
			Usage: "Set the notAfter field in the certificate (RFC3339 format)",
			Config: cli.TimestampConfig{
				Layouts: []string{time.RFC3339},
			},
		},
		&cli.StringFlag{
			Name: flgPreferredChain,
			Usage: "If the CA offers multiple certificate chains, prefer the chain with an issuer matching this Subject Common Name." +
				" If no match, the default offered chain will be used.",
		},
		&cli.StringFlag{
			Name:  flgProfile,
			Usage: "If the CA offers multiple certificate profiles (draft-ietf-acme-profiles), choose this one.",
		},
		&cli.StringFlag{
			Name:  flgAlwaysDeactivateAuthorizations,
			Usage: "Force the authorizations to be relinquished even if the certificate request was successful.",
		},
	}
}

func CreateFlags() []cli.Flag {
	var flags []cli.Flag

	flags = append(flags, CreateDomainFlag())
	flags = append(flags, CreateAccountFlags()...)
	flags = append(flags, CreateACMEClientFlags()...)
	flags = append(flags, CreateOutputFlags()...)
	flags = append(flags, CreateHTTPChallengeFlags()...)
	flags = append(flags, CreateTLSChallengeFlags()...)
	flags = append(flags, CreateDNSChallengeFlags()...)

	return flags
}

func CreateDomainFlag() cli.Flag {
	return &cli.StringSliceFlag{
		Name:    flgDomains,
		Aliases: []string{"d"},
		Usage:   "Add a domain to the process. Can be specified multiple times or use comma as a separator.",
	}
}

func CreatePathFlag(forceCreation bool) cli.Flag {
	return &cli.StringFlag{
		Name:    flgPath,
		Sources: cli.NewValueSourceChain(cli.EnvVar(envPath), &defaultPathValueSource{}),
		Usage:   "Directory to use for storing the data.",
		Validator: func(s string) error {
			if !forceCreation {
				return nil
			}

			err := storage.CreateNonExistingFolder(s)
			if err != nil {
				return fmt.Errorf("could not check/create the path %q: %w", s, err)
			}

			return nil
		},
		Required: true,
	}
}

// defaultPathValueSource gets the default path based on the current working directory.
// The field value is only here because clihelp/generator.
type defaultPathValueSource struct{}

func (p *defaultPathValueSource) String() string {
	return "default path"
}

func (p *defaultPathValueSource) GoString() string {
	return "&defaultPathValueSource{}"
}

func (p *defaultPathValueSource) Lookup() (string, bool) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", false
	}

	return filepath.Join(cwd, ".lego"), true
}
