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
				createStubAuthorizationHTTP01("acme.wtf", acme.StatusProcessing),
				createStubAuthorizationHTTP01("lego.wtf", acme.StatusProcessing),
				createStubAuthorizationHTTP01("mydomain.wtf", acme.StatusProcessing),
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
				createStubAuthorizationHTTP01("acme.wtf", acme.StatusValid),
				createStubAuthorizationHTTP01("lego.wtf", acme.StatusValid),
				createStubAuthorizationHTTP01("mydomain.wtf", acme.StatusValid),
			},
		},
		{
			desc: "when preSolve fail, auth is flagged as error and skipped",
			solvers: map[challenge.Type]solver{
				challenge.HTTP01: &preSolverMock{
					preSolve: map[string]error{
						"acme.wtf": errors.New("preSolve error acme.wtf"),
					},
					solve: map[string]error{
						"acme.wtf": errors.New("solve error acme.wtf"),
					},
					cleanUp: map[string]error{
						"acme.wtf": errors.New("clean error acme.wtf"),
					},
				},
			},
			authz: []acme.Authorization{
				createStubAuthorizationHTTP01("acme.wtf", acme.StatusProcessing),
				createStubAuthorizationHTTP01("lego.wtf", acme.StatusProcessing),
				createStubAuthorizationHTTP01("mydomain.wtf", acme.StatusProcessing),
			},
			expectedError: `error: one or more domains had a problem:
[acme.wtf] preSolve error acme.wtf
`,
		},
		{
			desc: "errors at different stages",
			solvers: map[challenge.Type]solver{
				challenge.HTTP01: &preSolverMock{
					preSolve: map[string]error{
						"acme.wtf": errors.New("preSolve error acme.wtf"),
					},
					solve: map[string]error{
						"acme.wtf": errors.New("solve error acme.wtf"),
						"lego.wtf": errors.New("solve error lego.wtf"),
					},
					cleanUp: map[string]error{
						"mydomain.wtf": errors.New("clean error mydomain.wtf"),
					},
				},
			},
			authz: []acme.Authorization{
				createStubAuthorizationHTTP01("acme.wtf", acme.StatusProcessing),
				createStubAuthorizationHTTP01("lego.wtf", acme.StatusProcessing),
				createStubAuthorizationHTTP01("mydomain.wtf", acme.StatusProcessing),
			},
			expectedError: `error: one or more domains had a problem:
[acme.wtf] preSolve error acme.wtf
[lego.wtf] solve error lego.wtf
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
