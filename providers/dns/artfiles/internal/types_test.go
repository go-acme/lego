package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordValue_RemoveValue(t *testing.T) {
	testCases := []struct {
		desc     string
		data     map[string][]string
		toRemove map[string][]string
		expected string
	}{
		{
			desc: "remove the only value",
			data: map[string][]string{
				"a": {"b"},
			},
			toRemove: map[string][]string{
				"a": {"b"},
			},
			expected: ``,
		},
		{
			desc: "remove value in the middle",
			data: map[string][]string{
				"a": {"b", "c", "d"},
			},
			toRemove: map[string][]string{
				"a": {"c"},
			},
			expected: `a "b" "d"`,
		},
		{
			desc: "remove value at the beginning",
			data: map[string][]string{
				"a": {"b", "c", "d"},
			},
			toRemove: map[string][]string{
				"a": {"b"},
			},
			expected: `a "c" "d"`,
		},
		{
			desc: "remove value at the end",
			data: map[string][]string{
				"a": {"b", "c", "d"},
			},
			toRemove: map[string][]string{
				"a": {"d"},
			},
			expected: `a "b" "c"`,
		},
		{
			desc: "remove all (delete)",
			data: map[string][]string{
				"a": {"b", "c", "d"},
			},
			toRemove: map[string][]string{
				"a": {"b", "c", "d"},
			},
			expected: ``,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rv := make(RecordValue)

			for k, values := range test.data {
				for _, v := range values {
					rv.Add(k, v)
				}
			}

			for k, values := range test.toRemove {
				for _, v := range values {
					rv.RemoveValue(k, v)
				}
			}

			assert.Equal(t, test.expected, rv.String())
		})
	}
}

func TestParseRecordValue(t *testing.T) {
	file, err := os.ReadFile(filepath.FromSlash("./fixtures/txt_record.txt"))
	require.NoError(t, err)

	data := ParseRecordValue(string(file))

	expected := RecordValue{
		"@":                      "\"v=spf1 a mx ~all\"",
		"_acme-challenge":        "\"TheAcmeChallenge\"",
		"_dmarc":                 "\"v=DMARC1;p=reject;sp=reject;adkim=r;aspf=r;pct=100;rua=mailto:someone@in.mailhardener.com,mailto:postmaster@example.tld;ri=86400;ruf=mailto:someone@in.mailhardener.com,mailto:postmaster@example.tld;fo=1;rf=afrf\"",
		"_mta-sts":               "\"v=STSv1;id=yyyymmddTHHMMSS;\"",
		"_smtp._tls":             "\"v=TLSRPTv1;rua=mailto:someone@in.mailhardener.com\"",
		"selector._domainkey":    "\"v=DKIM1;k=rsa;p=Base64Stuff\" \"MoreBase64Stuff\" \"Even++MoreBase64Stuff\" \"YesMoreBase64Stuff\" \"And+Yes+Even+MoreBase64Stuff\" \"Sure++MoreBase64Stuff\" \"LastBase64Stuff\"",
		"selectorecc._domainkey": "\"v=DKIM1;k=ed25519;p=Base64Stuff\"",
	}

	assert.Equal(t, expected, data)
}

func Test_parseDomains(t *testing.T) {
	file, err := os.ReadFile(filepath.FromSlash("./fixtures/domains.txt"))
	require.NoError(t, err)

	domains, err := parseDomains(string(file))
	require.NoError(t, err)

	expected := []string{"example.com", "example.org", "example.net"}

	assert.Equal(t, expected, domains)
}
