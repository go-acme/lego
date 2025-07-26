package resolver

import (
	"errors"
	"testing"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/stretchr/testify/require"
)

func TestProber_Solve(t *testing.T) {
	testCases := []struct {
		desc          string
		solvers       map[challenge.Type]solver
		authz         []acme.Authorization
		expectedError string
	}{
		{
			desc: "success",
			solvers: map[challenge.Type]solver{
				challenge.HTTP01: &preSolverMock{
					preSolve: map[string]error{},
					solve:    map[string]error{},
					cleanUp:  map[string]error{},
				},
			},
			authz: []acme.Authorization{
				createStubAuthorizationHTTP01("example.com", acme.StatusProcessing),
				createStubAuthorizationHTTP01("example.org", acme.StatusProcessing),
				createStubAuthorizationHTTP01("example.net", acme.StatusProcessing),
			},
		},
		{
			desc: "already valid",
			solvers: map[challenge.Type]solver{
				challenge.HTTP01: &preSolverMock{
					preSolve: map[string]error{},
					solve:    map[string]error{},
					cleanUp:  map[string]error{},
				},
			},
			authz: []acme.Authorization{
				createStubAuthorizationHTTP01("example.com", acme.StatusValid),
				createStubAuthorizationHTTP01("example.org", acme.StatusValid),
				createStubAuthorizationHTTP01("example.net", acme.StatusValid),
			},
		},
		{
			desc: "when preSolve fail, auth is flagged as error and skipped",
			solvers: map[challenge.Type]solver{
				challenge.HTTP01: &preSolverMock{
					preSolve: map[string]error{
						"example.com": errors.New("preSolve error example.com"),
					},
					solve: map[string]error{
						"example.com": errors.New("solve error example.com"),
					},
					cleanUp: map[string]error{
						"example.com": errors.New("clean error example.com"),
					},
				},
			},
			authz: []acme.Authorization{
				createStubAuthorizationHTTP01("example.com", acme.StatusProcessing),
				createStubAuthorizationHTTP01("example.org", acme.StatusProcessing),
				createStubAuthorizationHTTP01("example.net", acme.StatusProcessing),
			},
			expectedError: `error: one or more domains had a problem:
[example.com] preSolve error example.com
`,
		},
		{
			desc: "errors at different stages",
			solvers: map[challenge.Type]solver{
				challenge.HTTP01: &preSolverMock{
					preSolve: map[string]error{
						"example.com": errors.New("preSolve error example.com"),
					},
					solve: map[string]error{
						"example.com": errors.New("solve error example.com"),
						"example.org": errors.New("solve error example.org"),
					},
					cleanUp: map[string]error{
						"example.net": errors.New("clean error example.net"),
					},
				},
			},
			authz: []acme.Authorization{
				createStubAuthorizationHTTP01("example.com", acme.StatusProcessing),
				createStubAuthorizationHTTP01("example.org", acme.StatusProcessing),
				createStubAuthorizationHTTP01("example.net", acme.StatusProcessing),
			},
			expectedError: `error: one or more domains had a problem:
[example.com] preSolve error example.com
[example.org] solve error example.org
`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			prober := &Prober{
				solverManager: &SolverManager{solvers: test.solvers},
			}

			err := prober.Solve(test.authz)
			if test.expectedError != "" {
				require.EqualError(t, err, test.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
