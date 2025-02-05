package cmd

import (
	"fmt"
	"os/exec"
	"strings"
)

type Commander interface {
	Run() error
}

type realCmd struct {
	cmd *exec.Cmd
}

func (r *realCmd) Run() error {
	return r.cmd.Run()
}

var commandRunner = func(name string, args ...string) Commander {
	return &realCmd{cmd: exec.Command(name, args...)}
}

func runDeploy(meta map[string]string) error {
	deployHosts := meta["DEPLOY_HOSTS"]
	remoteDir := meta["DEPLOY_REMOTE_DIR"]
	sshKey := meta["DEPLOY_SSH_KEY"]
	nomadJob := meta["DEPLOY_NOMAD_JOB"]
	domain := meta[hookEnvCertDomain]
	basePath := meta["BASE_PATH"]

	certPath := fmt.Sprintf("%s/certificates/%s.crt", basePath, domain)
	keyPath := fmt.Sprintf("%s/certificates/%s.key", basePath, domain)

	hosts := strings.Split(deployHosts, ",")
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		fmt.Printf("Deploying to host %s\n", host)

		scpKey := commandRunner("scp", "-i", sshKey, keyPath, fmt.Sprintf("uc@%s:%s", host, remoteDir))
		if err := scpKey.Run(); err != nil {
			return fmt.Errorf("failed to copy key to host %s: %v", host, err)
		}

		scpCert := commandRunner("scp", "-i", sshKey, certPath, fmt.Sprintf("uc@%s:%s", host, remoteDir))
		if err := scpCert.Run(); err != nil {
			return fmt.Errorf("failed to copy cert to host %s: %v", host, err)
		}
	}

	nomadCmd := commandRunner("nomad", "job", "restart", nomadJob)
	if err := nomadCmd.Run(); err != nil {
		return fmt.Errorf("failed to restart Nomad job %s: %v", nomadJob, err)
	}

	return nil
}
