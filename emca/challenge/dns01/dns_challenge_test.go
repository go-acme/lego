package dns01

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/emca/api"
	"github.com/xenolf/lego/emca/le"
)

func TestDNSValidServerResponse(t *testing.T) {
	backupPreCheckDNS := PreCheckDNS
	defer func() {
		PreCheckDNS = backupPreCheckDNS
	}()

	PreCheckDNS = func(fqdn, value string) (bool, error) {
		return true, nil
	}

	privKey, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err)

	serverURL, tearDown := mockAPIEndpoint()
	defer tearDown()

	go func() {
		time.Sleep(time.Second * 2)
		f := bufio.NewWriter(os.Stdout)
		defer f.Flush()
		_, _ = f.WriteString("\n")
	}()

	manualProvider, err := NewDNSProviderManual()
	require.NoError(t, err)

	clientChallenge := le.Challenge{Type: "dns01", Status: "pending", URL: serverURL + "/chlg", Token: "http8"}

	core, err := api.New(http.DefaultClient, "lego-test", serverURL, "", privKey)
	require.NoError(t, err)

	solver := &Challenge{
		core:     core,
		validate: stubValidate,
		provider: manualProvider,
	}

	err = solver.Solve(clientChallenge, "example.com")

	require.NoError(t, err)
}

func mockAPIEndpoint() (string, func()) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := writeJSONResponse(w, le.Directory{
			NewNonceURL:   ts.URL + "/nonce",
			NewAccountURL: ts.URL + "/account",
			NewOrderURL:   ts.URL + "/newOrder",
			RevokeCertURL: ts.URL + "/revokeCert",
			KeyChangeURL:  ts.URL + "/keyChange",
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	return ts.URL, ts.Close
}

// FIXME remove?
// stubValidate is like validate, except it does nothing.
func stubValidate(_ *api.Core, _, _ string, _ le.Challenge) error {
	return nil
}

// writeJSONResponse marshals the body as JSON and writes it to the response.
func writeJSONResponse(w http.ResponseWriter, body interface{}) error {
	bs, err := json.Marshal(body)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(bs); err != nil {
		return err
	}

	return nil
}
