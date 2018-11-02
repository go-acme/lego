package acme

import (
	"time"

	"github.com/xenolf/lego/log"
)

// Get the challenges needed to proof our identifier to the ACME server.
func (c *Client) getAuthzForOrder(order orderResource) ([]authorization, error) {
	resc, errc := make(chan authorization), make(chan domainError)

	delay := time.Second / overallRequestLimit

	for _, authzURL := range order.Authorizations {
		time.Sleep(delay)

		go func(authzURL string) {
			var authz authorization
			_, err := getJSON(authzURL, &authz)
			if err != nil {
				errc <- domainError{Domain: authz.Identifier.Value, Error: err}
				return
			}

			resc <- authz
		}(authzURL)
	}

	var responses []authorization
	failures := make(ObtainError)
	for i := 0; i < len(order.Authorizations); i++ {
		select {
		case res := <-resc:
			responses = append(responses, res)
		case err := <-errc:
			failures[err.Domain] = err.Error
		}
	}

	logAuthz(order)

	close(resc)
	close(errc)

	// be careful to not return an empty failures map;
	// even if empty, they become non-nil error values
	if len(failures) > 0 {
		return responses, failures
	}
	return responses, nil
}

func logAuthz(order orderResource) {
	for i, auth := range order.Authorizations {
		log.Infof("[%s] AuthURL: %s", order.Identifiers[i].Value, auth)
	}
}

// disableAuthz loops through the passed in slice and disables any auths which are not "valid"
func (c *Client) disableAuthz(authURL string) error {
	var disabledAuth authorization
	_, err := c.jws.postJSON(authURL, deactivateAuthMessage{Status: "deactivated"}, &disabledAuth)
	return err
}
