package internal

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("user", "secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithBasicAuth("user", "secret"),
	)
}

func TestClient_GetDomains(t *testing.T) {
	client := mockBuilder().
		Route("GET /domain/get_domains.html",
			servermock.ResponseFromFixture("domains.txt"),
		).
		Build(t)

	zones, err := client.GetDomains(t.Context())
	require.NoError(t, err)

	expected := []string{"example.com", "example.org", "example.net"}

	assert.Equal(t, expected, zones)
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /domain/get_dns.html",
			servermock.ResponseFromFixture("get_dns.json"),
			servermock.CheckQueryParameter().Strict().
				With("domain", "example.com"),
		).
		Build(t)

	records, err := client.GetRecords(t.Context(), "example.com")
	require.NoError(t, err)

	expected := map[string]json.RawMessage{
		"A":          json.RawMessage(strconv.Quote("sub1   1.2.3.4\nsub2     1.2.3.4\nsub3 1.2.3.4\nsub4    1.2.3.4\nsub5 1.2.3.4\nsub6      1.2.3.4\nsub7    1.2.3.4\nsub8     1.2.3.4\nsub9  1.2.3.4\nsub10    1.2.3.4\nsub11      1.2.3.4\nsub12    1.2.3.4\nsub13   1.2.3.4\nsub14     1.2.3.4\nsub15      1.2.3.4\nsub16      1.2.3.4\nsub17      1.2.3.4\nsub18      1.2.3.4\n@        1.2.3.4")),
		"AAAA":       json.RawMessage(strconv.Quote("")),
		"CAA":        json.RawMessage(strconv.Quote("@ 128 iodef \"mailto:someone@example.tld\"\n@ 128 issue \"letsencrypt.org\"\n@ 128 issuewild \"letsencrypt.org\"")),
		"CName":      json.RawMessage(strconv.Quote("some cname.to.example.tld.")),
		"MX":         json.RawMessage(strconv.Quote("10 mail.example.tld.")),
		"SRV":        json.RawMessage(strconv.Quote("_imap._tcp 0 0 0 .\n_imaps._tcp 0 1 993 mail.example.tld.\n_pop3._tcp 0 0 0 .\n_pop3s._tcp 0 0 0 .")),
		"TLSA":       json.RawMessage(strconv.Quote("_25._tcp.mail.example.tld. 2 1 1 CBBC559B44D524D6A132BDAC672744DA3407F12AAE5D5F722C5F6C7913871C75\n_25._tcp.mail.example.tld. 2 1 1 885BF0572252C6741DC9A52F5044487FEF2A93B811CDEDFAD7624CC283B7CDD5\n_25._tcp.mail.example.tld. 2 1 1 F1440A9B76E1E41E53A4CB461329BF6337B419726BE513E42E19F1C691C5D4B2\n_465._tcp.mail.example.tld. 2 1 1 CBBC559B44D524D6A132BDAC672744DA3407F12AAE5D5F722C5F6C7913871C75\n_465._tcp.mail.example.tld. 2 1 1 885BF0572252C6741DC9A52F5044487FEF2A93B811CDEDFAD7624CC283B7CDD5\n_465._tcp.mail.example.tld. 2 1 1 F1440A9B76E1E41E53A4CB461329BF6337B419726BE513E42E19F1C691C5D4B2\n_587._tcp.mail.example.tld. 2 1 1 CBBC559B44D524D6A132BDAC672744DA3407F12AAE5D5F722C5F6C7913871C75\n_587._tcp.mail.example.tld. 2 1 1 885BF0572252C6741DC9A52F5044487FEF2A93B811CDEDFAD7624CC283B7CDD5\n_587._tcp.mail.example.tld. 2 1 1 F1440A9B76E1E41E53A4CB461329BF6337B419726BE513E42E19F1C691C5D4B2")),
		"TXT":        json.RawMessage(strconv.Quote("_dmarc \"v=DMARC1;p=reject;sp=reject;adkim=r;aspf=r;pct=100;rua=mailto:someone@in.mailhardener.com,mailto:postmaster@example.tld;ri=86400;ruf=mailto:someone@in.mailhardener.com,mailto:postmaster@example.tld;fo=1;rf=afrf\"\n_mta-sts \"v=STSv1;id=yyyymmddTHHMMSS;\"\n_smtp._tls \"v=TLSRPTv1;rua=mailto:someone@in.mailhardener.com\"\n@ \"v=spf1 a mx ~all\"\nselector._domainkey \"v=DKIM1;k=rsa;p=Base64Stuff\" \"MoreBase64Stuff\" \"Even++MoreBase64Stuff\" \"YesMoreBase64Stuff\" \"And+Yes+Even+MoreBase64Stuff\" \"Sure++MoreBase64Stuff\" \"LastBase64Stuff\"\nselectorecc._domainkey \"v=DKIM1;k=ed25519;p=Base64Stuff\"\n_acme-challenge \"TheAcmeChallenge\"")),
		"TTL":        json.RawMessage("3600"),
		"comment":    json.RawMessage(strconv.Quote("TLSA RR:\nInfo  -> https://dnssec-stats.ant.isi.edu/~viktor/x3hosts.html\nTest 1 -> https://stats.dnssec-tools.org/explore/?example.tld\nTest 2 -> https://dane.sys4.de/smtp/example.tld\n\nSMIMEA RR:\nGenerator -> https://www.smimea.info/smimea-generator.php\nTest      -> https://www.smimea.info/smimea-test.php")),
		"nameserver": json.RawMessage(strconv.Quote("auth1.artfiles.de.\nauth2.artfiles.de.")),
	}

	assert.Equal(t, expected, records)
}

func TestClient_SetRecords(t *testing.T) {
	client := mockBuilder().
		Route("POST /domain/set_dns.html",
			servermock.ResponseFromFixture("set_dns.json"),
			servermock.CheckQueryParameter().Strict().
				With("TXT", "a b\nc \"d\"").
				With("domain", "example.com"),
		).
		Build(t)

	err := client.SetRecords(t.Context(), "example.com", "TXT", RecordValue{"c": `"d"`, "a": "b"})
	require.NoError(t, err)
}
