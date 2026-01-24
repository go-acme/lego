package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-acme/lego/v5/acme"
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

// Flag names related to the network stack.
const (
	flgIPv4Only = "ipv4only"
	flgIPv6Only = "ipv6only"
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

// Flags names related to hooks.
const (
	flgDeployHook        = "deploy-hook"
	flgDeployHookTimeout = "deploy-hook-timeout"
)

// Flag names related to the specific run command.
const (
	flgPrivateKey = "private-key"
)

// Flag names related to the specific renew command.
const (
	flgRenewDays              = "days"
	flgRenewDynamic           = "dynamic"
	flgARIDisable             = "ari-disable"
	flgARIWaitToRenewDuration = "ari-wait-to-renew-duration"
	flgReuseKey               = "reuse-key"
	flgNoRandomSleep          = "no-random-sleep"
	flgForceCertDomains       = "force-cert-domains"
)

// Flag names related to the specific revoke command.
const (
	flgKeep   = "keep"
	flgReason = "reason"
)

// Flag names related to the list command.
const (
	flgAccounts = "accounts"
	flgNames    = "names"
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

func CreateChallengesFlags() []cli.Flag {
	var flags []cli.Flag

	flags = append(flags, CreateHTTPChallengeFlags()...)
	flags = append(flags, CreateTLSChallengeFlags()...)
	flags = append(flags, CreateDNSChallengeFlags()...)
	flags = append(flags, CreateNetworkStackFlags()...)

	return flags
}

func CreateNetworkStackFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    flgIPv4Only,
			Aliases: []string{"4"},
			Usage:   "Use IPv4 only.",
		},
		&cli.BoolFlag{
			Name:    flgIPv6Only,
			Aliases: []string{"6"},
			Usage:   "Use IPv6 only.",
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

func CreateHookFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  flgDeployHook,
			Usage: "Define a hook. The hook is executed only when the certificates are effectively created/renewed.",
		},
		&cli.DurationFlag{
			Name:  flgDeployHookTimeout,
			Usage: "Define the timeout for the hook execution.",
			Value: 2 * time.Minute,
		},
	}
}

func CreateBaseFlags() []cli.Flag {
	var flags []cli.Flag

	flags = append(flags, CreateDomainFlag())
	flags = append(flags, CreateAccountFlags()...)
	flags = append(flags, CreateACMEClientFlags()...)
	flags = append(flags, CreateOutputFlags()...)

	return flags
}

func createRunFlags() []cli.Flag {
	flags := CreateBaseFlags()

	flags = append(flags, CreateChallengesFlags()...)
	flags = append(flags, CreateObtainFlags()...)
	flags = append(flags, CreateHookFlags()...)

	flags = append(flags,
		&cli.StringFlag{
			Name:  flgPrivateKey,
			Usage: "Path to private key (in PEM encoding) for the certificate. By default, the private key is generated.",
		},
	)

	return flags
}

func createRenewFlags() []cli.Flag {
	flags := CreateBaseFlags()

	flags = append(flags, CreateChallengesFlags()...)
	flags = append(flags, CreateObtainFlags()...)
	flags = append(flags, CreateHookFlags()...)

	flags = append(flags,
		&cli.IntFlag{
			Name:  flgRenewDays,
			Value: 30,
			Usage: "The number of days left on a certificate to renew it.",
		},
		// TODO(ldez): in v5, remove this flag, use this behavior as default.
		&cli.BoolFlag{
			Name:  flgRenewDynamic,
			Value: false,
			Usage: "Compute dynamically, based on the lifetime of the certificate(s), when to renew: use 1/3rd of the lifetime left, or 1/2 of the lifetime for short-lived certificates). This supersedes --days and will be the default behavior in Lego v5.",
		},
		&cli.BoolFlag{
			Name:  flgARIDisable,
			Usage: "Do not use the renewalInfo endpoint (RFC9773) to check if a certificate should be renewed.",
		},
		&cli.DurationFlag{
			Name:  flgARIWaitToRenewDuration,
			Usage: "The maximum duration you're willing to sleep for a renewal time returned by the renewalInfo endpoint.",
		},
		&cli.BoolFlag{
			Name:  flgReuseKey,
			Usage: "Used to indicate you want to reuse your current private key for the new certificate.",
		},
		&cli.BoolFlag{
			Name: flgNoRandomSleep,
			Usage: "Do not add a random sleep before the renewal." +
				" We do not recommend using this flag if you are doing your renewals in an automated way.",
		},
		&cli.BoolFlag{
			Name:  flgForceCertDomains,
			Usage: "Check and ensure that the cert's domain list matches those passed in the domains argument.",
		},
	)

	return flags
}

func createRevokeFlags() []cli.Flag {
	flags := CreateBaseFlags()

	flags = append(flags,
		&cli.BoolFlag{
			Name:    flgKeep,
			Aliases: []string{"k"},
			Usage:   "Keep the certificates after the revocation instead of archiving them.",
		},
		&cli.UintFlag{
			Name: flgReason,
			Usage: "Identifies the reason for the certificate revocation." +
				" See https://www.rfc-editor.org/rfc/rfc5280.html#section-5.3.1." +
				" Valid values are:" +
				" 0 (unspecified), 1 (keyCompromise), 2 (cACompromise), 3 (affiliationChanged)," +
				" 4 (superseded), 5 (cessationOfOperation), 6 (certificateHold), 8 (removeFromCRL)," +
				" 9 (privilegeWithdrawn), or 10 (aACompromise).",
			Value: acme.CRLReasonUnspecified,
		},
	)

	return flags
}

func createListFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    flgAccounts,
			Aliases: []string{"a"},
			Usage:   "Display accounts.",
		},
		&cli.BoolFlag{
			Name:    flgNames,
			Aliases: []string{"n"},
			Usage:   "Display certificate names only.",
		},
		CreatePathFlag(false),
	}
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
