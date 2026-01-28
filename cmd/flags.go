package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
	"software.sslmate.com/src/go-pkcs12"
)

const (
	categoryHTTP01Challenge    = "Flags related to the HTTP-01 challenge:"
	categoryTLSALPN01Challenge = "Flags related to the TLS-ALPN-01 challenge:"
	categoryDNS01Challenge     = "Flags related to the DNS-01 challenge:"
	categoryStorage            = "Flags related to the storage:"
	categoryHooks              = "Flags related to hooks:"
	categoryEAB                = "Flags related to External Account Binding:"
	categoryACMEClient         = "Flags related to the ACME client:"
	categoryAdvanced           = "Flags related to advanced options:"
	categoryARI                = "Flags related to ACME Renewal Information (ARI) Extension:"
)

// Flag names related to the account and domains.
const (
	flgDomains   = "domains"
	flgAcceptTOS = "accept-tos"
	flgEmail     = "email"
	flgAccountID = "account-id"
	flgEAB       = "eab"
	flgEABKID    = "eab.kid"
	flgEABHMAC   = "eab.hmac"
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

// Flag names related to the storage.
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
	flgEnableCommonName    = "enable-cn"
	flgKeyType             = "key-type"
	flgHTTPTimeout         = "http-timeout"
	flgTLSSkipVerify       = "tls-skip-verify"
	flgOverallRequestLimit = "overall-request-limit"
	flgUserAgent           = "user-agent"
)

// Flag names related to certificates.
const (
	flgCertTimeout = "cert.timeout"
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
	flgDNSPropagationWait       = "dns.propagation.wait"
	flgDNSPropagationDisableANS = "dns.propagation.disable-ans"
	flgDNSPropagationDisableRNS = "dns.propagation.disable-rns"
	flgDNSResolvers             = "dns.resolvers"
	flgDNSTimeout               = "dns.timeout"
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

func toEnvName(flg string) string {
	fields := strings.FieldsFunc(flg, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	return "LEGO_" + strings.ToUpper(strings.Join(fields, "_"))
}

func createACMEClientFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			// NOTE(ldez): if Required is true, then the default value is not display in the help.
			Name:    flgServer,
			Aliases: []string{"s"},
			Sources: cli.EnvVars(toEnvName(flgServer)),
			Usage: fmt.Sprintf("CA (ACME server). It can be either a URL or a shortcode."+
				"\n\t(available shortcodes: %s)", strings.Join(lego.GetAllCodes(), ", ")),
			Value: lego.DirectoryURLLetsEncrypt,
			Action: func(ctx context.Context, cmd *cli.Command, s string) error {
				directoryURL, err := lego.GetDirectoryURL(s)
				if err != nil {
					log.Debug("Server shortcode not found. Use the value as URL.", slog.String("value", s), log.ErrorAttr(err))

					directoryURL = s
				}

				return cmd.Set(flgServer, directoryURL)
			},
		},
		&cli.BoolFlag{
			Category: categoryAdvanced,
			Name:     flgEnableCommonName,
			Usage:    "Enable the use of the common name. (Not recommended)",
		},
		&cli.StringFlag{
			Name:    flgKeyType,
			Aliases: []string{"k"},
			Value:   "ec256",
			Usage:   "Key type to use for private keys. Supported: rsa2048, rsa3072, rsa4096, rsa8192, ec256, ec384.",
		},
		&cli.IntFlag{
			Category: categoryACMEClient,
			Name:     flgHTTPTimeout,
			Usage:    "Set the HTTP timeout value to a specific value in seconds.",
		},
		&cli.BoolFlag{
			Category: categoryACMEClient,
			Name:     flgTLSSkipVerify,
			Usage:    "Skip the TLS verification of the ACME server.",
		},
		&cli.IntFlag{
			Category: categoryAdvanced,
			Name:     flgCertTimeout,
			Usage:    "Set the certificate timeout value to a specific value in seconds. Only used when obtaining certificates.",
			Value:    30,
		},
		&cli.IntFlag{
			Category: categoryACMEClient,
			Name:     flgOverallRequestLimit,
			Usage:    "ACME overall requests limit.",
			Value:    certificate.DefaultOverallRequestLimit,
		},
		&cli.StringFlag{
			Category: categoryACMEClient,
			Name:     flgUserAgent,
			Usage:    "Add to the user-agent sent to the CA to identify an application embedding lego-cli",
		},
	}
}

