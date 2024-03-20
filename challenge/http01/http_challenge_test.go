package http01

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderServer_GetAddress(t *testing.T) {
	dir := t.TempDir()
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	sock := filepath.Join(dir, "var", "run", "test")

	testCases := []struct {
		desc     string
		server   *ProviderServer
		expected string
	}{
		{
			desc:     "TCP default address",
			server:   NewProviderServer("", ""),
			expected: ":80",
		},
		{
			desc:     "TCP with explicit port",
			server:   NewProviderServer("", "8080"),
			expected: ":8080",
		},
		{
			desc:     "TCP with host and port",
			server:   NewProviderServer("localhost", "8080"),
			expected: "localhost:8080",
		},
		{
			desc:     "UDS socket",
			server:   NewUnixProviderServer(sock, fs.ModeSocket|0o666),
			expected: sock,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			address := test.server.GetAddress()
			assert.Equal(t, test.expected, address)
		})
	}
}

func TestChallenge(t *testing.T) {
	_, apiURL := tester.SetupFakeAPI(t)

	providerServer := NewProviderServer("", "23457")

	validate := func(_ *api.Core, _ string, chlng acme.Challenge) error {
		uri := "http://localhost" + providerServer.GetAddress() + ChallengePath(chlng.Token)

		resp, err := http.DefaultClient.Get(uri)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if want := "text/plain"; resp.Header.Get("Content-Type") != want {
			t.Errorf("Get(%q) Content-Type: got %q, want %q", uri, resp.Header.Get("Content-Type"), want)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		bodyStr := string(body)

		if bodyStr != chlng.KeyAuthorization {
			t.Errorf("Get(%q) Body: got %q, want %q", uri, bodyStr, chlng.KeyAuthorization)
		}

		return nil
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", privateKey)
	require.NoError(t, err)

	solver := NewChallenge(core, validate, providerServer)

	authz := acme.Authorization{
		Identifier: acme.Identifier{
			Value: "localhost:23457",
		},
		Challenges: []acme.Challenge{
			{Type: challenge.HTTP01.String(), Token: "http1"},
		},
	}

	err = solver.Solve(authz)
	require.NoError(t, err)
}

func TestChallengeUnix(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("only for UNIX systems")
	}

	_, apiURL := tester.SetupFakeAPI(t)

	dir := t.TempDir()
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	socket := filepath.Join(dir, "lego-challenge-test.sock")

	providerServer := NewUnixProviderServer(socket, fs.ModeSocket|0o666)

	validate := func(_ *api.Core, _ string, chlng acme.Challenge) error {
		// any uri will do, as we hijack the dial
		uri := "http://localhost" + ChallengePath(chlng.Token)

		client := &http.Client{Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", socket)
			},
		}}

		resp, err := client.Get(uri)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		if want := "text/plain"; resp.Header.Get("Content-Type") != want {
			t.Errorf("Get(%q) Content-Type: got %q, want %q", uri, resp.Header.Get("Content-Type"), want)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		bodyStr := string(body)

		if bodyStr != chlng.KeyAuthorization {
			t.Errorf("Get(%q) Body: got %q, want %q", uri, bodyStr, chlng.KeyAuthorization)
		}

		return nil
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", privateKey)
	require.NoError(t, err)

	solver := NewChallenge(core, validate, providerServer)

	authz := acme.Authorization{
		Identifier: acme.Identifier{
			Value: "localhost",
		},
		Challenges: []acme.Challenge{
			{Type: challenge.HTTP01.String(), Token: "http1"},
		},
	}

	err = solver.Solve(authz)
	require.NoError(t, err)
}

func TestChallengeInvalidPort(t *testing.T) {
	_, apiURL := tester.SetupFakeAPI(t)

	privateKey, err := rsa.GenerateKey(rand.Reader, 128)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", privateKey)
	require.NoError(t, err)

	validate := func(_ *api.Core, _ string, _ acme.Challenge) error { return nil }

	solver := NewChallenge(core, validate, NewProviderServer("", "123456"))

	authz := acme.Authorization{
		Identifier: acme.Identifier{
			Value: "localhost:123456",
		},
		Challenges: []acme.Challenge{
			{Type: challenge.HTTP01.String(), Token: "http2"},
		},
	}

	err = solver.Solve(authz)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid port")
	assert.Contains(t, err.Error(), "123456")
}

type testProxyHeader struct {
	name   string
	values []string
}

func (h *testProxyHeader) update(r *http.Request) {
	if h == nil || len(h.values) == 0 {
		return
	}
	if h.name == "Host" {
		r.Host = h.values[0]
	} else if h.name != "" {
		r.Header[h.name] = h.values
	}
}

func TestChallengeWithProxy(t *testing.T) {
	h := func(name string, values ...string) *testProxyHeader {
		name = textproto.CanonicalMIMEHeaderKey(name)
		return &testProxyHeader{name, values}
	}

	const (
		ok   = "localhost:23457"
		nook = "example.com"
	)

	testCases := []struct {
		name   string
		header *testProxyHeader
		extra  *testProxyHeader
		isErr  bool
	}{
		// tests for hostMatcher
		{
			name: "no proxy",
		},
		{
			name:   "empty string",
			header: h(""),
		},
		{
			name:   "empty Host",
			header: h("host"),
		},
		{
			name:   "matching Host",
			header: h("host", ok),
		},
		{
			name:   "Host mismatch",
			header: h("host", nook),
			isErr:  true,
		},
		{
			name:   "Host mismatch (ignoring forwarding header)",
			header: h("host", nook),
			extra:  h("X-Forwarded-Host", ok),
			isErr:  true,
		},
		// test for arbitraryMatcher
		{
			name:   "matching X-Forwarded-Host",
			header: h("X-Forwarded-Host", ok),
		},
		{
			name:   "matching X-Forwarded-Host (multiple fields)",
			header: h("X-Forwarded-Host", ok, nook),
		},
		{
			name:   "matching X-Forwarded-Host (chain value)",
			header: h("X-Forwarded-Host", ok+", "+nook),
		},
		{
			name:   "X-Forwarded-Host mismatch",
			header: h("X-Forwarded-Host", nook),
			extra:  h("host", ok),
			isErr:  true,
		},
		{
			name:   "X-Forwarded-Host mismatch (multiple fields)",
			header: h("X-Forwarded-Host", nook, ok),
			isErr:  true,
		},
		{
			name:   "matching X-Something-Else",
			header: h("X-Something-Else", ok),
		},
		{
			name:   "matching X-Something-Else (multiple fields)",
			header: h("X-Something-Else", ok, nook),
		},
		{
			name:   "matching X-Something-Else (chain value)",
			header: h("X-Something-Else", ok+", "+nook),
		},
		{
			name:   "X-Something-Else mismatch",
			header: h("X-Something-Else", nook),
			isErr:  true,
		},
		{
			name:   "X-Something-Else mismatch (multiple fields)",
			header: h("X-Something-Else", nook, ok),
			isErr:  true,
		},
		{
			name:   "X-Something-Else mismatch (chain value)",
			header: h("X-Something-Else", nook+", "+ok),
			isErr:  true,
		},
		// tests for forwardedHeader
		{
			name:   "matching Forwarded",
			header: h("Forwarded", fmt.Sprintf("host=%q;foo=bar", ok)),
		},
		{
			name:   "matching Forwarded (multiple fields)",
			header: h("Forwarded", fmt.Sprintf("host=%q", ok), "host="+nook),
		},
		{
			name:   "matching Forwarded (chain value)",
			header: h("Forwarded", fmt.Sprintf("host=%q, host=%s", ok, nook)),
		},
		{
			name:   "Forwarded mismatch",
			header: h("Forwarded", "host="+nook),
			isErr:  true,
		},
		{
			name:   "Forwarded mismatch (missing information)",
			header: h("Forwarded", "for=127.0.0.1"),
			isErr:  true,
		},
		{
			name:   "Forwarded mismatch (multiple fields)",
			header: h("Forwarded", "host="+nook, fmt.Sprintf("host=%q", ok)),
			isErr:  true,
		},
		{
			name:   "Forwarded mismatch (chain value)",
			header: h("Forwarded", fmt.Sprintf("host=%s, host=%q", nook, ok)),
			isErr:  true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			testServeWithProxy(t, test.header, test.extra, test.isErr)
		})
	}
}

