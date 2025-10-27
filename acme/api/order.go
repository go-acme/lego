package api

import (
	"encoding/base64"
	"errors"
	"fmt"
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
	// - https://www.ietf.org/id/draft-ietf-acme-profiles-00.html#section-4
	Profile string

	// A string uniquely identifying a previously-issued certificate which this
	// order is intended to replace.
	// - https://www.rfc-editor.org/rfc/rfc9773.html#section-5
	ReplacesCertID string
}

type OrderService service

// New Creates a new order.
func (o *OrderService) New(domains []string) (acme.ExtendedOrder, error) {
	return o.NewWithOptions(domains, nil)
}

// NewWithOptions Creates a new order.
func (o *OrderService) NewWithOptions(domains []string, opts *OrderOptions) (acme.ExtendedOrder, error) {
	orderReq := acme.Order{Identifiers: createIdentifiers(domains)}

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
		// https://www.rfc-editor.org/rfc/rfc9773.html#section-5
		orderReq.Replaces = ""

		resp, err = o.core.post(o.core.GetDirectory().NewOrderURL, orderReq, &order)
		if err != nil {
			return acme.ExtendedOrder{}, err
		}
	}

	// The server MUST return an error if it cannot fulfill the request as specified,
	// and it MUST NOT issue a certificate with contents other than those requested.
	// If the server requires the request to be modified in a certain way,
	// it should indicate the required changes using an appropriate error type and description.
	// https://www.rfc-editor.org/rfc/rfc8555#section-7.4
	//
	// Some ACME servers don't return an error,
	// and/or change the order identifiers in the response,
	// so we need to ensure that the identifiers are the same as requested.
	// Deduplication by the server is allowed.
	if compareIdentifiers(orderReq.Identifiers, order.Identifiers) != 0 {
		// Sorts identifiers to avoid error message ambiguities about the order of the identifiers.
		slices.SortStableFunc(orderReq.Identifiers, compareIdentifier)
		slices.SortStableFunc(order.Identifiers, compareIdentifier)

		return acme.ExtendedOrder{},
			fmt.Errorf("order identifiers have been modified by the ACME server (RFC8555 §7.4): %+v != %+v",
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
