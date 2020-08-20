package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
)

type Passport struct {
	SubjectID     string `json:"subject_id"`
	CertificateID string `json:"certificate_id"`
	Issuer        string `json:"issuer"`
	PrivateKey    string `json:"private_key"`
	PublicKey     string `json:"public_key"`
}

func LoadPassportFile(location string) (*Passport, error) {
	file, err := os.Open(location)
	if err != nil {
		return nil, fmt.Errorf("failed to open passport file: %w", err)
	}

	defer func() { _ = file.Close() }()

	var passport Passport
	err = json.NewDecoder(file).Decode(&passport)
	if err != nil {
		return nil, fmt.Errorf("failed to parse passport file: %w", err)
	}

	err = passport.validate()
	if err != nil {
		return nil, fmt.Errorf("passport file validation failed: %w", err)
	}

	return &passport, nil
}

func (passport *Passport) validate() error {
	if passport.Issuer == "" {
		return errors.New("issuer is empty")
	}

	if passport.CertificateID == "" {
		return errors.New("certificate ID is empty")
	}

	if passport.PrivateKey == "" {
		return errors.New("private key is missing")
	}

	if passport.SubjectID == "" {
		return errors.New("subject is empty")
	}

	return nil
}

func (passport *Passport) ExtractProjectID() (string, error) {
	re := regexp.MustCompile("iam/project/([a-zA-Z0-9]+)")

	parts := re.FindStringSubmatch(passport.SubjectID)
	if len(parts) != 2 {
		return "", fmt.Errorf("failed to extract project ID from subject ID: %s", passport.SubjectID)
	}

	return parts[1], nil
}
