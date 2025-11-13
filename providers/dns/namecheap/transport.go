package namecheap

import (
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/go-acme/lego/v4/platform/config/env"
	"golang.org/x/net/http/httpproxy"
)

const (
	envHTTPProxy       = "HTTP_PROXY"
	envHTTPProxyLower  = "http_proxy"
	envHTTPSProxy      = "HTTPS_PROXY"
	envHTTPSProxyLower = "https_proxy"
	envNoProxy         = "NO_PROXY"
	envNoProxyLower    = "no_proxy"
	envRequestMethod   = "REQUEST_METHOD"
)

// Allows lazy loading of the proxy.
var (
	envProxyOnce      sync.Once
	envProxyFuncValue func(*url.URL) (*url.URL, error)
)

func defaultTransport(namespace string) http.RoundTripper {
	tr, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil
	}

	clone := tr.Clone()
	clone.Proxy = proxyFromEnvironment(namespace)

	return clone
}

// Inspired by:
// - https://pkg.go.dev/net/http#ProxyFromEnvironment
// - https://pkg.go.dev/golang.org/x/net/http/httpproxy#FromEnvironment
func envProxyFunc(namespace string) func(*url.URL) (*url.URL, error) {
	envProxyOnce.Do(func() {
		cfg := &httpproxy.Config{
			HTTPProxy:  getEnv(namespace, envHTTPProxy, envHTTPProxyLower),
			HTTPSProxy: getEnv(namespace, envHTTPSProxy, envHTTPSProxyLower),
			NoProxy:    getEnv(namespace, envNoProxy, envNoProxyLower),
			CGI:        env.GetOneWithFallback(namespace+envRequestMethod, "", env.ParseString, envRequestMethod) != "",
		}

		envProxyFuncValue = cfg.ProxyFunc()
	})

	return envProxyFuncValue
}

// Inspired by:
// - https://pkg.go.dev/net/http#ProxyFromEnvironment
// - https://pkg.go.dev/golang.org/x/net/http/httpproxy#FromEnvironment
func proxyFromEnvironment(namespace string) func(req *http.Request) (*url.URL, error) {
	return func(req *http.Request) (*url.URL, error) {
		return envProxyFunc(namespace)(req.URL)
	}
}

func getEnv(namespace, baseEnvName, baseEnvNameLower string) string {
	return env.GetOneWithFallback(namespace+baseEnvName, "", env.ParseString,
		strings.ToLower(namespace)+baseEnvNameLower, baseEnvName, baseEnvNameLower)
}
