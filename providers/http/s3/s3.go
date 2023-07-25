// Package s3 implements a HTTP provider for solving the HTTP-01 challenge using web server's root path.
package s3

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-acme/lego/v4/challenge/http01"
)

// Environment variables names.
const (
	envNamespace = "AWS_"

	EnvRegionKey = envNamespace + "REGION"
	EnvAccessKey = envNamespace + "ACCESS_KEY_ID"
	EnvSecretKey = envNamespace + "SECRET_ACCESS_KEY"
)

// HTTPProvider implements ChallengeProvider for `http-01` challenge.
type HTTPProvider struct {
	bucket string
	client *s3.Client
}

// NewHTTPProvider returns a HTTPProvider instance with a configured s3 bucket and aws session.
// Credentials must be passed in the environment variables.
func NewHTTPProvider(bucket string) (*HTTPProvider, error) {
	if bucket == "" {
		return nil, fmt.Errorf("S3 bucket name missing")
	}
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to create AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &HTTPProvider{
		bucket: bucket,
		client: client,
	}, nil
}

// Present makes the token available at `HTTP01ChallengePath(token)` by creating a file in the given s3 bucket.
func (s *HTTPProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()
	path := strings.Trim(http01.ChallengePath(token), "/")
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		ACL:    "public-read",
		Bucket: &s.bucket,
		Key:    &path,
		Body:   bytes.NewReader([]byte(keyAuth)),
	})
	if err != nil {
		return fmt.Errorf("failed to upload token to s3: %w", err)
	}
	return nil
}

// CleanUp removes the file created for the challenge.
func (s *HTTPProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	path := strings.Trim(http01.ChallengePath(token), "/")
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &path,
	})
	if err != nil {
		return fmt.Errorf("could not remove file in s3 bucket after HTTP challenge: %w", err)
	}

	return nil
}
