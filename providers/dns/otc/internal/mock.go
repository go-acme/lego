package internal

import (
	"fmt"
	"net/http"
)

const fakeOTCToken = "62244bc21da68d03ebac94e6636ff01f"

func IdentityHandlerMock() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("X-Subject-Token", fakeOTCToken)

		_, _ = fmt.Fprintf(w, `{
		  "token": {
		    "catalog": [
		      {
			"type": "dns",
			"id": "56cd81db1f8445d98652479afe07c5ba",
			"name": "",
			"endpoints": [
			  {
			    "url": "http://%s",
			    "region": "eu-de",
			    "region_id": "eu-de",
			    "interface": "public",
			    "id": "0047a06690484d86afe04877074efddf"
			  }
			]
		      }
		    ]
		  }}`, req.Context().Value(http.LocalAddrContextKey))
	}
}
