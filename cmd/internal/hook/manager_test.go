package hook

import (
	"maps"
	"regexp"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Manager(t *testing.T) {
	certificatesStorage := storage.NewCertificatesStorage(t.TempDir())

	testCases := []struct {
		desc           string
		options        []Option
		metadataPre    map[string]*regexp.Regexp
		metadataDeploy map[string]*regexp.Regexp
	}{
		{
			desc: "all hooks",
			options: []Option{
				WithPre("echo Pre Hook", 1*time.Second),
				WithDeploy("echo Deploy Hook", 1*time.Second),
				WithPost("echo Post Hook", 1*time.Second),
			},
			metadataPre: map[string]*regexp.Regexp{
				"LEGO_HOOK_CERT_DOMAINS":        regexp.MustCompile(`example\.com,example\.org`),
				"LEGO_HOOK_CERT_KEY_TYPE":       regexp.MustCompile("EC256"),
				"LEGO_HOOK_CERT_NAME":           regexp.MustCompile("a"),
				"LEGO_HOOK_CERT_NAME_SANITIZED": regexp.MustCompile("a"),
			},
			metadataDeploy: map[string]*regexp.Regexp{
				"LEGO_HOOK_CERT_DOMAINS":        regexp.MustCompile(`example\.net`),
				"LEGO_HOOK_CERT_KEY_TYPE":       regexp.MustCompile("EC384"),
				"LEGO_HOOK_CERT_NAME":           regexp.MustCompile("b"),
				"LEGO_HOOK_CERT_NAME_SANITIZED": regexp.MustCompile("b"),
				"LEGO_HOOK_CERT_PATH":           regexp.MustCompile(`.+[/\\]certificates[/\\]b\.crt`),
				"LEGO_HOOK_CERT_KEY_PATH":       regexp.MustCompile(`.+[/\\]certificates[/\\]b\.key`),
			},
		},
		{
			desc: "pre-hook only",
			options: []Option{
				WithPre("echo Pre Hook", 1*time.Second),
			},
			metadataPre: map[string]*regexp.Regexp{
				"LEGO_HOOK_CERT_DOMAINS":        regexp.MustCompile(`example\.com,example\.org`),
				"LEGO_HOOK_CERT_KEY_TYPE":       regexp.MustCompile("EC256"),
				"LEGO_HOOK_CERT_NAME":           regexp.MustCompile("a"),
				"LEGO_HOOK_CERT_NAME_SANITIZED": regexp.MustCompile("a"),
			},
		},
		{
			desc: "deploy-hook only",
			options: []Option{
				WithDeploy("echo Deploy Hook", 1*time.Second),
			},
			metadataDeploy: map[string]*regexp.Regexp{
				"LEGO_HOOK_CERT_DOMAINS":        regexp.MustCompile(`example\.net`),
				"LEGO_HOOK_CERT_KEY_TYPE":       regexp.MustCompile("EC384"),
				"LEGO_HOOK_CERT_NAME":           regexp.MustCompile("b"),
				"LEGO_HOOK_CERT_NAME_SANITIZED": regexp.MustCompile("b"),
				"LEGO_HOOK_CERT_PATH":           regexp.MustCompile(`.+[/\\]certificates[/\\]b\.crt`),
				"LEGO_HOOK_CERT_KEY_PATH":       regexp.MustCompile(`.+[/\\]certificates[/\\]b\.key`),
			},
		},
		{
			desc: "post-hook only",
			options: []Option{
				WithPost("echo Post Hook", 1*time.Second),
			},
		},
		{
			desc: "no hook",
		},
		{
			desc: "all hooks (metadata)",
			options: []Option{
				WithPre("echo Pre Hook", 1*time.Second),
				WithDeploy("echo Deploy Hook", 1*time.Second),
				WithPost("echo Post Hook", 1*time.Second),
				WithAccountMetadata(&storage.Account{ID: "foo@example.com", Email: "bar@example.com"}),
			},
			metadataPre: map[string]*regexp.Regexp{
				"LEGO_HOOK_CERT_DOMAINS":        regexp.MustCompile(`example\.com,example\.org`),
				"LEGO_HOOK_CERT_KEY_TYPE":       regexp.MustCompile("EC256"),
				"LEGO_HOOK_CERT_NAME":           regexp.MustCompile("a"),
				"LEGO_HOOK_CERT_NAME_SANITIZED": regexp.MustCompile("a"),
				"LEGO_HOOK_ACCOUNT_EMAIL":       regexp.MustCompile(`bar@example\.com`),
				"LEGO_HOOK_ACCOUNT_ID":          regexp.MustCompile(`foo@example\.co`),
				"LEGO_HOOK_ACCOUNT_SERVER":      regexp.MustCompile(`^$`),
			},
			metadataDeploy: map[string]*regexp.Regexp{
				"LEGO_HOOK_CERT_DOMAINS":        regexp.MustCompile(`example\.net`),
				"LEGO_HOOK_CERT_KEY_TYPE":       regexp.MustCompile("EC384"),
				"LEGO_HOOK_CERT_NAME":           regexp.MustCompile("b"),
				"LEGO_HOOK_CERT_NAME_SANITIZED": regexp.MustCompile("b"),
				"LEGO_HOOK_CERT_PATH":           regexp.MustCompile(`.+[/\\]certificates[/\\]b\.crt`),
				"LEGO_HOOK_CERT_KEY_PATH":       regexp.MustCompile(`.+[/\\]certificates[/\\]b\.key`),
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			expectedMetadata := map[string]*regexp.Regexp{}

			manager := NewManager(certificatesStorage, test.options...)

			request := certificate.ObtainRequest{
				Domains: []string{"example.com", "example.org"},
				KeyType: certcrypto.EC256,
			}

			err := manager.PreForDomains(t.Context(), "a", request)
			require.NoError(t, err)

			t.Log("pre", manager.metadata)

			maps.Copy(expectedMetadata, test.metadataPre)

			assertMetadata(t, manager.metadata, expectedMetadata)

			resource := &certificate.Resource{
				ID:      "b",
				Domains: []string{"example.net"},
				KeyType: certcrypto.EC384,
			}

			err = manager.Deploy(t.Context(), resource, &storage.SaveOptions{})
			require.NoError(t, err)

			t.Log("deploy", manager.metadata)

			maps.Copy(expectedMetadata, test.metadataDeploy)

			assertMetadata(t, manager.metadata, expectedMetadata)

			err = manager.Post(t.Context())
			require.NoError(t, err)

			t.Log("post", manager.metadata)

			assertMetadata(t, manager.metadata, expectedMetadata)
		})
	}
}

func assertMetadata(t *testing.T, metadata map[string]string, expected map[string]*regexp.Regexp) {
	t.Helper()

	require.NotNil(t, metadata)

	for k, exp := range expected {
		v, ok := metadata[k]
		require.True(t, ok, k)
		assert.Regexp(t, exp, v, k)
	}

	assert.Len(t, metadata, len(expected))
}

func Test_Manager_errors(t *testing.T) {
	certificatesStorage := storage.NewCertificatesStorage(t.TempDir())

	testCases := []struct {
		desc          string
		options       []Option
		requirePre    require.ErrorAssertionFunc
		requireDeploy require.ErrorAssertionFunc
		requirePost   require.ErrorAssertionFunc
	}{
		{
			desc: "pre-hook error",
			options: []Option{
				WithPre("thisappdoesnotexistpre", 1*time.Second),
				WithDeploy("echo Deploy Hook", 1*time.Second),
				WithPost("echo Post Hook", 1*time.Second),
			},
			requirePre:    require.Error,
			requireDeploy: require.NoError,
			requirePost:   require.NoError,
		},
		{
			desc: "deploy-hook error",
			options: []Option{
				WithPre("echo Pre Hook", 1*time.Second),
				WithDeploy("thiscommanddoesnotexistdeploy", 1*time.Second),
				WithPost("echo Post Hook", 1*time.Second),
			},
			requirePre:    require.NoError,
			requireDeploy: require.Error,
			requirePost:   require.NoError,
		},
		{
			desc: "post-hook error",
			options: []Option{
				WithPre("echo Pre Hook", 1*time.Second),
				WithDeploy("echo Deploy Hook", 1*time.Second),
				WithPost("thiscommanddoesnotexistpost", 1*time.Second),
			},
			requirePre:    require.NoError,
			requireDeploy: require.NoError,
			requirePost:   require.Error,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			manager := NewManager(certificatesStorage, test.options...)

			request := certificate.ObtainRequest{
				Domains: []string{"example.com", "example.org"},
				KeyType: certcrypto.EC256,
			}

			err := manager.PreForDomains(t.Context(), "a", request)
			test.requirePre(t, err)

			err = manager.Deploy(t.Context(), &certificate.Resource{ID: "example.org"}, &storage.SaveOptions{})
			test.requireDeploy(t, err)

			err = manager.Post(t.Context())
			test.requirePost(t, err)
		})
	}
}
