package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const certResponseNoBundleMock = `-----BEGIN CERTIFICATE-----
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
`

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

const certResponseMock2 = `
-----BEGIN CERTIFICATE-----
MIIFUzCCBDugAwIBAgISA/z9btaZCSo/qlVwmJrHpoyPMA0GCSqGSIb3DQEBCwUA
MEoxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MSMwIQYDVQQD
ExpMZXQncyBFbmNyeXB0IEF1dGhvcml0eSBYMzAeFw0yMDA3MjUwNjUxNDRaFw0y
MDEwMjMwNjUxNDRaMBgxFjAUBgNVBAMTDW5hdHVyZS5nbG9iYWwwggEiMA0GCSqG
SIb3DQEBAQUAA4IBDwAwggEKAoIBAQDN/PF8lWub3i+lO3CLl/HJAM86pQH9hWej
Whci1PPNzKyEByJq2psNLCO1W1mXK3ClWSyifptCf7+AAFAOoBojPMwjaKMziw1M
BxAQiX8MzZLv4Hr4Uk08cQX31QHiEpOv4pMHqB0UpodTYY10dZnDdyJHaGKzxfJh
nQPYIVto+UegcVu9iZIDow7ugoT2Gh8nB8jOAc4wtBgmylgeAFmYR6QZ4PYSYFh0
DLZGGB1WuU/4YC5OciwTDv5EiqP3KM3NdkmGhPY0A3jcTrjN+HhcE4pYBtG1wHi8
PEuqqKyCLa3AjHq4WrZyCCkCMXPbIDS1Qt7botDmUZr/26xJZnl5AgMBAAGjggJj
MIICXzAOBgNVHQ8BAf8EBAMCBaAwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUF
BwMCMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYEFFm72Cv7LnjVhcLqUujrykUr70lF
MB8GA1UdIwQYMBaAFKhKamMEfd265tE5t6ZFZe/zqOyhMG8GCCsGAQUFBwEBBGMw
YTAuBggrBgEFBQcwAYYiaHR0cDovL29jc3AuaW50LXgzLmxldHNlbmNyeXB0Lm9y
ZzAvBggrBgEFBQcwAoYjaHR0cDovL2NlcnQuaW50LXgzLmxldHNlbmNyeXB0Lm9y
Zy8wGAYDVR0RBBEwD4INbmF0dXJlLmdsb2JhbDBMBgNVHSAERTBDMAgGBmeBDAEC
ATA3BgsrBgEEAYLfEwEBATAoMCYGCCsGAQUFBwIBFhpodHRwOi8vY3BzLmxldHNl
bmNyeXB0Lm9yZzCCAQUGCisGAQQB1nkCBAIEgfYEgfMA8QB3ALIeBcyLos2KIE6H
ZvkruYolIGdr2vpw57JJUy3vi5BeAAABc4T006IAAAQDAEgwRgIhAPEEvCEMkekD
8XLDaxHPnJ85UZL72JqGgNK+7I/NdFNuAiEA5D78b4V1YsD8wvWz/sk6Ks8VgjED
eKGl/TyXwKEpzEIAdgBvU3asMfAxGdiZAKRRFf93FRwR2QLBACkGjbIImjfZEwAA
AXOE9NPrAAAEAwBHMEUCIAu4YFfGZIN/P+0eRG0krSddHKCSf6rqr6aVqUWkJY3F
AiEAz0HkTe0alED1gW9nEAJ1qqK1MLMjRM8SsUv9Is86+CwwDQYJKoZIhvcNAQEL
BQADggEBAGriSVi9YuBnm50w84gjlinmeGdvxgugblIoEqKoXd3d5/zx0DvW9Tm6
YGfXsvAJUSCag7dZ/s/PEu23jKNdFoaBmDaUHHKnUwbWWF7/ptYZ+YuDVGOJo8PL
CULNfUMon20rPU9smzW4BFDBZ6KmX/r4Q8cQ7FLOqKdcng0yMcqIfq4cBxEvd0uQ
pHR3AwCjAIGpV6Q9WHHiHx+SEd/Xc18Z5pXa9m3Rz4i6Mfv+AYLtnsZDxcH81cVM
7rYp80vhXM9tFd4wyrqLuaVZgYD1ylxTYpTI7sijIq4Sl984f3IPA/olN+zK6E8d
EbiufIcKeju/aSellDzzBabEo80YT4o=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIDSjCCAjKgAwIBAgIQRK+wgNajJ7qJMDmGLvhAazANBgkqhkiG9w0BAQUFADA/
MSQwIgYDVQQKExtEaWdpdGFsIFNpZ25hdHVyZSBUcnVzdCBDby4xFzAVBgNVBAMT
DkRTVCBSb290IENBIFgzMB4XDTAwMDkzMDIxMTIxOVoXDTIxMDkzMDE0MDExNVow
PzEkMCIGA1UEChMbRGlnaXRhbCBTaWduYXR1cmUgVHJ1c3QgQ28uMRcwFQYDVQQD
Ew5EU1QgUm9vdCBDQSBYMzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEB
AN+v6ZdQCINXtMxiZfaQguzH0yxrMMpb7NnDfcdAwRgUi+DoM3ZJKuM/IUmTrE4O
rz5Iy2Xu/NMhD2XSKtkyj4zl93ewEnu1lcCJo6m67XMuegwGMoOifooUMM0RoOEq
OLl5CjH9UL2AZd+3UWODyOKIYepLYYHsUmu5ouJLGiifSKOeDNoJjj4XLh7dIN9b
xiqKqy69cK3FCxolkHRyxXtqqzTWMIn/5WgTe1QLyNau7Fqckh49ZLOMxt+/yUFw
7BZy1SbsOFU5Q9D8/RhcQPGX69Wam40dutolucbY38EVAjqr2m7xPi71XAicPNaD
aeQQmxkqtilX4+U9m5/wAl0CAwEAAaNCMEAwDwYDVR0TAQH/BAUwAwEB/zAOBgNV
HQ8BAf8EBAMCAQYwHQYDVR0OBBYEFMSnsaR7LHH62+FLkHX/xBVghYkQMA0GCSqG
SIb3DQEBBQUAA4IBAQCjGiybFwBcqR7uKGY3Or+Dxz9LwwmglSBd49lZRNI+DT69
ikugdB/OEIKcdBodfpga3csTS7MgROSR6cz8faXbauX+5v3gTt23ADq1cEmv8uXr
AvHRAosZy5Q6XkjEGB5YGV8eAlrwDPGxrancWYaLbumR9YbK+rlmM6pZW87ipxZz
R8srzJmwN0jP41ZL9c8PDHIyh8bwRLtTcm1D9SZImlJnt1ir/md2cXjbDaJWFBM5
JDGFoqgCWjBH4d1QB7wCCZAA62RjYJsWvIjJEubSfZGL+T0yjWW06XyxV3bqxbYo
Ob8VZRzI9neWagqNdwvYkQsEjgfbKbYK7p2CNTUQ
-----END CERTIFICATE-----
`

