package errutils

import (
	"errors"
	"testing"

	"github.com/go-acme/lego/v5/acme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomainsError_Error(t *testing.T) {
	err := &DomainsError{
		prefix: "resolver",
		data: map[string]error{
			"a": &acme.ProblemDetails{Type: "001"},
			"b": errors.New("oops"),
			"c": errors.New("I did it again"),
		},
	}

	require.EqualError(t, err.Join(), `resolver: one or more domains had a problem: [a: acme: error: 0 :: 001 :: ] [b: oops] [c: I did it again]`)
}

func TestDomainsError_Unwrap(t *testing.T) {
	testCases := []struct {
		desc   string
		err    *DomainsError
		assert assert.BoolAssertionFunc
	}{
		{
			desc: "one ok",
			err: &DomainsError{
				prefix: "resolver",
				data: map[string]error{
					"a": &acme.ProblemDetails{},
					"b": errors.New("oops"),
					"c": errors.New("I did it again"),
				},
			},
			assert: assert.True,
		},
		{
			desc: "all ok",
			err: &DomainsError{
				prefix: "resolver",
				data: map[string]error{
					"a": &acme.ProblemDetails{Type: "001"},
					"b": &acme.ProblemDetails{Type: "002"},
					"c": &acme.ProblemDetails{Type: "002"},
				},
			},
			assert: assert.True,
		},
		{
			desc: "nope",
			err: &DomainsError{
				prefix: "resolver",
				data: map[string]error{
					"a": errors.New("hello"),
					"b": errors.New("oops"),
					"c": errors.New("I did it again"),
				},
			},
			assert: assert.False,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var pd *acme.ProblemDetails

			test.assert(t, errors.As(test.err, &pd))
		})
	}
}

type TomatoError struct{}

func (t TomatoError) Error() string {
	return "tomato"
}

type CarrotError struct{}

func (t CarrotError) Error() string {
	return "carrot"
}

func TestDomainsError_Join(t *testing.T) {
	failures := NewDomainsError("certificates")

	failures.Add("example.com", &TomatoError{})

	err := failures.Join()

	assert.EqualError(t, err, "certificates: one or more domains had a problem: [example.com: tomato]")

	to := &TomatoError{}
	require.ErrorAs(t, err, &to)
}

func TestDomainsError_Join_multiple_domains(t *testing.T) {
	failures := NewDomainsError("certificates")

	failures.Add("example.com", &TomatoError{})
	failures.Add("example.org", &CarrotError{})

	err := failures.Join()

	assert.EqualError(t, err, "certificates: one or more domains had a problem: [example.com: tomato] [example.org: carrot]")

	to := &TomatoError{}
	require.ErrorAs(t, err, &to)

	ca := &CarrotError{}
	require.ErrorAs(t, err, &ca)
}

func TestDomainsError_Join_no_error(t *testing.T) {
	failures := NewDomainsError("certificates")

	require.NoError(t, failures.Join())
}

func TestDomainsError_Join_same_domain(t *testing.T) {
	failures := NewDomainsError("certificates")

	failures.Add("example.com", &TomatoError{})
	failures.Add("example.com", &CarrotError{})

	err := failures.Join()

	assert.EqualError(t, err, "certificates: one or more domains had a problem: [example.com: carrot]")

	to := &TomatoError{}
	if errors.As(err, &to) {
		require.Fail(t, "TomatoError should be overridden by CarrotError")
	}

	ca := &CarrotError{}
	require.ErrorAs(t, err, &ca)
}
