package certificate

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenewalInfoResponse_ShouldRenew(t *testing.T) {
	now := time.Now().UTC()

	t.Run("Window is in the past", func(t *testing.T) {
		ri := RenewalInfo{
			ExtendedRenewalInfo: &acme.ExtendedRenewalInfo{
				RenewalInfo: acme.RenewalInfo{
					SuggestedWindow: acme.Window{
						Start: now.Add(-2 * time.Hour),
						End:   now.Add(-1 * time.Hour),
					},
					ExplanationURL: "",
				},
				RetryAfter: 0,
			},
		}

		rt := ri.ShouldRenewAt(now, 0)
		require.NotNil(t, rt)
		assert.Equal(t, now, *rt)
	})

	t.Run("Window is in the future", func(t *testing.T) {
		ri := RenewalInfo{
			ExtendedRenewalInfo: &acme.ExtendedRenewalInfo{
				RenewalInfo: acme.RenewalInfo{
					SuggestedWindow: acme.Window{
						Start: now.Add(1 * time.Hour),
						End:   now.Add(2 * time.Hour),
					},
					ExplanationURL: "",
				},
				RetryAfter: 0,
			},
		}

		rt := ri.ShouldRenewAt(now, 0)
		assert.Nil(t, rt)
	})

	t.Run("Window is in the future, but caller is willing to sleep", func(t *testing.T) {
		ri := RenewalInfo{
			ExtendedRenewalInfo: &acme.ExtendedRenewalInfo{
				RenewalInfo: acme.RenewalInfo{
					SuggestedWindow: acme.Window{
						Start: now.Add(1 * time.Hour),
						End:   now.Add(2 * time.Hour),
					},
					ExplanationURL: "",
				},
				RetryAfter: 0,
			},
		}

		rt := ri.ShouldRenewAt(now, 2*time.Hour)
		require.NotNil(t, rt)
		assert.True(t, rt.Before(now.Add(2*time.Hour)))
	})

	t.Run("Window is in the future, but caller isn't willing to sleep long enough", func(t *testing.T) {
		ri := RenewalInfo{
			ExtendedRenewalInfo: &acme.ExtendedRenewalInfo{
				RenewalInfo: acme.RenewalInfo{
					SuggestedWindow: acme.Window{
						Start: now.Add(1 * time.Hour),
						End:   now.Add(2 * time.Hour),
					},
					ExplanationURL: "",
				},
				RetryAfter: 0,
			},
		}

		rt := ri.ShouldRenewAt(now, 59*time.Minute)
		assert.Nil(t, rt)
	})
}
