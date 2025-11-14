package namecheap

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_defaultTransport(t *testing.T) {
	client := servermock.NewBuilder(
		func(server *httptest.Server) (*http.Client, error) {
			cl := server.Client()

			t.Setenv("NAMECHEAP_HTTP_PROXY", server.URL)

			cl.Transport = defaultTransport(envNamespace)

			return cl, nil
		}).
		Route("/",
			servermock.Noop().WithStatusCode(http.StatusTeapot)).
		Build(t)

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = resp.Body.Close()
	})

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}
