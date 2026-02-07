package hook

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
)

// TODO(ldez) rename the env vars with LEGO_HOOK_ prefix to avoid collisions with flag names.
const (
	EnvAccountEmail      = "LEGO_ACCOUNT_EMAIL"
	EnvCertDomain        = "LEGO_CERT_DOMAIN"
	EnvCertPath          = "LEGO_CERT_PATH"
	EnvCertKeyPath       = "LEGO_CERT_KEY_PATH"
	EnvIssuerCertKeyPath = "LEGO_ISSUER_CERT_PATH"
	EnvCertPEMPath       = "LEGO_CERT_PEM_PATH"
	EnvCertPFXPath       = "LEGO_CERT_PFX_PATH"
)

// TODO(ldez): merge this with the previous constant block.
const (
	EnvCertNameSanitized = "LEGO_HOOK_CERT_NAME_SANITIZED"
	EnvCertID            = "LEGO_HOOK_CERT_ID"
	EnvCertDomains       = "LEGO_HOOK_CERT_DOMAINS"
)

func Launch(ctx context.Context, hook string, timeout time.Duration, meta map[string]string) error {
	if hook == "" {
		return nil
	}

	ctxCmd, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	parts := strings.Fields(hook)

	cmd := exec.CommandContext(ctxCmd, parts[0], parts[1:]...)

	cmd.Env = append(os.Environ(), metaToEnv(meta)...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create pipe: %w", err)
	}

	cmd.Stderr = cmd.Stdout

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	go func() {
		<-ctxCmd.Done()

		if ctxCmd.Err() != nil {
			_ = cmd.Process.Kill()
			_ = stdout.Close()
		}
	}()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	err = cmd.Wait()
	if err != nil {
		if errors.Is(ctxCmd.Err(), context.DeadlineExceeded) {
			return errors.New("hook timed out")
		}

		return fmt.Errorf("wait command: %w", err)
	}

	return nil
}

func metaToEnv(meta map[string]string) []string {
	var envs []string

	for k, v := range meta {
		envs = append(envs, k+"="+v)
	}

	return envs
}

// AddPathToMetadata adds information about the certificate to the metadata map.
func AddPathToMetadata(meta map[string]string, certRes *certificate.Resource, certsStorage *storage.CertificatesStorage, options *storage.SaveOptions) {
	meta[EnvCertID] = certRes.ID
	meta[EnvCertNameSanitized] = storage.SanitizedName(certRes.ID)

	meta[EnvCertDomains] = strings.Join(certRes.Domains, ",")

	meta[EnvCertPath] = certsStorage.GetFileName(certRes.ID, storage.ExtCert)
	meta[EnvCertKeyPath] = certsStorage.GetFileName(certRes.ID, storage.ExtKey)

	if certRes.IssuerCertificate != nil {
		meta[EnvIssuerCertKeyPath] = certsStorage.GetFileName(certRes.ID, storage.ExtIssuer)
	}

	if options.PEM {
		meta[EnvCertPEMPath] = certsStorage.GetFileName(certRes.ID, storage.ExtPEM)
	}

	if options.PFX {
		meta[EnvCertPFXPath] = certsStorage.GetFileName(certRes.ID, storage.ExtPFX)
	}
}