func createChallengesFlags() []cli.Flag {
	var flags []cli.Flag

	flags = append(flags, createHTTPChallengeFlags()...)
	flags = append(flags, createTLSChallengeFlags()...)
	flags = append(flags, createDNSChallengeFlags()...)
	flags = append(flags, createNetworkStackFlags()...)

	return flags
}

func createNetworkStackFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Category: categoryAdvanced,
			Name:     flgIPv4Only,
			Aliases:  []string{"4"},
			Usage:    "Use IPv4 only.",
		},
		&cli.BoolFlag{
			Category: categoryAdvanced,
			Name:     flgIPv6Only,
			Aliases:  []string{"6"},
			Usage:    "Use IPv6 only.",
		},
	}
}

func createHTTPChallengeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Category: categoryHTTP01Challenge,
			Name:     flgHTTP,
			Usage:    "Use the HTTP-01 challenge to solve challenges. Can be mixed with other types of challenges.",
		},
		&cli.StringFlag{
			Category: categoryHTTP01Challenge,
			Name:     flgHTTPPort,
			Usage:    "Set the port and interface to use for HTTP-01 based challenges to listen on. Supported: interface:port or :port.",
			Value:    ":80",
		},
		&cli.DurationFlag{
			Category: categoryHTTP01Challenge,
			Name:     flgHTTPDelay,
			Usage:    "Delay between the starts of the HTTP server (use for HTTP-01 based challenges) and the validation of the challenge.",
			Value:    0,
		},
		&cli.StringFlag{
			Category: categoryHTTP01Challenge,
			Name:     flgHTTPProxyHeader,
			Usage:    "Validate against this HTTP header when solving HTTP-01 based challenges behind a reverse proxy.",
			Value:    "Host",
		},
		&cli.StringFlag{
			Category: categoryHTTP01Challenge,
			Name:     flgHTTPWebroot,
			Usage: "Set the webroot folder to use for HTTP-01 based challenges to write directly to the .well-known/acme-challenge file." +
				" This disables the built-in server and expects the given directory to be publicly served with access to .well-known/acme-challenge",
		},
		&cli.StringSliceFlag{
			Category: categoryHTTP01Challenge,
			Name:     flgHTTPMemcachedHost,
			Usage:    "Set the memcached host(s) to use for HTTP-01 based challenges. Challenges will be written to all specified hosts.",
		},
		&cli.StringFlag{
			Category: categoryHTTP01Challenge,
			Name:     flgHTTPS3Bucket,
			Usage:    "Set the S3 bucket name to use for HTTP-01 based challenges. Challenges will be written to the S3 bucket.",
		},
	}
}

func createTLSChallengeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Category: categoryTLSALPN01Challenge,
			Name:     flgTLS,
			Usage:    "Use the TLS-ALPN-01 challenge to solve challenges. Can be mixed with other types of challenges.",
		},
		&cli.StringFlag{
			Category: categoryTLSALPN01Challenge,
			Name:     flgTLSPort,
			Usage:    "Set the port and interface to use for TLS-ALPN-01 based challenges to listen on. Supported: interface:port or :port.",
			Value:    ":443",
		},
		&cli.DurationFlag{
			Category: categoryTLSALPN01Challenge,
			Name:     flgTLSDelay,
			Usage:    "Delay between the start of the TLS listener (use for TLSALPN-01 based challenges) and the validation of the challenge.",
			Value:    0,
		},
	}
}

func createDNSChallengeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Category: categoryDNS01Challenge,
			Name:     flgDNS,
			Usage:    "Solve a DNS-01 challenge using the specified provider. Can be mixed with other types of challenges. Run 'lego dnshelp' for help on usage.",
		},
		&cli.BoolFlag{
			Category: categoryDNS01Challenge,
			Name:     flgDNSPropagationDisableANS,
			Usage:    "By setting this flag to true, disables the need to await propagation of the TXT record to all authoritative name servers.",
		},
		&cli.BoolFlag{
			Category: categoryDNS01Challenge,
			Name:     flgDNSPropagationDisableRNS,
			Usage:    "By setting this flag to true, disables the need to await propagation of the TXT record to all recursive name servers (aka resolvers).",
		},
		&cli.DurationFlag{
			Category: categoryDNS01Challenge,
			Name:     flgDNSPropagationWait,
			Usage:    "By setting this flag, disables all the propagation checks of the TXT record and uses a wait duration instead.",
			Validator: func(d time.Duration) error {
				if d < 0 {
					return errors.New("it cannot be negative")
				}

				return nil
			},
		},
		&cli.StringSliceFlag{
			Category: categoryDNS01Challenge,
			Name:     flgDNSResolvers,
			Usage: "Set the resolvers to use for performing (recursive) CNAME resolving and apex domain determination." +
				" For DNS-01 challenge verification, the authoritative DNS server is queried directly." +
				" Supported: host:port." +
				" The default is to use the system resolvers, or Google's DNS resolvers if the system's cannot be determined.",
		},
		&cli.IntFlag{
			Category: categoryDNS01Challenge,
			Name:     flgDNSTimeout,
			Usage:    "Set the DNS timeout value to a specific value in seconds. Used only when performing authoritative name server queries.",
			Value:    10,
		},
	}
}

