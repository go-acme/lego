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
	privKey, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err)

	_, apiURL, tearDown := tester.SetupFakeAPI()
	defer tearDown()

	go func() {
		time.Sleep(time.Second * 2)
		f := bufio.NewWriter(os.Stdout)
		defer f.Flush()
		_, _ = f.WriteString("\n")
	}()

	manualProvider, err := NewDNSProviderManual()
	require.NoError(t, err)

	clientChallenge := le.Challenge{Type: "dns01", Status: "pending", URL: apiURL + "/chlg", Token: "http8"}

	core, err := api.New(http.DefaultClient, "lego-test", apiURL, "", privKey)
	require.NoError(t, err)

	solver := NewChallenge(
		core,
		func(_ *api.Core, _, _ string, _ le.Challenge) error { return nil },
		manualProvider,
		AddPreCheck(func(fqdn, value string) (bool, error) {
			return true, nil
		}),
	)

	err = solver.Solve(clientChallenge, "example.com")

	require.NoError(t, err)
}
