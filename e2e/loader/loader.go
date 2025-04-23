package loader

import (
	"bufio"
	"bytes"
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

	"github.com/go-acme/lego/v4/platform/wait"
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

func (l *EnvLoader) MainTest(m *testing.M) int {
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

	pebbleTearDown := l.launchPebble()
	defer pebbleTearDown()

	challSrvTearDown := l.launchChallSrv()
	defer challSrvTearDown()

	legoBinary, tearDown, err := buildLego()
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

func (l *EnvLoader) RunLegoCombinedOutput(arg ...string) ([]byte, error) {
	cmd := exec.Command(l.lego, arg...)
	cmd.Env = l.LegoOptions

	fmt.Printf("$ %s\n", strings.Join(cmd.Args, " "))

	return cmd.CombinedOutput()
}

func (l *EnvLoader) RunLego(arg ...string) error {
	cmd := exec.Command(l.lego, arg...)
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

func (l *EnvLoader) launchPebble() func() {
	if l.PebbleOptions == nil {
		return func() {}
	}

	pebble, outPebble := l.cmdPebble()
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

func (l *EnvLoader) cmdPebble() (*exec.Cmd, *bytes.Buffer) {
	cmd := exec.Command(cmdNamePebble, l.PebbleOptions.Args...)
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

func (l *EnvLoader) launchChallSrv() func() {
	if l.ChallSrv == nil {
		return func() {}
	}

	challtestsrv, outChalSrv := l.cmdChallSrv()
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

func (l *EnvLoader) cmdChallSrv() (*exec.Cmd, *bytes.Buffer) {
	cmd := exec.Command(cmdNameChallSrv, l.ChallSrv.Args...)

	fmt.Printf("$ %s\n", strings.Join(cmd.Args, " "))

	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b

	return cmd, &b
}

func buildLego() (string, func(), error) {
	here, err := os.Getwd()
	if err != nil {
		return "", func() {}, err
	}
	defer func() { _ = os.Chdir(here) }()

	buildPath, err := os.MkdirTemp("", "lego_test")
	if err != nil {
		return "", func() {}, err
	}

	projectRoot, err := getProjectRoot()
	if err != nil {
		return "", func() {}, err
	}

	mainFolder := filepath.Join(projectRoot, "cmd", "lego")

	err = os.Chdir(mainFolder)
	if err != nil {
		return "", func() {}, err
	}

	binary := filepath.Join(buildPath, "lego")

	err = build(binary)
	if err != nil {
		return "", func() {}, err
	}

	err = os.Chdir(here)
	if err != nil {
		return "", func() {}, err
	}

	return binary, func() {
		_ = os.RemoveAll(buildPath)
		CleanLegoFiles()
	}, nil
}

func getProjectRoot() (string, error) {
	git := exec.Command("git", "rev-parse", "--show-toplevel")

	output, err := git.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", output)
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func build(binary string) error {
	toolPath, err := goToolPath()
	if err != nil {
		return err
	}
	cmd := exec.Command(toolPath, "build", "-o", binary)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", output)
		return err
	}

	return nil
}

func goToolPath() (string, error) {
	// inspired by go1.11.1/src/internal/testenv/testenv.go
	if os.Getenv("GO_GCFLAGS") != "" {
		return "", errors.New("'go build' not compatible with setting $GO_GCFLAGS")
	}

	if runtime.GOOS == "darwin" && strings.HasPrefix(runtime.GOARCH, "arm") {
		return "", fmt.Errorf("skipping test: 'go build' not available on %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	return goTool()
}

func goTool() (string, error) {
	var exeSuffix string
	if runtime.GOOS == "windows" {
		exeSuffix = ".exe"
	}

	path := filepath.Join(runtime.GOROOT(), "bin", "go"+exeSuffix)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	goBin, err := exec.LookPath("go" + exeSuffix)
	if err != nil {
		return "", fmt.Errorf("cannot find go tool: %w", err)
	}

	return goBin, nil
}

func CleanLegoFiles() {
	cmd := exec.Command("rm", "-rf", ".lego")
	fmt.Printf("$ %s\n", strings.Join(cmd.Args, " "))
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
	}
}
