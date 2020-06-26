package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os/user"
	"path"
	"regexp"
	"strings"
)

type Passport struct {
	SubjectID     string `json:"subject_id"`
	CertificateID string `json:"certificate_id"`
	Issuer        string `json:"issuer"`
	PrivateKey    string `json:"private_key"`
	PublicKey     string `json:"public_key"`
}

func GetDefaultPassportLocation() string {
	usr, err := user.Current()
	if err != nil {
		return ""
	}

	return path.Join(usr.HomeDir, ".h1/passport.json")
}

func LoadPassportFile(location string) (*Passport, error) {
	byteFileValue, err := ioutil.ReadFile(location)
	if err != nil {
		return nil, fmt.Errorf("Error when opening passport file:%+v", err)
	}

	var passport Passport
	err = json.Unmarshal(byteFileValue, &passport)
	if err != nil {
		return nil, fmt.Errorf("Error when parsing passport file:%+v", err)
	}

	err = passport.validate()
	if err != nil {
		return nil, fmt.Errorf("Error when validating passport file:%+v", err)
	}

	return &passport, nil
}

func (passport *Passport) validate() error {
	if passport.Issuer == "" {
		return errors.New("Issuer can't be empty")
	}

	if passport.CertificateID == "" {
		return errors.New("CertificateID can't be empty")
	}

	if passport.PrivateKey == "" {
		return errors.New("Private key must be present")
	}

	if passport.SubjectID == "" {
		return errors.New("Subject can't be empty")
	}

	return nil
}

func (passport *Passport) ExtractProjectID() (string, error) {
	re := regexp.MustCompile("iam/project/[a-zA-Z0-9]+")
	byteProjectIam := re.Find([]byte(passport.SubjectID))
	if len(byteProjectIam) == 0 {
		return "", errors.New("Error when extracting projectID")
	}
	projectIamString := string(byteProjectIam)
	segments := strings.Split(projectIamString, "/")
	projectID := segments[len(segments)-1]
	return projectID, nil
}
