package storage

import "github.com/go-acme/lego/v5/certificate"

type Certificate struct {
	*certificate.Resource

	Origin string `json:"origin,omitempty"`
}
