package gandiv5

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

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

// serverResponses is the JSON Request->Response map used by the
// fake JSON server.
var serverResponses = map[string]string{
	// Present Request->Response (addTXTRecord)
	`{"rrset_ttl":300,"rrset_values":["TOKEN"]}`: `{"message": "Zone Record Created"}`,
	// CleanUp Request->Response (deleteTXTRecord)
	`{"delete":true}`: ``,
}
