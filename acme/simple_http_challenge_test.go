package acme

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestSimpleHTTPCanSolve(t *testing.T) {
	challenge := &simpleHTTPChallenge{}

	// determine public ip
	resp, err := http.Get("https://icanhazip.com/")
	if err != nil {
		t.Errorf("Could not get public IP -> %v", err)
	}

	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Could not get public IP -> %v", err)
	}
	ipStr := string(ip)

	if expected, actual := false, challenge.CanSolve("google.com"); expected != actual {
		t.Errorf("Expected CanSolve to return %t for domain 'google.com' but was %t", expected, actual)
	}

	localResolv := strings.Replace(ipStr, "\n", "", -1) + ".xip.io"
	if expected, actual := true, challenge.CanSolve(localResolv); expected != actual {
		t.Errorf("Expected CanSolve to return %t for domain 'localhost' but was %t", expected, actual)
	}
}
