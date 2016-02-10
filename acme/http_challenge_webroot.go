package acme

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// httpChallengeWebRoot implements ChallengeProvider for `http-01` challenge
type httpChallengeWebRoot struct {
	path string
}

// Present makes the token available at `HTTP01ChallengePath(token)`
func (w *httpChallengeWebRoot) Present(domain, token, keyAuth string) error {
	var err error
	err = ioutil.WriteFile(path.Join(w.path, HTTP01ChallengePath(token)), []byte(keyAuth), 0777)
	if err != nil {
		return fmt.Errorf("Could not write file in webroot for HTTP challenge -> %v", err)
	}

	return nil
}

func (w *httpChallengeWebRoot) CleanUp(domain, token, keyAuth string) error {
	var err error
	err = os.Remove(path.Join(w.path, HTTP01ChallengePath(token)))
	if err != nil {
		return fmt.Errorf("Could not remove file in webroot after HTTP challenge -> %v", err)
	}

	return nil
}
