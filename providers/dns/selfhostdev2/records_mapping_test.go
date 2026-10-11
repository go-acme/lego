package selfhostdev2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseRecordsMapping(t *testing.T) {
	testCases := []struct {
		desc     string
		rawData  string
		expected map[string]*Seq
	}{
		{
			desc:    "one domain, one record id",
			rawData: "example.com:123",
			expected: map[string]*Seq{
				"example.com": NewSeq(123),
			},
		},
		{
			desc:    "several domain, one record id",
			rawData: "example.com:123, example.org:456,foo.example.com:789",
			expected: map[string]*Seq{
				"example.com":     NewSeq(123),
				"example.org":     NewSeq(456),
				"foo.example.com": NewSeq(789),
			},
		},
		{
			desc:    "one domain, 2 record ids",
			rawData: "example.com:123:456",
			expected: map[string]*Seq{
				"example.com": NewSeq(123, 456),
			},
		},
		{
			desc:    "several domain, 2 record ids",
			rawData: "example.com:123:321, example.org:456:654,foo.example.com:789:987",
			expected: map[string]*Seq{
				"example.com":     NewSeq(123, 321),
				"example.org":     NewSeq(456, 654),
				"foo.example.com": NewSeq(789, 987),
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			mapping, err := parseRecordsMapping(test.rawData)
			require.NoError(t, err)

			assert.Equal(t, test.expected, mapping)
		})
	}
}

func Test_parseRecordsMapping_error(t *testing.T) {
	testCases := []struct {
		desc     string
		rawData  string
		expected string
	}{
		{
			desc:     "empty",
			rawData:  "",
			expected: "empty mapping",
		},
		{
			desc:     "only spaces",
			rawData:  "    ",
			expected: "empty mapping",
		},
		{
			desc:     "one domain, no record id",
			rawData:  "example.com",
			expected: `missing ":": example.com`,
		},
		{
			desc:     "one domain, invalid record id",
			rawData:  "example.com:abc",
			expected: `invalid record ID (abc): strconv.ParseInt: parsing "abc": invalid syntax`,
		},
		{
			desc:     "one domain, more than 2 record ids",
			rawData:  "example.com:123:456:789",
			expected: "too many record IDs for one domain: example.com:123:456:789",
		},
		{
			desc:     "several domain, more than 2 record ids",
			rawData:  "example.com:123, example.org:456:789:147",
			expected: "too many record IDs for one domain: example.org:456:789:147",
		},
		{
			desc:     "no ids, ends with 2 dots",
			rawData:  "example.com:",
			expected: `last char is ":": example.com:`,
		},
		{
			desc:     "no ids,starts with 2 dots",
			rawData:  ":example.com",
			expected: `first char is ":": :example.com`,
		},
		{
			desc:     "with ids but ends with 2 dots",
			rawData:  "example.com:123:",
			expected: `last char is ":": 123:`,
		},
		{
			desc:     "only 2 dots",
			rawData:  ":",
			expected: `first char is ":": :`,
		},
		{
			desc:     "only comma",
			rawData:  ",",
			expected: `first char is ",": ,`,
		},
		{
			desc:     "ends with comma",
			rawData:  "example.com,",
			expected: `last char is ",": example.com,`,
		},
		{
			desc:     "combo",
			rawData:  "::::,::",
			expected: `first char is ":": ::::`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, err := parseRecordsMapping(test.rawData)
			require.EqualError(t, err, test.expected)
		})
	}
}

func TestSeq_Next(t *testing.T) {
	testCases := []struct {
		desc     string
		ids      []int64
		expected []int64
	}{
		{
			desc:     "one value",
			ids:      []int64{1},
			expected: []int64{1, 1, 1},
		},
		{
			desc:     "two values",
			ids:      []int64{1, 2},
			expected: []int64{1, 2, 1, 2},
		},
		{
			desc:     "three values",
			ids:      []int64{1, 2, 3},
			expected: []int64{1, 2, 3, 1, 2, 3, 1},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			seq := NewSeq(test.ids...)
			for _, s := range test.expected {
				assert.Equal(t, s, seq.Next())
			}
		})
	}
}
