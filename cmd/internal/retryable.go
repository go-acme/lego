package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/log"
	"github.com/hashicorp/go-retryablehttp"
)

func NewRetryableClient(client *http.Client) *http.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	retryClient.HTTPClient = client
	retryClient.CheckRetry = checkRetry
	retryClient.Logger = nil

	if _, v := os.LookupEnv("LEGO_DEBUG_ACME_HTTP_CLIENT"); v {
		retryClient.Logger = log.Default()
	}

	return retryClient.StandardClient()
}

func checkRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {
	rt, err := retryablehttp.ErrorPropagatedRetryPolicy(ctx, resp, err)
	if err != nil {
		return rt, err
	}

	if resp == nil {
		return rt, nil
	}

	if resp.StatusCode/100 == 2 {
		return rt, nil
	}

	all, err := io.ReadAll(resp.Body)
	if err == nil {
		var errorDetails *acme.ProblemDetails

		err = json.Unmarshal(all, &errorDetails)
		if err != nil {
			return rt, fmt.Errorf("%s %s: %s", resp.Request.Method, resp.Request.URL.Redacted(), string(all))
		}

		switch errorDetails.Type {
		case acme.BadNonceErrorType:
			return false, &acme.NonceError{
				ProblemDetails: errorDetails,
			}

		case acme.AlreadyReplacedErrorType:
			if errorDetails.HTTPStatus == http.StatusConflict {
				return false, &acme.AlreadyReplacedError{
					ProblemDetails: errorDetails,
				}
			}

		default:
			log.Warnf(log.LazySprintf("retry: %v", errorDetails))

			return rt, errorDetails
		}
	}

	return rt, nil
}
