package resolver

import (
	"errors"
	"testing"

	"github.com/go-acme/lego/v4/acme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_obtainError_Error(t *testing.T) {
	err := obtainError{
		"a": &acme.ProblemDetails{Type: "001"},
		"b": errors.New("oops"),
		"c": errors.New("I did it again"),
	}

	require.EqualError(t, err, `error: one or more domains had a problem:
[a] acme: error: 0 :: 001 :: 
[b] oops
[c] I did it again
`)
}

func Test_obtainError_Unwrap(t *testing.T) {
	testCases := []struct {
		desc   string
		err    obtainError
		assert assert.BoolAssertionFunc
	}{
		{
			desc: "one ok",
			err: obtainError{
				"a": &acme.ProblemDetails{},
				"b": errors.New("oops"),
				"c": errors.New("I did it again"),
			},
			assert: assert.True,
		},
		{
			desc: "all ok",
			err: obtainError{
				"a": &acme.ProblemDetails{Type: "001"},
				"b": &acme.ProblemDetails{Type: "002"},
				"c": &acme.ProblemDetails{Type: "002"},
			},
			assert: assert.True,
		},
		{
			desc: "nope",
			err: obtainError{
				"a": errors.New("hello"),
				"b": errors.New("oops"),
				"c": errors.New("I did it again"),
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
