package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"slices"
	"time"

	"github.com/go-acme/lego/v4/acme"
)

// OrderOptions used to create an order (optional).
type OrderOptions struct {
	NotBefore time.Time
	NotAfter  time.Time

	// A string uniquely identifying the profile
	// which will be used to affect issuance of the certificate requested by this Order.
	// - https://www.ietf.org/id/draft-aaron-acme-profiles-00.html#section-4
	Profile string

	// A string uniquely identifying a previously-issued certificate which this
	// order is intended to replace.
	// - https://datatracker.ietf.org/doc/html/draft-ietf-acme-ari-03#section-5
	ReplacesCertID string
}

type OrderService service

// New Creates a new order.
func (o *OrderService) New(domains []string) (acme.ExtendedOrder, error) {
	return o.NewWithOptions(domains, nil)
}

// NewWithOptions Creates a new order.
func (o *OrderService) NewWithOptions(domains []string, opts *OrderOptions) (acme.ExtendedOrder, error) {
	var identifiers []acme.Identifier
	for _, domain := range domains {
		ident := acme.Identifier{Value: domain, Type: "dns"}

		if net.ParseIP(domain) != nil {
			ident.Type = "ip"
		}

		identifiers = append(identifiers, ident)
	}

	orderReq := acme.Order{Identifiers: identifiers}

	if opts != nil {
		if !opts.NotAfter.IsZero() {
			orderReq.NotAfter = opts.NotAfter.Format(time.RFC3339)
		}

		if !opts.NotBefore.IsZero() {
			orderReq.NotBefore = opts.NotBefore.Format(time.RFC3339)
		}

		if o.core.GetDirectory().RenewalInfo != "" {
			orderReq.Replaces = opts.ReplacesCertID
		}

		if opts.Profile != "" {
			orderReq.Profile = opts.Profile
		}
	}

	var order acme.Order
	resp, err := o.core.post(o.core.GetDirectory().NewOrderURL, orderReq, &order)
	if err != nil {
		are := &acme.AlreadyReplacedError{}
		if !errors.As(err, &are) {
			return acme.ExtendedOrder{}, err
		}

		// If the Server rejects the request because the identified certificate has already been marked as replaced,
		// it MUST return an HTTP 409 (Conflict) with a problem document of type "alreadyReplaced" (see Section 7.4).
		// https://datatracker.ietf.org/doc/html/draft-ietf-acme-ari-08#section-5
		orderReq.Replaces = ""

		resp, err = o.core.post(o.core.GetDirectory().NewOrderURL, orderReq, &order)
		if err != nil {
			return acme.ExtendedOrder{}, err
		}
	}

	// The elements of the "authorizations" and "identifiers" arrays are immutable once set.
	// The server MUST NOT change the contents of either array after they are created.
	// If a client observes a change in the contents of either array,
	// then it SHOULD consider the order invalid.
	// https://www.rfc-editor.org/rfc/rfc8555#section-7.1.3
	if compareIdentifiers(orderReq.Identifiers, order.Identifiers) != 0 {
		// Sorts identifiers to avoid error message ambiguities about the order of the identifiers.
		slices.SortStableFunc(orderReq.Identifiers, compareIdentifier)
		slices.SortStableFunc(order.Identifiers, compareIdentifier)

		return acme.ExtendedOrder{},
			fmt.Errorf("order identifiers have been by the ACME server (RFC8555 §7.1.3): %+v != %+v",
				orderReq.Identifiers, order.Identifiers)
	}

	return acme.ExtendedOrder{
		Order:    order,
		Location: resp.Header.Get("Location"),
	}, nil
}

// Get Gets an order.
func (o *OrderService) Get(orderURL string) (acme.ExtendedOrder, error) {
	if orderURL == "" {
		return acme.ExtendedOrder{}, errors.New("order[get]: empty URL")
	}

	var order acme.Order
	_, err := o.core.postAsGet(orderURL, &order)
	if err != nil {
		return acme.ExtendedOrder{}, err
	}

	return acme.ExtendedOrder{Order: order}, nil
}

// UpdateForCSR Updates an order for a CSR.
func (o *OrderService) UpdateForCSR(orderURL string, csr []byte) (acme.ExtendedOrder, error) {
	csrMsg := acme.CSRMessage{
		Csr: base64.RawURLEncoding.EncodeToString(csr),
	}

	var order acme.Order
	_, err := o.core.post(orderURL, csrMsg, &order)
	if err != nil {
		return acme.ExtendedOrder{}, err
	}

	if order.Status == acme.StatusInvalid {
		return acme.ExtendedOrder{}, fmt.Errorf("invalid order: %w", order.Err())
	}

	return acme.ExtendedOrder{Order: order}, nil
}
