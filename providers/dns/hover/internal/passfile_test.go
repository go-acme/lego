package internal_test

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"testing"

	internal "github.com/go-acme/lego/v4/providers/dns/hover/internal" // to ensure testing without extra access
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
)

var BasicTests = []struct {
	Name     string
	Config   string
	Expected internal.PlaintextAuth
}{
	{"scott/tiger", `{"username": "scott", "plaintextpassword": "tiger"}`, internal.PlaintextAuth{Username: "scott", PlaintextPassword: "tiger"}},
}

func TestParseJson(t *testing.T) {
	for _, tt := range BasicTests {
		t.Run(tt.Name, func(t *testing.T) {
			var observed internal.PlaintextAuth
			err := json.Unmarshal([]byte(tt.Config), &observed)
			if assert.NoErrorf(t, err, "Unmarshal returned an error: %s", "formatted") {
				if diff := deep.Equal(tt.Expected, observed); diff != nil {
					t.Error(diff)
				}
			}
		})
	}
}

// TestParseJsonFile leverages ReadConfigFile tested in TestParseJson to parse an on-disk passwd
// file by writing a test file, reading it, and validating the parsed result against the expected
// result as written to the file.  This should allow for any corner cases to be appended to
// BasicTests above and validated herein automatically on every build.
func TestParseJsonFile(t *testing.T) {
	for _, tt := range BasicTests {
		t.Run(tt.Name, func(t *testing.T) {
			tmpfile, err := ioutil.TempFile("", "testfile")
			if assert.NoErrorf(t, err, "Error creating temp file: %s", "formatted") {
				defer os.Remove(tmpfile.Name())

				_, err = tmpfile.Write([]byte(tt.Config))
				if assert.NoErrorf(t, err, "Error writing temp file: %s", "formatted") {
					observed, err := internal.ReadConfigFile(tmpfile.Name())
					if assert.NoErrorf(t, err, "Unmarshal returned an error: %s", "formatted") {
						if diff := deep.Equal(tt.Expected, *observed); diff != nil {
							t.Error(diff)
						}
					}
				}
			}
		})
	}
}
