//go:build lego_dnslight

package registry

import (
	"slices"
	"testing"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func noop() (challenge.Provider, error) {
	return nil, nil
}

func TestStore_Add(t *testing.T) {
	s := New()

	s.Add("a", noop, "b", "c")
	s.Add("x", noop)
	s.Add("y", noop)

	assert.Len(t, s.providers, 3)
	assert.Len(t, s.providerAliases, 2)
}

func TestStore_Get(t *testing.T) {
	s := New()

	s.Add("a", noop, "b", "c")

	prv, err := s.Get("a")
	require.NoError(t, err)
	assert.NotNil(t, prv)

	prv, err = s.Get("b")
	require.NoError(t, err)
	assert.NotNil(t, prv)

	prv, err = s.Get("c")
	require.NoError(t, err)
	assert.NotNil(t, prv)

	prv, err = s.Get("z")
	require.EqualError(t, err, "not found")
	assert.Nil(t, prv)
}

func TestStore_GetCodes(t *testing.T) {
	s := New()

	s.Add("a", noop, "b", "c")
	s.Add("x", noop)
	s.Add("y", noop)

	codes := s.GetCodes()
	slices.Sort(codes)

	assert.Equal(t, []string{"a", "x", "y"}, codes)
}