func testServeWithProxy(t *testing.T, header, extra *testProxyHeader, expectError bool) {
	t.Helper()

	_, apiURL := tester.SetupFakeAPI(t)

	providerServer := NewProviderServer("localhost", "23457")
	if header != nil {
		providerServer.SetProxyHeader(header.name)
	}

	validate := func(_ *api.Core, _ string, chlng acme.Challenge) error {
		uri := "http://" + providerServer.GetAddress() + ChallengePath(chlng.Token)

		req, err := http.NewRequest(http.MethodGet, uri, nil)
		if err != nil {
			return err
		}
		header.update(req)
		extra.update(req)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if want := "text/plain"; resp.Header.Get("Content-Type") != want {
			return fmt.Errorf("Get(%q) Content-Type: got %q, want %q", uri, resp.Header.Get("Content-Type"), want)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		bodyStr := string(body)

		if bodyStr != chlng.KeyAuthorization {
			return fmt.Errorf("Get(%q) Body: got %q, want %q", uri, bodyStr, chlng.KeyAuthorization)
		}

		return nil
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", privateKey)
	require.NoError(t, err)

	solver := NewChallenge(core, validate, providerServer)

	authz := acme.Authorization{
		Identifier: acme.Identifier{
			Value: "localhost:23457",
		},
		Challenges: []acme.Challenge{
			{Type: challenge.HTTP01.String(), Token: "http1"},
		},
	}

	err = solver.Solve(authz)
	if expectError {
		require.Error(t, err)
	} else {
		require.NoError(t, err)
	}
}
