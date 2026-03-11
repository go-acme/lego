package root

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
)

func Process(ctx context.Context, cfg *configuration.Configuration) error {
	archiver := storage.NewArchiver(cfg.Storage)

	err := archiver.Accounts(cfg)
	if err != nil {
		return err
	}

	err = archiver.Certificates(cfg.Certificates)
	if err != nil {
		return err
	}

	return obtain(ctx, cfg)
}

// NOTE(ldez): this is partially a duplication with flags parsing, but the errors are slightly different.
func parseAddress(address string) (string, string, error) {
	if !strings.Contains(address, ":") {
		return "", "", fmt.Errorf("the address only accepts 'interface:port' or ':port' for its argument: '%s'",
			address)
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", "", fmt.Errorf("could not split address '%s': %w", address, err)
	}

	return host, port, nil
}
