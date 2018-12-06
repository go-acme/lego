package resolver

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/le"

	"github.com/xenolf/lego/challenge"
)

func TestProber_Solve(t *testing.T) {
	testCases := []struct {
		desc          string
		solvers       map[challenge.Type]solver
		authz         []le.Authorization
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
			authz: []le.Authorization{
				createStubAuthorizationHTTP01("acme.wtf", le.StatusProcessing),
				createStubAuthorizationHTTP01("lego.wtf", le.StatusProcessing),
				createStubAuthorizationHTTP01("mydomain.wtf", le.StatusProcessing),
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
			authz: []le.Authorization{
				createStubAuthorizationHTTP01("acme.wtf", le.StatusValid),
				createStubAuthorizationHTTP01("lego.wtf", le.StatusValid),
				createStubAuthorizationHTTP01("mydomain.wtf", le.StatusValid),
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
			authz: []le.Authorization{
				createStubAuthorizationHTTP01("acme.wtf", le.StatusProcessing),
				createStubAuthorizationHTTP01("lego.wtf", le.StatusProcessing),
				createStubAuthorizationHTTP01("mydomain.wtf", le.StatusProcessing),
			},
			expectedError: `acme: Error -> One or more domains had a problem:
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
			authz: []le.Authorization{
				createStubAuthorizationHTTP01("acme.wtf", le.StatusProcessing),
				createStubAuthorizationHTTP01("lego.wtf", le.StatusProcessing),
				createStubAuthorizationHTTP01("mydomain.wtf", le.StatusProcessing),
			},
			expectedError: `acme: Error -> One or more domains had a problem:
[acme.wtf] preSolve error acme.wtf
[lego.wtf] solve error lego.wtf
`,
		},
	}

	for _, test := range testCases {
		test := test
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
