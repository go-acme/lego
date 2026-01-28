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
)

const (
	EnvAccountEmail      = "LEGO_ACCOUNT_EMAIL"
	EnvCertDomain        = "LEGO_CERT_DOMAIN"
	EnvCertPath          = "LEGO_CERT_PATH"
	EnvCertKeyPath       = "LEGO_CERT_KEY_PATH"
	EnvIssuerCertKeyPath = "LEGO_ISSUER_CERT_PATH"
	EnvCertPEMPath       = "LEGO_CERT_PEM_PATH"
	EnvCertPFXPath       = "LEGO_CERT_PFX_PATH"
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
