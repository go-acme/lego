package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func launchHook(hook string, meta map[string]string) error {
	if hook == "" {
		return nil
	}

	ctxCmd, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	parts := strings.Fields(hook)

	cmdCtx := exec.CommandContext(ctxCmd, parts[0], parts[1:]...)
	cmdCtx.Env = append(os.Environ(), metaToEnv(meta)...)

	output, err := cmdCtx.CombinedOutput()

	if len(output) > 0 {
		fmt.Println(string(output))
	}

	if errors.Is(ctxCmd.Err(), context.DeadlineExceeded) {
		return errors.New("hook timed out")
	}

	return err
}

func metaToEnv(meta map[string]string) []string {
	var envs []string

	for k, v := range meta {
		envs = append(envs, k+"="+v)
	}

	return envs
}