const issuerMock2 = `-----BEGIN CERTIFICATE-----
MIIDSjCCAjKgAwIBAgIQRK+wgNajJ7qJMDmGLvhAazANBgkqhkiG9w0BAQUFADA/
MSQwIgYDVQQKExtEaWdpdGFsIFNpZ25hdHVyZSBUcnVzdCBDby4xFzAVBgNVBAMT
DkRTVCBSb290IENBIFgzMB4XDTAwMDkzMDIxMTIxOVoXDTIxMDkzMDE0MDExNVow
PzEkMCIGA1UEChMbRGlnaXRhbCBTaWduYXR1cmUgVHJ1c3QgQ28uMRcwFQYDVQQD
Ew5EU1QgUm9vdCBDQSBYMzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEB
AN+v6ZdQCINXtMxiZfaQguzH0yxrMMpb7NnDfcdAwRgUi+DoM3ZJKuM/IUmTrE4O
rz5Iy2Xu/NMhD2XSKtkyj4zl93ewEnu1lcCJo6m67XMuegwGMoOifooUMM0RoOEq
OLl5CjH9UL2AZd+3UWODyOKIYepLYYHsUmu5ouJLGiifSKOeDNoJjj4XLh7dIN9b
xiqKqy69cK3FCxolkHRyxXtqqzTWMIn/5WgTe1QLyNau7Fqckh49ZLOMxt+/yUFw
7BZy1SbsOFU5Q9D8/RhcQPGX69Wam40dutolucbY38EVAjqr2m7xPi71XAicPNaD
aeQQmxkqtilX4+U9m5/wAl0CAwEAAaNCMEAwDwYDVR0TAQH/BAUwAwEB/zAOBgNV
HQ8BAf8EBAMCAQYwHQYDVR0OBBYEFMSnsaR7LHH62+FLkHX/xBVghYkQMA0GCSqG
SIb3DQEBBQUAA4IBAQCjGiybFwBcqR7uKGY3Or+Dxz9LwwmglSBd49lZRNI+DT69
ikugdB/OEIKcdBodfpga3csTS7MgROSR6cz8faXbauX+5v3gTt23ADq1cEmv8uXr
AvHRAosZy5Q6XkjEGB5YGV8eAlrwDPGxrancWYaLbumR9YbK+rlmM6pZW87ipxZz
R8srzJmwN0jP41ZL9c8PDHIyh8bwRLtTcm1D9SZImlJnt1ir/md2cXjbDaJWFBM5
JDGFoqgCWjBH4d1QB7wCCZAA62RjYJsWvIjJEubSfZGL+T0yjWW06XyxV3bqxbYo
Ob8VZRzI9neWagqNdwvYkQsEjgfbKbYK7p2CNTUQ
-----END CERTIFICATE-----
`

