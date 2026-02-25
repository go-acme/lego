package dnschallenge

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v5"
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
)

const (
	testPersistCLIDomain         = "persist-cli.localhost"
	testPersistCLIWildcardDomain = "*." + testPersistCLIDomain

	testPersistCLIEmail      = "persist-e2e@example.com"
	testPersistCLIFreshEmail = "persist-e2e-fresh@example.com"
	testPersistCLIRenewEmail = "persist-e2e-renew@example.com"
)

func TestChallengeDNSPersist_Client_Obtain(t *testing.T) {
	t.Setenv("LEGO_CA_CERTIFICATES", "../fixtures/certs/pebble.minica.pem")

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

	updateDNS(t, reg.URI, testPersistBaseDomain)

	mockDefaultPersist(t)

	err = client.Challenge.SetDNSPersist01(
		dnspersist01.WithAccountURI(reg.URI),
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

	t.Setenv("LEGO_CA_CERTIFICATES", "../fixtures/certs/pebble.minica.pem")

	accountURI := createCLIAccountState(t, testPersistCLIEmail)
	require.NotEmpty(t, accountURI)

	updateDNS(t, accountURI, testPersistCLIDomain)

	err := load.RunLego(
		context.Background(),
		"run",
		"--email", testPersistCLIEmail,
		"--accept-tos",
		"--dns-persist",
		"--dns-persist.resolvers", ":8053",
		"--dns-persist.propagation.disable-ans",
		"--dns-persist.issuer-domain-name", testPersistIssuer,
		"--server", "https://localhost:15000/dir",
		"--domains", testPersistCLIWildcardDomain,
		"--domains", testPersistCLIDomain,
	)
	require.NoError(t, err)
}

func TestChallengeDNSPersist_Run_NewAccount(t *testing.T) {
	loader.CleanLegoFiles(context.Background())

	t.Setenv("LEGO_CA_CERTIFICATES", "../fixtures/certs/pebble.minica.pem")

	client := newChallTestSrvClient()

	defer func() {
		err := client.ClearPersistRecord(testPersistCLIDomain)
		require.NoError(t, err)
	}()

	stdinReader, stdinWriter := io.Pipe()

	defer func() { _ = stdinReader.Close() }()

	errChan := make(chan error, 1)

	go func() {
		defer func() { _ = stdinWriter.Close() }()

		accountURI, waitErr := waitForAccountFile(t.Context(), testPersistCLIFreshEmail)
		if waitErr != nil {
			errChan <- fmt.Errorf("wait for account URI: %w", waitErr)
			return
		}

		txtValue, err := dnspersist01.BuildIssueValue(testPersistIssuer, accountURI, true, time.Time{})
		require.NoError(t, err)

		err = client.SetPersistRecord(testPersistCLIDomain, txtValue)
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

	err := load.RunLegoWithInput(
		context.Background(),
		stdinReader,
		"run",
		"--email", testPersistCLIFreshEmail,
		"--accept-tos",
		"--dns-persist",
		"--dns-persist.resolvers", ":8053",
		"--dns-persist.propagation.disable-ans",
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

	t.Setenv("LEGO_CA_CERTIFICATES", "../fixtures/certs/pebble.minica.pem")

	accountURI := createCLIAccountState(t, testPersistCLIRenewEmail)
	require.NotEmpty(t, accountURI)

	updateDNS(t, accountURI, testPersistCLIDomain)

	err := load.RunLego(
		context.Background(),
		"run",
		"--email", testPersistCLIRenewEmail,
		"--accept-tos",
		"--dns-persist",
		"--dns-persist.resolvers", ":8053",
		"--dns-persist.propagation.disable-ans",
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
		"--dns-persist.propagation.disable-ans",
		"--dns-persist.issuer-domain-name", testPersistIssuer,
		"--server", "https://localhost:15000/dir",
		"--domains", testPersistCLIWildcardDomain,
		"--domains", testPersistCLIDomain,
		"--renew-force",
		"--no-random-sleep",
	)
	require.NoError(t, err)
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

	accountPathRoot := getAccountPath(email, keyType)

	err = os.MkdirAll(accountPathRoot, 0o700)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(accountPathRoot, email+".key"), certcrypto.PEMEncode(privateKey), 0o600)
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

func waitForAccountFile(ctx context.Context, email string) (string, error) {
	accountPath := filepath.Join(getAccountPath(email, certcrypto.EC256), "account.json")

	type accountFile struct {
		Registration *registration.Resource `json:"registration"`
	}

	return backoff.Retry(ctx,
		func() (string, error) {
			content, err := os.ReadFile(accountPath)
			if err != nil {
				if !os.IsNotExist(err) {
					return "", backoff.Permanent(err)
				}

				return "", err
			}

			var account accountFile

			err = json.Unmarshal(content, &account)
			if err != nil {
				return "", err
			}

			if account.Registration != nil && account.Registration.URI != "" {
				return account.Registration.URI, nil
			}

			return "", errors.New("account URI not found")
		},
		backoff.WithBackOff(backoff.NewConstantBackOff(50*time.Millisecond)),
		backoff.WithMaxElapsedTime(10*time.Second))
}

func getAccountPath(accountID string, keyType certcrypto.KeyType) string {
	return filepath.Join(".lego", "accounts", "localhost_15000", accountID, string(keyType))
}

func mockDefaultPersist(t *testing.T) {
	t.Helper()

	backup := dnspersist01.DefaultClient()

	t.Cleanup(func() {
		dnspersist01.SetDefaultClient(backup)
	})

	dnspersist01.SetDefaultClient(dnspersist01.NewClient(&dnspersist01.Options{RecursiveNameservers: []string{":8053"}}))
}
