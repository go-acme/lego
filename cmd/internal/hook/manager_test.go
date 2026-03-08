package hook

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/stretchr/testify/require"
)

func Test_Manager(t *testing.T) {
	certificatesStorage := storage.NewCertificatesStorage(t.TempDir())

	testCases := []struct {
		desc    string
		options []Option
	}{
		{
			desc: "all hooks",
			options: []Option{
				WithPre("echo Pre Hook", 1*time.Second),
				WithDeploy("echo Deploy Hook", 1*time.Second),
				WithPost("echo Post Hook", 1*time.Second),
			},
		},
		{
			desc: "pre-hook only",
			options: []Option{
				WithPre("echo Pre Hook", 1*time.Second),
			},
		},
		{
			desc: "deploy-hook only",
			options: []Option{
				WithDeploy("echo Deploy Hook", 1*time.Second),
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
				WithAccountMetadata(&storage.Account{ID: "foo@exmaple.com", Email: "bar@example.com"}),
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			manager := NewManager(certificatesStorage, test.options...)

			err := manager.Pre(t.Context(), "a", []string{"example.com", "example.org"})
			require.NoError(t, err)

			t.Log(manager.metadata)

			err = manager.Deploy(t.Context(), &certificate.Resource{ID: "example.org"}, &storage.SaveOptions{})
			require.NoError(t, err)

			t.Log(manager.metadata)

			err = manager.Post(t.Context())
			require.NoError(t, err)

			t.Log(manager.metadata)
		})
	}
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

			err := manager.Pre(t.Context(), "a", []string{"example.com", "example.org"})
			test.requirePre(t, err)

			err = manager.Deploy(t.Context(), &certificate.Resource{ID: "example.org"}, &storage.SaveOptions{})
			test.requireDeploy(t, err)

			err = manager.Post(t.Context())
			test.requirePost(t, err)
		})
	}
}
