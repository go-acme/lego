package api

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"net/http"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const certResponseMock = `-----BEGIN CERTIFICATE-----
MIIDEDCCAfigAwIBAgIHPhckqW5fPDANBgkqhkiG9w0BAQsFADAoMSYwJAYDVQQD
Ex1QZWJibGUgSW50ZXJtZWRpYXRlIENBIDM5NWU2MTAeFw0xODExMDcxNzQ2NTZa
Fw0yMzExMDcxNzQ2NTZaMBMxETAPBgNVBAMTCGFjbWUud3RmMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwtLNKvZXD20XPUQCWYSK9rUSKxD9Eb0c9fag
bxOxOkLRTgL8LH6yln+bxc3MrHDou4PpDUdeo2CyOQu3CKsTS5mrH3NXYHu0H7p5
y3riOJTHnfkGKLT9LciGz7GkXd62nvNP57bOf5Sk4P2M+Qbxd0hPTSfu52740LSy
144cnxe2P1aDYehrEp6nYCESuyD/CtUHTo0qwJmzIy163Sp3rSs15BuCPyhySnE3
BJ8Ggv+qC6D5I1932DfSqyQJ79iq/HRm0Fn84am3KwvRlUfWxabmsUGARXoqCgnE
zcbJVOZKewv0zlQJpfac+b+Imj6Lvt1TGjIz2mVyefYgLx8gwwIDAQABo1QwUjAO
BgNVHQ8BAf8EBAMCBaAwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMAwG
A1UdEwEB/wQCMAAwEwYDVR0RBAwwCoIIYWNtZS53dGYwDQYJKoZIhvcNAQELBQAD
ggEBABB/0iYhmfPSQot5RaeeovQnsqYjI5ryQK2cwzW6qcTJfv8N6+p6XkqF1+W4
jXZjrQP8MvgO9KNWlvx12vhINE6wubk88L+2piAi5uS2QejmZbXpyYB9s+oPqlk9
IDvfdlVYOqvYAhSx7ggGi+j73mjZVtjAavP6dKuu475ZCeq+NIC15RpbbikWKtYE
HBJ7BW8XQKx67iHGx8ygHTDLbREL80Bck3oUm7wIYGMoNijD6RBl25p4gYl9dzOd
TqGl5hW/1P5hMbgEzHbr4O3BfWqU2g7tV36TASy3jbC3ONFRNNYrpEZ1AL3+cUri
OPPkKtAKAbQkKbUIfsHpBZjKZMU=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIDDDCCAfSgAwIBAgIIOV5hkYJx0JwwDQYJKoZIhvcNAQELBQAwIDEeMBwGA1UE
AxMVUGViYmxlIFJvb3QgQ0EgNTBmZmJkMB4XDTE4MTEwNzE3NDY0N1oXDTQ4MTEw
NzE3NDY0N1owKDEmMCQGA1UEAxMdUGViYmxlIEludGVybWVkaWF0ZSBDQSAzOTVl
NjEwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCacwXN4LmyRTgYS8TT
SZYgz758npHiPTBDKgeN5WVmkkwW0TuN4W2zXhEmcM82uxOEjWS2drvK0+iJKneh
0fQR8ZF35dIYFe8WXTg3kEmqcizSgh4LxlOntsXvatfX/6GU/ADo3xAFoBKCijen
SRBIY65yq5m00cWx3RMIcQq1B0X8nJS0O1P7MYE/Vvidz5St/36RXVu1oWLeS5Fx
HAezW0lqxEUzvC+uLTFWC6f/CilzmI7SsPAkZBk7dO5Qs0d7m/zWF588vlGS+0pt
D1on+lU85Ma2zuAd0qmB6LY66N8pEKKtMk93wF/o4Z5i58ahbwNvTKAzz4JSRWSu
mB9LAgMBAAGjQjBAMA4GA1UdDwEB/wQEAwIChDAdBgNVHSUEFjAUBggrBgEFBQcD
AQYIKwYBBQUHAwIwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEA
upU0DjzvIvoCOYKbq1RRN7rPdYad39mfjxgkeV0iOF5JoIdO6y1C7XAm9lT69Wjm
iUPvnCTMDYft40N2SvmXuuMaPOm4zjNwn4K33jw5XBnpwxC7By/Y0oV+Sl10fBsd
QqXC6H7LcSGkv+4eJbgY33P4uH5ZAy+2TkHUuZDkpufkAshzBust7nDAjfv3AIuQ
wlPoyZfI11eqyiOqRzOq+B5dIBr1JzKnEzSL6n0JLNQiPO7iN03rud/wYD3gbmcv
rzFL1KZfz+HZdnFwFW2T2gVW8L3ii1l9AJDuKzlvjUH3p6bgihVq02sjT8mx+GM2
7R4IbHGnj0BJA2vMYC4hSw==
-----END CERTIFICATE-----
`

