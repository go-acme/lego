package internal

import (
	"net/url"
	"testing"

	querystring "github.com/google/go-querystring/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilters_encode(t *testing.T) {
	f := Filters{&RecordFilter{
		Name: "_acme-challenge.foo",
		Type: []string{"TXT"},
	}}

	values, err := querystring.Values(f)
	require.NoError(t, err)

	unescape, err := url.QueryUnescape(values.Encode())
	require.NoError(t, err)

	assert.Equal(t, "filters[name]=_acme-challenge.foo&filters[type][]=TXT", unescape)
}
