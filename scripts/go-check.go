package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const golangciLintVersion = "2.15.0"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(steps []string) error {
	if len(steps) == 0 {
		steps = []string{"all"}
	}
	dir, err := apiDir()
	if err != nil {
		return err
	}
	for _, step := range steps {
		if err := runStep(dir, step); err != nil {
			return err
		}
	}
	return nil
}

func apiDir() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("cannot locate scripts directory")
	}
	return filepath.Join(filepath.Dir(file), "..", "apps", "api"), nil
}

func runStep(dir, step string) error {
	switch step {
	case "fmt":
		return checkFmt(dir)
	case "vet":
		return command(dir, "go", "vet", "./...")
	case "lint":
		return checkLint(dir)
	case "test":
		return runTests(dir)
	case "build":
		return command(dir, "go", "build", "./...")
	case "static":
		return runAll(dir, "fmt", "vet", "lint")
	case "all":
		return runAll(dir, "fmt", "vet", "lint", "test", "build")
	default:
		return fmt.Errorf("usage: go run scripts/go-check.go {fmt|vet|lint|test|build|static|all}...")
	}
}

func runTests(dir string) error {
	args := []string{"test"}
	switch {
	case os.Getenv("CI") != "" || hasCCompiler():
		args = append(args, "-race")
	default:
		fmt.Fprintln(os.Stderr, "go-check: no C compiler found, running tests without -race (CI still runs them with it)")
	}
	return command(dir, "go", append(args, "./...")...)
}

func hasCCompiler() bool {
	if os.Getenv("CGO_ENABLED") == "0" {
		return false
	}
	for _, name := range []string{"gcc", "cc", "clang"} {
		if _, err := exec.LookPath(name); err == nil {
			return true
		}
	}
	return false
}

func runAll(dir string, steps ...string) error {
	for _, step := range steps {
		if err := runStep(dir, step); err != nil {
			return err
		}
	}
	return nil
}

func command(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func checkFmt(dir string) error {
	var out bytes.Buffer
	cmd := exec.Command("gofmt", "-l", ".")
	cmd.Dir = dir
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	if unformatted := strings.TrimSpace(out.String()); unformatted != "" {
		return fmt.Errorf("%s\ngofmt: run 'gofmt -w' on the files above", unformatted)
	}
	return nil
}

func checkLint(dir string) error {
	binary, err := findGolangciLint(dir)
	if err != nil {
		return err
	}
	version, _ := exec.Command(binary, "version").CombinedOutput()
	if !strings.Contains(string(version), "version "+golangciLintVersion) {
		first, _, _ := strings.Cut(strings.TrimSpace(string(version)), "\n")
		return fmt.Errorf("golangci-lint v%s required, found: %s", golangciLintVersion, first)
	}
	return command(dir, binary, "run", "./...")
}

func findGolangciLint(dir string) (string, error) {
	if path, err := exec.LookPath("golangci-lint"); err == nil {
		return path, nil
	}
	cmd := exec.Command("go", "env", "GOPATH")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err == nil {
		gopath := strings.TrimSpace(string(out))
		for _, name := range []string{"golangci-lint", "golangci-lint.exe"} {
			candidate := filepath.Join(gopath, "bin", name)
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate, nil
			}
		}
	}
	return "", fmt.Errorf("golangci-lint v%s not found. Install: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v%s", golangciLintVersion, golangciLintVersion)
}
