package certificate

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type TomatoError struct{}

func (t TomatoError) Error() string {
	return "tomato"
}

type CarrotError struct{}

func (t CarrotError) Error() string {
	return "carrot"
}

func Test_obtainError_Join(t *testing.T) {
	failures := newObtainError()

	failures.Add("example.com", &TomatoError{})

	err := failures.Join()

	to := &TomatoError{}
	require.ErrorAs(t, err, &to)
}

func Test_obtainError_Join_multiple_domains(t *testing.T) {
	failures := newObtainError()

	failures.Add("example.com", &TomatoError{})
	failures.Add("example.org", &CarrotError{})

	err := failures.Join()

	to := &TomatoError{}
	require.ErrorAs(t, err, &to)

	ca := &CarrotError{}
	require.ErrorAs(t, err, &ca)
}

func Test_obtainError_Join_no_error(t *testing.T) {
	failures := newObtainError()

	require.NoError(t, failures.Join())
}

func Test_obtainError_Join_same_domain(t *testing.T) {
	failures := newObtainError()

	failures.Add("example.com", &TomatoError{})
	failures.Add("example.com", &CarrotError{})

	err := failures.Join()

	to := &TomatoError{}
	if errors.As(err, &to) {
		require.Fail(t, "TomatoError should be overridden by CarrotError")
	}

	ca := &CarrotError{}
	require.ErrorAs(t, err, &ca)
}
