package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordValue_Set(t *testing.T) {
	rv := make(RecordValue)

	rv.Set("a", "1")
	rv.Set("b", "2")
	rv.Set("b", "3")

	assert.Equal(t, "a \"1\"\nb \"3\"", rv.String())
}

func TestRecordValue_Add(t *testing.T) {
	rv := make(RecordValue)

	rv.Add("a", "1")
	rv.Add("b", "2")
	rv.Add("b", "3")

	assert.Equal(t, "a \"1\"\nb \"2\"\nb \"3\"", rv.String())
}

func TestRecordValue_Delete(t *testing.T) {
	rv := make(RecordValue)

	rv.Set("a", "1")
	rv.Add("b", "2")

	rv.Delete("b")

	assert.Equal(t, "a \"1\"", rv.String())
}

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
				"a": {"1"},
			},
			toRemove: map[string][]string{
				"a": {"1"},
			},
			expected: ``,
		},
		{
			desc: "remove value in the middle",
			data: map[string][]string{
				"a": {"1", "2", "3"},
			},
			toRemove: map[string][]string{
				"a": {"2"},
			},
			expected: "a \"1\"\na \"3\"",
		},
		{
			desc: "remove value at the beginning",
			data: map[string][]string{
				"a": {"1", "2", "3"},
			},
			toRemove: map[string][]string{
				"a": {"1"},
			},
			expected: "a \"2\"\na \"3\"",
		},
		{
			desc: "remove value at the end",
			data: map[string][]string{
				"a": {"1", "2", "3"},
			},
			toRemove: map[string][]string{
				"a": {"3"},
			},
			expected: "a \"1\"\na \"2\"",
		},
		{
			desc: "remove all (delete)",
			data: map[string][]string{
				"a": {"1", "2", "3"},
			},
			toRemove: map[string][]string{
				"a": {"1", "2", "3"},
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
	testCases := []struct {
		desc     string
		filename string
		expected RecordValue
	}{
		{
			desc:     "simple",
			filename: "txt_record.txt",
			expected: RecordValue{
				"@":                      []string{"\"v=spf1 a mx ~all\""},
				"_acme-challenge":        []string{"\"TheAcmeChallenge\""},
				"_dmarc":                 []string{"\"v=DMARC1;p=reject;sp=reject;adkim=r;aspf=r;pct=100;rua=mailto:someone@in.mailhardener.com,mailto:postmaster@example.tld;ri=86400;ruf=mailto:someone@in.mailhardener.com,mailto:postmaster@example.tld;fo=1;rf=afrf\""},
				"_mta-sts":               []string{"\"v=STSv1;id=yyyymmddTHHMMSS;\""},
				"_smtp._tls":             []string{"\"v=TLSRPTv1;rua=mailto:someone@in.mailhardener.com\""},
				"selector._domainkey":    []string{"\"v=DKIM1;k=rsa;p=Base64Stuff\" \"MoreBase64Stuff\" \"Even++MoreBase64Stuff\" \"YesMoreBase64Stuff\" \"And+Yes+Even+MoreBase64Stuff\" \"Sure++MoreBase64Stuff\" \"LastBase64Stuff\""},
				"selectorecc._domainkey": []string{"\"v=DKIM1;k=ed25519;p=Base64Stuff\""},
			},
		},
		{
			desc:     "multiple values with the same key",
			filename: "txt_record-multiple.txt",
			expected: RecordValue{
				"@":                      []string{"\"v=spf1 a mx ~all\""},
				"_acme-challenge":        []string{"\"xxx\"", "\"yyy\""},
				"_dmarc":                 []string{"\"v=DMARC1;p=reject;sp=reject;adkim=r;aspf=r;pct=100;rua=mailto:someone@in.mailhardener.com,mailto:postmaster@example.tld;ri=86400;ruf=mailto:someone@in.mailhardener.com,mailto:postmaster@example.tld;fo=1;rf=afrf\""},
				"_mta-sts":               []string{"\"v=STSv1;id=yyyymmddTHHMMSS;\""},
				"_smtp._tls":             []string{"\"v=TLSRPTv1;rua=mailto:someone@in.mailhardener.com\""},
				"selector._domainkey":    []string{"\"v=DKIM1;k=rsa;p=Base64Stuff\" \"MoreBase64Stuff\" \"Even++MoreBase64Stuff\" \"YesMoreBase64Stuff\" \"And+Yes+Even+MoreBase64Stuff\" \"Sure++MoreBase64Stuff\" \"LastBase64Stuff\""},
				"selectorecc._domainkey": []string{"\"v=DKIM1;k=ed25519;p=Base64Stuff\""},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			file, err := os.ReadFile(filepath.Join("fixtures", test.filename))
			require.NoError(t, err)

			data := ParseRecordValue(string(file))

			assert.Equal(t, test.expected, data)
		})
	}
}

func Test_parseDomains(t *testing.T) {
	file, err := os.ReadFile(filepath.FromSlash("./fixtures/domains.txt"))
	require.NoError(t, err)

	domains, err := parseDomains(string(file))
	require.NoError(t, err)

	expected := []string{"example.com", "example.org", "example.net"}

	assert.Equal(t, expected, domains)
}
