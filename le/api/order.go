package api

import (
	"encoding/base64"
	"errors"

	"github.com/xenolf/lego/le"
)

type OrderService service

func (o *OrderService) New(domains []string) (le.ExtendedOrder, error) {
	var identifiers []le.Identifier
	for _, domain := range domains {
		identifiers = append(identifiers, le.Identifier{Type: "dns", Value: domain})
	}

	orderReq := le.Order{Identifiers: identifiers}

	var order le.Order
	resp, err := o.core.post(o.core.GetDirectory().NewOrderURL, orderReq, &order)
	if err != nil {
		return le.ExtendedOrder{}, err
	}

	return le.ExtendedOrder{
		Location: resp.Header.Get("Location"),
		Order:    order,
	}, nil
}

func (o *OrderService) Get(orderURL string) (le.Order, error) {
	if len(orderURL) == 0 {
		return le.Order{}, errors.New("order[get]: empty URL")
	}

	var order le.Order
	_, err := o.core.postAsGet(orderURL, &order)
	if err != nil {
		return le.Order{}, err
	}

	return order, nil
}

func (o *OrderService) UpdateForCSR(orderURL string, csr []byte) (le.Order, error) {
	csrMsg := le.CSRMessage{
		Csr: base64.RawURLEncoding.EncodeToString(csr),
	}

	var order le.Order
	_, err := o.core.post(orderURL, csrMsg, &order)
	if err != nil {
		return le.Order{}, err
	}

	if order.Status == le.StatusInvalid {
		return le.Order{}, order.Error
	}

	return order, nil
}
