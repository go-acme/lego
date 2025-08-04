package dns01

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
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
	server := tester.MockACMEServer().BuildHTTPS(t)

	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	core, err := api.New(server.Client(), "lego-test", server.URL+"/dir", "", privateKey)
	require.NoError(t, err)

	testCases := []struct {
		desc        string
		validate    ValidateFunc
		preCheck    WrapPreCheckFunc
		provider    challenge.Provider
		expectError bool
	}{
		{
			desc:     "success",
			validate: func(_ *api.Core, _ string, _ acme.Challenge) error { return nil },
			preCheck: func(_, _, _ string, _ PreCheckFunc) (bool, error) { return true, nil },
			provider: &providerMock{},
		},
		{
			desc:     "validate fail",
			validate: func(_ *api.Core, _ string, _ acme.Challenge) error { return errors.New("OOPS") },
			preCheck: func(_, _, _ string, _ PreCheckFunc) (bool, error) { return true, nil },
			provider: &providerMock{
				present: nil,
				cleanUp: nil,
			},
		},
		{
			desc:     "preCheck fail",
			validate: func(_ *api.Core, _ string, _ acme.Challenge) error { return nil },
			preCheck: func(_, _, _ string, _ PreCheckFunc) (bool, error) { return false, errors.New("OOPS") },
			provider: &providerTimeoutMock{
				timeout:  2 * time.Second,
				interval: 500 * time.Millisecond,
			},
		},
		{
			desc:     "present fail",
			validate: func(_ *api.Core, _ string, _ acme.Challenge) error { return nil },
			preCheck: func(_, _, _ string, _ PreCheckFunc) (bool, error) { return true, nil },
			provider: &providerMock{
				present: errors.New("OOPS"),
			},
			expectError: true,
		},
		{
			desc:     "cleanUp fail",
			validate: func(_ *api.Core, _ string, _ acme.Challenge) error { return nil },
			preCheck: func(_, _, _ string, _ PreCheckFunc) (bool, error) { return true, nil },
			provider: &providerMock{
				cleanUp: errors.New("OOPS"),
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			chlg := NewChallenge(core, test.validate, test.provider, WrapPreCheck(test.preCheck))

			authz := acme.Authorization{
				Identifier: acme.Identifier{
					Value: "example.com",
				},
				Challenges: []acme.Challenge{
					{Type: challenge.DNS01.String()},
				},
			}

			err = chlg.PreSolve(authz)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestChallenge_Solve(t *testing.T) {
	server := tester.MockACMEServer().BuildHTTPS(t)

	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	core, err := api.New(server.Client(), "lego-test", server.URL+"/dir", "", privateKey)
	require.NoError(t, err)

	testCases := []struct {
		desc        string
		validate    ValidateFunc
		preCheck    WrapPreCheckFunc
		provider    challenge.Provider
		expectError bool
	}{
		{
			desc:     "success",
			validate: func(_ *api.Core, _ string, _ acme.Challenge) error { return nil },
			preCheck: func(_, _, _ string, _ PreCheckFunc) (bool, error) { return true, nil },
			provider: &providerMock{},
		},
		{
			desc:     "validate fail",
			validate: func(_ *api.Core, _ string, _ acme.Challenge) error { return errors.New("OOPS") },
			preCheck: func(_, _, _ string, _ PreCheckFunc) (bool, error) { return true, nil },
			provider: &providerMock{
				present: nil,
				cleanUp: nil,
			},
			expectError: true,
		},
		{
			desc:     "preCheck fail",
			validate: func(_ *api.Core, _ string, _ acme.Challenge) error { return nil },
			preCheck: func(_, _, _ string, _ PreCheckFunc) (bool, error) { return false, errors.New("OOPS") },
			provider: &providerTimeoutMock{
				timeout:  2 * time.Second,
				interval: 500 * time.Millisecond,
			},
			expectError: true,
		},
		{
			desc:     "present fail",
			validate: func(_ *api.Core, _ string, _ acme.Challenge) error { return nil },
			preCheck: func(_, _, _ string, _ PreCheckFunc) (bool, error) { return true, nil },
			provider: &providerMock{
				present: errors.New("OOPS"),
			},
		},
		{
			desc:     "cleanUp fail",
			validate: func(_ *api.Core, _ string, _ acme.Challenge) error { return nil },
			preCheck: func(_, _, _ string, _ PreCheckFunc) (bool, error) { return true, nil },
			provider: &providerMock{
				cleanUp: errors.New("OOPS"),
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			var options []ChallengeOption
			if test.preCheck != nil {
				options = append(options, WrapPreCheck(test.preCheck))
			}
			chlg := NewChallenge(core, test.validate, test.provider, options...)

			authz := acme.Authorization{
				Identifier: acme.Identifier{
					Value: "example.com",
				},
				Challenges: []acme.Challenge{
					{Type: challenge.DNS01.String()},
				},
			}

			err = chlg.Solve(authz)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestChallenge_CleanUp(t *testing.T) {
	server := tester.MockACMEServer().BuildHTTPS(t)

	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	core, err := api.New(server.Client(), "lego-test", server.URL+"/dir", "", privateKey)
	require.NoError(t, err)

	testCases := []struct {
		desc        string
		validate    ValidateFunc
		preCheck    WrapPreCheckFunc
		provider    challenge.Provider
		expectError bool
	}{
		{
			desc:     "success",
			validate: func(_ *api.Core, _ string, _ acme.Challenge) error { return nil },
			preCheck: func(_, _, _ string, _ PreCheckFunc) (bool, error) { return true, nil },
			provider: &providerMock{},
		},
		{
			desc:     "validate fail",
			validate: func(_ *api.Core, _ string, _ acme.Challenge) error { return errors.New("OOPS") },
			preCheck: func(_, _, _ string, _ PreCheckFunc) (bool, error) { return true, nil },
			provider: &providerMock{
				present: nil,
				cleanUp: nil,
			},
		},
		{
			desc:     "preCheck fail",
			validate: func(_ *api.Core, _ string, _ acme.Challenge) error { return nil },
			preCheck: func(_, _, _ string, _ PreCheckFunc) (bool, error) { return false, errors.New("OOPS") },
			provider: &providerTimeoutMock{
				timeout:  2 * time.Second,
				interval: 500 * time.Millisecond,
			},
		},
		{
			desc:     "present fail",
			validate: func(_ *api.Core, _ string, _ acme.Challenge) error { return nil },
			preCheck: func(_, _, _ string, _ PreCheckFunc) (bool, error) { return true, nil },
			provider: &providerMock{
				present: errors.New("OOPS"),
			},
		},
		{
			desc:     "cleanUp fail",
			validate: func(_ *api.Core, _ string, _ acme.Challenge) error { return nil },
			preCheck: func(_, _, _ string, _ PreCheckFunc) (bool, error) { return true, nil },
			provider: &providerMock{
				cleanUp: errors.New("OOPS"),
			},
			expectError: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			chlg := NewChallenge(core, test.validate, test.provider, WrapPreCheck(test.preCheck))

			authz := acme.Authorization{
				Identifier: acme.Identifier{
					Value: "example.com",
				},
				Challenges: []acme.Challenge{
					{Type: challenge.DNS01.String()},
				},
			}

			err = chlg.CleanUp(authz)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
