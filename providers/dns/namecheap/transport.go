package namecheap

import (
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/go-acme/lego/v4/platform/config/env"
	"golang.org/x/net/http/httpproxy"
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
			HTTPProxy: env.GetOneWithFallback(namespace+"HTTP_PROXY", "", env.ParseString,
				strings.ToLower(namespace)+"http_proxy", "HTTP_PROXY", "http_proxy"),
			HTTPSProxy: env.GetOneWithFallback(namespace+"HTTPS_PROXY", "", env.ParseString,
				strings.ToLower(namespace)+"https_proxy", "HTTPS_PROXY", "https_proxy"),
			NoProxy: env.GetOneWithFallback(namespace+"NO_PROXY", "", env.ParseString,
				strings.ToLower(namespace)+"no_proxy", "NO_PROXY", "no_proxy"),
			CGI: env.GetOneWithFallback(namespace+"REQUEST_METHOD", "", env.ParseString, "REQUEST_METHOD") != "",
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
