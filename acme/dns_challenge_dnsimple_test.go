package acme

import (
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestNewDNSProviderDNSimpleValid(t *testing.T) {
	_, err := NewDNSProviderDNSimple("example@example.com", "123")
	assert.NoError(t, err)
}
