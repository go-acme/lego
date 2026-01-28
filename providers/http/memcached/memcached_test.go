package memcached

import (
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-acme/lego/v5/challenge/http01"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	domain  = "lego.test"
	token   = "foo"
	keyAuth = "bar"
)

func getMemcachedHosts() []string {
	memcachedHostsStr := os.Getenv("MEMCACHED_HOSTS")

	if memcachedHostsStr != "" {
		return strings.Split(memcachedHostsStr, ",")
	}

	return nil
}

func TestNewMemcachedProviderEmpty(t *testing.T) {
	emptyHosts := make([]string, 0)
	_, err := NewMemcachedProvider(emptyHosts)
	require.EqualError(t, err, "no memcached hosts provided")
}

func TestNewMemcachedProviderValid(t *testing.T) {
	memcachedHosts := getMemcachedHosts()

	if len(memcachedHosts) == 0 {
		t.Skip("Skipping memcached tests")
	}

	_, err := NewMemcachedProvider(memcachedHosts)
	require.NoError(t, err)
}

func TestMemcachedPresentSingleHost(t *testing.T) {
	memcachedHosts := getMemcachedHosts()

	if len(memcachedHosts) == 0 {
		t.Skip("Skipping memcached tests")
	}

	p, err := NewMemcachedProvider(memcachedHosts[0:1])
	require.NoError(t, err)

	challengePath := path.Join("/", http01.ChallengePath(token))

	err = p.Present(t.Context(), domain, token, keyAuth)
	require.NoError(t, err)

	mc := memcache.New(memcachedHosts[0])

	i, err := mc.Get(challengePath)
	require.NoError(t, err)
	assert.Equal(t, i.Value, []byte(keyAuth))
}

func TestMemcachedPresentMultiHost(t *testing.T) {
	memcachedHosts := getMemcachedHosts()

	if len(memcachedHosts) <= 1 {
		t.Skip("Skipping memcached multi-host tests")
	}

	p, err := NewMemcachedProvider(memcachedHosts)
	require.NoError(t, err)

	challengePath := path.Join("/", http01.ChallengePath(token))

	err = p.Present(t.Context(), domain, token, keyAuth)
	require.NoError(t, err)

	for _, host := range memcachedHosts {
		mc := memcache.New(host)

		// Only because this is slow on GitHub action.
		mc.Timeout = 1 * time.Second
		mc.MaxIdleConns = memcache.DefaultMaxIdleConns * 2

		i, err := mc.Get(challengePath)
		require.NoError(t, err)
		assert.Equal(t, i.Value, []byte(keyAuth))
	}
}

func TestMemcachedPresentPartialFailureMultiHost(t *testing.T) {
	memcachedHosts := getMemcachedHosts()

	if len(memcachedHosts) == 0 {
		t.Skip("Skipping memcached tests")
	}

	hosts := append(memcachedHosts, "5.5.5.5:11211")
	p, err := NewMemcachedProvider(hosts)
	require.NoError(t, err)

	challengePath := path.Join("/", http01.ChallengePath(token))

	err = p.Present(t.Context(), domain, token, keyAuth)
	require.NoError(t, err)

	for _, host := range memcachedHosts {
		mc := memcache.New(host)
		// Only because this is slow on GitHub action.
		mc.Timeout = 1 * time.Second
		mc.MaxIdleConns = memcache.DefaultMaxIdleConns * 2

		i, err := mc.Get(challengePath)
		require.NoError(t, err)
		assert.Equal(t, i.Value, []byte(keyAuth))
	}
}

func TestMemcachedCleanup(t *testing.T) {
	memcachedHosts := getMemcachedHosts()

	if len(memcachedHosts) == 0 {
		t.Skip("Skipping memcached tests")
	}

	p, err := NewMemcachedProvider(memcachedHosts)
	require.NoError(t, err)
	require.NoError(t, p.CleanUp(t.Context(), domain, token, keyAuth))
}