func Test_checkResponse(t *testing.T) {
	mux, apiURL := tester.SetupFakeAPI(t)

	mux.HandleFunc("/certificate", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte(certResponseMock))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", key)
	require.NoError(t, err)

	certifier := NewCertifier(core, &resolverMock{}, CertifierOptions{KeyType: certcrypto.RSA2048})

	order := acme.ExtendedOrder{
		Order: acme.Order{
			Status:      acme.StatusValid,
			Certificate: apiURL + "/certificate",
		},
	}
	certRes := &Resource{}

	valid, err := certifier.checkResponse(order, certRes, true, "")
	require.NoError(t, err)
	assert.True(t, valid)
	assert.NotNil(t, certRes)
	assert.Equal(t, "", certRes.Domain)
	assert.Contains(t, certRes.CertStableURL, "/certificate")
	assert.Contains(t, certRes.CertURL, "/certificate")
	assert.Nil(t, certRes.CSR)
	assert.Nil(t, certRes.PrivateKey)
	assert.Equal(t, certResponseMock, string(certRes.Certificate), "Certificate")
	assert.Equal(t, issuerMock, string(certRes.IssuerCertificate), "IssuerCertificate")
}

func Test_checkResponse_issuerRelUp(t *testing.T) {
	mux, apiURL := tester.SetupFakeAPI(t)

	mux.HandleFunc("/certificate", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Link", "<"+apiURL+`/issuer>; rel="up"`)
		_, err := w.Write([]byte(certResponseMock))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/issuer", func(w http.ResponseWriter, _ *http.Request) {
		p, _ := pem.Decode([]byte(issuerMock))
		_, err := w.Write(p.Bytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", key)
	require.NoError(t, err)

	certifier := NewCertifier(core, &resolverMock{}, CertifierOptions{KeyType: certcrypto.RSA2048})

	order := acme.ExtendedOrder{
		Order: acme.Order{
			Status:      acme.StatusValid,
			Certificate: apiURL + "/certificate",
		},
	}
	certRes := &Resource{}

	valid, err := certifier.checkResponse(order, certRes, true, "")
	require.NoError(t, err)
	assert.True(t, valid)
	assert.NotNil(t, certRes)
	assert.Equal(t, "", certRes.Domain)
	assert.Contains(t, certRes.CertStableURL, "/certificate")
	assert.Contains(t, certRes.CertURL, "/certificate")
	assert.Nil(t, certRes.CSR)
	assert.Nil(t, certRes.PrivateKey)
	assert.Equal(t, certResponseMock, string(certRes.Certificate), "Certificate")
	assert.Equal(t, issuerMock, string(certRes.IssuerCertificate), "IssuerCertificate")
}

func Test_checkResponse_no_bundle(t *testing.T) {
	mux, apiURL := tester.SetupFakeAPI(t)

	mux.HandleFunc("/certificate", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte(certResponseMock))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", key)
	require.NoError(t, err)

	certifier := NewCertifier(core, &resolverMock{}, CertifierOptions{KeyType: certcrypto.RSA2048})

	order := acme.ExtendedOrder{
		Order: acme.Order{
			Status:      acme.StatusValid,
			Certificate: apiURL + "/certificate",
		},
	}
	certRes := &Resource{}

	valid, err := certifier.checkResponse(order, certRes, false, "")
	require.NoError(t, err)
	assert.True(t, valid)
	assert.NotNil(t, certRes)
	assert.Equal(t, "", certRes.Domain)
	assert.Contains(t, certRes.CertStableURL, "/certificate")
	assert.Contains(t, certRes.CertURL, "/certificate")
	assert.Nil(t, certRes.CSR)
	assert.Nil(t, certRes.PrivateKey)
	assert.Equal(t, certResponseNoBundleMock, string(certRes.Certificate), "Certificate")
	assert.Equal(t, issuerMock, string(certRes.IssuerCertificate), "IssuerCertificate")
}

func Test_checkResponse_alternate(t *testing.T) {
	mux, apiURL := tester.SetupFakeAPI(t)

	mux.HandleFunc("/certificate", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Link", fmt.Sprintf(`<%s/certificate/1>;title="foo";rel="alternate"`, apiURL))

		_, err := w.Write([]byte(certResponseMock))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/certificate/1", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte(certResponseMock2))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", key)
	require.NoError(t, err)

	certifier := NewCertifier(core, &resolverMock{}, CertifierOptions{KeyType: certcrypto.RSA2048})

	order := acme.ExtendedOrder{
		Order: acme.Order{
			Status:      acme.StatusValid,
			Certificate: apiURL + "/certificate",
		},
	}
	certRes := &Resource{
		Domain: "example.com",
	}

	valid, err := certifier.checkResponse(order, certRes, true, "DST Root CA X3")
	require.NoError(t, err)

	assert.True(t, valid)
	assert.NotNil(t, certRes)
	assert.Equal(t, "example.com", certRes.Domain)
	assert.Contains(t, certRes.CertStableURL, "/certificate/1")
	assert.Contains(t, certRes.CertURL, "/certificate/1")
	assert.Nil(t, certRes.CSR)
	assert.Nil(t, certRes.PrivateKey)
	assert.Equal(t, certResponseMock2, string(certRes.Certificate), "Certificate")
	assert.Equal(t, issuerMock2, string(certRes.IssuerCertificate), "IssuerCertificate")
}

func Test_Get(t *testing.T) {
	mux, apiURL := tester.SetupFakeAPI(t)

	mux.HandleFunc("/acme/cert/test-cert", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte(certResponseMock))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", key)
	require.NoError(t, err)

	certifier := NewCertifier(core, &resolverMock{}, CertifierOptions{KeyType: certcrypto.RSA2048})

	certRes, err := certifier.Get(apiURL+"/acme/cert/test-cert", true)
	require.NoError(t, err)

	assert.NotNil(t, certRes)
	assert.Equal(t, "acme.wtf", certRes.Domain)
	assert.Equal(t, apiURL+"/acme/cert/test-cert", certRes.CertStableURL)
	assert.Equal(t, apiURL+"/acme/cert/test-cert", certRes.CertURL)
	assert.Nil(t, certRes.CSR)
	assert.Nil(t, certRes.PrivateKey)
	assert.Equal(t, certResponseMock, string(certRes.Certificate), "Certificate")
	assert.Equal(t, issuerMock, string(certRes.IssuerCertificate), "IssuerCertificate")
}

type resolverMock struct {
	error error
}

func (r *resolverMock) Solve(_ []acme.Authorization) error {
	return r.error
}

func Test_makeCertID(t *testing.T) {
	leafPEM := `-----BEGIN CERTIFICATE-----
MIIDMDCCAhigAwIBAgIIPqNFaGVEHxwwDQYJKoZIhvcNAQELBQAwIDEeMBwGA1UE
AxMVbWluaWNhIHJvb3QgY2EgM2ExMzU2MB4XDTIyMDMxNzE3NTEwOVoXDTI0MDQx
NjE3NTEwOVowFjEUMBIGA1UEAxMLZXhhbXBsZS5jb20wggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQCgm9K/c+il2Pf0f8qhgxn9SKqXq88cOm9ov9AVRbPA
OWAAewqX2yUAwI4LZBGEgzGzTATkiXfoJ3cN3k39cH6tBbb3iSPuEn7OZpIk9D+e
3Q9/hX+N/jlWkaTB/FNA+7aE5IVWhmdczYilXa10V9r+RcvACJt0gsipBZVJ4jfJ
HnWJJGRZzzxqG/xkQmpXxZO7nOPFc8SxYKWdfcgp+rjR2ogYhSz7BfKoVakGPbpX
vZOuT9z4kkHra/WjwlkQhtHoTXdAxH3qC2UjMzO57Tx+otj0CxAv9O7CTJXISywB
vEVcmTSZkHS3eZtvvIwPx7I30ITRkYk/tLl1MbyB3SiZAgMBAAGjeDB2MA4GA1Ud
DwEB/wQEAwIFoDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwDAYDVR0T
AQH/BAIwADAfBgNVHSMEGDAWgBQ4zzDRUaXHVKqlSTWkULGU4zGZpTAWBgNVHREE
DzANggtleGFtcGxlLmNvbTANBgkqhkiG9w0BAQsFAAOCAQEAx0aYvmCk7JYGNEXe
+hrOfKawkHYzWvA92cI/Oi6h+oSdHZ2UKzwFNf37cVKZ37FCrrv5pFP/xhhHvrNV
EnOx4IaF7OrnaTu5miZiUWuvRQP7ZGmGNFYbLTEF6/dj+WqyYdVaWzxRqHFu1ptC
TXysJCeyiGnR+KOOjOOQ9ZlO5JUK3OE4hagPLfaIpDDy6RXQt3ss0iNLuB1+IOtp
1URpvffLZQ8xPsEgOZyPWOcabTwJrtqBwily+lwPFn2mChUx846LwQfxtsXU/lJg
HX2RteNJx7YYNeX3Uf960mgo5an6vE8QNAsIoNHYrGyEmXDhTRe9mCHyiW2S7fZq
o9q12g==
-----END CERTIFICATE-----`
	issuerPEM := `-----BEGIN CERTIFICATE-----
MIIDSzCCAjOgAwIBAgIIOhNWtJ7Igr0wDQYJKoZIhvcNAQELBQAwIDEeMBwGA1UE
AxMVbWluaWNhIHJvb3QgY2EgM2ExMzU2MCAXDTIyMDMxNzE3NTEwOVoYDzIxMjIw
MzE3MTc1MTA5WjAgMR4wHAYDVQQDExVtaW5pY2Egcm9vdCBjYSAzYTEzNTYwggEi
MA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDc3P6cxcCZ7FQOQrYuigReSa8T
IOPNKmlmX9OrTkPwjThiMNEETYKO1ea99yXPK36LUHC6OLmZ9jVQW2Ny1qwQCOy6
TrquhnwKgtkBMDAZBLySSEXYdKL3r0jA4sflW130/OLwhstU/yv0J8+pj7eSVOR3
zJBnYd1AqnXHRSwQm299KXgqema7uwsa8cgjrXsBzAhrwrvYlVhpWFSv3lQRDFQg
c5Z/ZDV9i26qiaJsCCmdisJZWN7N2luUgxdRqzZ4Cr2Xoilg3T+hkb2y/d6ttsPA
kaSA+pq3q6Qa7/qfGdT5WuUkcHpvKNRWqnwT9rCYlmG00r3hGgc42D/z1VvfAgMB
AAGjgYYwgYMwDgYDVR0PAQH/BAQDAgKEMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggr
BgEFBQcDAjASBgNVHRMBAf8ECDAGAQH/AgEAMB0GA1UdDgQWBBQ4zzDRUaXHVKql
STWkULGU4zGZpTAfBgNVHSMEGDAWgBQ4zzDRUaXHVKqlSTWkULGU4zGZpTANBgkq
hkiG9w0BAQsFAAOCAQEArbDHhEjGedjb/YjU80aFTPWOMRjgyfQaPPgyxwX6Dsid
1i2H1x4ud4ntz3sTZZxdQIrOqtlIWTWVCjpStwGxaC+38SdreiTTwy/nikXGa/6W
ZyQRppR3agh/pl5LHVO6GsJz3YHa7wQhEhj3xsRwa9VrRXgHbLGbPOFVRTHPjaPg
Gtsv2PN3f67DsPHF47ASqyOIRpLZPQmZIw6D3isJwfl+8CzvlB1veO0Q3uh08IJc
fspYQXvFBzYa64uKxNAJMi4Pby8cf4r36Wnb7cL4ho3fOHgAltxdW8jgibRzqZpQ
QKyxn2jX7kxeUDt0hFDJE8lOrhP73m66eBNzxe//FQ==
-----END CERTIFICATE-----`
	leafBlock, _ := pem.Decode([]byte(leafPEM))
	leaf, err := x509.ParseCertificate(leafBlock.Bytes)
	require.NoError(t, err)
	issuerBlock, _ := pem.Decode([]byte(issuerPEM))
	issuer, err := x509.ParseCertificate(issuerBlock.Bytes)
	require.NoError(t, err)

	actual, err := makeCertID(leaf, issuer)
	require.NoError(t, err)

	expected := "MFswCwYJYIZIAWUDBAIBBCCeWLRusNLb--vmWOkxm34qDjTMWkc3utIhOMoMwKDqbgQg2iiKWySZrD-6c88HMZ6vhIHZPamChLlzGHeZ7pTS8jYCCD6jRWhlRB8c"
	assert.Equal(t, expected, actual)
}

func TestRetrieveRenewalInfo(t *testing.T) {
	leafPEM := `-----BEGIN CERTIFICATE-----
MIIDMDCCAhigAwIBAgIIPqNFaGVEHxwwDQYJKoZIhvcNAQELBQAwIDEeMBwGA1UE
AxMVbWluaWNhIHJvb3QgY2EgM2ExMzU2MB4XDTIyMDMxNzE3NTEwOVoXDTI0MDQx
NjE3NTEwOVowFjEUMBIGA1UEAxMLZXhhbXBsZS5jb20wggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQCgm9K/c+il2Pf0f8qhgxn9SKqXq88cOm9ov9AVRbPA
OWAAewqX2yUAwI4LZBGEgzGzTATkiXfoJ3cN3k39cH6tBbb3iSPuEn7OZpIk9D+e
3Q9/hX+N/jlWkaTB/FNA+7aE5IVWhmdczYilXa10V9r+RcvACJt0gsipBZVJ4jfJ
HnWJJGRZzzxqG/xkQmpXxZO7nOPFc8SxYKWdfcgp+rjR2ogYhSz7BfKoVakGPbpX
vZOuT9z4kkHra/WjwlkQhtHoTXdAxH3qC2UjMzO57Tx+otj0CxAv9O7CTJXISywB
vEVcmTSZkHS3eZtvvIwPx7I30ITRkYk/tLl1MbyB3SiZAgMBAAGjeDB2MA4GA1Ud
DwEB/wQEAwIFoDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwDAYDVR0T
AQH/BAIwADAfBgNVHSMEGDAWgBQ4zzDRUaXHVKqlSTWkULGU4zGZpTAWBgNVHREE
DzANggtleGFtcGxlLmNvbTANBgkqhkiG9w0BAQsFAAOCAQEAx0aYvmCk7JYGNEXe
+hrOfKawkHYzWvA92cI/Oi6h+oSdHZ2UKzwFNf37cVKZ37FCrrv5pFP/xhhHvrNV
EnOx4IaF7OrnaTu5miZiUWuvRQP7ZGmGNFYbLTEF6/dj+WqyYdVaWzxRqHFu1ptC
TXysJCeyiGnR+KOOjOOQ9ZlO5JUK3OE4hagPLfaIpDDy6RXQt3ss0iNLuB1+IOtp
1URpvffLZQ8xPsEgOZyPWOcabTwJrtqBwily+lwPFn2mChUx846LwQfxtsXU/lJg
HX2RteNJx7YYNeX3Uf960mgo5an6vE8QNAsIoNHYrGyEmXDhTRe9mCHyiW2S7fZq
o9q12g==
-----END CERTIFICATE-----`
	issuerPEM := `-----BEGIN CERTIFICATE-----
MIIDSzCCAjOgAwIBAgIIOhNWtJ7Igr0wDQYJKoZIhvcNAQELBQAwIDEeMBwGA1UE
AxMVbWluaWNhIHJvb3QgY2EgM2ExMzU2MCAXDTIyMDMxNzE3NTEwOVoYDzIxMjIw
MzE3MTc1MTA5WjAgMR4wHAYDVQQDExVtaW5pY2Egcm9vdCBjYSAzYTEzNTYwggEi
MA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDc3P6cxcCZ7FQOQrYuigReSa8T
IOPNKmlmX9OrTkPwjThiMNEETYKO1ea99yXPK36LUHC6OLmZ9jVQW2Ny1qwQCOy6
TrquhnwKgtkBMDAZBLySSEXYdKL3r0jA4sflW130/OLwhstU/yv0J8+pj7eSVOR3
zJBnYd1AqnXHRSwQm299KXgqema7uwsa8cgjrXsBzAhrwrvYlVhpWFSv3lQRDFQg
c5Z/ZDV9i26qiaJsCCmdisJZWN7N2luUgxdRqzZ4Cr2Xoilg3T+hkb2y/d6ttsPA
kaSA+pq3q6Qa7/qfGdT5WuUkcHpvKNRWqnwT9rCYlmG00r3hGgc42D/z1VvfAgMB
AAGjgYYwgYMwDgYDVR0PAQH/BAQDAgKEMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggr
BgEFBQcDAjASBgNVHRMBAf8ECDAGAQH/AgEAMB0GA1UdDgQWBBQ4zzDRUaXHVKql
STWkULGU4zGZpTAfBgNVHSMEGDAWgBQ4zzDRUaXHVKqlSTWkULGU4zGZpTANBgkq
hkiG9w0BAQsFAAOCAQEArbDHhEjGedjb/YjU80aFTPWOMRjgyfQaPPgyxwX6Dsid
1i2H1x4ud4ntz3sTZZxdQIrOqtlIWTWVCjpStwGxaC+38SdreiTTwy/nikXGa/6W
ZyQRppR3agh/pl5LHVO6GsJz3YHa7wQhEhj3xsRwa9VrRXgHbLGbPOFVRTHPjaPg
Gtsv2PN3f67DsPHF47ASqyOIRpLZPQmZIw6D3isJwfl+8CzvlB1veO0Q3uh08IJc
fspYQXvFBzYa64uKxNAJMi4Pby8cf4r36Wnb7cL4ho3fOHgAltxdW8jgibRzqZpQ
QKyxn2jX7kxeUDt0hFDJE8lOrhP73m66eBNzxe//FQ==
-----END CERTIFICATE-----`
	leafBlock, _ := pem.Decode([]byte(leafPEM))
	leaf, err := x509.ParseCertificate(leafBlock.Bytes)
	require.NoError(t, err)
	issuerBlock, _ := pem.Decode([]byte(issuerPEM))
	issuer, err := x509.ParseCertificate(issuerBlock.Bytes)
	require.NoError(t, err)

	leafCertID := "MFswCwYJYIZIAWUDBAIBBCCeWLRusNLb--vmWOkxm34qDjTMWkc3utIhOMoMwKDqbgQg2iiKWySZrD-6c88HMZ6vhIHZPamChLlzGHeZ7pTS8jYCCD6jRWhlRB8c"

	// Test with a fake API.
	mux, apiURL := tester.SetupFakeAPI(t)
	mux.HandleFunc("/renewalInfo/"+leafCertID, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{
				"suggestedWindow": {
					"start": "2020-03-17T17:51:09Z",
					"end": "2020-03-17T18:21:09Z"
				},
				"explanationUrl": "https://aricapable.ca/docs/renewal-advice/"
			}
		}`))
		require.NoError(t, err)
	})

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", key)
	require.NoError(t, err)

	certifier := NewCertifier(core, &resolverMock{}, CertifierOptions{KeyType: certcrypto.RSA2048})

	ri, err := certifier.RetrieveRenewalInfo(CheckRenewalInfoRequest{leaf, issuer})
	require.NoError(t, err)
	assert.Equal(t, "2020-03-17T17:51:09Z", ri.SuggestedWindow.Start.Format(time.RFC3339))
	assert.Equal(t, "2020-03-17T18:21:09Z", ri.SuggestedWindow.End.Format(time.RFC3339))
	assert.Equal(t, "https://aricapable.ca/docs/renewal-advice/", ri.ExplanationURL)
}

func TestRenewalInfo_ShouldRenew(t *testing.T) {
	now := time.Now().UTC()

	// Window is in the past.
	ri := RenewalInfo{
		SuggestedWindow{
			Start: now.Add(-1 * time.Hour),
			End:   now.Add(-30 * time.Minute),
		},
		"",
	}
	rt := ri.ShouldRenewAt(now, 0)
	assert.Equal(t, now, *rt)

	// Window is in the future.
	ri = RenewalInfo{
		SuggestedWindow{
			Start: now.Add(1 * time.Hour),
			End:   now.Add(2 * time.Hour),
		},
		"",
	}
	rt = ri.ShouldRenewAt(now, 0)
	assert.Nil(t, rt)

	// Window is in the future, but caller is willing to sleep.
	rt = ri.ShouldRenewAt(now, 2*time.Hour)
	assert.NotNil(t, rt)
	assert.True(t, rt.Before(now.Add(2*time.Hour)))

	// Window is in the future, but caller isn't willing to sleep long enough.
	rt = ri.ShouldRenewAt(now, 59*time.Minute)
	assert.Nil(t, rt)
}
