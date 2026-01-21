package certificate

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/log"
)

func (c *Certifier) getAuthorizations(ctx context.Context, order acme.ExtendedOrder) ([]acme.Authorization, error) {
	resc, errc := make(chan acme.Authorization), make(chan domainError)

	delay := time.Second / time.Duration(c.overallRequestLimit)

	for _, authzURL := range order.Authorizations {
		time.Sleep(delay)

		go func(authzURL string) {
			authz, err := c.core.Authorizations.Get(ctx, authzURL)
			if err != nil {
				errc <- domainError{Domain: authz.Identifier.Value, Error: err}
				return
			}

			resc <- authz
		}(authzURL)
	}

	var responses []acme.Authorization

	failures := newObtainError()

	for range len(order.Authorizations) {
		select {
		case res := <-resc:
			responses = append(responses, res)
		case err := <-errc:
			failures.Add(err.Domain, err.Error)
		}
	}

	for i, auth := range order.Authorizations {
		log.Info("Authorization",
			slog.String("url", order.Identifiers[i].Value),
			slog.String("authz", auth),
		)
	}

	close(resc)
	close(errc)

	return responses, failures.Join()
}

func (c *Certifier) deactivateAuthorizations(ctx context.Context, order acme.ExtendedOrder, force bool) {
	for _, authzURL := range order.Authorizations {
		auth, err := c.core.Authorizations.Get(ctx, authzURL)
		if err != nil {
			log.Info("Unable to get the authorization.",
				slog.String("url", authzURL),
				log.ErrorAttr(err),
			)

			continue
		}

		if auth.Status == acme.StatusValid && !force {
			log.Info("Skipping deactivating of valid authorization.", slog.String("url", authzURL))

			continue
		}

		log.Info("Deactivating authorization.", slog.String("url", authzURL))

		if c.core.Authorizations.Deactivate(ctx, authzURL) != nil {
			log.Info("Unable to deactivate the authorization.", slog.String("url", authzURL))
		}
	}
}
