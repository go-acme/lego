package internal

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValue_MarshalJSON(t *testing.T) {
	testCases := []struct {
		desc     string
		record   Record
		expected string
	}{
		{
			desc: "empty content",
			record: Record{
				Name:    "_acme-challenge.www",
				Type:    "TXT",
				TTL:     30,
				Content: Value{},
			},
			expected: `{"name":"_acme-challenge.www","type":"TXT","ttl":30}`,
		},
		{
			desc: "content with a single value",
			record: Record{
				Name:    "_acme-challenge.www",
				Type:    "TXT",
				TTL:     30,
				Content: Value{"aaa"},
			},
			expected: `{"name":"_acme-challenge.www","type":"TXT","ttl":30,"content":"aaa"}`,
		},
		{
			desc: "content with multiple values",
			record: Record{
				Name:    "_acme-challenge.www",
				Type:    "TXT",
				TTL:     30,
				Content: Value{"aaa", "bbb", "ccc"},
			},
			expected: `{"name":"_acme-challenge.www","type":"TXT","ttl":30,"content":["aaa","bbb","ccc"]}`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			content, err := json.Marshal(test.record)
			require.NoError(t, err)

			assert.JSONEq(t, test.expected, string(content))
		})
	}
}

func TestValue_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		desc     string
		data     string
		expected Record
	}{
		{
			desc: "empty content",
			data: `{"name":"_acme-challenge.www","type":"TXT","ttl":30}`,
			expected: Record{
				Name:    "_acme-challenge.www",
				Type:    "TXT",
				TTL:     30,
				Content: Value(nil),
			},
		},
		{
			desc: "content with a single value",
			data: `{"name":"_acme-challenge.www","type":"TXT","ttl":30,"content":"aaa"}`,
			expected: Record{
				Name:    "_acme-challenge.www",
				Type:    "TXT",
				TTL:     30,
				Content: Value{"aaa"},
			},
		},
		{
			desc: "content with multiple values",
			data: `{"name":"_acme-challenge.www","type":"TXT","ttl":30,"content":["aaa","bbb","ccc"]}`,
			expected: Record{
				Name:    "_acme-challenge.www",
				Type:    "TXT",
				TTL:     30,
				Content: Value{"aaa", "bbb", "ccc"},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			record := Record{}
			err := json.Unmarshal([]byte(test.data), &record)
			require.NoError(t, err)

			assert.Equal(t, test.expected, record)
		})
	}
}
