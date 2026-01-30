package resolver

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProber_Solve(t *testing.T) {
	testCases := []struct {
		desc             string
		solvers          map[challenge.Type]solver
		authz            []acme.Authorization
		expectedError    string
		expectedCounters map[challenge.Type]string
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
			expectedCounters: map[challenge.Type]string{
				challenge.HTTP01: "PreSolve: 3, Solve: 3, CleanUp: 3",
			},
		},
		{
			desc: "DNS-01 deduplicate",
			solvers: map[challenge.Type]solver{
				challenge.DNS01: &preSolverMock{
					preSolve: map[string]error{},
					solve:    map[string]error{},
					cleanUp:  map[string]error{},
				},
			},
			authz: []acme.Authorization{
				createStubAuthorizationDNS01("a.example", false),
				createStubAuthorizationDNS01("a.example", true),
				createStubAuthorizationDNS01("b.example", false),
				createStubAuthorizationDNS01("b.example", true),
				createStubAuthorizationDNS01("c.example", true),
				createStubAuthorizationDNS01("d.example", false),
			},
			expectedCounters: map[challenge.Type]string{
				challenge.DNS01: "PreSolve: 4, Solve: 6, CleanUp: 4",
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
			expectedCounters: map[challenge.Type]string{
				challenge.HTTP01: "PreSolve: 0, Solve: 0, CleanUp: 0",
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
			expectedCounters: map[challenge.Type]string{
				challenge.HTTP01: "PreSolve: 3, Solve: 2, CleanUp: 3",
			},
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
			expectedCounters: map[challenge.Type]string{
				challenge.HTTP01: "PreSolve: 3, Solve: 2, CleanUp: 3",
			},
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

			for n, s := range test.solvers {
				assert.Equal(t, test.expectedCounters[n], fmt.Sprintf("%s", s))
			}
		})
	}
}
