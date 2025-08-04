package dns01

import (
	"io"
	"os"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/dnsmock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

func TestDNSProviderManual(t *testing.T) {
	useAsNameserver(t, dnsmock.NewServer().
		Query("_acme-challenge.example.com. CNAME", dnsmock.Noop).
		Query("_acme-challenge.example.com. SOA", dnsmock.Error(dns.RcodeNameError)).
		Query("example.com. SOA", dnsmock.SOA("")).
		Build(t))

	backupStdin := os.Stdin
	defer func() { os.Stdin = backupStdin }()

	testCases := []struct {
		desc        string
		input       string
		expectError bool
	}{
		{
			desc:  "Press enter",
			input: "ok\n",
		},
		{
			desc:        "Missing enter",
			input:       "ok",
			expectError: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			file, err := os.CreateTemp(t.TempDir(), "lego_test")
			require.NoError(t, err)

			t.Cleanup(func() { _ = file.Close() })

			_, err = file.WriteString(test.input)
			require.NoError(t, err)

			_, err = file.Seek(0, io.SeekStart)
			require.NoError(t, err)

			os.Stdin = file

			manualProvider, err := NewDNSProviderManual()
			require.NoError(t, err)

			err = manualProvider.Present("example.com", "", "")
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				err = manualProvider.CleanUp("example.com", "", "")
				require.NoError(t, err)
			}
		})
	}
}
