package selectelv2

import (
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(envUsernameOS, envPasswordOS, envAccount, envProjectId).
	WithDomain(envDomain).
	WithLiveTestRequirements(envUsernameOS, envPasswordOS, envAccount, envProjectId)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc    string
		envVars map[string]string
		err     error
	}{
		{
			desc: "OK",
			envVars: map[string]string{
				envUsernameOS: "someName",
				envPasswordOS: "qwerty",
				envAccount:    "1",
				envProjectId:  "111a11111aaa11aa1a11aaa11111aa1a",
			},
		},
		{
			desc: "Fail;No username",
			envVars: map[string]string{
				envPasswordOS: "qwerty",
				envAccount:    "1",
				envProjectId:  "111a11111aaa11aa1a11aaa11111aa1a",
			},
			err: UsernameMissingErr,
		},
		{
			desc: "Fail;No password",
			envVars: map[string]string{
				envUsernameOS: "someName",
				envAccount:    "1",
				envProjectId:  "111a11111aaa11aa1a11aaa11111aa1a",
			},
			err: PasswordMissingErr,
		},
		{
			desc: "Fail;No account",
			envVars: map[string]string{
				envUsernameOS: "someName",
				envPasswordOS: "qwerty",
				envProjectId:  "111a11111aaa11aa1a11aaa11111aa1a",
			},
			err: AccountMissingErr,
		},
		{
			desc: "Fail;No project",
			envVars: map[string]string{
				envUsernameOS: "someName",
				envPasswordOS: "qwerty",
				envAccount:    "1",
			},
			err: ProjectMissingErr,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()
			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()
			if test.err != nil {
				assert.Nil(t, p)
				assert.EqualError(t, &SelectelError{test.err}, err.Error())
			} else {
				//
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

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
