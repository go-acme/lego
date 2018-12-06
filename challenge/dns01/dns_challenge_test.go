package dns01

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/challenge"
	"github.com/xenolf/lego/le"
	"github.com/xenolf/lego/le/api"
	"github.com/xenolf/lego/platform/tester"
)

type providerMock struct {
	present, cleanUp error
}

func (p *providerMock) Present(domain, token, keyAuth string) error { return p.present }
func (p *providerMock) CleanUp(domain, token, keyAuth string) error { return p.cleanUp }

type providerTimeoutMock struct {
	present, cleanUp  error
	timeout, interval time.Duration
}

func (p *providerTimeoutMock) Present(domain, token, keyAuth string) error { return p.present }
func (p *providerTimeoutMock) CleanUp(domain, token, keyAuth string) error { return p.cleanUp }
func (p *providerTimeoutMock) Timeout() (time.Duration, time.Duration)     { return p.timeout, p.interval }

func TestChallenge_PreSolve(t *testing.T) {
	_, apiURL, tearDown := tester.SetupFakeAPI()
	defer tearDown()

	privKey, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err)

	core, err := api.New(http.DefaultClient, "lego-test", apiURL, "", privKey)
	require.NoError(t, err)

	testCases := []struct {
		desc        string
		validate    ValidateFunc
		preCheck    PreCheckFunc
		provider    challenge.Provider
		expectError bool
	}{
		{
			desc:     "success",
			validate: func(_ *api.Core, _, _ string, _ le.Challenge) error { return nil },
			preCheck: func(_, _ string) (bool, error) { return true, nil },
			provider: &providerMock{},
		},
		{
			desc:     "validate fail",
			validate: func(_ *api.Core, _, _ string, _ le.Challenge) error { return errors.New("OOPS") },
			preCheck: func(_, _ string) (bool, error) { return true, nil },
			provider: &providerMock{
				present: nil,
				cleanUp: nil,
			},
		},
		{
			desc:     "preCheck fail",
			validate: func(_ *api.Core, _, _ string, _ le.Challenge) error { return nil },
			preCheck: func(_, _ string) (bool, error) { return false, errors.New("OOPS") },
			provider: &providerTimeoutMock{
				timeout:  2 * time.Second,
				interval: 500 * time.Millisecond,
			},
		},
		{
			desc:     "present fail",
			validate: func(_ *api.Core, _, _ string, _ le.Challenge) error { return nil },
			preCheck: func(_, _ string) (bool, error) { return true, nil },
			provider: &providerMock{
				present: errors.New("OOPS"),
			},
			expectError: true,
		},
		{
			desc:     "cleanUp fail",
			validate: func(_ *api.Core, _, _ string, _ le.Challenge) error { return nil },
			preCheck: func(_, _ string) (bool, error) { return true, nil },
			provider: &providerMock{
				cleanUp: errors.New("OOPS"),
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			chlg := NewChallenge(core, test.validate, test.provider, AddPreCheck(test.preCheck))

			err = chlg.PreSolve(le.Challenge{}, "example.com")
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestChallenge_Solve(t *testing.T) {
	_, apiURL, tearDown := tester.SetupFakeAPI()
	defer tearDown()

	privKey, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err)

	core, err := api.New(http.DefaultClient, "lego-test", apiURL, "", privKey)
	require.NoError(t, err)

	testCases := []struct {
		desc        string
		validate    ValidateFunc
		preCheck    PreCheckFunc
		provider    challenge.Provider
		expectError bool
	}{
		{
			desc:     "success",
			validate: func(_ *api.Core, _, _ string, _ le.Challenge) error { return nil },
			preCheck: func(_, _ string) (bool, error) { return true, nil },
			provider: &providerMock{},
		},
		{
			desc:     "validate fail",
			validate: func(_ *api.Core, _, _ string, _ le.Challenge) error { return errors.New("OOPS") },
			preCheck: func(_, _ string) (bool, error) { return true, nil },
			provider: &providerMock{
				present: nil,
				cleanUp: nil,
			},
			expectError: true,
		},
		{
			desc:     "preCheck fail",
			validate: func(_ *api.Core, _, _ string, _ le.Challenge) error { return nil },
			preCheck: func(_, _ string) (bool, error) { return false, errors.New("OOPS") },
			provider: &providerTimeoutMock{
				timeout:  2 * time.Second,
				interval: 500 * time.Millisecond,
			},
			expectError: true,
		},
		{
			desc:     "present fail",
			validate: func(_ *api.Core, _, _ string, _ le.Challenge) error { return nil },
			preCheck: func(_, _ string) (bool, error) { return true, nil },
			provider: &providerMock{
				present: errors.New("OOPS"),
			},
		},
		{
			desc:     "cleanUp fail",
			validate: func(_ *api.Core, _, _ string, _ le.Challenge) error { return nil },
			preCheck: func(_, _ string) (bool, error) { return true, nil },
			provider: &providerMock{
				cleanUp: errors.New("OOPS"),
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			chlg := NewChallenge(core, test.validate, test.provider, AddPreCheck(test.preCheck))

			err = chlg.Solve(le.Challenge{}, "example.com")
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestChallenge_CleanUp(t *testing.T) {
	_, apiURL, tearDown := tester.SetupFakeAPI()
	defer tearDown()

	privKey, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err)

	core, err := api.New(http.DefaultClient, "lego-test", apiURL, "", privKey)
	require.NoError(t, err)

	testCases := []struct {
		desc        string
		validate    ValidateFunc
		preCheck    PreCheckFunc
		provider    challenge.Provider
		expectError bool
	}{
		{
			desc:     "success",
			validate: func(_ *api.Core, _, _ string, _ le.Challenge) error { return nil },
			preCheck: func(_, _ string) (bool, error) { return true, nil },
			provider: &providerMock{},
		},
		{
			desc:     "validate fail",
			validate: func(_ *api.Core, _, _ string, _ le.Challenge) error { return errors.New("OOPS") },
			preCheck: func(_, _ string) (bool, error) { return true, nil },
			provider: &providerMock{
				present: nil,
				cleanUp: nil,
			},
		},
		{
			desc:     "preCheck fail",
			validate: func(_ *api.Core, _, _ string, _ le.Challenge) error { return nil },
			preCheck: func(_, _ string) (bool, error) { return false, errors.New("OOPS") },
			provider: &providerTimeoutMock{
				timeout:  2 * time.Second,
				interval: 500 * time.Millisecond,
			},
		},
		{
			desc:     "present fail",
			validate: func(_ *api.Core, _, _ string, _ le.Challenge) error { return nil },
			preCheck: func(_, _ string) (bool, error) { return true, nil },
			provider: &providerMock{
				present: errors.New("OOPS"),
			},
		},
		{
			desc:     "cleanUp fail",
			validate: func(_ *api.Core, _, _ string, _ le.Challenge) error { return nil },
			preCheck: func(_, _ string) (bool, error) { return true, nil },
			provider: &providerMock{
				cleanUp: errors.New("OOPS"),
			},
			expectError: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			chlg := NewChallenge(core, test.validate, test.provider, AddPreCheck(test.preCheck))

			err = chlg.CleanUp(le.Challenge{}, "example.com")
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
