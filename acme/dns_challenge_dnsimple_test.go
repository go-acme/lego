package acme

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDNSProviderDNSimpleValid(t *testing.T) {
	_, err := NewDNSProviderDNSimple("example@example.com", "123")
	assert.NoError(t, err)
}

func TestNewDNSProviderDNSimpleMissingCredErr(t *testing.T) {
	_, err := NewDNSProviderDNSimple("", "")
	assert.EqualError(t, err, "DNSimple credentials missing")
}
