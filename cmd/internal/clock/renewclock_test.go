package clock

import (
	"crypto/rand"
	"crypto/x509"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenewClock_IsInRenewalPeriod_days(t *testing.T) {
	now := time.Date(2025, 1, 19, 1, 1, 1, 1, time.UTC)

	oneDay := 24 * time.Hour

	testCases := []struct {
		desc     string
		cert     *x509.Certificate
		noDays   int
		days     int
		expected assert.BoolAssertionFunc
	}{
		{
			desc: "days: 30 days, NotAfter now",
			cert: &x509.Certificate{
				NotAfter: now,
			},
			days:     30,
			expected: assert.True,
		},
		{
			desc: "days: 30 days, NotAfter 31 days",
			cert: &x509.Certificate{
				NotAfter: now.Add(31*oneDay + 1*time.Second),
			},
			days:     30,
			expected: assert.False,
		},
		{
			desc: "days: 30 days, NotAfter 30 days",
			cert: &x509.Certificate{
				NotAfter: now.Add(30 * oneDay),
			},
			days:     30,
			expected: assert.True,
		},
		{
			desc: "days: 0 days, NotAfter 30 days: only the day of the expiration (custom nodays)",
			cert: &x509.Certificate{
				NotAfter: now.Add(30 * oneDay),
			},
			days:     0,
			noDays:   -666,
			expected: assert.False,
		},
		{
			desc: "days: 0 days, NotAfter 30 days: only the day of the expiration",
			cert: &x509.Certificate{
				NotAfter: now.Add(30 * oneDay),
			},
			days:     0,
			expected: assert.True,
		},
		{
			desc: "days: -1 days, NotAfter 30 days: always renew",
			cert: &x509.Certificate{
				NotAfter: now.Add(30 * oneDay),
			},
			days:     -1,
			expected: assert.True,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Context()

			test.cert.SerialNumber = randBigInt(t)

			clock := NewRenewClock(test.noDays)
			clock.now = func() time.Time { return now }
			clock.jitter = func(n *big.Int) float64 { return 0 }

			actual := clock.IsInRenewalPeriod(test.cert, "example.com", test.days)

			test.expected(t, actual)
		})
	}
}

func TestRenewClock_IsInRenewalPeriod_dynamic(t *testing.T) {
	const noDays = 0

	testCases := []struct {
		desc                string
		now                 time.Time
		notBefore, notAfter time.Time
		expected            assert.BoolAssertionFunc
	}{
		{
			desc:      "higher than 1/3 of the certificate lifetime left (lifetime > 10 days)",
			now:       time.Date(2025, 1, 19, 1, 1, 1, 1, time.UTC),
			notBefore: time.Date(2025, 1, 1, 1, 1, 1, 1, time.UTC),
			notAfter:  time.Date(2025, 1, 30, 1, 1, 1, 1, time.UTC),
			expected:  assert.False,
		},
		{
			desc:      "lower than 1/3 of the certificate lifetime left(lifetime > 10 days)",
			now:       time.Date(2025, 1, 21, 1, 1, 1, 1, time.UTC),
			notBefore: time.Date(2025, 1, 1, 1, 1, 1, 1, time.UTC),
			notAfter:  time.Date(2025, 1, 30, 1, 1, 1, 1, time.UTC),
			expected:  assert.True,
		},
		{
			desc:      "higher than 1/2 of the certificate lifetime left (lifetime < 10 days)",
			now:       time.Date(2025, 1, 4, 1, 1, 1, 1, time.UTC),
			notBefore: time.Date(2025, 1, 1, 1, 1, 1, 1, time.UTC),
			notAfter:  time.Date(2025, 1, 10, 1, 1, 1, 1, time.UTC),
			expected:  assert.False,
		},
		{
			desc:      "lower than 1/2 of the certificate lifetime left (lifetime < 10 days)",
			now:       time.Date(2025, 1, 6, 1, 1, 1, 1, time.UTC),
			notBefore: time.Date(2025, 1, 1, 1, 1, 1, 1, time.UTC),
			notAfter:  time.Date(2025, 1, 10, 1, 1, 1, 1, time.UTC),
			expected:  assert.True,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			clock := NewRenewClock(noDays)
			clock.now = func() time.Time { return test.now }
			clock.jitter = func(n *big.Int) float64 { return 0 }

			cert := &x509.Certificate{
				SerialNumber: randBigInt(t),
				NotBefore:    test.notBefore,
				NotAfter:     test.notAfter,
			}

			actual := clock.IsInRenewalPeriod(cert, "example.com", noDays)

			test.expected(t, actual)
		})
	}
}

func Test_getDueDate(t *testing.T) {
	const oneDay = 24 * time.Hour

	clock := NewRenewClock(0)

	cert := &x509.Certificate{
		SerialNumber: randBigInt(t),
		NotBefore:    time.Date(2025, time.January, 1, 1, 1, 1, 1, time.UTC),
		NotAfter:     time.Date(2025, time.January, 1, 1, 1, 1, 1, time.UTC).Add(100 * oneDay),
	}

	actual := clock.getDueDate(cert, 0, time.Now())

	expected := time.Date(2025, time.March, 8, 17, 1, 1, 1, time.UTC)

	t.Log(actual)
	assert.WithinRange(t, actual, expected.Add(-5*oneDay), expected.Add(5*oneDay))
}

func Test_deterministicJitter(t *testing.T) {
	testCases := []struct {
		desc     string
		n        *big.Int
		expected float64
	}{
		{
			desc:     "network",
			n:        big.NewInt(23),
			expected: 0.01836374862784526,
		},
		{
			desc:     "number of the beast",
			n:        big.NewInt(666),
			expected: -0.028234840693323684,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.InEpsilon(t, test.expected, deterministicJitter(test.n), 0.000001)

			// 2 calls with the same value must return the same jitter
			assert.InEpsilon(t, test.expected, deterministicJitter(test.n), 0.000001)
		})
	}
}

func Test_deterministicJitter_deterministic(t *testing.T) {
	for range 10 {
		n := randBigInt(t)

		a := deterministicJitter(n)
		b := deterministicJitter(n)

		// 2 calls with the same value must return the same jitter
		assert.InEpsilon(t, a, b, 0.000001)
	}
}

func randBigInt(t *testing.T) *big.Int {
	t.Helper()

	n, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 256))
	require.NoError(t, err)

	return n
}
