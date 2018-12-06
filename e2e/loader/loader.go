package loader

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/xenolf/lego/platform/wait"
)

const (
	cmdNamePebble   = "pebble"
	cmdNameChallSrv = "challtestsrv"
)

type CmdOption struct {
	Args []string
	Env  []string
	Dir  string
}

type EnvLoader struct {
	PebbleOptions *CmdOption
	LegoOptions   []string
	ChallSrv      *CmdOption
	lego          string
}

func (l *EnvLoader) MainTest(m *testing.M) int {
	_, force := os.LookupEnv("LEGO_E2E_TESTS")
	if _, ci := os.LookupEnv("CI"); !ci && !force {
		fmt.Fprintln(os.Stderr, "skipping test: e2e tests are disabled. (no 'CI' or 'LEGO_E2E_TESTS' env var)")
		fmt.Println("PASS")
		return 0
	}

	if _, err := exec.LookPath("git"); err != nil {
		fmt.Fprintln(os.Stderr, "skipping because git command not found")
		fmt.Println("PASS")
		return 0
	}

	if _, err := exec.LookPath(cmdNamePebble); err != nil {
		fmt.Fprintln(os.Stderr, "skipping because pebble binary not found")
		fmt.Println("PASS")
		return 0
	}

	if _, err := exec.LookPath(cmdNameChallSrv); err != nil {
		fmt.Fprintln(os.Stderr, "skipping because challtestsrv binary not found")
		fmt.Println("PASS")
		return 0
	}

	pebbleTearDown := func() {}
	if l.PebbleOptions != nil {
		pebble, outPebble := l.cmdPebble()
		go func() {
			err := pebble.Run()
			if err != nil {
				fmt.Println(err)
			}
		}()
		pebbleTearDown = func() {
			pebble.Process.Kill()
			fmt.Println(outPebble.String())
			CleanLegoFiles()
		}
	}
	defer pebbleTearDown()

	challSrvTearDown := func() {}
	if l.ChallSrv != nil {
		challtestsrv, outChalSrv := l.cmdChallSrv()
		go func() {
			err := challtestsrv.Run()
			if err != nil {
				fmt.Println(err)
			}
		}()
		challSrvTearDown = func() {
			challtestsrv.Process.Kill()
			fmt.Println(outChalSrv.String())
			CleanLegoFiles()
		}
	}
	defer challSrvTearDown()

	legoBinary, tearDown, err := buildLego()
	defer tearDown()
	if err != nil {
		return 1
	}

	l.lego = legoBinary

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	wait.For(10*time.Second, 500*time.Millisecond, func() (bool, error) {
		resp, err := client.Get("https://localhost:14000/dir")
		if err != nil {
			return false, err
		}

		if resp.StatusCode != http.StatusOK {
			return false, nil
		}

		return true, nil
	})

	return m.Run()
}

func (l *EnvLoader) RunLego(arg ...string) ([]byte, error) {
	cmd := exec.Command(l.lego, arg...)
	cmd.Env = l.LegoOptions

	fmt.Printf("$ %s\n", strings.Join(cmd.Args, " "))

	return cmd.CombinedOutput()
}

func (l *EnvLoader) cmdPebble() (*exec.Cmd, *bytes.Buffer) {
	cmd := exec.Command(cmdNamePebble, l.PebbleOptions.Args...)
	cmd.Env = l.PebbleOptions.Env
	cmd.Dir, _ = filepath.Abs(l.PebbleOptions.Dir)

	fmt.Printf("$ %s\n", strings.Join(cmd.Args, " "))

	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b

	return cmd, &b
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

	buildPath, err := ioutil.TempDir("", "lego_test")
	if err != nil {
		return "", func() {}, err
	}

	projectRoot, err := getProjectRoot()
	if err != nil {
		return "", func() {}, err
	}

	err = os.Chdir(projectRoot)
	if err != nil {
		return "", func() {}, err
	}

	binary := filepath.Join(buildPath, "lego")

	build(binary)

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
		return "", fmt.Errorf("cannot find go tool: %v", err)
	}

	return goBin, nil
}

func CleanLegoFiles() {
	cmd := exec.Command("rm", "-rf", ".lego")
	fmt.Printf("$ %s\n", strings.Join(cmd.Args, " "))
	cmd.CombinedOutput()
}
