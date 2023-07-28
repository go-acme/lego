// Package s3 implements an HTTP provider for solving the HTTP-01 challenge using AWS S3.
package s3

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-acme/lego/v4/challenge/http01"
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
		return nil, fmt.Errorf("s3: bucket name missing")
	}

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("s3: unable to create AWS config: %w", err)
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

	params := &s3.PutObjectInput{
		ACL:    "public-read",
		Bucket: aws.String(s.bucket),
		Key:    aws.String(strings.Trim(http01.ChallengePath(token), "/")),
		Body:   bytes.NewReader([]byte(keyAuth)),
	}

	_, err := s.client.PutObject(ctx, params)
	if err != nil {
		return fmt.Errorf("s3: failed to upload token to s3: %w", err)
	}
	return nil
}

// CleanUp removes the file created for the challenge.
func (s *HTTPProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	params := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(strings.Trim(http01.ChallengePath(token), "/")),
	}

	_, err := s.client.DeleteObject(ctx, params)
	if err != nil {
		return fmt.Errorf("s3: could not remove file in s3 bucket after HTTP challenge: %w", err)
	}

	return nil
}