func createStorageFlags() []cli.Flag {
	return []cli.Flag{
		createPathFlag(true),
		&cli.BoolFlag{
			Category: categoryStorage,
			Name:     flgPEM,
			Usage:    "Generate an additional .pem (base64) file by concatenating the .key and .crt files together.",
		},
		&cli.BoolFlag{
			Category: categoryStorage,
			Name:     flgPFX,
			Usage:    "Generate an additional .pfx (PKCS#12) file by concatenating the .key and .crt and issuer .crt files together.",
			Sources:  cli.EnvVars(toEnvName(flgPFX)),
		},
		&cli.StringFlag{
			Category: categoryStorage,
			Name:     flgPFXPass,
			Usage:    "The password used to encrypt the .pfx (PCKS#12) file.",
			Value:    pkcs12.DefaultPassword,
			Sources:  cli.EnvVars(toEnvName(flgPFXPass)),
		},
		&cli.StringFlag{
			Category: categoryStorage,
			Name:     flgPFXFormat,
			Usage:    "The encoding format to use when encrypting the .pfx (PCKS#12) file. Supported: RC2, DES, SHA256.",
			Value:    "RC2",
			Sources:  cli.EnvVars(toEnvName(flgPFXFormat)),
		},
	}
}

func createAccountFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    flgEmail,
			Aliases: []string{"m"},
			Sources: cli.EnvVars(toEnvName(flgEmail)),
			Usage:   "Email used for registration and recovery contact.",
		},
		&cli.StringFlag{
			Category: categoryStorage,
			Name:     flgAccountID,
			Aliases:  []string{"a"},
			Sources:  cli.EnvVars(toEnvName(flgAccountID)),
			Usage:    "Account identifier (The email is used if there is account ID is undefined).",
		},
		&cli.BoolFlag{
			Category: categoryEAB,
			Name:     flgEAB,
			Sources:  cli.EnvVars(toEnvName(flgEAB)),
			Usage:    fmt.Sprintf("Use External Account Binding for account registration. Requires %s and %s.", flgEABKID, flgEABHMAC),
		},
		&cli.StringFlag{
			Category: categoryEAB,
			Name:     flgEABKID,
			Sources:  cli.EnvVars(toEnvName(flgEABKID)),
			Usage:    "Key identifier for External Account Binding.",
		},
		&cli.StringFlag{
			Category: categoryEAB,
			Name:     flgEABHMAC,
			Sources:  cli.EnvVars(toEnvName(flgEABHMAC)),
			Usage:    "MAC key for External Account Binding. Should be in Base64 URL Encoding without padding format.",
		},
	}
}

func createObtainFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Category: categoryAdvanced,
			Name:     flgCSR,
			Aliases:  []string{"c"},
			Usage:    "Certificate signing request filename, if an external CSR is to be used.",
		},
		&cli.BoolFlag{
			Category: categoryAdvanced,
			Name:     flgNoBundle,
			Usage:    "Do not create a certificate bundle by adding the issuers certificate to the new certificate.",
		},
		&cli.BoolFlag{
			Category: categoryAdvanced,
			Name:     flgMustStaple,
			Usage: "Include the OCSP must staple TLS extension in the CSR and generated certificate." +
				" Only works if the CSR is generated by lego.",
		},
		&cli.TimestampFlag{
			Category: categoryAdvanced,
			Name:     flgNotBefore,
			Usage:    "Set the notBefore field in the certificate (RFC3339 format)",
			Config: cli.TimestampConfig{
				Layouts: []string{time.RFC3339},
			},
		},
		&cli.TimestampFlag{
			Category: categoryAdvanced,
			Name:     flgNotAfter,
			Usage:    "Set the notAfter field in the certificate (RFC3339 format)",
			Config: cli.TimestampConfig{
				Layouts: []string{time.RFC3339},
			},
		},
		&cli.StringFlag{
			Category: categoryAdvanced,
			Name:     flgPreferredChain,
			Usage: "If the CA offers multiple certificate chains, prefer the chain with an issuer matching this Subject Common Name." +
				" If no match, the default offered chain will be used.",
		},
		&cli.StringFlag{
			Category: categoryAdvanced,
			Name:     flgProfile,
			Usage:    "If the CA offers multiple certificate profiles (draft-ietf-acme-profiles), choose this one.",
		},
		&cli.StringFlag{
			Category: categoryAdvanced,
			Name:     flgAlwaysDeactivateAuthorizations,
			Usage:    "Force the authorizations to be relinquished even if the certificate request was successful.",
		},
	}
}

func createHookFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Category: categoryHooks,
			Name:     flgDeployHook,
			Usage:    "Define a hook. The hook is executed only when the certificates are effectively created/renewed.",
		},
		&cli.DurationFlag{
			Category: categoryHooks,
			Name:     flgDeployHookTimeout,
			Usage:    "Define the timeout for the hook execution.",
			Value:    2 * time.Minute,
		},
	}
}

func createRunFlags() []cli.Flag {
	flags := []cli.Flag{
		createDomainFlag(),
	}

	flags = append(flags, createAccountFlags()...)
	flags = append(flags, createACMEClientFlags()...)
	flags = append(flags, createStorageFlags()...)
	flags = append(flags, createAcceptFlag())
	flags = append(flags, createChallengesFlags()...)
	flags = append(flags, createObtainFlags()...)
	flags = append(flags, createHookFlags()...)

	flags = append(flags,
		&cli.StringFlag{
			Category: categoryAdvanced,
			Name:     flgPrivateKey,
			Usage:    "Path to a private key (in PEM encoding) for the certificate. By default, a private key is generated.",
		},
	)

	return flags
}

func createRenewFlags() []cli.Flag {
	flags := []cli.Flag{
		createDomainFlag(),
	}

	flags = append(flags, createAccountFlags()...)
	flags = append(flags, createACMEClientFlags()...)
	flags = append(flags, createStorageFlags()...)
	flags = append(flags, createChallengesFlags()...)
	flags = append(flags, createObtainFlags()...)
	flags = append(flags, createHookFlags()...)

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
			Category: categoryARI,
			Name:     flgARIDisable,
			Usage:    "Do not use the renewalInfo endpoint (RFC9773) to check if a certificate should be renewed.",
		},
		&cli.DurationFlag{
			Category: categoryARI,
			Name:     flgARIWaitToRenewDuration,
			Usage:    "The maximum duration you're willing to sleep for a renewal time returned by the renewalInfo endpoint.",
		},
		&cli.BoolFlag{
			Category: categoryAdvanced,
			Name:     flgReuseKey,
			Usage:    "Used to indicate you want to reuse your current private key for the new certificate.",
		},
		&cli.BoolFlag{
			Category: categoryAdvanced,
			Name:     flgNoRandomSleep,
			Usage: "Do not add a random sleep before the renewal." +
				" We do not recommend using this flag if you are doing your renewals in an automated way.",
		},
		&cli.BoolFlag{
			Category: categoryAdvanced,
			Name:     flgForceCertDomains,
			Usage:    "Check and ensure that the cert's domain list matches those passed in the domains argument.",
		},
	)

	return flags
}

func createRevokeFlags() []cli.Flag {
	flags := []cli.Flag{
		createPathFlag(false),
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
	}

	flags = append(flags, createDomainFlag())
	flags = append(flags, createAccountFlags()...)
	flags = append(flags, createACMEClientFlags()...)

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
		createPathFlag(false),
	}
}

func createRegisterFlags() []cli.Flag {
	flags := []cli.Flag{
		createPathFlag(true),
		createAcceptFlag(),
	}

	flags = append(flags, createACMEClientFlags()...)
	flags = append(flags, createAccountFlags()...)

	return flags
}

func createAcceptFlag() cli.Flag {
	return &cli.BoolFlag{
		Name:    flgAcceptTOS,
		Aliases: []string{"a"},
		Usage:   "By setting this flag to true you indicate that you accept the current Let's Encrypt terms of service.",
	}
}

func createDomainFlag() cli.Flag {
	return &cli.StringSliceFlag{
		Name:    flgDomains,
		Aliases: []string{"d"},
		Usage:   "Add a domain. For multiple domains either repeat the option or provide a comma-separated list.",
	}
}

func createPathFlag(forceCreation bool) cli.Flag {
	return &cli.StringFlag{
		Category: categoryStorage,
		Name:     flgPath,
		Sources:  cli.NewValueSourceChain(cli.EnvVar(toEnvName(flgPath)), &defaultPathValueSource{}),
		Usage:    "Directory to use for storing the data.",
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
