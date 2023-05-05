package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const getIPURL = "https://dynamicdns.park-your-domain.com/getip"

// GetClientIP returns the client's public IP address.
// It uses namecheap's IP discovery service to perform the lookup.
func GetClientIP(ctx context.Context, client *http.Client, debug bool) (addr string, err error) {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

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

	if debug {
		log.Println("Client IP:", string(clientIP))
	}

	return string(clientIP), nil
}
