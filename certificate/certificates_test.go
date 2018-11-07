package certificate

import (
	"github.com/xenolf/lego/le"
)

// func TestCertifier_getNewOrderForDomains(t *testing.T) {
// 	mux, apiURL, tearDown := tester.SetupFakeAPI()
// 	defer tearDown()
//
// 	mux.HandleFunc("/newOrder", func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodPost {
// 			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
// 			return
// 		}
//
// 		err := tester.WriteJSONResponse(w, le.Order{})
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	})
//
// 	mux.HandleFunc("/nonce", func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodHead {
// 			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
// 			return
// 		}
//
// 		w.Header().Add("Replay-Nonce", "12345")
// 		w.Header().Add("Retry-After", "0")
// 	})
//
// 	// small value keeps test fast
// 	key, err := rsa.GenerateKey(rand.Reader, 512)
// 	require.NoError(t, err, "Could not generate test key")
//
// 	core, err := api.New(http.DefaultClient, "lego-test", apiURL, "", key)
// 	require.NoError(t, err)
//
// 	certifier := NewCertifier(core, certcrypto.RSA2048, &mockResolver{})
//
// 	_, err = certifier.getNewOrderForDomains([]string{"example.com"})
// 	require.NoError(t, err)
// }

type mockResolver struct{}

func (*mockResolver) Solve(authorizations []le.Authorization) error {
	return nil
}
