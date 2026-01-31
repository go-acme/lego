package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/mattn/go-zglob"
	"github.com/urfave/cli/v3"
)

type ListCertificate struct {
	Name           string   `json:"name,omitempty"`
	Domains        []string `json:"domains,omitempty"`
	IPs            []string `json:"ips,omitempty"`
	ExpirationDate string   `json:"expirationDate,omitempty"`
	Expired        bool     `json:"expired"`
	Issuer         string   `json:"issuer,omitempty"`
	Path           string   `json:"path,omitempty"`
}

func createListCertificates() *cli.Command {
	return &cli.Command{
		Name:   "certificates",
		Usage:  "Display information about certificates.",
		Action: listCertificates,
		Flags:  createListFlags(),
	}
}

func listCertificates(ctx context.Context, cmd *cli.Command) error {
	if cmd.Bool(flgFormatJSON) {
		return listCertificatesJSON(ctx, cmd)
	}

	return listCertificatesText(ctx, cmd)
}

func listCertificatesText(_ context.Context, cmd *cli.Command) error {
	certs, err := readCertificates(cmd)
	if err != nil {
		return err
	}

	if len(certs) == 0 {
		fmt.Println("No certificates were found.")

		return nil
	}

	fmt.Println("Found the following certificates:")

	for _, info := range certs {
		fmt.Println(info.Name)

		if info.Expired {
			fmt.Println("├── Status: this certificate is expired.")
		}

		if len(info.Domains) > 0 {
			fmt.Println("├── Domains:", strings.Join(info.Domains, ", "))
		}

		if len(info.IPs) > 0 {
			fmt.Println("├── IPs:", strings.Join(info.IPs, ","))
		}

		fmt.Println("├── Expiration Date:", info.ExpirationDate)
		fmt.Println("├── Issuer:", info.Issuer)
		fmt.Println("└── Certificate Path:", info.Path)
		fmt.Println()
	}

	return nil
}

func listCertificatesJSON(_ context.Context, cmd *cli.Command) error {
	certs, err := readCertificates(cmd)
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(certs)
}

func readCertificates(cmd *cli.Command) ([]ListCertificate, error) {
	certsStorage := storage.NewCertificatesStorage(cmd.String(flgPath))

	matches, err := zglob.Glob(filepath.Join(certsStorage.GetRootPath(), "**", "*.json"))
	if err != nil {
		return nil, err
	}

	var certificates []ListCertificate

	for _, filename := range matches {
		certFilename := strings.TrimSuffix(filename, storage.ExtResource) + storage.ExtCert

		data, err := os.ReadFile(certFilename)
		if err != nil {
			return nil, err
		}

		pCert, err := certcrypto.ParsePEMCertificate(data)
		if err != nil {
			return nil, err
		}

		name := strings.TrimSuffix(filepath.Base(certFilename), storage.ExtCert)

		certificates = append(certificates, ListCertificate{
			Name:           name,
			Domains:        pCert.DNSNames,
			IPs:            toStringSlice(pCert.IPAddresses),
			ExpirationDate: pCert.NotAfter.String(),
			Expired:        pCert.NotAfter.Before(time.Now()),
			Issuer:         pCert.Issuer.String(),
			Path:           certFilename,
		})
	}

	return certificates, nil
}

func toStringSlice[T fmt.Stringer](values []T) []string {
	var s []string

	for _, value := range values {
		s = append(s, value.String())
	}

	return s
}
