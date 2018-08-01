// Package s3 implements a HTTP provider for solving the HTTP-01 challenge
// using AWS S3 in combination with AWS CloudFront.
package s3

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/xenolf/lego/acme"
)

// HTTPProvider implements ChallengeProvider for `http-01` challenge
type HTTPProvider struct {
	bucket string
	client *s3.S3
}

// NewHTTPProvider returns a HTTPProvider instance with a configured s3 bucket
func NewHTTPProvider(bucket, region string) (*HTTPProvider, error) {
	if bucket == "" {
		return nil, fmt.Errorf("S3 bucket name missing")
	}
	if region == "" {
		return nil, fmt.Errorf("S3 region name missing")
	}

	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, err
	}
	client := s3.New(sess)
	return &HTTPProvider{
		bucket: bucket,
		client: client,
	}, nil
}

// Present makes the token available at `HTTP01ChallengePath(token)` by creating a file in the given s3 bucket
func (s *HTTPProvider) Present(domain, token, keyAuth string) error {
	params := &s3.PutObjectInput{
		Body:   strings.NewReader(keyAuth),
		Bucket: aws.String(s.bucket),
		Key:    aws.String(acme.HTTP01ChallengePath(token)),
		ACL:    aws.String("public-read"),
	}
	_, err := s.client.PutObject(params)
	if err != nil {
		return fmt.Errorf("failed to upload file to S3 bucket: %v", err)
	}
	return nil
}

// CleanUp removes the file created for the challenge
func (s *HTTPProvider) CleanUp(domain, token, keyAuth string) error {
	params := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(acme.HTTP01ChallengePath(token)),
	}
	_, err := s.client.DeleteObject(params)
	if err != nil {
		return fmt.Errorf("failed to remove file from S3 bucket: %v", err)
	}
	return nil
}
