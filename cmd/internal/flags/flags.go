package flags

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/internal"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
	"software.sslmate.com/src/go-pkcs12"
)

func CreateRootFlags() []cli.Flag {
	flags := []cli.Flag{
		createConfigFlag(),
	}

	flags = append(flags, createLogFlags()...)

	return flags
}

func CreateRunFlags() []cli.Flag {
	flags := []cli.Flag{
		createDomainFlag(),
		createCertNameFlag(),
		createAcceptFlag(),
	}

	flags = append(flags, createAccountFlags()...)
	flags = append(flags, createACMEClientFlags()...)
	flags = append(flags, createStorageFlags()...)
	flags = append(flags, createChallengesFlags()...)
	flags = append(flags, createObtainFlags()...)
	flags = append(flags, createPreHookFlags()...)
	flags = append(flags, createDeployHookFlags()...)
	flags = append(flags, createPostHookFlags()...)
	flags = append(flags, CreateRenewFlags()...)

	flags = append(flags,
		&cli.StringFlag{
			Category: categoryAdvanced,
			Name:     FlgPrivateKey,
			Sources:  cli.EnvVars(toEnvName(FlgPrivateKey)),
			Usage:    "Path to a private key (in PEM encoding) for the certificate. By default, a private key is generated.",
		},
	)

	return flags
}

func CreateRenewFlags() []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Category: categoryRenew,
			Name:     FlgRenewDays,
			Sources:  cli.EnvVars(toEnvName(FlgRenewDays)),
			Usage: "The number of days left on a certificate to renew it." +
				"\n\tBy default, compute dynamically, based on the lifetime of the certificate(s), when to renew: use 1/3rd of the lifetime left, or 1/2 of the lifetime for short-lived certificates).",
		},
		&cli.BoolFlag{
			Category: categoryRenew,
			Name:     FlgRenewForce,
			Sources:  cli.EnvVars(toEnvName(FlgRenewForce)),
			Usage:    "Force the renewal of the certificate even if it is not due for renewal yet.",
		},
		&cli.BoolFlag{
			Category: categoryRenew,
			Name:     FlgARIDisable,
			Sources:  cli.EnvVars(toEnvName(FlgARIDisable)),
			Usage:    "(ARI) Do not use the renewalInfo endpoint (RFC9773) to check if a certificate should be renewed.",
		},
		&cli.DurationFlag{
			Category: categoryRenew,
			Name:     FlgARIWaitToRenewDuration,
			Sources:  cli.EnvVars(toEnvName(FlgARIWaitToRenewDuration)),
			Usage:    "(ARI) The maximum duration you're willing to sleep for a renewal time returned by the renewalInfo endpoint.",
		},
		&cli.BoolFlag{
			Category: categoryRenew,
			Name:     FlgReuseKey,
			Sources:  cli.EnvVars(toEnvName(FlgReuseKey)),
			Usage:    "Used to indicate you want to reuse the current certificate private key for the new certificate.",
		},
		&cli.BoolFlag{
			Category: categoryRenew,
			Name:     FlgNoRandomSleep,
			Sources:  cli.EnvVars(toEnvName(FlgNoRandomSleep)),
			Usage: "Do not add a random sleep before the renewal." +
				" We do not recommend using this flag if you are doing your renewals in an automated way.",
		},
		&cli.BoolFlag{
			Category: categoryRenew,
			Name:     FlgForceCertDomains,
			Sources:  cli.EnvVars(toEnvName(FlgForceCertDomains)),
			Usage:    "Check and ensure that the cert's domain list matches those passed in the domains argument.",
		},
	}
}

func CreateRevokeFlags() []cli.Flag {
	flags := []cli.Flag{
		CreatePathFlag(false),
		createCertNamesFlag(),
		&cli.BoolFlag{
			Name:    FlgKeep,
			Sources: cli.EnvVars(toEnvName(FlgKeep)),
			Usage:   "Keep the certificates after the revocation instead of archiving them.",
		},
		&cli.UintFlag{
			Name:    FlgReason,
			Sources: cli.EnvVars(toEnvName(FlgReason)),
			Usage: "Identifies the reason for the certificate revocation." +
				" See https://www.rfc-editor.org/rfc/rfc5280.html#section-5.3.1." +
				"\n\tValid values are:" +
				" 0 (unspecified), 1 (keyCompromise), 2 (cACompromise), 3 (affiliationChanged)," +
				" 4 (superseded), 5 (cessationOfOperation), 6 (certificateHold), 8 (removeFromCRL)," +
				" 9 (privilegeWithdrawn), or 10 (aACompromise).",
			Value: acme.CRLReasonUnspecified,
		},
		createConfigFlag(),
	}

	flags = append(flags, createAccountFlags()...)
	flags = append(flags, createACMEClientFlags()...)

	return flags
}

