package gandiv5

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(EnvAPIKey, EnvPersonalAccessToken)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIKey: "123",
			},
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				EnvAPIKey: "",
			},
			expected: "gandiv5: credentials information are missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()

			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		apiKey   string
		expected string
	}{
		{
			desc:   "success",
			apiKey: "123",
		},
		{
			desc:     "missing credentials",
			expected: "gandiv5: credentials information are missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.PersonalAccessToken = "123412341234123412341234"
			config.BaseURL = server.URL
			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			return p, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Bearer 123412341234123412341234"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /domains/example.com/verification/_acme-challenge.abc.def/TXT",
			servermock.ResponseFromInternal("post_txt_verification_record.json"),
			servermock.CheckRequestJSONBodyFromInternal("post_txt_verification_record-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "abc.def.example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp_delete(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domains/example.com/records/_acme-challenge.abc.def/TXT",
			servermock.ResponseFromInternal("api_response.json").
				WithStatusCode(http.StatusNotFound),
		).
		Route("DELETE /domains/example.com/records/_acme-challenge.abc.def/TXT",
			servermock.Noop().
				WithStatusCode(http.StatusCreated),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "abc.def.example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp_update(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domains/example.com/records/_acme-challenge.abc.def/TXT",
			servermock.ResponseFromInternal("get_txt_record_existing.json"),
		).
		Route("PUT /domains/example.com/records/_acme-challenge.abc.def/TXT",
			servermock.ResponseFromInternal("put_txt_record_pristine-request.json"),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "abc.def.example.com", "abc", "123d==")
	require.NoError(t, err)
}
