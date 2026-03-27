package configuration

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/log"
)

const (
	LogFormatText    = "text"
	LogFormatJSON    = "json"
	LogFormatColored = "colored"
)

func Validate(cfg *Configuration) error {
	validators := []func(*Configuration) error{
		validateLog,
		validateChallenges,
		validateCertificates,
		validateServers,
		validateAccounts,
	}

	for _, validator := range validators {
		err := validator(cfg)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateCertificates(cfg *Configuration) error {
	if len(cfg.Certificates) == 0 {
		return errors.New("no certificate configurations found")
	}

	for name, cert := range cfg.Certificates {
		if strings.TrimSpace(name) == "" {
			return errors.New("the certificate name cannot be empty")
		}

		err := validateCertificate(cert)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}

		err = existInMap(cfg.Accounts, cert.Account)
		if err != nil {
			return fmt.Errorf("%s: account: %w", name, err)
		}

		err = existInMap(cfg.Challenges, cert.Challenge)
		if err != nil {
			return fmt.Errorf("%s: challenge: %w", name, err)
		}
	}

	return nil
}

func validateCertificate(cert *Certificate) error {
	if len(cert.Domains) == 0 && cert.CSR == "" {
		return errors.New("at least one domain or CSR must be provided")
	}

	if cert.CSR != "" && len(cert.Domains) > 0 {
		return errors.New("domains and CSR are mutually exclusive")
	}

	if cert.Account == "" {
		return errors.New("an account is required")
	}

	if cert.Challenge == "" {
		return errors.New("a challenge is required")
	}

	if !certcrypto.IsSupported(cert.KeyType) {
		return fmt.Errorf("unsupported key type: %s", cert.KeyType)
	}

	return nil
}

func validateChallenges(cfg *Configuration) error {
	if len(cfg.Challenges) == 0 {
		return errors.New("no challenge configurations found")
	}

	for name, challenge := range cfg.Challenges {
		if strings.TrimSpace(name) == "" {
			return errors.New("the challenge name cannot be empty")
		}

		err := validateChallenge(challenge)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}

	return nil
}

func validateChallenge(chlg *Challenge) error {
	hasTLSChallenge := chlg.TLS != nil
	hasHTTPChallenge := chlg.HTTP != nil
	hasDNSChallenge := chlg.DNS != nil
	hasDNSPersistChallenge := chlg.DNSPersist != nil

	if !hasTLSChallenge && !hasHTTPChallenge && !hasDNSChallenge && !hasDNSPersistChallenge {
		return errors.New("at least one challenge type must be defined")
	}

	if hasDNSChallenge {
		if chlg.DNS.Provider == "" {
			return errors.New("a provider is required")
		}

		err := validatePropagationExclusiveOptions(chlg.DNS.Propagation)
		if err != nil {
			return err
		}
	}

	if hasDNSPersistChallenge {
		err := validatePropagationExclusiveOptions(chlg.DNSPersist.Propagation)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateAccounts(cfg *Configuration) error {
	if len(cfg.Accounts) == 0 {
		return errors.New("no account configurations found")
	}

	for name, account := range cfg.Accounts {
		err := validateAccount(name, account)
		if err != nil {
			return fmt.Errorf("account '%s': %w", name, err)
		}
	}

	return nil
}

func validateAccount(name string, account *Account) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("the account name cannot be empty")
	}

	if !certcrypto.IsSupported(account.KeyType) {
		return fmt.Errorf("unsupported key type: %s", account.KeyType)
	}

	if account.ExternalAccountBinding != nil {
		if account.ExternalAccountBinding.KID == "" || account.ExternalAccountBinding.HmacKey == "" {
			return errors.New("KID and HMAC key must be provided for External Account Binding")
		}
	}

	return nil
}

func validateServers(cfg *Configuration) error {
	serverUsageCount := make(map[string]int)

	for name := range cfg.Servers {
		if strings.TrimSpace(name) == "" {
			return errors.New("the server name cannot be empty")
		}

		serverUsageCount[name] = 0
	}

	for _, account := range cfg.Accounts {
		if _, ok := serverUsageCount[account.Server]; ok {
			serverUsageCount[account.Server]++
		}
	}

	for name, count := range serverUsageCount {
		if count == 0 {
			log.Warn("Server configuration unused.", slog.String("server", name))
		}
	}

	return nil
}

func validateLog(cfg *Configuration) error {
	if cfg.Log == nil {
		return nil
	}

	formats := []string{LogFormatText, LogFormatJSON, LogFormatColored, ""}

	if !slices.Contains(formats, cfg.Log.Format) {
		return fmt.Errorf("invalid log format '%s'", cfg.Log.Format)
	}

	return nil
}

func validatePropagationExclusiveOptions(cfg *Propagation) error {
	if cfg == nil || cfg.Wait == 0 {
		return nil
	}

	if cfg.Wait < 0 {
		return errors.New("'wait' must be a positive integer")
	}

	if cfg.DisableAuthoritativeNameservers {
		return errors.New("'wait' and 'disableAuthoritativeNameservers' are mutually exclusive")
	}

	if cfg.DisableRecursiveNameservers {
		return errors.New("'wait' and 'disableRecursiveNameservers' are mutually exclusive")
	}

	return nil
}

func existInMap[T map[string]V, V any](m T, key string) error {
	if key == "" {
		return nil
	}

	if _, ok := m[key]; !ok {
		return fmt.Errorf("'%s' not found", key)
	}

	return nil
}
