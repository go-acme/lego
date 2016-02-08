package acme

import (
	"bufio"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

var lookupNameserversTestsOK = []struct {
	fqdn string
	nss  []string
}{
	{"books.google.com.ng.",
		[]string{"ns1.google.com.", "ns2.google.com.", "ns3.google.com.", "ns4.google.com."},
	},
	{"www.google.com.",
		[]string{"ns1.google.com.", "ns2.google.com.", "ns3.google.com.", "ns4.google.com."},
	},
	{"physics.georgetown.edu.",
		[]string{"ns1.georgetown.edu.", "ns2.georgetown.edu.", "ns3.georgetown.edu."},
	},
}

var lookupNameserversTestsErr = []struct {
	fqdn  string
	error string
}{
	// invalid tld
	{"_null.n0n0.",
		"Could not resolve TLD",
	},
	// invalid domain
	{"_null.com.",
		"Could not resolve NXDOMAIN",
	},
	// invalid subdomain
	{"_null.google.com.",
		"Could not resolve NXDOMAIN",
	},
}

var checkAuthoritativeNssTests = []struct {
	fqdn, value string
	ns          []string
	ok          bool
}{
	// TXT RR w/ expected value
	{"8.8.8.8.asn.routeviews.org.", "151698.8.8.024", []string{"asnums.routeviews.org."},
		true,
	},
	// No TXT RR
	{"ns1.google.com.", "", []string{"ns2.google.com."},
		false,
	},
}

var checkAuthoritativeNssTestsErr = []struct {
	fqdn, value string
	ns          []string
	error       string
}{
	// TXT RR /w unexpected value
	{"8.8.8.8.asn.routeviews.org.", "fe01=", []string{"asnums.routeviews.org."},
		"did not return the expected TXT record",
	},
	// No TXT RR
	{"ns1.google.com.", "fe01=", []string{"ns2.google.com."},
		"did not return the expected TXT record",
	},
}

func TestDNSValidServerResponse(t *testing.T) {
	preCheckDNS = func(domain, fqdn, value string) error {
		return nil
	}
	privKey, _ := generatePrivateKey(rsakey, 512)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Replay-Nonce", "12345")
		w.Write([]byte("{\"type\":\"dns01\",\"status\":\"valid\",\"uri\":\"http://some.url\",\"token\":\"http8\"}"))
	}))

	manualProvider, _ := NewDNSProviderManual()
	jws := &jws{privKey: privKey.(*rsa.PrivateKey), directoryURL: ts.URL}
	solver := &dnsChallenge{jws: jws, validate: validate, provider: manualProvider}
	clientChallenge := challenge{Type: "dns01", Status: "pending", URI: ts.URL, Token: "http8"}

	go func() {
		time.Sleep(time.Second * 2)
		f := bufio.NewWriter(os.Stdout)
		defer f.Flush()
		f.WriteString("\n")
	}()

	if err := solver.Solve(clientChallenge, "example.com"); err != nil {
		t.Errorf("VALID: Expected Solve to return no error but the error was -> %v", err)
	}
}

func TestPreCheckDNS(t *testing.T) {
	err := preCheckDNS("api.letsencrypt.org", "acme-staging.api.letsencrypt.org", "fe01=")
	if err != nil {
		t.Errorf("preCheckDNS failed for acme-staging.api.letsencrypt.org")
	}
}

func TestLookupNameserversOK(t *testing.T) {
	for _, tt := range lookupNameserversTestsOK {
		nss, err := lookupNameservers(tt.fqdn)
		if err != nil {
			t.Fatalf("#%s: got %q; want nil", tt.fqdn, err)
		}

		sort.Strings(nss)
		sort.Strings(tt.nss)

		if !reflect.DeepEqual(nss, tt.nss) {
			t.Errorf("#%s: got %v; want %v", tt.fqdn, nss, tt.nss)
		}
	}
}

func TestLookupNameserversErr(t *testing.T) {
	for _, tt := range lookupNameserversTestsErr {
		_, err := lookupNameservers(tt.fqdn)
		if err == nil {
			t.Fatalf("#%s: expected %q (error); got <nil>", tt.fqdn, tt.error)
		}

		if !strings.Contains(err.Error(), tt.error) {
			t.Errorf("#%s: expected %q (error); got %q", tt.fqdn, tt.error, err)
			continue
		}
	}
}

func TestCheckAuthoritativeNss(t *testing.T) {
	for _, tt := range checkAuthoritativeNssTests {
		ok, _ := checkAuthoritativeNss(tt.fqdn, tt.value, tt.ns)
		if ok != tt.ok {
			t.Errorf("#%s: got %t; want %t", tt.fqdn, tt.ok)
		}
	}
}

func TestCheckAuthoritativeNssErr(t *testing.T) {
	for _, tt := range checkAuthoritativeNssTestsErr {
		_, err := checkAuthoritativeNss(tt.fqdn, tt.value, tt.ns)
		if err == nil {
			t.Fatalf("#%s: expected %q (error); got <nil>", tt.fqdn, tt.error)
		}
		if !strings.Contains(err.Error(), tt.error) {
			t.Errorf("#%s: expected %q (error); got %q", tt.fqdn, tt.error, err)
			continue
		}
	}
}

func TestWaitForTimeout(t *testing.T) {
	c := make(chan error)
	go func() {
		err := waitFor(3, 1, func() (bool, error) {
			return false, nil
		})
		c <- err
	}()

	timeout := time.After(4 * time.Second)
	select {
	case <-timeout:
		t.Fatal("timeout exceeded")
	case err := <-c:
		if err == nil {
			t.Errorf("expected timeout error; got <nil>", err)
		}
	}
}
