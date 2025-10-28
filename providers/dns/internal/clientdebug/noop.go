//go:build !lego.debug

package clientdebug

import "net/http"

func Wrap(client *http.Client) *http.Client {
	return client
}
