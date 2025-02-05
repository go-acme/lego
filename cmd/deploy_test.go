package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCmd struct {
	runFunc func() error
}

func (f *fakeCmd) Run() error {
	return f.runFunc()
}

func TestRunDeploy_Success(t *testing.T) {
	originalRunner := commandRunner
	defer func() { commandRunner = originalRunner }()

	commandRunner = func(name string, args ...string) Commander {
		return &fakeCmd{
			runFunc: func() error {
				return nil
			},
		}
	}

	meta := map[string]string{
		"DEPLOY_HOSTS":      "10.10.1.6,10.10.1.8",
		"DEPLOY_REMOTE_DIR": "~/traefik/certificates",
		"DEPLOY_SSH_KEY":    "~/.ssh/node",
		"DEPLOY_NOMAD_JOB":  "traefik",
		hookEnvCertDomain:   "reports.uc-test.ch",
		"BASE_PATH":         "/etc/lego",
	}

	err := runDeploy(meta)
	require.NoError(t, err, "runDeploy should succeed when scp/nomad are mocked as success")
}

func TestRunDeploy_SCPKeyFails(t *testing.T) {
	originalRunner := commandRunner
	defer func() { commandRunner = originalRunner }()

	callCount := 0

	commandRunner = func(name string, args ...string) Commander {
		return &fakeCmd{
			runFunc: func() error {
				callCount++
				if callCount == 1 && strings.Contains(strings.Join(args, " "), ".key") {
					return errors.New("scp key error")
				}
				return nil
			},
		}
	}

	meta := map[string]string{
		"DEPLOY_HOSTS":      "10.10.1.6",
		"DEPLOY_REMOTE_DIR": "/some/dir",
		"DEPLOY_SSH_KEY":    "~/.ssh/node",
		"DEPLOY_NOMAD_JOB":  "traefik",
		hookEnvCertDomain:   "reports.uc-test.ch",
		"BASE_PATH":         "/etc/lego",
	}

	err := runDeploy(meta)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scp key error", "Expected runDeploy to fail on scp key")
}

func TestRunDeploy_NomadFails(t *testing.T) {
	originalRunner := commandRunner
	defer func() { commandRunner = originalRunner }()

	commandRunner = func(name string, args ...string) Commander {
		return &fakeCmd{
			runFunc: func() error {
				if name == "nomad" {
					return errors.New("nomad error")
				}
				return nil
			},
		}
	}

	meta := map[string]string{
		"DEPLOY_HOSTS":      "10.10.1.6",
		"DEPLOY_REMOTE_DIR": "~/traefik/certificates",
		"DEPLOY_SSH_KEY":    "~/.ssh/node",
		"DEPLOY_NOMAD_JOB":  "traefik",
		hookEnvCertDomain:   "reports.uc-test.ch",
		"BASE_PATH":         "/etc/lego",
	}

	err := runDeploy(meta)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nomad error", "Expected runDeploy to fail on nomad restart")
}
