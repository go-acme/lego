package internal

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
)

const getIPURL = "https://ifconfig.co/ip"

const (
	envLegoIPv4Only = "LEGO_IPV4ONLY"
	envLegoIPv6Only = "LEGO_IPV6ONLY"
)

// NetworkStack represents the IP network stack preference.
type NetworkStack int

const (
	StackDual NetworkStack = iota
	StackIPv4
	StackIPv6
)

// GetNetworkStackFromEnv reads LEGO_IPV4ONLY / LEGO_IPV6ONLY to determine the network stack.
func GetNetworkStackFromEnv() NetworkStack {
	if env.GetOrDefaultBool(envLegoIPv4Only, false) {
		return StackIPv4
	}
	if env.GetOrDefaultBool(envLegoIPv6Only, false) {
		return StackIPv6
	}
	return StackDual
}

// GetClientIP returns the client's public IP address.
// It uses namecheap's IP discovery service to perform the lookup.
// If stack is StackIPv4 or StackIPv6, the HTTP dialer is restricted accordingly.
func GetClientIP(ctx context.Context, client *http.Client) (addr string, err error) {
	stack := GetNetworkStackFromEnv()

	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	client = applyNetworkStack(client, stack)
	client = clientdebug.Wrap(client)

	log.Infof(log.LazySprintf("namecheap: detecting client IP via %s (network stack: %s)", getIPURL, stackName(stack)))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, getIPURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("unable to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer func() { _ = resp.Body.Close() }()

	clientIP, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	return string(clientIP), nil
}

// applyNetworkStack wraps the client's transport with a dialer restricted to the given network stack.
func applyNetworkStack(client *http.Client, stack NetworkStack) *http.Client {
	if stack == StackDual {
		return client
	}

	network := "tcp4"
	if stack == StackIPv6 {
		network = "tcp6"
	}

	dialer := &net.Dialer{Timeout: 5 * time.Second}

	var transport http.RoundTripper
	if t, ok := client.Transport.(*http.Transport); ok {
		clone := t.Clone()
		clone.DialContext = func(ctx context.Context, network2, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, addr)
		}
		transport = clone
	} else {
		transport = &http.Transport{
			DialContext: func(ctx context.Context, network2, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, addr)
			},
		}
	}

	client.Transport = transport
	return client
}

func stackName(s NetworkStack) string {
	switch s {
	case StackIPv4:
		return "ipv4"
	case StackIPv6:
		return "ipv6"
	default:
		return "dual"
	}
}