func CreateRegisterFlags() []cli.Flag {
	flags := []cli.Flag{
		CreatePathFlag(true),
		createAcceptFlag(),
	}

	flags = append(flags, createACMEClientFlags()...)
	flags = append(flags, createAccountFlags()...)

	return flags
}

func CreateRecoverFlags() []cli.Flag {
	flags := []cli.Flag{
		CreatePathFlag(true),
		&cli.StringFlag{
			Name:     FlgPrivateKey,
			Sources:  cli.EnvVars(toEnvName(FlgPrivateKey)),
			Usage:    "Path to the account private key (PEM encoded).",
			Required: true,
		},
	}

	flags = append(flags, createACMEClientFlags()...)
	flags = append(flags, createAccountFlags()...)

	return flags
}

func CreateKeyRolloverFlags() []cli.Flag {
	flags := []cli.Flag{
		CreatePathFlag(true),
		&cli.StringFlag{
			Name:    FlgPrivateKey,
			Sources: cli.EnvVars(toEnvName(FlgPrivateKey)),
			Usage:   "Path to the new account private key (PEM encoded). If not specified, the private key will be generated.",
		},
		createKeyTypeFlag("Key type to use for the new private key of the account."),
	}

	flags = append(flags, createACMEClientFlags()...)
	flags = append(flags, createAccountFlags()...)

	return flags
}

func CreateListFlags() []cli.Flag {
	return []cli.Flag{
		CreatePathFlag(false),
		&cli.BoolFlag{
			Name:  FlgFormatJSON,
			Usage: "Format the output as JSON.",
		},
	}
}

func CreateMigrateFlags() []cli.Flag {
	return []cli.Flag{
		CreatePathFlag(false),
		&cli.BoolFlag{
			Name:    FlgAccountOnly,
			Sources: cli.EnvVars(toEnvName(FlgAccountOnly)),
			Usage:   "Only migrate accounts.",
			Value:   false,
		},
	}
}

func createACMEClientFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			// NOTE(ldez): if Required is true, then the default value is not display in the help.
			Name:    FlgServer,
			Aliases: []string{flgAliasServer},
			Sources: cli.EnvVars(toEnvName(FlgServer)),
			Usage: fmt.Sprintf("CA (ACME server). It can be either a URL or a shortcode."+
				"\n\t(available shortcodes: %s)", strings.Join(lego.GetAllCodes(), ", ")),
			Value: lego.DirectoryURLLetsEncrypt,
			Action: func(ctx context.Context, cmd *cli.Command, s string) error {
				directoryURL, err := lego.GetDirectoryURL(s)
				if err != nil {
					log.Debug("Server shortcode not found. Use the value as URL.", slog.String("value", s), log.ErrorAttr(err))

					directoryURL = s
				}

				return cmd.Set(FlgServer, directoryURL)
			},
		},
		&cli.BoolFlag{
			Category: categoryAdvanced,
			Name:     FlgEnableCommonName,
			Sources:  cli.EnvVars(toEnvName(FlgEnableCommonName)),
			Usage:    "Enable the use of the common name. (Not recommended)",
		},
		createKeyTypeFlag("Key type to use for private keys."),
		&cli.IntFlag{
			Category: categoryACMEClient,
			Name:     FlgHTTPTimeout,
			Sources:  cli.EnvVars(toEnvName(FlgHTTPTimeout)),
			Usage:    "Set the HTTP timeout value to a specific value in seconds.",
		},
		&cli.BoolFlag{
			Category: categoryACMEClient,
			Name:     FlgTLSSkipVerify,
			Sources:  cli.EnvVars(toEnvName(FlgTLSSkipVerify)),
			Usage:    "Skip the TLS verification of the ACME server.",
		},
		&cli.IntFlag{
			Category: categoryAdvanced,
			Name:     FlgCertTimeout,
			Sources:  cli.EnvVars(toEnvName(FlgCertTimeout)),
			Usage:    "Set the certificate timeout value to a specific value in seconds. Only used when obtaining certificates.",
			Value:    30,
		},
		&cli.IntFlag{
			Category: categoryACMEClient,
			Name:     FlgOverallRequestLimit,
			Sources:  cli.EnvVars(toEnvName(FlgOverallRequestLimit)),
			Usage:    "ACME overall requests limit.",
			Value:    certificate.DefaultOverallRequestLimit,
		},
		&cli.StringFlag{
			Category: categoryACMEClient,
			Name:     FlgUserAgent,
			Sources:  cli.EnvVars(toEnvName(FlgUserAgent)),
			Usage:    "Add to the user-agent sent to the CA to identify an application embedding lego-cli",
		},
	}
}

