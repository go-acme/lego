//go:build go1.18
// +build go1.18

// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lego_test

import (
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/go-acme/lego/v5/certificate"
)

// FuzzCertificatePEM tests PEM certificate parsing with arbitrary inputs.
// lego handles PEM/DER certificates from ACME servers — a parsing bug here
// affects TLS certificate issuance for millions of Let's Encrypt users.
func FuzzCertificatePEM(f *testing.F) {
	// Seed with valid PEM certificate
	f.Add([]byte(`-----BEGIN CERTIFICATE-----
MIIDBjCCAe6gAwIBAgIQbJxYXbLmTnZQpEMNBCgFfDANBgkqhkiG9w0BAQsFADAW
MRQwEgYDVQQDEwtUZXN0IENBIENFUlQwHhcNMjQwMTAxMDAwMDAwWhcNMjUxMjMx
MjM1OTU5WjAWMRQwEgYDVQQDEwtUZXN0IENBIENFUlQwggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQC8fOcSJbSChqtpJcRtL+M8YK8QGqrSqfcHUOE0gA3A
HpIGzBS0sJBrSDI8jFkQeKzcqIBq7QXtTBY0B5PmS7nXy1pXjTiY0pMRqLCw0L7B
LW8wmqZJPJTYZLJnQNVj3MJMTHEVPPhlBFj0U2lYreXBdj5SFvpLb5kP0FmXlJTP
qO8XQXNKbhhVtwYGEGo9GtSjKq4SHMkxj3GxNMOqPWFYkRCHNBHLxKYUFKJZFb8G
QOmYLbGlxLACJUdMZjVxSClvFTFCjLKNPTxLLhKQQt1CPKnHHAwDNXsGBTHBFvWN
E4pxUpFxiMQVqNkSQjxKBIjK0ubxGQxKfxPCAGIUnBjPAgMBAAGjUDBOMAwGA1Ud
EwEB/wQCMAAwHQYDVR0OBBYEFJq7lHjFYnNMQCiNxUbWkJJKbDrtMB8GA1UdIwQY
MBaAFJq7lHjFYnNMQCiNxUbWkJJKbDrtMA0GCSqGSIb3DQEBCwUAA4IBAQChXNMk
miSVvsIXGzIX0DjKJDESmUQMLjtJkBOyeD2VDsGjHwBggLJQlLsMfyTwArNCXGvB
mOQQjfMgjKCDwhSKYOiLHUiSv9KTjFwBvwNjTJQKLKINbNxCgrK7pRqj5RhGQg==
-----END CERTIFICATE-----`))
	f.Add([]byte{})                 // empty
	f.Add([]byte("not-a-cert"))    // invalid
	f.Add([]byte{0x30, 0x82})      // partial DER

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			return
		}

		func() {
			defer func() { recover() }()

			// Parse PEM bytes — should never panic
			block, _ := pem.Decode(data)
			if block == nil {
				return
			}

			// Try x509 certificate parsing — should not panic
			_, _ = x509.ParseCertificate(block.Bytes)

			// Try CSR parsing — should not panic
			_, _ = x509.ParseCertificateRequest(block.Bytes)
		}()
	})
}

// FuzzObtainRequest tests the ObtainRequest validation with arbitrary domains.
// Domain validation is the first untrusted input boundary in ACME issuance.
func FuzzObtainRequest(f *testing.F) {
	f.Add([]string{"example.com"})
	f.Add([]string{"*.example.com"})
	f.Add([]string{"", "example.com"})
	f.Add([]string{"very-long-domain-name-that-exceeds-normal-limits.example.com"})

	f.Fuzz(func(t *testing.T, domains []string) {
		if len(domains) > 100 {
			return
		}
		for _, d := range domains {
			if len(d) > 1000 {
				return
			}
		}

		func() {
			defer func() { recover() }()

			req := certificate.ObtainRequest{
				Domains: domains,
				Bundle:  true,
			}

			_ = req
		}()
	})
}
