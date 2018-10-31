package emca

import (
	"time"

	"github.com/xenolf/lego/emca/le"
	"github.com/xenolf/lego/log"
)

// Get the challenges needed to proof our identifier to the ACME server.
func (c *Client) getAuthzForOrder(order le.OrderResource) ([]le.Authorization, error) {
	resc, errc := make(chan le.Authorization), make(chan domainError)

	delay := time.Second / overallRequestLimit

	for _, authzURL := range order.Authorizations {
		time.Sleep(delay)

		go func(authzURL string) {
			var authz le.Authorization
			_, err := c.jws.PostAsGet(authzURL, &authz)
			if err != nil {
				errc <- domainError{Domain: authz.Identifier.Value, Error: err}
				return
			}

			resc <- authz
		}(authzURL)
	}

	var responses []le.Authorization
	failures := make(ObtainError)
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

func logAuthorizations(order le.OrderResource) {
	for i, auth := range order.Authorizations {
		log.Infof("[%s] AuthURL: %s", order.Identifiers[i].Value, auth)
	}
}

// disableAuthz loops through the passed in slice and disables any auths which are not "valid"
func (c *Client) disableAuthz(authURL string) error {
	var disabledAuth le.Authorization
	_, err := c.jws.PostJSON(authURL, le.DeactivateAuthMessage{Status: "deactivated"}, &disabledAuth)
	return err
}
