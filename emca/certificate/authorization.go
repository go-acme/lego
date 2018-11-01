package certificate

import (
	"time"

	"github.com/xenolf/lego/emca/le"
	"github.com/xenolf/lego/log"
)

const (
	// overallRequestLimit is the overall number of request per second
	// limited on the "new-reg", "new-authz" and "new-cert" endpoints.
	// From the documentation the limitation is 20 requests per second,
	// but using 20 as value doesn't work but 18 do
	overallRequestLimit = 18
)

// Get the challenges needed to proof our identifier to the ACME server.
func (c *Certifier) getAuthzForOrder(order orderResource) ([]le.Authorization, error) {
	resc, errc := make(chan le.Authorization), make(chan domainError)

	delay := time.Second / overallRequestLimit

	for _, authzURL := range order.Authorizations {
		time.Sleep(delay)

		go func(authzURL string) {
			var authz le.Authorization
			_, err := c.core.PostAsGet(authzURL, &authz)
			if err != nil {
				errc <- domainError{Domain: authz.Identifier.Value, Error: err}
				return
			}

			resc <- authz
		}(authzURL)
	}

	var responses []le.Authorization
	failures := make(obtainError)
	for i := 0; i < len(order.Authorizations); i++ {
		select {
		case res := <-resc:
			responses = append(responses, res)
		case err := <-errc:
			failures[err.Domain] = err.Error
		}
	}

	logAuthorizations(order)

	close(resc)
	close(errc)

	// be careful to not return an empty failures map;
	// even if empty, they become non-nil error values
	if len(failures) > 0 {
		return responses, failures
	}
	return responses, nil
}

func logAuthorizations(order orderResource) {
	for i, auth := range order.Authorizations {
		log.Infof("[%s] AuthURL: %s", order.Identifiers[i].Value, auth)
	}
}

// disableAuthz loops through the passed in slice and disables any auths which are not "valid"
func (c *Certifier) disableAuthz(authzURL string) error {
	var disabledAuth le.Authorization
	_, err := c.core.Post(authzURL, le.Authorization{Status: le.StatusDeactivated}, &disabledAuth)
	return err
}
