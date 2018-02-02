package gandiv5

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/xenolf/lego/acme"
)

// stagingServer is the Let's Encrypt staging server used by the live test
const stagingServer = "https://acme-staging.api.letsencrypt.org/directory"

// user implements acme.User and is used by the live test
type user struct {
	Email        string
	Registration *acme.RegistrationResource
	key          crypto.PrivateKey
}

func (u *user) GetEmail() string {
	return u.Email
}
func (u *user) GetRegistration() *acme.RegistrationResource {
	return u.Registration
}
func (u *user) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// TestDNSProvider runs Present and CleanUp against a fake Gandi RPC
// Server, whose responses are predetermined for particular requests.
func TestDNSProvider(t *testing.T) {
	fakeAPIKey := "123412341234123412341234"
	fakeKeyAuth := "XXXX"
	provider, err := NewDNSProviderCredentials(fakeAPIKey)
	if err != nil {
		t.Fatal(err)
	}
	regexpToken, err := regexp.Compile(`"rrset_values":\[".+"\]`)
	if err != nil {
		t.Fatal(err)
	}
	// start fake RPC server
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("Content-Type: application/json header not found")
		}
		req, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		req = regexpToken.ReplaceAllLiteral(
			req, []byte(`"rrset_values":["TOKEN"]`))
		resp, ok := serverResponses[string(req)]
		if !ok {
			t.Fatalf("Server response for request not found")
		}
		_, err = io.Copy(w, strings.NewReader(resp))
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer fakeServer.Close()
	// define function to override findZoneByFqdn with
	fakeFindZoneByFqdn := func(fqdn string, nameserver []string) (string, error) {
		return "example.com.", nil
	}
	// override gandi endpoint and findZoneByFqdn function
	savedEndpoint, savedFindZoneByFqdn := endpoint, findZoneByFqdn
	defer func() {
		endpoint, findZoneByFqdn = savedEndpoint, savedFindZoneByFqdn
	}()
	endpoint, findZoneByFqdn = fakeServer.URL, fakeFindZoneByFqdn
	// run Present
	err = provider.Present("abc.def.example.com", "", fakeKeyAuth)
	if err != nil {
		t.Fatal(err)
	}
	// run CleanUp
	err = provider.CleanUp("abc.def.example.com", "", fakeKeyAuth)
	if err != nil {
		t.Fatal(err)
	}
}

// TestDNSProviderLive performs a live test to obtain a certificate
// using the Let's Encrypt staging server. It runs provided that both
// the environment variables GANDIV5_API_KEY and GANDI_TEST_DOMAIN are
// set. Otherwise the test is skipped.
//
// To complete this test, go test must be run with the -timeout=40m
// flag, since the default timeout of 10m is insufficient.
func TestDNSProviderLive(t *testing.T) {
	apiKey := os.Getenv("GANDIV5_API_KEY")
	domain := os.Getenv("GANDI_TEST_DOMAIN")
	if apiKey == "" || domain == "" {
		t.Skip("skipping live test")
	}
	// create a user.
	const rsaKeySize = 2048
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		t.Fatal(err)
	}
	myUser := user{
		Email: "test@example.com",
		key:   privateKey,
	}
	// create a client using staging server
	client, err := acme.NewClient(stagingServer, &myUser, acme.RSA2048)
	if err != nil {
		t.Fatal(err)
	}
	provider, err := NewDNSProviderCredentials(apiKey)
	if err != nil {
		t.Fatal(err)
	}
	err = client.SetChallengeProvider(acme.DNS01, provider)
	if err != nil {
		t.Fatal(err)
	}
	client.ExcludeChallenges([]acme.Challenge{acme.HTTP01, acme.TLSSNI01})
	// register and agree tos
	reg, err := client.Register()
	if err != nil {
		t.Fatal(err)
	}
	myUser.Registration = reg
	err = client.AgreeToTOS()
	if err != nil {
		t.Fatal(err)
	}
	// complete the challenge
	bundle := false
	_, failures := client.ObtainCertificate([]string{domain}, bundle, nil, false)
	if len(failures) > 0 {
		t.Fatal(failures)
	}
}

// serverResponses is the JSON Request->Response map used by the
// fake JSON server.
var serverResponses = map[string]string{
	// Present Request->Response (addTXTRecord)
	`{"rrset_ttl":300,"rrset_values":["TOKEN"]}`: `{"message": "Zone Record Created"}`,
	// CleanUp Request->Response (deleteTXTRecord)
	`{"delete":true}`: ``,
}
