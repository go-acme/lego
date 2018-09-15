package gandiv5

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDNSProvider runs Present and CleanUp against a fake Gandi RPC
// Server, whose responses are predetermined for particular requests.
func TestDNSProvider(t *testing.T) {
	fakeKeyAuth := "XXXX"

	regexpToken, err := regexp.Compile(`"rrset_values":\[".+"\]`)
	require.NoError(t, err)

	// start fake RPC server
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "application/json", r.Header.Get("Content-Type"), "invalid content type")

		req, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		req = regexpToken.ReplaceAllLiteral(req, []byte(`"rrset_values":["TOKEN"]`))

		resp, ok := serverResponses[string(req)]
		require.True(t, ok, "Server response for request not found")

		_, err = io.Copy(w, strings.NewReader(resp))
		require.NoError(t, err)
	}))
	defer fakeServer.Close()

	// define function to override findZoneByFqdn with
	fakeFindZoneByFqdn := func(fqdn string, nameserver []string) (string, error) {
		return "example.com.", nil
	}

	config := NewDefaultConfig()
	config.APIKey = "123412341234123412341234"
	config.BaseURL = fakeServer.URL

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	// override findZoneByFqdn function
	savedFindZoneByFqdn := findZoneByFqdn
	defer func() {
		findZoneByFqdn = savedFindZoneByFqdn
	}()
	findZoneByFqdn = fakeFindZoneByFqdn

	// run Present
	err = provider.Present("abc.def.example.com", "", fakeKeyAuth)
	require.NoError(t, err)

	// run CleanUp
	err = provider.CleanUp("abc.def.example.com", "", fakeKeyAuth)
	require.NoError(t, err)
}

// serverResponses is the JSON Request->Response map used by the
// fake JSON server.
var serverResponses = map[string]string{
	// Present Request->Response (addTXTRecord)
	`{"rrset_ttl":300,"rrset_values":["TOKEN"]}`: `{"message": "Zone Record Created"}`,
	// CleanUp Request->Response (deleteTXTRecord)
	`{"delete":true}`: ``,
}