const issuerMock = `-----BEGIN CERTIFICATE-----
MIIDDDCCAfSgAwIBAgIIOV5hkYJx0JwwDQYJKoZIhvcNAQELBQAwIDEeMBwGA1UE
AxMVUGViYmxlIFJvb3QgQ0EgNTBmZmJkMB4XDTE4MTEwNzE3NDY0N1oXDTQ4MTEw
NzE3NDY0N1owKDEmMCQGA1UEAxMdUGViYmxlIEludGVybWVkaWF0ZSBDQSAzOTVl
NjEwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCacwXN4LmyRTgYS8TT
SZYgz758npHiPTBDKgeN5WVmkkwW0TuN4W2zXhEmcM82uxOEjWS2drvK0+iJKneh
0fQR8ZF35dIYFe8WXTg3kEmqcizSgh4LxlOntsXvatfX/6GU/ADo3xAFoBKCijen
SRBIY65yq5m00cWx3RMIcQq1B0X8nJS0O1P7MYE/Vvidz5St/36RXVu1oWLeS5Fx
HAezW0lqxEUzvC+uLTFWC6f/CilzmI7SsPAkZBk7dO5Qs0d7m/zWF588vlGS+0pt
D1on+lU85Ma2zuAd0qmB6LY66N8pEKKtMk93wF/o4Z5i58ahbwNvTKAzz4JSRWSu
mB9LAgMBAAGjQjBAMA4GA1UdDwEB/wQEAwIChDAdBgNVHSUEFjAUBggrBgEFBQcD
AQYIKwYBBQUHAwIwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEA
upU0DjzvIvoCOYKbq1RRN7rPdYad39mfjxgkeV0iOF5JoIdO6y1C7XAm9lT69Wjm
iUPvnCTMDYft40N2SvmXuuMaPOm4zjNwn4K33jw5XBnpwxC7By/Y0oV+Sl10fBsd
QqXC6H7LcSGkv+4eJbgY33P4uH5ZAy+2TkHUuZDkpufkAshzBust7nDAjfv3AIuQ
wlPoyZfI11eqyiOqRzOq+B5dIBr1JzKnEzSL6n0JLNQiPO7iN03rud/wYD3gbmcv
rzFL1KZfz+HZdnFwFW2T2gVW8L3ii1l9AJDuKzlvjUH3p6bgihVq02sjT8mx+GM2
7R4IbHGnj0BJA2vMYC4hSw==
-----END CERTIFICATE-----
`

func TestCertificateService_Get_issuerRelUp(t *testing.T) {
	mux, apiURL := tester.SetupFakeAPI(t)

	mux.HandleFunc("/certificate", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Link", "<"+apiURL+`/issuer>; rel="up"`)
		_, err := w.Write([]byte(certResponseMock))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/issuer", func(w http.ResponseWriter, _ *http.Request) {
		p, _ := pem.Decode([]byte(issuerMock))
		_, err := w.Write(p.Bytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	core, err := New(http.DefaultClient, "lego-test", apiURL+"/dir", "", key)
	require.NoError(t, err)

	cert, issuer, err := core.Certificates.Get(apiURL+"/certificate", true)
	require.NoError(t, err)
	assert.Equal(t, certResponseMock, string(cert), "Certificate")
	assert.Equal(t, issuerMock, string(issuer), "IssuerCertificate")
}

func TestCertificateService_Get_embeddedIssuer(t *testing.T) {
	mux, apiURL := tester.SetupFakeAPI(t)

	mux.HandleFunc("/certificate", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte(certResponseMock))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	core, err := New(http.DefaultClient, "lego-test", apiURL+"/dir", "", key)
	require.NoError(t, err)

	cert, issuer, err := core.Certificates.Get(apiURL+"/certificate", true)
	require.NoError(t, err)
	assert.Equal(t, certResponseMock, string(cert), "Certificate")
	assert.Equal(t, issuerMock, string(issuer), "IssuerCertificate")
}