func createChallengesFlags() []cli.Flag {
	flags := []cli.Flag{
		CreateEnvFileFlag(),
	}

	flags = append(flags, createHTTPChallengeFlags()...)
	flags = append(flags, createTLSChallengeFlags()...)
	flags = append(flags, createDNSChallengeFlags()...)
	flags = append(flags, createDNSPersistChallengeFlags()...)
	flags = append(flags, createNetworkStackFlags()...)

	return flags
}

func createNetworkStackFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Category: categoryAdvanced,
			Name:     FlgIPv4Only,
			Aliases:  []string{flgAliasIPv4Only},
			Sources:  cli.EnvVars(toEnvName(FlgIPv4Only)),
			Usage:    "Use IPv4 only.",
		},
		&cli.BoolFlag{
			Category: categoryAdvanced,
			Name:     FlgIPv6Only,
			Aliases:  []string{flgAliasIPv6Only},
			Sources:  cli.EnvVars(toEnvName(FlgIPv6Only)),
			Usage:    "Use IPv6 only.",
		},
	}
}

func createHTTPChallengeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Category: categoryHTTP01Challenge,
			Name:     FlgHTTP,
			Sources:  cli.EnvVars(toEnvName(FlgHTTP)),
			Usage:    "Use the HTTP-01 challenge to solve challenges. Can be mixed with other types of challenges.",
		},
		&cli.StringFlag{
			Category: categoryHTTP01Challenge,
			Name:     FlgHTTPAddress,
			Sources:  cli.EnvVars(toEnvName(FlgHTTPAddress)),
			Usage:    "Set the address to use for HTTP-01 based challenges to listen on. Supported: interface:port or :port.",
			Value:    ":80",
		},
		&cli.DurationFlag{
			Category: categoryHTTP01Challenge,
			Name:     FlgHTTPDelay,
			Sources:  cli.EnvVars(toEnvName(FlgHTTPDelay)),
			Usage:    "Delay between the starts of the HTTP server (use for HTTP-01 based challenges) and the validation of the challenge.",
			Value:    0,
		},
		&cli.StringFlag{
			Category: categoryHTTP01Challenge,
			Name:     FlgHTTPProxyHeader,
			Sources:  cli.EnvVars(toEnvName(FlgHTTPProxyHeader)),
			Usage:    "Validate against this HTTP header when solving HTTP-01 based challenges behind a reverse proxy.",
			Value:    "Host",
		},
		&cli.StringFlag{
			Category: categoryHTTP01Challenge,
			Name:     FlgHTTPWebroot,
			Sources:  cli.EnvVars(toEnvName(FlgHTTPWebroot)),
			Usage: "Set the webroot folder to use for HTTP-01 based challenges to write directly to the .well-known/acme-challenge file." +
				" This disables the built-in server and expects the given directory to be publicly served with access to .well-known/acme-challenge",
		},
		&cli.StringSliceFlag{
			Category: categoryHTTP01Challenge,
			Name:     FlgHTTPMemcachedHost,
			Sources:  cli.EnvVars(toEnvName(FlgHTTPMemcachedHost)),
			Usage:    "Set the memcached host(s) to use for HTTP-01 based challenges. Challenges will be written to all specified hosts.",
		},
		&cli.StringFlag{
			Category: categoryHTTP01Challenge,
			Name:     FlgHTTPS3Bucket,
			Sources:  cli.EnvVars(toEnvName(FlgHTTPS3Bucket)),
			Usage:    "Set the S3 bucket name to use for HTTP-01 based challenges. Challenges will be written to the S3 bucket.",
		},
	}
}

func createTLSChallengeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Category: categoryTLSALPN01Challenge,
			Name:     FlgTLS,
			Sources:  cli.EnvVars(toEnvName(FlgTLS)),
			Usage:    "Use the TLS-ALPN-01 challenge to solve challenges. Can be mixed with other types of challenges.",
		},
		&cli.StringFlag{
			Category: categoryTLSALPN01Challenge,
			Name:     FlgTLSAddress,
			Sources:  cli.EnvVars(toEnvName(FlgTLSAddress)),
			Usage:    "Set the address to use for TLS-ALPN-01 based challenges to listen on. Supported: interface:port or :port.",
			Value:    ":443",
		},
		&cli.DurationFlag{
			Category: categoryTLSALPN01Challenge,
			Name:     FlgTLSDelay,
			Sources:  cli.EnvVars(toEnvName(FlgTLSDelay)),
			Usage:    "Delay between the start of the TLS listener (use for TLSALPN-01 based challenges) and the validation of the challenge.",
			Value:    0,
		},
	}
}

func createDNSChallengeFlags() []cli.Flag {
	flags := []cli.Flag{
		&cli.StringFlag{
			Category: categoryDNS01Challenge,
			Name:     FlgDNS,
			Sources:  cli.EnvVars(toEnvName(FlgDNS)),
			Usage:    "Solve a DNS-01 challenge using the specified provider. Can be mixed with other types of challenges. Run 'lego dnshelp' for help on usage.",
		},
		&cli.StringSliceFlag{
			Category: categoryDNS01Challenge,
			Name:     FlgDNSResolvers,
			Sources:  cli.EnvVars(toEnvName(FlgDNSResolvers)),
			Usage: "Set the nameservers to use for performing (recursive) CNAME resolving and apex domain determination." +
				" For DNS-01 challenge verification, the authoritative DNS server is queried directly." +
				" Supported: host:port." +
				" The default is to use the system nameservers, or Cloudflare's nameservers if the system's cannot be determined.",
		},
		&cli.IntFlag{
			Category: categoryDNS01Challenge,
			Name:     FlgDNSTimeout,
			Sources:  cli.EnvVars(toEnvName(FlgDNSTimeout)),
			Usage:    "Set the DNS timeout value to a specific value in seconds. Used only when performing authoritative name server queries.",
			Value:    10,
		},
	}

	flags = append(flags,
		createDNSPropagationFlags(
			categoryDNS01Challenge,
			FlgDNSPropagationWait,
			FlgDNSPropagationDisableANS,
			FlgDNSPropagationDisableRNS,
		)...,
	)

	return flags
}

func createDNSPersistChallengeFlags() []cli.Flag {
	flags := []cli.Flag{
		&cli.BoolFlag{
			Category: categoryDNSPersist01Challenge,
			Name:     FlgDNSPersist,
			Sources:  cli.EnvVars(toEnvName(FlgDNSPersist)),
			Usage:    "Use the DNS-PERSIST-01 challenge to solve challenges. Manual verification only. Can be mixed with other types of challenges.",
		},
		&cli.StringFlag{
			Category: categoryDNSPersist01Challenge,
			Name:     FlgDNSPersistIssuerDomainName,
			Sources:  cli.EnvVars(toEnvName(FlgDNSPersistIssuerDomainName)),
			Usage:    "Override the issuer-domain-name to use for DNS-PERSIST-01 when multiple are offered. Must be offered by the challenge.",
		},
		&cli.TimestampFlag{
			Name:     FlgDNSPersistPersistUntil,
			Category: categoryDNSPersist01Challenge,
			Usage:    "Set the optional persistUntil for DNS-PERSIST-01 records as an RFC3339 timestamp (for example, 2026-03-01T00:00:00Z).",
			Sources:  cli.EnvVars(toEnvName(FlgDNSPersistPersistUntil)),
			Config: cli.TimestampConfig{
				Layouts: []string{time.RFC3339},
			},
		},
		&cli.StringSliceFlag{
			Category: categoryDNSPersist01Challenge,
			Name:     FlgDNSPersistResolvers,
			Sources:  cli.EnvVars(toEnvName(FlgDNSPersistResolvers)),
			Usage: "Set the resolvers to use for DNS-PERSIST-01 TXT lookups." +
				" Supported: host:port." +
				" The default is to use the system nameservers, or Cloudflare's nameservers if the system's cannot be determined.",
		},
		&cli.IntFlag{
			Category: categoryDNSPersist01Challenge,
			Name:     FlgDNSPersistTimeout,
			Sources:  cli.EnvVars(toEnvName(FlgDNSPersistTimeout)),
			Usage:    "Set the DNS timeout value to a specific value in seconds. Used for DNS-PERSIST-01 lookups.",
		},
	}

	flags = append(flags,
		createDNSPropagationFlags(
			categoryDNSPersist01Challenge,
			FlgDNSPersistPropagationWait,
			FlgDNSPersistPropagationDisableANS,
			FlgDNSPersistPropagationDisableRNS,
		)...,
	)

	return flags
}

