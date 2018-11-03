package dns01

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/le"
	"github.com/xenolf/lego/le/api"
	"github.com/xenolf/lego/platform/tester"
)

func TestDNSValidServerResponse(t *testing.T) {
	_, apiURL, tearDown := tester.SetupFakeAPI()
	defer tearDown()

	go func() {
		time.Sleep(time.Second * 2)
		f := bufio.NewWriter(os.Stdout)
		defer f.Flush()
		_, _ = f.WriteString("\n")
	}()

	privKey, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err)

	core, err := api.New(http.DefaultClient, "lego-test", apiURL, "", privKey)
	require.NoError(t, err)

	manualProvider, err := NewDNSProviderManual()
	require.NoError(t, err)

	validate := func(_ *api.Core, _, _ string, _ le.Challenge) error { return nil }
	preCheck := func(fqdn, value string) (bool, error) { return true, nil }

	chlg := NewChallenge(core, validate, manualProvider, AddPreCheck(preCheck))

	clientChallenge := le.Challenge{Type: "dns01", Status: "pending", URL: apiURL + "/chlg", Token: "http8"}

	err = chlg.Solve(clientChallenge, "example.com")
	require.NoError(t, err)
}
