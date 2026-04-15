package flags

import (
	"strings"
	"unicode"
)

const (
	categoryHTTP01Challenge       = "Flags related to the HTTP-01 challenge:"
	categoryTLSALPN01Challenge    = "Flags related to the TLS-ALPN-01 challenge:"
	categoryDNS01Challenge        = "Flags related to the DNS-01 challenge:"
	categoryDNSPersist01Challenge = "Flags related to the DNS-PERSIST-01 challenge:"
	categoryStorage               = "Flags related to the storage:"
	categoryHooks                 = "Flags related to hooks:"
	categoryEAB                   = "Flags related to External Account Binding:"
	categoryACMEClient            = "Flags related to the ACME client:"
	categoryAdvanced              = "Flags related to advanced options:"
	categoryRenew                 = "Flags related to certificate renewal:"
	categoryLogs                  = "Flags related to logs:"
	categoryConfiguration         = "Flags related to the configuration file:"
)

// Flag aliases (short-codes).
const (
	flgAliasAcceptTOS = "a"
	flgAliasCertName  = "c"
	flgAliasDomains   = "d"
	flgAliasEmail     = "m"
	flgAliasIPv4Only  = "4"
	flgAliasIPv6Only  = "6"
	flgAliasKeyType   = "k"
	flgAliasServer    = "s"
)

// Flag names related to the account.
const (
	FlgAcceptTOS = "accept-tos"
	FlgEmail     = "email"
	FlgKeyType   = "key-type"
	FlgAccountID = "account-id"
	FlgEAB       = "eab"
	FlgEABKID    = "eab.kid"
	FlgEABHMAC   = "eab.hmac"
)

// Flag names related to Obtain certificates.
const (
	FlgDomains                        = "domains"
	FlgCSR                            = "csr"
	FlgNoBundle                       = "no-bundle"
	FlgMustStaple                     = "must-staple"
	FlgNotBefore                      = "not-before"
	FlgNotAfter                       = "not-after"
	FlgPreferredChain                 = "preferred-chain"
	FlgProfile                        = "profile"
	FlgAlwaysDeactivateAuthorizations = "always-deactivate-authorizations"
)

// Flag names related to the storage.
const (
	FlgPath      = "path"
	FlgEnvFile   = "env-file"
	FlgPEM       = "pem"
	FlgPFX       = "pfx"
	FlgPFXPass   = "pfx.password"
	FlgPFXFormat = "pfx.format"
)

// Flag names related to the ACME client.
const (
	FlgServer              = "server"
	FlgEnableCommonName    = "enable-cn"
	FlgHTTPTimeout         = "http-timeout"
	FlgTLSSkipVerify       = "tls-skip-verify"
	FlgOverallRequestLimit = "overall-request-limit"
	FlgUserAgent           = "user-agent"
)

// Flag names related to certificates.
const (
	FlgCertTimeout = "cert.timeout"
	FlgCertName    = "cert.name"
)

// Flag names related to the network stack.
const (
	FlgIPv4Only = "ipv4only"
	FlgIPv6Only = "ipv6only"
)

// Flag names related to HTTP-01 challenge.
const (
	FlgHTTP              = "http"
	FlgHTTPAddress       = "http.address"
	FlgHTTPDelay         = "http.delay"
	FlgHTTPProxyHeader   = "http.proxy-header"
	FlgHTTPWebroot       = "http.webroot"
	FlgHTTPMemcachedHost = "http.memcached-host"
	FlgHTTPS3Bucket      = "http.s3-bucket"
)

// Flag names related to TLS-ALPN-01 challenge.
const (
	FlgTLS        = "tls"
	FlgTLSAddress = "tls.address"
	FlgTLSDelay   = "tls.delay"
)

// Flag names related to DNS-01 challenge.
const (
	FlgDNS                      = "dns"
	FlgDNSPropagationWait       = "dns.propagation.wait"
	FlgDNSPropagationDisableANS = "dns.propagation.disable-ans"
	FlgDNSPropagationDisableRNS = "dns.propagation.disable-rns"
	FlgDNSResolvers             = "dns.resolvers"
	FlgDNSTimeout               = "dns.timeout"
)

// Flag names related to the DNS-PERSIST-01 challenge.
const (
	FlgDNSPersist                      = "dns-persist"
	FlgDNSPersistIssuerDomainName      = "dns-persist.issuer-domain-name"
	FlgDNSPersistPersistUntil          = "dns-persist.persist-until"
	FlgDNSPersistPropagationWait       = "dns-persist.propagation.wait"
	FlgDNSPersistPropagationDisableANS = "dns-persist.propagation.disable-ans"
	FlgDNSPersistPropagationDisableRNS = "dns-persist.propagation.disable-rns"
	FlgDNSPersistResolvers             = "dns-persist.resolvers"
	FlgDNSPersistTimeout               = "dns-persist.timeout"
)

// Flags names related to hooks.
const (
	FlgPreHook           = "pre-hook"
	FlgPreHookTimeout    = "pre-hook-timeout"
	FlgDeployHook        = "deploy-hook"
	FlgDeployHookTimeout = "deploy-hook-timeout"
	FlgPostHook          = "post-hook"
	FlgPostHookTimeout   = "post-hook-timeout"
)

// Flag names related to logs.
const (
	FlgLogLevel  = "log.level"
	FlgLogFormat = "log.format"
)

// Flag names related to the configuration file.
const (
	FlgConfig = "config"
)

// Flag names related to the specific run command.
const (
	FlgPrivateKey = "private-key"
)

// Flag names related to the specific renew command.
const (
	FlgRenewDays              = "renew-days"
	FlgRenewForce             = "renew-force"
	FlgARIDisable             = "ari-disable"
	FlgARIWaitToRenewDuration = "ari-wait-to-renew-duration"
	FlgReuseKey               = "reuse-key"
	FlgNoRandomSleep          = "no-random-sleep"
	FlgForceCertDomains       = "force-cert-domains"
)

// Flag names related to the specific revoke command.
const (
	FlgKeep   = "keep"
	FlgReason = "reason"
)

// Flag names related to the list commands.
const (
	FlgFormatJSON = "json"
)

// Flag names related to the migrate command.
const (
	FlgAccountOnly = "account-only"
)

func toEnvName(flg string) string {
	fields := strings.FieldsFunc(flg, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	return "LEGO_" + strings.ToUpper(strings.Join(fields, "_"))
}
