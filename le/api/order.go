package api

import (
	"encoding/base64"

	"github.com/xenolf/lego/le"
)

type OrderService service

func (o *OrderService) New(domains []string) (le.OrderExtend, error) {
	var identifiers []le.Identifier
	for _, domain := range domains {
		identifiers = append(identifiers, le.Identifier{Type: "dns", Value: domain})
	}

	orderReq := le.OrderMessage{Identifiers: identifiers}

	var order le.OrderMessage
	resp, err := o.core.post(o.core.GetDirectory().NewOrderURL, orderReq, &order)
	if err != nil {
		return le.OrderExtend{}, err
	}

	return le.OrderExtend{
		Location:     resp.Header.Get("Location"),
		OrderMessage: order,
	}, nil
}

func (o *OrderService) Get(orderURL string) (le.OrderMessage, error) {
	var order le.OrderMessage
	_, err := o.core.postAsGet(orderURL, &order)
	if err != nil {
		return le.OrderMessage{}, err
	}

	return order, nil
}

func (o *OrderService) UpdateForCSR(orderURL string, csr []byte) (le.OrderMessage, error) {
	csrMsg := le.CSRMessage{
		Csr: base64.RawURLEncoding.EncodeToString(csr),
	}

	var order le.OrderMessage
	_, err := o.core.post(orderURL, csrMsg, &order)
	if err != nil {
		return le.OrderMessage{}, err
	}

	if order.Status == le.StatusInvalid {
		return le.OrderMessage{}, order.Error
	}

	return order, nil
}
