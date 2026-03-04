package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ninjops/ninjops/internal/config"
)

func TestConfigureCmd_WritesDefaultConfigJSON(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	prevCfg := cfg
	cfg = config.DefaultConfig()
	t.Cleanup(func() {
		cfg = prevCfg
	})

	cmd := newConfigureCmd()
	cmd.SetArgs([]string{
		"--non-interactive",
		"--base-url", "https://example.invalid",
		"--api-token", "token-123",
		"--api-secret", "secret-123",
		"--provider", "offline",
		"--plan", "default",
		"--listen", "0.0.0.0",
		"--port", "9090",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("configure command failed: %v", err)
	}

	path := filepath.Join(tmp, ".config", "ninjops", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read config output: %v", err)
	}
	authPath := filepath.Join(tmp, ".config", "ninjops", "auth-creds.json")
	authData, err := os.ReadFile(authPath)
	if err != nil {
		t.Fatalf("failed to read auth-creds output: %v", err)
	}

	var written map[string]interface{}
	if err := json.Unmarshal(data, &written); err != nil {
		t.Fatalf("failed to parse written config: %v", err)
	}
	var authWritten map[string]interface{}
	if err := json.Unmarshal(authData, &authWritten); err != nil {
		t.Fatalf("failed to parse written auth-creds: %v", err)
	}

	ninja := written["ninja"].(map[string]interface{})
	agent := written["agent"].(map[string]interface{})
	serve := written["serve"].(map[string]interface{})
	authNinja := authWritten["ninja"].(map[string]interface{})
	authAgent := authWritten["agent"].(map[string]interface{})

	if ninja["base_url"] != "https://example.invalid" {
		t.Fatalf("unexpected ninja.base_url: %v", ninja["base_url"])
	}
	if _, exists := ninja["api_token"]; exists {
		t.Fatalf("ninja.api_token should not be written to main config: %v", ninja["api_token"])
	}
	if _, exists := ninja["api_secret"]; exists {
		t.Fatalf("ninja.api_secret should not be written to main config: %v", ninja["api_secret"])
	}
	if authNinja["api_token"] != "token-123" {
		t.Fatalf("unexpected auth ninja.api_token: %v", authNinja["api_token"])
	}
	if authNinja["api_secret"] != "secret-123" {
		t.Fatalf("unexpected auth ninja.api_secret: %v", authNinja["api_secret"])
	}
	if agent["provider"] != "offline" {
		t.Fatalf("unexpected agent.provider: %v", agent["provider"])
	}
	if agent["plan"] != "default" {
		t.Fatalf("unexpected agent.plan: %v", agent["plan"])
	}
	if agent["model"] != config.DefaultModelForProvider("offline") {
		t.Fatalf("unexpected agent.model: %v", agent["model"])
	}
	if _, exists := agent["provider_api_key"]; exists {
		t.Fatalf("agent.provider_api_key should not be written to main config: %v", agent["provider_api_key"])
	}
	if authAgent["provider_api_key"] != "" {
		t.Fatalf("unexpected auth agent.provider_api_key: %v", authAgent["provider_api_key"])
	}
	if written["auth_creds_file"] != authPath {
		t.Fatalf("unexpected auth_creds_file: %v", written["auth_creds_file"])
	}
	if serve["listen"] != "0.0.0.0" {
		t.Fatalf("unexpected serve.listen: %v", serve["listen"])
	}
	if serve["port"].(float64) != 9090 {
		t.Fatalf("unexpected serve.port: %v", serve["port"])
	}
}

func TestConfigureCmd_NonOfflineProviderRequiresAPIKey(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	prevCfg := cfg
	cfg = config.DefaultConfig()
	t.Cleanup(func() {
		cfg = prevCfg
	})

	cmd := newConfigureCmd()
	cmd.SetArgs([]string{
		"--non-interactive",
		"--provider", "openai",
		"--plan", "default",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected configure to fail without provider API key")
	}

	if got, want := err.Error(), "openai provider requires an API key"; !strings.Contains(got, want) {
		t.Fatalf("unexpected error %q, expected to contain %q", got, want)
	}
}

func TestConfigureCmd_SkipProviderTestWritesProviderKey(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	prevCfg := cfg
	cfg = config.DefaultConfig()
	t.Cleanup(func() {
		cfg = prevCfg
	})

	outputPath := filepath.Join(tmp, "configured.json")
	cmd := newConfigureCmd()
	cmd.SetArgs([]string{
		"--non-interactive",
		"--provider", "openai",
		"--plan", "default",
		"--provider-api-key", "sk-configured-key",
		"--skip-provider-test",
		"--output", outputPath,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("configure command failed: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read config output: %v", err)
	}
	authPath := filepath.Join(tmp, "auth-creds.json")
	authData, err := os.ReadFile(authPath)
	if err != nil {
		t.Fatalf("failed to read auth-creds output: %v", err)
	}

	var written map[string]interface{}
	if err := json.Unmarshal(data, &written); err != nil {
		t.Fatalf("failed to parse written config: %v", err)
	}
	var authWritten map[string]interface{}
	if err := json.Unmarshal(authData, &authWritten); err != nil {
		t.Fatalf("failed to parse written auth-creds: %v", err)
	}

	agent := written["agent"].(map[string]interface{})
	authAgent := authWritten["agent"].(map[string]interface{})
	if _, exists := agent["provider_api_key"]; exists {
		t.Fatalf("agent.provider_api_key should not be written to main config: %v", agent["provider_api_key"])
	}
	if authAgent["provider_api_key"] != "sk-configured-key" {
		t.Fatalf("unexpected auth agent.provider_api_key: %v", authAgent["provider_api_key"])
	}
	if written["auth_creds_file"] != authPath {
		t.Fatalf("unexpected auth_creds_file: %v", written["auth_creds_file"])
	}
}

func TestConfigureCmd_UsesEnvProviderKeyWhenFlagMissing(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("NINJOPS_OPENAI_API_KEY", "sk-from-env")

	prevCfg := cfg
	cfg = config.DefaultConfig()
	t.Cleanup(func() {
		cfg = prevCfg
	})

	outputPath := filepath.Join(tmp, "configured-env.json")
	cmd := newConfigureCmd()
	cmd.SetArgs([]string{
		"--non-interactive",
		"--provider", "openai",
		"--plan", "default",
		"--skip-provider-test",
		"--output", outputPath,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("configure command failed: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read config output: %v", err)
	}
	authPath := filepath.Join(tmp, "auth-creds.json")
	authData, err := os.ReadFile(authPath)
	if err != nil {
		t.Fatalf("failed to read auth-creds output: %v", err)
	}

	var written map[string]interface{}
	if err := json.Unmarshal(data, &written); err != nil {
		t.Fatalf("failed to parse written config: %v", err)
	}
	var authWritten map[string]interface{}
	if err := json.Unmarshal(authData, &authWritten); err != nil {
		t.Fatalf("failed to parse written auth-creds: %v", err)
	}

	authAgent := authWritten["agent"].(map[string]interface{})
	if authAgent["provider_api_key"] != "sk-from-env" {
		t.Fatalf("unexpected auth agent.provider_api_key: %v", authAgent["provider_api_key"])
	}
	if written["auth_creds_file"] != authPath {
		t.Fatalf("unexpected auth_creds_file: %v", written["auth_creds_file"])
	}
}

func TestConfigureCmd_UsesExplicitAuthCredsOutput(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	prevCfg := cfg
	cfg = config.DefaultConfig()
	t.Cleanup(func() {
		cfg = prevCfg
	})

	outputPath := filepath.Join(tmp, "custom", "configured.json")
	authOutputPath := filepath.Join(tmp, "secrets", "custom-auth.json")

	cmd := newConfigureCmd()
	cmd.SetArgs([]string{
		"--non-interactive",
		"--provider", "offline",
		"--plan", "default",
		"--api-token", "token-custom",
		"--api-secret", "secret-custom",
		"--output", outputPath,
		"--auth-creds-output", authOutputPath,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("configure command failed: %v", err)
	}

	configData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read config output: %v", err)
	}
	if _, err := os.ReadFile(authOutputPath); err != nil {
		t.Fatalf("failed to read auth output: %v", err)
	}

	var written map[string]interface{}
	if err := json.Unmarshal(configData, &written); err != nil {
		t.Fatalf("failed to parse written config: %v", err)
	}

	if written["auth_creds_file"] != authOutputPath {
		t.Fatalf("unexpected auth_creds_file: %v", written["auth_creds_file"])
	}
}

func TestConfigureCmd_OpenAICodexAliasPersistsOpencodeProviderAndModel(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	prevCfg := cfg
	cfg = config.DefaultConfig()
	t.Cleanup(func() {
		cfg = prevCfg
	})

	outputPath := filepath.Join(tmp, "configured-alias.json")
	cmd := newConfigureCmd()
	cmd.SetArgs([]string{
		"--non-interactive",
		"--provider", "openai",
		"--plan", "default",
		"--model", "openai-codex",
		"--provider-api-key", "sk-configured-key",
		"--skip-provider-test",
		"--output", outputPath,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("configure command failed: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read config output: %v", err)
	}

	var written map[string]interface{}
	if err := json.Unmarshal(data, &written); err != nil {
		t.Fatalf("failed to parse written config: %v", err)
	}

	agent := written["agent"].(map[string]interface{})
	if agent["provider"] != "opencode" {
		t.Fatalf("unexpected resolved provider: %v", agent["provider"])
	}
	if agent["model"] != "gpt-5.3-codex" {
		t.Fatalf("unexpected normalized model: %v", agent["model"])
	}
}
