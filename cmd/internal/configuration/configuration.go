package configuration

import (
	"time"

	"github.com/go-acme/lego/v5/certcrypto"
)

type Configuration struct {
	Storage      string                  `yaml:"storage,omitempty"`
	NetworkStack string                  `yaml:"networkStack,omitempty"`
	UserAgent    string                  `yaml:"userAgent,omitempty"`
	Servers      map[string]*Server      `yaml:"servers,omitempty"`
	Accounts     map[string]*Account     `yaml:"accounts,omitempty"`
	Challenges   map[string]*Challenge   `yaml:"challenges,omitempty"`
	Certificates map[string]*Certificate `yaml:"certificates,omitempty"`
	Hooks        *Hooks                  `yaml:"hooks,omitempty"`
	Log          *Log                    `yaml:"log,omitempty"`
}

type Server struct {
	URL                 string `yaml:"url,omitempty"`
	TLSSkipVerify       bool   `yaml:"tlsSkipVerify,omitempty"`
	OverallRequestLimit int    `yaml:"overallRequestLimit,omitempty"`
	HTTPTimeout         int    `yaml:"httpTimeout,omitempty"`
	CertTimeout         int    `yaml:"certTimeout,omitempty"`
}

type Account struct {
	ID string `yaml:"-"`

	Server                 string                  `yaml:"server,omitempty"`
	Email                  string                  `yaml:"email,omitempty"`
	KeyType                certcrypto.KeyType      `yaml:"keyType,omitempty"`
	AcceptsTermsOfService  bool                    `yaml:"acceptsTermsOfService,omitempty"`
	ExternalAccountBinding *ExternalAccountBinding `yaml:"eab,omitempty"`
}

type ExternalAccountBinding struct {
	KID     string `yaml:"kid,omitempty"`
	HmacKey string `yaml:"hmacKey,omitempty"`
}

type Challenge struct {
	ID string `yaml:"-"`

	HTTP       *HTTPChallenge       `yaml:"http,omitempty"`
	TLS        *TLSChallenge        `yaml:"tls,omitempty"`
	DNS        *DNSChallenge        `yaml:"dns,omitempty"`
	DNSPersist *DNSPersistChallenge `yaml:"dnsPersist,omitempty"`
}

type HTTPChallenge struct {
	Address        string        `yaml:"address,omitempty"`
	Delay          time.Duration `yaml:"delay,omitempty"`
	ProxyHeader    string        `yaml:"proxyHeader,omitempty"`
	Webroot        string        `yaml:"webroot,omitempty"`
	MemcachedHosts []string      `yaml:"memcachedHosts,omitempty"`
	S3Bucket       string        `yaml:"s3Bucket,omitempty"`
}

type TLSChallenge struct {
	Address string        `yaml:"address,omitempty"`
	Delay   time.Duration `yaml:"delay,omitempty"`
}

type DNSChallenge struct {
	Provider    string       `yaml:"provider,omitempty"`
	Propagation *Propagation `yaml:"propagation,omitempty"`
	DNSTimeout  int          `yaml:"dnsTimeout,omitempty"`
	Resolvers   []string     `yaml:"resolvers,omitempty"`
	EnvFile     string       `yaml:"envFile,omitempty"`
}

type DNSPersistChallenge struct {
	IssuerDomainName string       `yaml:"issuerDomainName,omitempty"`
	PersistUntil     time.Time    `yaml:"persistUntil,omitempty"`
	Propagation      *Propagation `yaml:"propagation,omitempty"`
	DNSTimeout       int          `yaml:"dnsTimeout,omitempty"`
	Resolvers        []string     `yaml:"resolvers,omitempty"`
}

type Propagation struct {
	DisableAuthoritativeNameservers bool          `yaml:"disableAuthoritativeNameservers,omitempty"`
	DisableRecursiveNameservers     bool          `yaml:"disableRecursiveNameservers,omitempty"`
	Wait                            time.Duration `yaml:"wait,omitempty"`
}

type Certificate struct {
	ID string `yaml:"-"`

	Domains []string `yaml:"domains,omitempty"`
	CSR     string   `yaml:"csr,omitempty"`

	KeyType certcrypto.KeyType `yaml:"keyType,omitempty"`

	Challenge string `yaml:"challenge,omitempty"`
	Account   string `yaml:"account,omitempty"`

	EnableCommonName bool `yaml:"enableCommonName,omitempty"`

	PreferredChain string `yaml:"preferredChain,omitempty"`
	Profile        string `yaml:"profile,omitempty"`

	NotBefore  time.Time `yaml:"notBefore,omitempty"`
	NotAfter   time.Time `yaml:"notAfter,omitempty"`
	NoBundle   bool      `yaml:"noBundle,omitempty"`
	MustStaple bool      `yaml:"mustStaple,omitempty"`

	AlwaysDeactivateAuthorizations bool `yaml:"alwaysDeactivateAuthorizations,omitempty"`

	Renew *RenewConfiguration `yaml:"renew,omitempty"`

	PFX *PFX `yaml:"pfx,omitempty"`
}

type RenewConfiguration struct {
	ARI *ARIConfiguration `yaml:"ari,omitempty"`

	Days int `yaml:"days,omitempty"`

	ReuseKey bool `yaml:"reuseKey,omitempty"`

	DisableRandomSleep bool `yaml:"disableRandomSleep,omitempty"`
}

// ARIConfiguration is the configuration for the Automatic Renewal Integration.
// NOTE(ldez): in the future, it may be also defined by server.
type ARIConfiguration struct {
	Disable             bool          `yaml:"disable,omitempty"`
	WaitToRenewDuration time.Duration `yaml:"waitToRenewDuration,omitempty"`
}

type PFX struct {
	Password string `yaml:"password,omitempty"`
	Format   string `yaml:"format,omitempty"`
}

type Hooks struct {
	Pre    *Hook `yaml:"pre,omitempty"`
	Deploy *Hook `yaml:"deploy,omitempty"`
	Post   *Hook `yaml:"post,omitempty"`
}

type Hook struct {
	Cmd     string        `yaml:"command,omitempty"`
	Timeout time.Duration `yaml:"timeout,omitempty"`
}

type Log struct {
	Level  string `yaml:"level,omitempty"`
	Format string `yaml:"format,omitempty"`
}
