package dnschallenge

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/challenge/dnspersist01"
	"github.com/go-acme/lego/v5/e2e/loader"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/registration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testPersistBaseDomain = "persist.localhost"
	testPersistDomain     = "*." + testPersistBaseDomain
	testPersistIssuer     = "pebble.letsencrypt.org"

	testPersistCLIDomain         = "persist-cli.localhost"
	testPersistCLIWildcardDomain = "*." + testPersistCLIDomain
	testPersistCLIEmail          = "persist-e2e@example.com"
	testPersistCLIFreshEmail     = "persist-e2e-fresh@example.com"
	testPersistCLIRenewEmail     = "persist-e2e-renew@example.com"
)

func setTXTRecord(t *testing.T, host, value string) {
	t.Helper()

	err := setTXTRecordRaw(host, value)
	require.NoError(t, err)
}

func setTXTRecordRaw(host, value string) error {
	body, err := json.Marshal(map[string]string{
		"host":  host,
		"value": value,
	})
	if err != nil {
		return err
	}

	resp, err := http.Post("http://localhost:8055/set-txt", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

func clearTXTRecord(t *testing.T, host string) {
	t.Helper()

	body, err := json.Marshal(map[string]string{
		"host": host,
	})
	require.NoError(t, err)

	resp, err := http.Post("http://localhost:8055/clear-txt", "application/json", bytes.NewReader(body))
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

//nolint:unparam // kept generic for future e2e tests.
func mustDNSPersistIssueValue(t *testing.T, issuerDomainName, accountURI string, wildcard bool, persistUntil *time.Time) string {
	t.Helper()

	value, err := dnspersist01.BuildIssueValue(issuerDomainName, accountURI, wildcard, persistUntil)
	require.NoError(t, err)

	return value
}

func createCLIAccountState(t *testing.T, email string) string {
	t.Helper()

	privateKey, err := certcrypto.GeneratePrivateKey(certcrypto.EC256)
	require.NoError(t, err)

	user := &fakeUser{
		email:      email,
		privateKey: privateKey,
	}
	config := lego.NewConfig(user)
	config.CADirURL = "https://localhost:15000/dir"

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	reg, err := client.Registration.Register(context.Background(), registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)
	require.NotEmpty(t, reg.URI)

	keyType := certcrypto.EC256
	accountPathRoot := filepath.Join(".lego", "accounts", "localhost_15000", email, string(keyType))
	err = os.MkdirAll(accountPathRoot, 0o700)
	require.NoError(t, err)

	err = saveAccountPrivateKey(filepath.Join(accountPathRoot, email+".key"), privateKey)
	require.NoError(t, err)

	accountPath := filepath.Join(accountPathRoot, "account.json")
	content, err := json.MarshalIndent(struct {
		ID           string                 `json:"id"`
		Email        string                 `json:"email"`
		KeyType      certcrypto.KeyType     `json:"keyType"`
		Registration *registration.Resource `json:"registration"`
	}{
		ID:           email,
		Email:        email,
		KeyType:      keyType,
		Registration: reg,
	}, "", "\t")
	require.NoError(t, err)

	err = os.WriteFile(accountPath, content, 0o600)
	require.NoError(t, err)

	return reg.URI
}

func saveAccountPrivateKey(path string, privateKey crypto.PrivateKey) error {
	return os.WriteFile(path, certcrypto.PEMEncode(privateKey), 0o600)
}

func cliAccountFilePath(email string) string {
	return filepath.Join(".lego", "accounts", "localhost_15000", email, string(certcrypto.EC256), "account.json")
}

func waitForCLIAccountURI(ctx context.Context, email string) (string, error) {
	accountPath := cliAccountFilePath(email)

	type accountFile struct {
		Registration *registration.Resource `json:"registration"`
	}

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
			content, err := os.ReadFile(accountPath)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}

				return "", err
			}

			var account accountFile

			err = json.Unmarshal(content, &account)
			if err != nil {
				continue
			}

			if account.Registration != nil && account.Registration.URI != "" {
				return account.Registration.URI, nil
			}
		}
	}
}

func TestChallengeDNSPersist_Client_Obtain(t *testing.T) {
	err := os.Setenv("LEGO_CA_CERTIFICATES", "../fixtures/certs/pebble.minica.pem")
	require.NoError(t, err)

	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	user := &fakeUser{privateKey: privateKey}
	config := lego.NewConfig(user)
	config.CADirURL = "https://localhost:15000/dir"

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	reg, err := client.Registration.Register(context.Background(), registration.RegisterOptions{TermsOfServiceAgreed: true})
	require.NoError(t, err)
	require.NotEmpty(t, reg.URI)

	user.registration = reg

	txtHost := fmt.Sprintf("_validation-persist.%s", testPersistBaseDomain)
	txtValue := mustDNSPersistIssueValue(t, testPersistIssuer, reg.URI, true, nil)

	setTXTRecord(t, txtHost, txtValue)
	defer clearTXTRecord(t, txtHost)

	err = client.Challenge.SetDNSPersist01(
		dnspersist01.WithAccountURI(reg.URI),
		dnspersist01.WithNameservers([]string{":8053"}),
		dnspersist01.AddRecursiveNameservers([]string{":8053"}),
		dnspersist01.DisableAuthoritativeNssPropagationRequirement(),
	)
	require.NoError(t, err)

	privateKeyCSR, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	request := certificate.ObtainRequest{
		Domains:    []string{testPersistDomain},
		Bundle:     true,
		PrivateKey: privateKeyCSR,
	}
	resource, err := client.Certificate.Obtain(context.Background(), request)
	require.NoError(t, err)

	require.NotNil(t, resource)
	assert.Equal(t, testPersistDomain, resource.Domains[0])
	assert.Regexp(t, `https://localhost:15000/certZ/[\w\d]{14,}`, resource.CertURL)
	assert.Regexp(t, `https://localhost:15000/certZ/[\w\d]{14,}`, resource.CertStableURL)
	assert.NotEmpty(t, resource.Certificate)
	assert.NotEmpty(t, resource.IssuerCertificate)
	assert.Empty(t, resource.CSR)
}