func createDNSPropagationFlags(category, flgWait, flgANS, flgRNS string) []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Category: category,
			Name:     flgANS,
			Sources:  cli.EnvVars(toEnvName(flgANS)),
			Usage:    "By setting this flag to true, disables the need to await propagation of the TXT record to all authoritative name servers.",
		},
		&cli.BoolFlag{
			Category: category,
			Name:     flgRNS,
			Sources:  cli.EnvVars(toEnvName(flgRNS)),
			Usage:    "By setting this flag to true, disables the need to await propagation of the TXT record to all recursive name servers (aka resolvers).",
		},
		&cli.DurationFlag{
			Category: category,
			Name:     flgWait,
			Sources:  cli.EnvVars(toEnvName(flgWait)),
			Usage:    "By setting this flag, disables all the propagation checks of the TXT record and uses a wait duration instead.",
			Validator: func(d time.Duration) error {
				if d < 0 {
					return errors.New("it cannot be negative")
				}

				return nil
			},
		},
	}
}

func createStorageFlags() []cli.Flag {
	return []cli.Flag{
		CreatePathFlag(true),
		&cli.BoolFlag{
			Category: categoryStorage,
			Name:     FlgPEM,
			Sources:  cli.EnvVars(toEnvName(FlgPEM)),
			Usage:    "Generate an additional .pem (base64) file by concatenating the .key and .crt files together.",
		},
		&cli.BoolFlag{
			Category: categoryStorage,
			Name:     FlgPFX,
			Sources:  cli.EnvVars(toEnvName(FlgPFX)),
			Usage:    "Generate an additional .pfx (PKCS#12) file by concatenating the .key and .crt and issuer .crt files together.",
		},
		&cli.StringFlag{
			Category: categoryStorage,
			Name:     FlgPFXPass,
			Sources:  cli.EnvVars(toEnvName(FlgPFXPass)),
			Usage:    "The password used to encrypt the .pfx (PCKS#12) file.",
			Value:    pkcs12.DefaultPassword,
		},
		&cli.StringFlag{
			Category: categoryStorage,
			Name:     FlgPFXFormat,
			Sources:  cli.EnvVars(toEnvName(FlgPFXFormat)),
			Usage: fmt.Sprintf("The encoding format to use when encrypting the .pfx (PCKS#12) file. Supported: %s.",
				strings.Join(certcrypto.AllPKCS12Formats(), ", "),
			),
			Value: certcrypto.PKCS12LegacyRC2,
			Validator: func(s string) error {
				if !certcrypto.IsPKCS12Supported(s) {
					return fmt.Errorf("invalid PFX format: %s", s)
				}

				return nil
			},
		},
	}
}

func createAccountFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    FlgEmail,
			Aliases: []string{flgAliasEmail},
			Sources: cli.EnvVars(toEnvName(FlgEmail)),
			Usage:   "Email used for registration and recovery contact.",
		},
		&cli.StringFlag{
			Category: categoryStorage,
			Name:     FlgAccountID,
			Sources:  cli.EnvVars(toEnvName(FlgAccountID)),
			Usage:    "Account identifier (The email is used if the account ID is undefined).",
		},
		&cli.BoolFlag{
			Category: categoryEAB,
			Name:     FlgEAB,
			Sources:  cli.EnvVars(toEnvName(FlgEAB)),
			Usage:    fmt.Sprintf("Use External Account Binding for account registration. Requires %s and %s.", FlgEABKID, FlgEABHMAC),
		},
		&cli.StringFlag{
			Category: categoryEAB,
			Name:     FlgEABKID,
			Sources:  cli.EnvVars(toEnvName(FlgEABKID)),
			Usage:    "Key identifier for External Account Binding.",
		},
		&cli.StringFlag{
			Category: categoryEAB,
			Name:     FlgEABHMAC,
			Sources:  cli.EnvVars(toEnvName(FlgEABHMAC)),
			Usage:    "MAC key for External Account Binding. Should be in Base64 URL Encoding without padding format.",
		},
	}
}

func createObtainFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Category: categoryAdvanced,
			Name:     FlgCSR,
			Sources:  cli.EnvVars(toEnvName(FlgCSR)),
			Usage:    "Certificate signing request filename, if an external CSR is to be used.",
		},
		&cli.BoolFlag{
			Category: categoryAdvanced,
			Name:     FlgNoBundle,
			Sources:  cli.EnvVars(toEnvName(FlgNoBundle)),
			Usage:    "Do not create a certificate bundle by adding the issuers certificate to the new certificate.",
		},
		&cli.BoolFlag{
			Category: categoryAdvanced,
			Name:     FlgMustStaple,
			Sources:  cli.EnvVars(toEnvName(FlgMustStaple)),
			Usage: "Include the OCSP must staple TLS extension in the CSR and generated certificate." +
				" Only works if the CSR is generated by lego.",
		},
		&cli.TimestampFlag{
			Category: categoryAdvanced,
			Name:     FlgNotBefore,
			Sources:  cli.EnvVars(toEnvName(FlgNotBefore)),
			Usage:    "Set the notBefore field in the certificate (RFC3339 format)",
			Config: cli.TimestampConfig{
				Layouts: []string{time.RFC3339},
			},
		},
		&cli.TimestampFlag{
			Category: categoryAdvanced,
			Name:     FlgNotAfter,
			Sources:  cli.EnvVars(toEnvName(FlgNotAfter)),
			Usage:    "Set the notAfter field in the certificate (RFC3339 format)",
			Config: cli.TimestampConfig{
				Layouts: []string{time.RFC3339},
			},
		},
		&cli.StringFlag{
			Category: categoryAdvanced,
			Name:     FlgPreferredChain,
			Sources:  cli.EnvVars(toEnvName(FlgPreferredChain)),
			Usage: "If the CA offers multiple certificate chains, prefer the chain with an issuer matching this Subject Common Name." +
				" If no match, the default offered chain will be used.",
		},
		&cli.StringFlag{
			Category: categoryAdvanced,
			Name:     FlgProfile,
			Sources:  cli.EnvVars(toEnvName(FlgProfile)),
			Usage:    "If the CA offers multiple certificate profiles (draft-ietf-acme-profiles), choose this one.",
		},
		&cli.StringFlag{
			Category: categoryAdvanced,
			Name:     FlgAlwaysDeactivateAuthorizations,
			Sources:  cli.EnvVars(toEnvName(FlgAlwaysDeactivateAuthorizations)),
			Usage:    "Force the authorizations to be relinquished even if the certificate request was successful.",
		},
	}
}

func createPreHookFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Category: categoryHooks,
			Name:     FlgPreHook,
			Sources:  cli.EnvVars(toEnvName(FlgPreHook)),
			Usage:    "Define a pre-hook. This hook runs, before the creation or the renewal, in cases where a certificate will be effectively created/renewed.",
		},
		&cli.DurationFlag{
			Category: categoryHooks,
			Name:     FlgPreHookTimeout,
			Sources:  cli.EnvVars(toEnvName(FlgPreHookTimeout)),
			Usage:    "Define the timeout for the pre-hook execution.",
			Value:    2 * time.Minute,
		},
	}
}

func createDeployHookFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Category: categoryHooks,
			Name:     FlgDeployHook,
			Sources:  cli.EnvVars(toEnvName(FlgDeployHook)),
			Usage:    "Define a hook. The hook runs, after the creation or the renewal, in cases where a certificate is successfully created/renewed.",
		},
		&cli.DurationFlag{
			Category: categoryHooks,
			Name:     FlgDeployHookTimeout,
			Sources:  cli.EnvVars(toEnvName(FlgDeployHookTimeout)),
			Usage:    "Define the timeout for the deploy-hook execution.",
			Value:    2 * time.Minute,
		},
	}
}

func createPostHookFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Category: categoryHooks,
			Name:     FlgPostHook,
			Sources:  cli.EnvVars(toEnvName(FlgPostHook)),
			Usage:    "Define a post-hook. This hook runs, after the creation or the renewal, in cases where a certificate is created/renewed, regardless of whether any errors occurred.",
		},
		&cli.DurationFlag{
			Category: categoryHooks,
			Name:     FlgPostHookTimeout,
			Sources:  cli.EnvVars(toEnvName(FlgPostHookTimeout)),
			Usage:    "Define the timeout for the post-hook execution.",
			Value:    2 * time.Minute,
		},
	}
}

func createConfigFlag() cli.Flag {
	return &cli.StringFlag{
		Category: categoryConfiguration,
		Name:     FlgConfig,
		Sources:  cli.EnvVars(toEnvName(FlgConfig)),
		Usage:    "Path to the configuration file.",
		Local:    true,
	}
}

func createAcceptFlag() cli.Flag {
	return &cli.BoolFlag{
		Name:    FlgAcceptTOS,
		Aliases: []string{flgAliasAcceptTOS},
		Sources: cli.EnvVars(toEnvName(FlgAcceptTOS)),
		Usage:   "By setting this flag to true, you indicate that you accept the current CA terms of service.",
	}
}

func createDomainFlag() cli.Flag {
	return &cli.StringSliceFlag{
		Name:    FlgDomains,
		Aliases: []string{flgAliasDomains},
		Sources: cli.EnvVars(toEnvName(FlgDomains)),
		Usage:   "Add a domain. For multiple domains either repeat the option or provide a comma-separated list.",
	}
}

func createCertNameFlag() cli.Flag {
	return &cli.StringFlag{
		Category: categoryStorage,
		Name:     FlgCertName,
		Aliases:  []string{flgAliasCertName},
		Sources:  cli.EnvVars(toEnvName(FlgCertName)),
		Usage:    "The certificate ID/Name, used to store and retrieve a certificate. By default, it uses the first domain name.",
	}
}

func createCertNamesFlag() cli.Flag {
	return &cli.StringSliceFlag{
		Name:    FlgCertName,
		Aliases: []string{flgAliasCertName},
		Sources: cli.EnvVars(toEnvName(FlgCertName)),
		Usage:   "The certificate IDs/Names, used to retrieve the certificates.",
	}
}

func createKeyTypeFlag(desc string) *cli.StringFlag {
	return &cli.StringFlag{
		Name:    FlgKeyType,
		Aliases: []string{flgAliasKeyType},
		Sources: cli.EnvVars(toEnvName(FlgKeyType)),
		Value:   string(certcrypto.EC256),
		Usage:   fmt.Sprintf("%s Supported: %s.", desc, internal.Join(certcrypto.AllKeyTypes(), ", ")),
	}
}

func CreatePathFlag(forceCreation bool) cli.Flag {
	return &cli.StringFlag{
		Category: categoryStorage,
		Name:     FlgPath,
		Sources:  cli.NewValueSourceChain(cli.EnvVar(toEnvName(FlgPath)), &defaultPathValueSource{}),
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

func CreateEnvFileFlag() cli.Flag {
	return &cli.StringFlag{
		Category: categoryStorage,
		Name:     FlgEnvFile,
		Sources:  cli.EnvVars(toEnvName(FlgEnvFile)),
		Usage:    "The path to the dotenv file.",
	}
}

func createLogFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Category: categoryLogs,
			Name:     FlgLogLevel,
			Sources:  cli.EnvVars(toEnvName(FlgLogLevel)),
			Usage:    "Set the logging level. Supported values: 'debug', 'info', 'warn', 'error'.",
			Value:    "info",
		},
		&cli.StringFlag{
			Category: categoryLogs,
			Name:     FlgLogFormat,
			Sources:  cli.EnvVars(toEnvName(FlgLogFormat)),
			Usage:    "Set the logging format. Supported values: 'colored', 'text', 'json'.",
			Value:    "colored",
		},
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
