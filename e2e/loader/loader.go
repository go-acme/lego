package loader

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/platform/wait"
	"github.com/ldez/grignotin/goenv"
)

const (
	cmdNamePebble   = "pebble"
	cmdNameChallSrv = "pebble-challtestsrv"
)

type CmdOption struct {
	HealthCheckURL string
	Args           []string
	Env            []string
	Dir            string
}

type EnvLoader struct {
	PebbleOptions *CmdOption
	LegoOptions   []string
	ChallSrv      *CmdOption
	lego          string
}

func (l *EnvLoader) MainTest(ctx context.Context, m *testing.M) int {
	if _, e2e := os.LookupEnv("LEGO_E2E_TESTS"); !e2e {
		fmt.Fprintln(os.Stderr, "skipping test: e2e tests are disabled. (no 'LEGO_E2E_TESTS' env var)")
		fmt.Println("PASS")

		return 0
	}

	if _, err := exec.LookPath("git"); err != nil {
		fmt.Fprintln(os.Stderr, "skipping because git command not found")
		fmt.Println("PASS")

		return 0
	}

	if l.PebbleOptions != nil {
		if _, err := exec.LookPath(cmdNamePebble); err != nil {
			fmt.Fprintln(os.Stderr, "skipping because pebble binary not found")
			fmt.Println("PASS")

			return 0
		}
	}

	if l.ChallSrv != nil {
		if _, err := exec.LookPath(cmdNameChallSrv); err != nil {
			fmt.Fprintln(os.Stderr, "skipping because challtestsrv binary not found")
			fmt.Println("PASS")

			return 0
		}
	}

	pebbleTearDown := l.launchPebble(ctx)
	defer pebbleTearDown()

	challSrvTearDown := l.launchChallSrv(ctx)
	defer challSrvTearDown()

	legoBinary, tearDown, err := buildLego(ctx)
	defer tearDown()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	l.lego = legoBinary

	if l.PebbleOptions != nil && l.PebbleOptions.HealthCheckURL != "" {
		pebbleHealthCheck(l.PebbleOptions)
	}

	return m.Run()
}

func (l *EnvLoader) RunLegoCombinedOutput(ctx context.Context, arg ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, l.lego, arg...)
	cmd.Env = l.LegoOptions

	fmt.Printf("$ %s\n", strings.Join(cmd.Args, " "))

	return cmd.CombinedOutput()
}

func (l *EnvLoader) RunLego(ctx context.Context, arg ...string) error {
	cmd := exec.CommandContext(ctx, l.lego, arg...)
	cmd.Env = l.LegoOptions

	fmt.Printf("$ %s\n", strings.Join(cmd.Args, " "))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create pipe: %w", err)
	}

	cmd.Stderr = cmd.Stdout

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		println(scanner.Text())
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("wait command: %w", err)
	}

	return nil
}

func (l *EnvLoader) launchPebble(ctx context.Context) func() {
	if l.PebbleOptions == nil {
		return func() {}
	}

	pebble, outPebble := l.cmdPebble(ctx)

	go func() {
		err := pebble.Run()
		if err != nil {
			fmt.Println(err)
		}
	}()

	return func() {
		err := pebble.Process.Kill()
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(outPebble.String())
	}
}

func (l *EnvLoader) cmdPebble(ctx context.Context) (*exec.Cmd, *bytes.Buffer) {
	cmd := exec.CommandContext(ctx, cmdNamePebble, l.PebbleOptions.Args...)
	cmd.Env = l.PebbleOptions.Env

	dir, err := filepath.Abs(l.PebbleOptions.Dir)
	if err != nil {
		panic(err)
	}

	cmd.Dir = dir

	fmt.Printf("$ %s\n", strings.Join(cmd.Args, " "))

	var b bytes.Buffer

	cmd.Stdout = &b
	cmd.Stderr = &b

	return cmd, &b
}

func pebbleHealthCheck(options *CmdOption) {
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	err := wait.For("pebble", 10*time.Second, 500*time.Millisecond, func() (bool, error) {
		resp, err := client.Get(options.HealthCheckURL)
		if err != nil {
			return false, err
		}

		if resp.StatusCode != http.StatusOK {
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		panic(err)
	}
}

func (l *EnvLoader) launchChallSrv(ctx context.Context) func() {
	if l.ChallSrv == nil {
		return func() {}
	}

	challtestsrv, outChalSrv := l.cmdChallSrv(ctx)

	go func() {
		err := challtestsrv.Run()
		if err != nil {
			fmt.Println(err)
		}
	}()

	return func() {
		err := challtestsrv.Process.Kill()
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(outChalSrv.String())
	}
}

func (l *EnvLoader) cmdChallSrv(ctx context.Context) (*exec.Cmd, *bytes.Buffer) {
	cmd := exec.CommandContext(ctx, cmdNameChallSrv, l.ChallSrv.Args...)

	fmt.Printf("$ %s\n", strings.Join(cmd.Args, " "))

	var b bytes.Buffer

	cmd.Stdout = &b
	cmd.Stderr = &b

	return cmd, &b
}

func buildLego(ctx context.Context) (string, func(), error) {
	here, err := os.Getwd()
	if err != nil {
		return "", func() {}, err
	}

	defer func() { _ = os.Chdir(here) }()

	buildPath, err := os.MkdirTemp("", "lego_test")
	if err != nil {
		return "", func() {}, err
	}

	projectRoot, err := getProjectRoot(ctx)
	if err != nil {
		return "", func() {}, err
	}

	err = os.Chdir(projectRoot)
	if err != nil {
		return "", func() {}, err
	}

	binary := filepath.Join(buildPath, "lego")

	err = build(ctx, binary)
	if err != nil {
		return "", func() {}, err
	}

	err = os.Chdir(here)
	if err != nil {
		return "", func() {}, err
	}

	return binary, func() {
		_ = os.RemoveAll(buildPath)

		CleanLegoFiles(ctx)
	}, nil
}

func getProjectRoot(ctx context.Context) (string, error) {
	git := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")

	output, err := git.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", output)
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func build(ctx context.Context, binary string) error {
	toolPath, err := goToolPath(ctx)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, toolPath, "build", "-o", binary)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", output)
		return err
	}

	return nil
}

func goToolPath(ctx context.Context) (string, error) {
	// inspired by go1.11.1/src/internal/testenv/testenv.go
	if os.Getenv("GO_GCFLAGS") != "" {
		return "", errors.New("'go build' not compatible with setting $GO_GCFLAGS")
	}

	if runtime.GOOS == "darwin" && strings.HasPrefix(runtime.GOARCH, "arm") {
		return "", fmt.Errorf("skipping test: 'go build' not available on %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	return goTool(ctx)
}

func goTool(ctx context.Context) (string, error) {
	var exeSuffix string
	if runtime.GOOS == "windows" {
		exeSuffix = ".exe"
	}

	goRoot, err := goenv.GetOne(ctx, goenv.GOROOT)
	if err != nil {
		return "", fmt.Errorf("cannot find go root: %w", err)
	}

	path := filepath.Join(goRoot, "bin", "go"+exeSuffix)
	if _, err = os.Stat(path); err == nil {
		return path, nil
	}

	goBin, err := exec.LookPath("go" + exeSuffix)
	if err != nil {
		return "", fmt.Errorf("cannot find go tool: %w", err)
	}

	return goBin, nil
}

func CleanLegoFiles(ctx context.Context) {
	cmd := exec.CommandContext(ctx, "rm", "-rf", ".lego")
	fmt.Printf("$ %s\n", strings.Join(cmd.Args, " "))

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
	}
}
