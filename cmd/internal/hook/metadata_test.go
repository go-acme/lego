package hook

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_metaToEnv(t *testing.T) {
	env := metaToEnv(map[string]string{
		"foo": "bar",
	})

	expected := []string{"foo=bar"}

	assert.Equal(t, expected, env)
}
