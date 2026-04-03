package configuration

import (
	"maps"
	"slices"
)

type AccountNode[T any] struct {
	*Account

	ServerConfig *Server
	Children     []T
}

type ChallengeNode struct {
	*Challenge

	Certificates []*Certificate
}

func LookupCertificates(cfg *Configuration, certIDs []string) []*AccountNode[*Certificate] {
	uniqAccounts := make(map[string]*AccountNode[*Certificate])

	for _, certID := range certIDs {
		cert, ok := cfg.Certificates[certID]
		if !ok {
			continue
		}

		if _, ok := uniqAccounts[cert.Account]; !ok {
			uniqAccounts[cert.Account] = &AccountNode[*Certificate]{
				Account:      cfg.Accounts[cert.Account],
				ServerConfig: GetServerConfig(cfg, cert.Account),
			}
		}

		uniqAccounts[cert.Account].Children = append(uniqAccounts[cert.Account].Children, cert)
	}

	return slices.Collect(maps.Values(uniqAccounts))
}

func LookupChallenges(cfg *Configuration, filter *Filter) []*AccountNode[*ChallengeNode] {
	certsMappings := createCertificatesMapping(cfg, filter)

	var accounts []*AccountNode[*ChallengeNode]

	for accountID, challengesInfo := range certsMappings {
		account := &AccountNode[*ChallengeNode]{
			Account:      cfg.Accounts[accountID],
			ServerConfig: GetServerConfig(cfg, accountID),
		}

		for challengeID, certIDs := range challengesInfo {
			chlg := &ChallengeNode{
				Challenge: cfg.Challenges[challengeID],
			}

			for _, certID := range certIDs {
				chlg.Certificates = append(chlg.Certificates, cfg.Certificates[certID])
			}

			account.Children = append(account.Children, chlg)
		}

		accounts = append(accounts, account)
	}

	return accounts
}

type Filter struct {
	// If empty, all certificates will be included.
	CertificateIDs []string

	// If empty, all accounts will be included.
	AccountIDs []string
}

// createCertificatesMapping creates a mapping of account -> challenge -> certificate IDs.
func createCertificatesMapping(cfg *Configuration, filter *Filter) map[string]map[string][]string {
	// Accounts -> Challenges -> Certificates
	certsMappings := make(map[string]map[string][]string)

	for certID, certDesc := range cfg.Certificates {
		if filter != nil && len(filter.CertificateIDs) > 0 && !slices.Contains(filter.CertificateIDs, certID) {
			continue
		}

		if filter != nil && len(filter.AccountIDs) > 0 && !slices.Contains(filter.AccountIDs, certDesc.Account) {
			continue
		}

		if _, ok := certsMappings[certDesc.Account]; !ok {
			certsMappings[certDesc.Account] = make(map[string][]string)
		}

		certsMappings[certDesc.Account][certDesc.Challenge] = append(certsMappings[certDesc.Account][certDesc.Challenge], certID)
	}

	return certsMappings
}
