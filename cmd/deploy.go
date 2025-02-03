package cmd

import (
	"fmt"
	"os/exec"
	"strings"
)

// runDeploy uses values from meta to deploy the renewed certificate.
func runDeploy(meta map[string]string) error {
	deployHosts := meta["DEPLOY_HOSTS"]    // e.g., "10.10.1.6,10.10.1.8"
	remoteDir := meta["DEPLOY_REMOTE_DIR"] // e.g., "~/traefik/certificates"
	sshKey := meta["DEPLOY_SSH_KEY"]       // e.g., "~/.ssh/node"
	nomadJob := meta["DEPLOY_NOMAD_JOB"]   // e.g., "traefik"
	domain := meta[hookEnvCertDomain]      // Should be set by addPathToMetadata
	basePath := meta["BASE_PATH"]          // Set this from a CLI flag or configuration

	certPath := fmt.Sprintf("%s/certificates/%s.crt", basePath, domain)
	keyPath := fmt.Sprintf("%s/certificates/%s.key", basePath, domain)

	hosts := strings.Split(deployHosts, ",")
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		fmt.Printf("Deploying to host %s\n", host)
		scpKey := exec.Command("scp", "-i", sshKey, keyPath, fmt.Sprintf("uc@%s:%s", host, remoteDir))
		if err := scpKey.Run(); err != nil {
			return fmt.Errorf("failed to copy key to host %s: %v", host, err)
		}
		scpCert := exec.Command("scp", "-i", sshKey, certPath, fmt.Sprintf("uc@%s:%s", host, remoteDir))
		if err := scpCert.Run(); err != nil {
			return fmt.Errorf("failed to copy cert to host %s: %v", host, err)
		}
	}

	nomadCmd := exec.Command("nomad", "job", "restart", nomadJob)
	if err := nomadCmd.Run(); err != nil {
		return fmt.Errorf("failed to restart Nomad job %s: %v", nomadJob, err)
	}
	return nil
}
