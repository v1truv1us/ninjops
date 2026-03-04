package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLISmoke(t *testing.T) {
	tmp := t.TempDir()
	repoRoot := mustRepoRoot(t)

	configPath := filepath.Join(tmp, "smoke-config.toml")
	writeFileOrFail(t, configPath, []byte(`[ninja]
base_url = "https://example.invalid"
api_token = ""
api_secret = ""

[agent]
provider = "offline"
plan = "default"

[serve]
listen = "127.0.0.1"
port = 8080
`))

	binaryPath := filepath.Join(tmp, "ninjops-smoke")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	build := exec.Command("go", "build", "-o", binaryPath, "./cmd/ninjops")
	build.Dir = repoRoot
	build.Env = sanitizedEnv("")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("failed to build CLI binary: %v\n%s", err, string(output))
	}

	quotePath := filepath.Join(tmp, "quote.json")
	outDir := filepath.Join(tmp, "artifacts")
	configuredPath := filepath.Join(tmp, "configured.json")
	configuredAuthPath := filepath.Join(tmp, "auth-creds.json")

	helpOutput := runCLI(t, repoRoot, tmp, binaryPath, configPath, "--help")
	assertContains(t, helpOutput, "ninjops")

	ninjaHelpOutput := runCLI(t, repoRoot, tmp, binaryPath, configPath, "ninja", "--help")
	assertContains(t, ninjaHelpOutput, "Invoice Ninja operations")

	newOutput := runCLI(t, repoRoot, tmp, binaryPath, configPath, "new", "quote", "--output", quotePath)
	assertContains(t, newOutput, "Created")

	validateOutput := runCLI(t, repoRoot, tmp, binaryPath, configPath, "validate", "--input", quotePath)
	assertContains(t, validateOutput, "Valid QuoteSpec")

	generateOutput := runCLI(t, repoRoot, tmp, binaryPath, configPath, "generate", "--input", quotePath, "--out-dir", outDir)
	assertContains(t, generateOutput, "Generated artifacts")

	configureOutput := runCLI(
		t,
		repoRoot,
		tmp,
		binaryPath,
		configPath,
		"configure",
		"--non-interactive",
		"--base-url", "https://configured.example.invalid",
		"--api-token", "token-smoke",
		"--api-secret", "secret-smoke",
		"--provider", "offline",
		"--plan", "default",
		"--listen", "127.0.0.1",
		"--port", "7070",
		"--output", configuredPath,
	)
	assertContains(t, configureOutput, "Wrote config")

	assertFileExists(t, quotePath)
	assertFileExists(t, filepath.Join(outDir, "proposal.md"))
	assertFileExists(t, filepath.Join(outDir, "terms.md"))
	assertFileExists(t, filepath.Join(outDir, "notes.txt"))
	assertFileExists(t, filepath.Join(outDir, "generated.json"))
	assertFileExists(t, configuredPath)
	assertFileExists(t, configuredAuthPath)
}

func runCLI(t *testing.T, dir string, tmp string, binaryPath string, configPath string, args ...string) string {
	t.Helper()

	baseArgs := append([]string{"--config", configPath}, args...)
	cmd := exec.Command(binaryPath, baseArgs...)
	cmd.Dir = dir
	cmd.Env = sanitizedEnv(tmp)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %s %s\n%v\n%s", binaryPath, strings.Join(baseArgs, " "), err, string(output))
	}

	return string(output)
}

func mustRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func writeFileOrFail(t *testing.T, path string, data []byte) {
	t.Helper()

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected file to exist: %s (%v)", path, err)
	}

	if info.IsDir() {
		t.Fatalf("expected file but found directory: %s", path)
	}
}

func assertContains(t *testing.T, output string, want string) {
	t.Helper()

	if !strings.Contains(output, want) {
		t.Fatalf("expected output to contain %q, got:\n%s", want, output)
	}
}

func sanitizedEnv(home string) []string {
	env := make([]string, 0)
	for _, item := range os.Environ() {
		if strings.HasPrefix(item, "NINJOPS_") {
			continue
		}
		if home != "" && strings.HasPrefix(item, "HOME=") {
			continue
		}
		env = append(env, item)
	}

	if home != "" {
		env = append(env, "HOME="+home)
	}
	return env
}
