package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/clientcredentials"
)

func mockOAuthServer[T any](t *testing.T, fn func(server *httptest.Server) (T, error)) T {
	t.Helper()

	return servermock.NewBuilder[T](fn).
		Route("POST /",
			servermock.ResponseFromFixture("auth.json"),
			servermock.CheckHeader().
				WithContentTypeFromURLEncoded(),
			servermock.CheckForm().Strict().
				With("client_id", "user").
				With("client_secret", "secret").
				With("grant_type", "access_key"),
		).
		Build(t)
}

func buildOAuthClient(server *httptest.Server) (*http.Client, error) {
	config, err := buildOAuthConfig(server)
	if err != nil {
		return nil, err
	}

	return config.Client(context.Background()), nil
}

func buildOAuthConfig(server *httptest.Server) (*clientcredentials.Config, error) {
	config := CreateOAuthConfig(server.URL, "user", "secret")

	return config, nil
}

func mockOAuthClient(t *testing.T) *http.Client {
	t.Helper()

	return mockOAuthServer(t, buildOAuthClient)
}

func Test_Token(t *testing.T) {
	token, err := mockOAuthServer(t, buildOAuthConfig).Token(t.Context())
	require.NoError(t, err)

	assert.Equal(t,
		"eyJhbGciOiJSUzI1NiIsImtpZCI6ImEyMjM3ZDhhLWQ0ZDQtNDA5Yi04ZTMxLWM3NGJhYTZhM2NjYiIsInR5cCI6IkpXVCJ9."+
			"eyJhdWQiOlsiaWFtIl0sImF1dGhfdGltZSI6MTY5MDA1Mzg1MiwiYXpwIjoiOWE0M2I0OTM1ZDRhMDc5NmRkYjE0Mjk0NjUxYjk2NzciLCJlbWFpbCI6ImxlZ29AOTlmN2I5NzItZmZlYS00OTkyLTgyN2EtY2M4MDYzOTg1MmNhLmlhbS5zYmVyY2xvdWQucnUiLCJleHAiOjE2OTAwNTc0NTIsImlhdCI6MTY5MDA1Mzg1MiwiaXNzIjoiaHR0cHM6Ly9hdXRoLmlhbS5zYmVyY2xvdWQucnUvYXV0aC9zeXN0ZW0iLCJqdGkiOiJlYzk0ZWJhNC03NzU2LTRjNjQtYmNmMC0zMzYxODIwNWM5ODkiLCJuYmYiOjE2OTAwNTM4NTIsIm5vbmNlIjoiIiwicmVhbG1fYWNjZXNzIjpudWxsLCJyZXNvdXJjZV9hY2Nlc3MiOnt9LCJzY29wZSI6Im9wZW5pZCBwcm9maWxlIGVtYWlsIHJvbGVzIiwic3ViIjoiZmVhMzRmYTUtZmE5ZS00OTdkLTk5MDAtNjdhYWJjOWNkZWJjIiwic3ViX2lkIjoiZmVhMzRmYTUtZmE5ZS00OTdkLTk5MDAtNjdhYWJjOWNkZWJjIiwic3ViX3R5cGUiOiJzZXJ2aWNlX2FjY291bnQiLCJ0eXAiOiJCZWFyZXIifQ."+
			"hhPr-Xr_NbyRwrqGoqeepthWfpfmD47RjzHUwo2lVPkeMiL8AMWzDPRxs-8gns9eTSHZCoAH0RjyrBnTaOrztInM72h8_rIIFr0MMPIIqrUkp2id_alya9eoiSWg_69PzNZ2CKWJDylL8o4Vi9_cSBYp-6H1xNcOAvO4a9xkNCoGGiogjHWNFq64qnS_P6fYY-pl9leuprCeq1GAKPODevHwzmc4gkEZIj_15SUh_ofJRJICgyLmkELQ8a0wDGYmZcdNKiGQDpd7rHaGrOvO1k8IJHfgs5aCMyuHXybTg6AMlodpYs8MBdk6K_VFY-cxSRB8ocq_Q7Hgt9qaRADg2Q",
		token.AccessToken,
	)
	assert.Equal(t, "Bearer", token.TokenType)
	assert.WithinDuration(t, time.Now().Add(3600*time.Second), token.Expiry, 2*time.Second)
}
