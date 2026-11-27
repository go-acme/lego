package wedos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/providers/dns/wedos/internal"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvUsername, EnvPassword).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvUsername: "admin@example.com",
				EnvPassword: "secret",
			},
		},
		{
			desc: "missing credentials: username",
			envVars: map[string]string{
				EnvUsername: "",
				EnvPassword: "secret",
			},
			expected: "wedos: some credentials information are missing: WEDOS_USERNAME",
		},
		{
			desc: "missing credentials: password",
			envVars: map[string]string{
				EnvUsername: "admin@example.com",
				EnvPassword: "",
			},
			expected: "wedos: some credentials information are missing: WEDOS_WAPI_PASSWORD",
		},
		{
			desc: "missing credentials: all",
			envVars: map[string]string{
				EnvUsername: "",
				EnvPassword: "",
			},
			expected: "wedos: some credentials information are missing: WEDOS_USERNAME,WEDOS_WAPI_PASSWORD",
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
		username string
		password string
		expected string
	}{
		{
			desc:     "success",
			username: "admin@example.com",
			password: "secret",
		},
		{
			desc:     "missing username",
			password: "secret",
			expected: "wedos: some credentials information are missing",
		},
		{
			desc:     "missing WAPI password",
			username: "admin@example.com",
			expected: "wedos: some credentials information are missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Username = test.username
			config.Password = test.password

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

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.CleanUp(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

type cmdMapping struct {
	request  string
	response string
}

func mockBuilder(mapping map[string]cmdMapping) *servermock.Builder[*DNSProvider] {
	if mapping == nil {
		mapping = make(map[string]cmdMapping)
	}

	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.Username = "user"
			config.Password = "secret"
			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL = server.URL

			return p, nil
		},
		servermock.CheckHeader().
			WithContentTypeFromURLEncoded(),
	).
		Route("POST /",
			http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				err := req.ParseForm()
				if err != nil {
					http.Error(rw, err.Error(), http.StatusBadRequest)
					return
				}

				data := req.PostForm.Get("request")

				var rr internal.RequestEnvelope[json.RawMessage]

				err = json.Unmarshal([]byte(data), &rr)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusBadRequest)
					return
				}

				refs, ok := mapping[rr.Request.Command]
				if !ok {
					http.Error(rw, fmt.Sprintf("command undefined: %s", rr.Request.Command), http.StatusNotFound)
					return
				}

				if refs.request == "" {
					refs.request = rr.Request.Command
				}

				if refs.response == "" {
					refs.response = rr.Request.Command
				}

				file, err := os.ReadFile(filepath.Join("internal", "fixtures", refs.request+"_request.json"))
				if err != nil {
					http.Error(rw, err.Error(), http.StatusBadRequest)
					return
				}

				var expected bytes.Buffer

				err = json.Compact(&expected, file)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusBadRequest)
					return
				}

				form := regexp.MustCompile(`"auth":"\w+",`).
					ReplaceAllString(data, `"auth":"xxx",`)

				if form != expected.String() {
					log.Error("invalid form data",
						slog.String("command", rr.Request.Command),
						slog.String("form", form),
						slog.String("expected", expected.String()),
					)
					http.Error(rw, fmt.Sprintf("invalid form data: %s", data), http.StatusBadRequest)

					return
				}

				servermock.ResponseFromInternal(refs.response+".json").ServeHTTP(rw, req)
			}),
		)
}

func TestDNSProvider_Present_add(t *testing.T) {
	provider := mockBuilder(
		map[string]cmdMapping{
			internal.CommandDNSRowsList:     {},
			internal.CommandDNSRowAdd:       {},
			internal.CommandDNSDomainCommit: {},
		}).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_Present_update(t *testing.T) {
	provider := mockBuilder(
		map[string]cmdMapping{
			internal.CommandDNSRowsList: {
				response: internal.CommandDNSRowsList + "_1",
			},
			internal.CommandDNSRowUpdate:    {},
			internal.CommandDNSDomainCommit: {},
		}).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder(
		map[string]cmdMapping{
			internal.CommandDNSRowsList: {
				response: internal.CommandDNSRowsList + "_1",
			},
			internal.CommandDNSRowDelete:    {},
			internal.CommandDNSDomainCommit: {},
		}).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
