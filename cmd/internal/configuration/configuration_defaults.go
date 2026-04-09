package configuration

import (
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/lego"
)

const DefaultAccountID = "noemail@example.com"

const defaultLegoDirectory = ".lego"

const (
	defaultHTTP01       = "http-01"
	defaultTLSALPN01    = "tls-alpn-01"
	defaultDNSPersist01 = "dns-persist-01"
)

const (
	defaultHTTPAddress = ":80"
	defaultTLSAddress  = ":443"
)

func ApplyDefaults(cfg *Configuration) {
	if cfg.Storage == "" {
		cfg.Storage = defaultLegoDirectory
	}

	if len(cfg.Servers) == 0 {
		cfg.Servers = make(map[string]*Server)
	}

	if len(cfg.Accounts) == 0 {
		cfg.Accounts = make(map[string]*Account)
	}

	if len(cfg.Challenges) == 0 {
		cfg.Challenges = make(map[string]*Challenge)
	}

	if len(cfg.Certificates) == 0 {
		cfg.Certificates = make(map[string]*Certificate)
	}

	applyServersDefaults(cfg)
	applyAccountsDefaults(cfg)
	applyChallengesDefaults(cfg)
	applyCertificatesDefaults(cfg)
}

func applyServersDefaults(cfg *Configuration) {
	for _, server := range cfg.Servers {
		if server.OverallRequestLimit <= 0 {
			server.OverallRequestLimit = certificate.DefaultOverallRequestLimit
		}
	}
}

func applyAccountsDefaults(cfg *Configuration) {
	for id, account := range cfg.Accounts {
		account.ID = id

		if account.KeyType == "" {
			account.KeyType = certcrypto.EC256
		}

		if account.Server == "" {
			account.Server = lego.DirectoryURLLetsEncrypt
		}
	}
}

func applyChallengesDefaults(cfg *Configuration) {
	for id, challenge := range cfg.Challenges {
		challenge.ID = id

		if challenge.TLS != nil && challenge.TLS.Address == "" {
			challenge.TLS.Address = defaultTLSAddress
		}

		if challenge.HTTP != nil && challenge.HTTP.Address == "" {
			challenge.HTTP.Address = defaultHTTPAddress
		}
	}
}

func applyCertificatesDefaults(cfg *Configuration) {
	defaultAccount := getDefaultAccountID(cfg)

	for id, cert := range cfg.Certificates {
		cert.ID = id

		if cert.Account == "" {
			cert.Account = defaultAccount
		}

		if cert.Account == DefaultAccountID {
			setDefaultAccount(cfg)
		}

		if cert.KeyType == "" {
			cert.KeyType = getDefaultCertificateKeyType(cfg, cert.Account)
		}

		applyRenewDefaults(cert)

		switch cert.Challenge {
		case defaultHTTP01:
			setDefaultHTTP01(cfg)

			continue

		case defaultTLSALPN01:
			setDefaultTLSALPN01(cfg)

			continue

		case defaultDNSPersist01:
			setDefaultDNSPersist01(cfg)

			continue

		default:
			if cert.Challenge == "" && len(cfg.Challenges) == 1 {
				// If there is only one challenge, use it by default.
				for c := range cfg.Challenges {
					cert.Challenge = c
				}
			}
		}
	}
}

func applyRenewDefaults(cert *Certificate) {
	if cert.Renew == nil {
		cert.Renew = &RenewConfiguration{}
	}

	if cert.Renew.ARI == nil {
		cert.Renew.ARI = &ARIConfiguration{}
	}
}

func setDefaultHTTP01(cfg *Configuration) {
	if _, ok := cfg.Challenges[defaultHTTP01]; ok {
		return
	}

	cfg.Challenges[defaultHTTP01] = &Challenge{HTTP: &HTTPChallenge{
		Address: defaultHTTPAddress,
	}}
}

func setDefaultTLSALPN01(cfg *Configuration) {
	if _, ok := cfg.Challenges[defaultTLSALPN01]; ok {
		return
	}

	cfg.Challenges[defaultTLSALPN01] = &Challenge{TLS: &TLSChallenge{
		Address: defaultTLSAddress,
	}}
}

func setDefaultDNSPersist01(cfg *Configuration) {
	if _, ok := cfg.Challenges[defaultDNSPersist01]; ok {
		return
	}

	cfg.Challenges[defaultDNSPersist01] = &Challenge{DNSPersist: &DNSPersistChallenge{}}
}

func setDefaultAccount(cfg *Configuration) {
	if _, ok := cfg.Accounts[DefaultAccountID]; ok {
		return
	}

	cfg.Accounts[DefaultAccountID] = &Account{
		Server:  lego.DirectoryURLLetsEncrypt,
		KeyType: certcrypto.EC256,
	}
}

func getDefaultAccountID(cfg *Configuration) string {
	switch len(cfg.Accounts) {
	case 0:
		return DefaultAccountID
	case 1:
		for a := range cfg.Accounts {
			return a
		}
	}

	return ""
}

func getDefaultCertificateKeyType(cfg *Configuration, acc string) certcrypto.KeyType {
	if cfg.Accounts == nil {
		return certcrypto.EC256
	}

	if account, ok := cfg.Accounts[acc]; ok {
		return account.KeyType
	}

	return certcrypto.EC256
}
