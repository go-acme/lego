package cmd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
	"github.com/xenolf/lego/certificate"
	"github.com/xenolf/lego/log"
)

const baseCertificatesFolderName = "certificates"

func saveCertRes(certRes *certificate.Resource, c *cli.Context) {
	getOrCreateCertFolder(c)

	domain := certRes.Domain

	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	err := writeFileCert(c, domain, ".crt", certRes.Certificate)
	if err != nil {
		log.Fatalf("Unable to save Certificate for domain %s\n\t%v", domain, err)
	}

	if certRes.IssuerCertificate != nil {
		err = writeFileCert(c, domain, ".issuer.crt", certRes.IssuerCertificate)
		if err != nil {
			log.Fatalf("Unable to save IssuerCertificate for domain %s\n\t%v", domain, err)
		}
	}

	if certRes.PrivateKey != nil {
		// if we were given a CSR, we don't know the private key
		err = writeFileCert(c, domain, ".key", certRes.PrivateKey)
		if err != nil {
			log.Fatalf("Unable to save PrivateKey for domain %s\n\t%v", domain, err)
		}

		if c.GlobalBool("pem") {
			err = writeFileCert(c, domain, ".pem", bytes.Join([][]byte{certRes.Certificate, certRes.PrivateKey}, nil))
			if err != nil {
				log.Fatalf("Unable to save Certificate and PrivateKey in .pem for domain %s\n\t%v", domain, err)
			}
		}
	} else if c.GlobalBool("pem") {
		// we don't have the private key; can't write the .pem file
		log.Fatalf("Unable to save pem without private key for domain %s\n\t%v; are you using a CSR?", domain, err)
	}

	jsonBytes, err := json.MarshalIndent(certRes, "", "\t")
	if err != nil {
		log.Fatalf("Unable to marshal CertResource for domain %s\n\t%v", domain, err)
	}

	err = writeFileCert(c, domain, ".json", jsonBytes)
	if err != nil {
		log.Fatalf("Unable to save CertResource for domain %s\n\t%v", domain, err)
	}
}

func writeFileCert(c *cli.Context, domain, extension string, data []byte) error {
	var baseFileName string

	// Check filename cli parameter
	if c.GlobalString("filename") == "" {
		baseFileName = sanitizedDomain(domain)
	} else {
		baseFileName = c.GlobalString("filename")
	}

	issuerOut := filepath.Join(getCertPath(c), baseFileName+extension)
	return ioutil.WriteFile(issuerOut, data, 0600)
}

func readStoredFileCert(c *cli.Context, domain, extension string) ([]byte, error) {
	filename := sanitizedDomain(domain) + extension
	certPath := filepath.Join(getCertPath(c), filename)
	return ioutil.ReadFile(certPath)
}

func getOrCreateCertFolder(c *cli.Context) string {
	folder := filepath.Join(getCertPath(c))
	err := createNonExistingFolder(folder)
	if err != nil {
		log.Fatalf("Could not check/create path: %v", err)
	}
	return folder
}

// getCertPath gets the path for certificates.
func getCertPath(c *cli.Context) string {
	return filepath.Join(c.GlobalString("path"), baseCertificatesFolderName)
}

// sanitizedDomain Make sure no funny chars are in the cert names (like wildcards ;))
func sanitizedDomain(domain string) string {
	return strings.Replace(domain, "*", "_", -1)
}
