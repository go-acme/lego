package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certificate"
)

const (
	hookEnvAccountEmail      = "LEGO_ACCOUNT_EMAIL"
	hookEnvCertDomain        = "LEGO_CERT_DOMAIN"
	hookEnvCertPath          = "LEGO_CERT_PATH"
	hookEnvCertKeyPath       = "LEGO_CERT_KEY_PATH"
	hookEnvIssuerCertKeyPath = "LEGO_ISSUER_CERT_PATH"
	hookEnvCertPEMPath       = "LEGO_CERT_PEM_PATH"
	hookEnvCertPFXPath       = "LEGO_CERT_PFX_PATH"
)

func launchHook(hook string, timeout time.Duration, meta map[string]string) error {
	if hook == "" {
		return nil
	}

	ctxCmd, cancel := context.WithTimeout(context.Background(), timeout)
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

func addPathToMetadata(meta map[string]string, domain string, certRes *certificate.Resource, certsStorage *CertificatesStorage) {
	meta[hookEnvCertDomain] = domain
	meta[hookEnvCertPath] = certsStorage.GetFileName(domain, certExt)
	meta[hookEnvCertKeyPath] = certsStorage.GetFileName(domain, keyExt)

	if certRes.IssuerCertificate != nil {
		meta[hookEnvIssuerCertKeyPath] = certsStorage.GetFileName(domain, issuerExt)
	}

	if certsStorage.pem {
		meta[hookEnvCertPEMPath] = certsStorage.GetFileName(domain, pemExt)
	}

	if certsStorage.pfx {
		meta[hookEnvCertPFXPath] = certsStorage.GetFileName(domain, pfxExt)
	}
}