func TestChallengeDNSPersist_Run(t *testing.T) {
	loader.CleanLegoFiles(context.Background())

	err := os.Setenv("LEGO_CA_CERTIFICATES", "../fixtures/certs/pebble.minica.pem")
	require.NoError(t, err)

	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	accountURI := createCLIAccountState(t, testPersistCLIEmail)
	require.NotEmpty(t, accountURI)

	txtHost := fmt.Sprintf("_validation-persist.%s", testPersistCLIDomain)
	txtValue := mustDNSPersistIssueValue(t, testPersistIssuer, accountURI, true, nil)

	setTXTRecord(t, txtHost, txtValue)
	defer clearTXTRecord(t, txtHost)

	err = load.RunLego(
		context.Background(),
		"run",
		"--email", testPersistCLIEmail,
		"--accept-tos",
		"--dns-persist",
		"--dns-persist.resolvers", ":8053",
		"--dns-persist.propagation-disable-ans",
		"--dns-persist.issuer-domain-name", testPersistIssuer,
		"--server", "https://localhost:15000/dir",
		"--domains", testPersistCLIWildcardDomain,
		"--domains", testPersistCLIDomain,
	)
	require.NoError(t, err)
}

func TestChallengeDNSPersist_Run_NewAccount(t *testing.T) {
	loader.CleanLegoFiles(context.Background())

	err := os.Setenv("LEGO_CA_CERTIFICATES", "../fixtures/certs/pebble.minica.pem")
	require.NoError(t, err)

	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	txtHost := fmt.Sprintf("_validation-persist.%s", testPersistCLIDomain)
	defer clearTXTRecord(t, txtHost)

	stdinReader, stdinWriter := io.Pipe()

	defer func() { _ = stdinReader.Close() }()

	errChan := make(chan error, 1)

	go func() {
		defer func() { _ = stdinWriter.Close() }()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		accountURI, waitErr := waitForCLIAccountURI(ctx, testPersistCLIFreshEmail)
		if waitErr != nil {
			errChan <- fmt.Errorf("wait for account URI: %w", waitErr)
			return
		}

		txtValue := mustDNSPersistIssueValue(t, testPersistIssuer, accountURI, true, nil)

		err = setTXTRecordRaw(txtHost, txtValue)
		if err != nil {
			errChan <- fmt.Errorf("set TXT record: %w", err)
			return
		}

		_, err = io.WriteString(stdinWriter, "\n")
		if err != nil {
			errChan <- fmt.Errorf("send enter to lego: %w", err)
			return
		}

		errChan <- nil
	}()

	err = load.RunLegoWithInput(
		context.Background(),
		stdinReader,
		"run",
		"--email", testPersistCLIFreshEmail,
		"--accept-tos",
		"--dns-persist",
		"--dns-persist.resolvers", ":8053",
		"--dns-persist.propagation-disable-ans",
		"--dns-persist.issuer-domain-name", testPersistIssuer,
		"--server", "https://localhost:15000/dir",
		"--domains", testPersistCLIWildcardDomain,
		"--domains", testPersistCLIDomain,
	)
	require.NoError(t, err)
	require.NoError(t, <-errChan)
}

func TestChallengeDNSPersist_Renew(t *testing.T) {
	loader.CleanLegoFiles(context.Background())

	err := os.Setenv("LEGO_CA_CERTIFICATES", "../fixtures/certs/pebble.minica.pem")
	require.NoError(t, err)

	defer func() { _ = os.Unsetenv("LEGO_CA_CERTIFICATES") }()

	accountURI := createCLIAccountState(t, testPersistCLIRenewEmail)
	require.NotEmpty(t, accountURI)

	txtHost := fmt.Sprintf("_validation-persist.%s", testPersistCLIDomain)
	txtValue := mustDNSPersistIssueValue(t, testPersistIssuer, accountURI, true, nil)

	setTXTRecord(t, txtHost, txtValue)
	defer clearTXTRecord(t, txtHost)

	err = load.RunLego(
		context.Background(),
		"run",
		"--email", testPersistCLIRenewEmail,
		"--accept-tos",
		"--dns-persist",
		"--dns-persist.resolvers", ":8053",
		"--dns-persist.propagation-disable-ans",
		"--dns-persist.issuer-domain-name", testPersistIssuer,
		"--server", "https://localhost:15000/dir",
		"--domains", testPersistCLIWildcardDomain,
		"--domains", testPersistCLIDomain,
	)
	require.NoError(t, err)

	err = load.RunLego(
		context.Background(),
		"renew",
		"--email", testPersistCLIRenewEmail,
		"--dns-persist",
		"--dns-persist.resolvers", ":8053",
		"--dns-persist.propagation-disable-ans",
		"--dns-persist.issuer-domain-name", testPersistIssuer,
		"--server", "https://localhost:15000/dir",
		"--domains", testPersistCLIWildcardDomain,
		"--domains", testPersistCLIDomain,
		"--renew-force",
		"--no-random-sleep",
	)
	require.NoError(t, err)
}
