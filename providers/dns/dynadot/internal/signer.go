package internal

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// generateSignature builds the X-Signature header value as required by the
// Dynadot RESTful v2 API. It signs the concatenation of:
//
//	apiKey + "\n" + fullPathAndQuery + "\n" + xRequestID + "\n" + requestBody
//
// using HMAC-SHA256 with the API secret as key, then encodes the result with standard Base64.
// https://www.dynadot.com/domain/api-document
func generateSignature(apiKey, apiSecret, fullPathAndQuery, xRequestID, requestBody string) string {
	stringToSign := apiKey + "\n" + fullPathAndQuery + "\n" + xRequestID + "\n" + requestBody

	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write([]byte(stringToSign))

	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
