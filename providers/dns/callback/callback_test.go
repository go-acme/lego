package callback

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc            string
		presentCallback func(fqdn, recordBody string) error
		cleanupCallback func(fqdn, recordBody string) error
		expectedError   error
	}{
		{
			desc:            "happy path",
			presentCallback: func(fqdn, recordBody string) error { return nil },
			cleanupCallback: func(fqdn, recordBody string) error { return nil },
			expectedError:   nil,
		},
		{
			desc:            "missing present callback",
			presentCallback: nil,
			cleanupCallback: func(fqdn, recordBody string) error { return nil },
			expectedError:   errors.New("callback: got nil presentCallback"),
		},
		{
			desc:            "happy path",
			presentCallback: func(fqdn, recordBody string) error { return nil },
			cleanupCallback: nil,
			expectedError:   errors.New("callback: got nil cleanupCallback"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			provider, actualError := NewDNSProvider(tc.presentCallback, tc.cleanupCallback)
			if tc.expectedError == nil {
				require.NotNil(t, provider)
			}
			require.Equal(t, tc.expectedError, actualError)
		})
	}
}
