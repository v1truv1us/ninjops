package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ReadsConfigJSONFromUserConfigDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	path := filepath.Join(tmp, ".config", "ninjops", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	data := []byte(`{
  "ninja": {
    "base_url": "https://json.example",
    "api_token": "token-json",
    "api_secret": "secret-json"
  },
  "agent": {
    "provider": "offline",
    "plan": "default"
  },
  "serve": {
    "listen": "127.0.0.1",
    "port": 9191
  }
}`)

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Ninja.BaseURL != "https://json.example" {
		t.Fatalf("unexpected base URL: %s", cfg.Ninja.BaseURL)
	}
	if cfg.Ninja.APIToken != "token-json" {
		t.Fatalf("unexpected API token: %s", cfg.Ninja.APIToken)
	}
	if cfg.Ninja.APISecret != "secret-json" {
		t.Fatalf("unexpected API secret: %s", cfg.Ninja.APISecret)
	}
	if cfg.Serve.Port != 9191 {
		t.Fatalf("unexpected serve port: %d", cfg.Serve.Port)
	}
}

func TestLoad_ReadsJSONCFromUserConfigDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	path := filepath.Join(tmp, ".config", "ninjops", "ninjops.jsonc")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	data := []byte(`{
  // Invoice Ninja settings
  "ninja": {
    "base_url": "https://jsonc.example",
    "api_token": "token-jsonc",
    "api_secret": "secret-jsonc" /* optional */
  },
  "agent": {
    "provider": "offline",
    "plan": "default"
  },
  "serve": {
    "listen": "127.0.0.1",
    "port": 9292
  }
}`)

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Ninja.BaseURL != "https://jsonc.example" {
		t.Fatalf("unexpected base URL: %s", cfg.Ninja.BaseURL)
	}
	if cfg.Ninja.APIToken != "token-jsonc" {
		t.Fatalf("unexpected API token: %s", cfg.Ninja.APIToken)
	}
	if cfg.Ninja.APISecret != "secret-jsonc" {
		t.Fatalf("unexpected API secret: %s", cfg.Ninja.APISecret)
	}
	if cfg.Serve.Port != 9292 {
		t.Fatalf("unexpected serve port: %d", cfg.Serve.Port)
	}
}

func TestResolveProviderAPIKey_UsesEnvBeforeConfiguredKey(t *testing.T) {
	t.Setenv("NINJOPS_OPENAI_API_KEY", "env-openai-key")

	got := ResolveProviderAPIKey("openai", "config-openai-key")
	if got != "env-openai-key" {
		t.Fatalf("unexpected key: %s", got)
	}
}

func TestResolveProviderAPIKey_UsesConfiguredWhenEnvMissing(t *testing.T) {
	t.Setenv("NINJOPS_OPENAI_API_KEY", "")

	got := ResolveProviderAPIKey("openai", "config-openai-key")
	if got != "config-openai-key" {
		t.Fatalf("unexpected key: %s", got)
	}
}

func TestResolveProviderAPIKey_OpencodePrefersNativeEnvVar(t *testing.T) {
	t.Setenv("OPENCODE_API_KEY", "native-opencode-key")
	t.Setenv("NINJOPS_OPENCODE_API_KEY", "legacy-opencode-key")

	got := ResolveProviderAPIKey("opencode", "config-opencode-key")
	if got != "native-opencode-key" {
		t.Fatalf("unexpected key: %s", got)
	}
}

func TestResolveProviderAPIKey_OpencodeFallsBackToLegacyEnvVar(t *testing.T) {
	t.Setenv("OPENCODE_API_KEY", "")
	t.Setenv("NINJOPS_OPENCODE_API_KEY", "legacy-opencode-key")

	got := ResolveProviderAPIKey("opencode", "config-opencode-key")
	if got != "legacy-opencode-key" {
		t.Fatalf("unexpected key: %s", got)
	}
}

func TestLoad_HydratesFromReferencedAuthCredsFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	configPath := filepath.Join(tmp, ".config", "ninjops", "config.json")
	authPath := filepath.Join(tmp, ".config", "ninjops", "auth-creds.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configData := []byte(`{
  "ninja": {
    "base_url": "https://json.example"
  },
  "agent": {
    "provider": "openai",
    "plan": "default"
  },
  "serve": {
    "listen": "127.0.0.1",
    "port": 9191
  },
  "auth_creds_file": "auth-creds.json"
}`)

	authData := []byte(`{
  "ninja": {
    "api_token": "token-from-auth",
    "api_secret": "secret-from-auth"
  },
  "agent": {
    "provider_api_key": "provider-key-from-auth"
  }
}`)

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	if err := os.WriteFile(authPath, authData, 0644); err != nil {
		t.Fatalf("failed to write auth creds file: %v", err)
	}

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Ninja.APIToken != "token-from-auth" {
		t.Fatalf("unexpected API token: %s", cfg.Ninja.APIToken)
	}
	if cfg.Ninja.APISecret != "secret-from-auth" {
		t.Fatalf("unexpected API secret: %s", cfg.Ninja.APISecret)
	}
	if cfg.Agent.ProviderAPIKey != "provider-key-from-auth" {
		t.Fatalf("unexpected provider API key: %s", cfg.Agent.ProviderAPIKey)
	}
}

func TestLoad_EnvOverridesAuthCredsValues(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("NINJOPS_NINJA_API_TOKEN", "token-from-env")
	t.Setenv("NINJOPS_NINJA_API_SECRET", "secret-from-env")
	t.Setenv("NINJOPS_AGENT_PROVIDER_API_KEY", "provider-key-from-env")

	configPath := filepath.Join(tmp, ".config", "ninjops", "config.json")
	authPath := filepath.Join(tmp, ".config", "ninjops", "auth-creds.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configData := []byte(`{
  "ninja": {
    "base_url": "https://json.example"
  },
  "agent": {
    "provider": "openai",
    "plan": "default"
  },
  "serve": {
    "listen": "127.0.0.1",
    "port": 9191
  },
  "auth_creds_file": "auth-creds.json"
}`)

	authData := []byte(`{
  "ninja": {
    "api_token": "token-from-auth",
    "api_secret": "secret-from-auth"
  },
  "agent": {
    "provider_api_key": "provider-key-from-auth"
  }
}`)

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	if err := os.WriteFile(authPath, authData, 0644); err != nil {
		t.Fatalf("failed to write auth creds file: %v", err)
	}

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Ninja.APIToken != "token-from-env" {
		t.Fatalf("unexpected API token: %s", cfg.Ninja.APIToken)
	}
	if cfg.Ninja.APISecret != "secret-from-env" {
		t.Fatalf("unexpected API secret: %s", cfg.Ninja.APISecret)
	}
	if cfg.Agent.ProviderAPIKey != "provider-key-from-env" {
		t.Fatalf("unexpected provider API key: %s", cfg.Agent.ProviderAPIKey)
	}
}

func TestLoad_MainConfigSecretsOverrideAuthCredsValues(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	configPath := filepath.Join(tmp, ".config", "ninjops", "config.json")
	authPath := filepath.Join(tmp, ".config", "ninjops", "auth-creds.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configData := []byte(`{
  "ninja": {
    "base_url": "https://json.example",
    "api_token": "token-from-main",
    "api_secret": "secret-from-main"
  },
  "agent": {
    "provider": "openai",
    "plan": "default",
    "provider_api_key": "provider-key-from-main"
  },
  "serve": {
    "listen": "127.0.0.1",
    "port": 9191
  },
  "auth_creds_file": "auth-creds.json"
}`)

	authData := []byte(`{
  "ninja": {
    "api_token": "token-from-auth",
    "api_secret": "secret-from-auth"
  },
  "agent": {
    "provider_api_key": "provider-key-from-auth"
  }
}`)

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	if err := os.WriteFile(authPath, authData, 0644); err != nil {
		t.Fatalf("failed to write auth creds file: %v", err)
	}

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Ninja.APIToken != "token-from-main" {
		t.Fatalf("unexpected API token: %s", cfg.Ninja.APIToken)
	}
	if cfg.Ninja.APISecret != "secret-from-main" {
		t.Fatalf("unexpected API secret: %s", cfg.Ninja.APISecret)
	}
	if cfg.Agent.ProviderAPIKey != "provider-key-from-main" {
		t.Fatalf("unexpected provider API key: %s", cfg.Agent.ProviderAPIKey)
	}
}

func TestNormalizeModelAlias_OpenAICodex(t *testing.T) {
	if got := NormalizeModelAlias("openai", "openai-codex"); got != ModelResolvedOpenAICodex {
		t.Fatalf("unexpected openai alias normalization: %s", got)
	}

	if got := NormalizeModelAlias("opencode", "openai-codex"); got != "gpt-5.3-codex" {
		t.Fatalf("unexpected opencode alias normalization: %s", got)
	}
}

func TestResolveProviderModel_OpenAICodexAliasUsesOpencode(t *testing.T) {
	provider, model := ResolveProviderModel("openai", "openai-codex")
	if provider != "opencode" {
		t.Fatalf("unexpected provider: %s", provider)
	}
	if model != "gpt-5.3-codex" {
		t.Fatalf("unexpected model: %s", model)
	}
}

func TestLoad_AcceptsOpenAICompatibleProvider(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	path := filepath.Join(tmp, ".config", "ninjops", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	data := []byte(`{
  "ninja": {
    "base_url": "https://json.example"
  },
  "agent": {
    "provider": "302ai",
    "plan": "default",
    "model": "qwen3-235b-a22b-instruct-2507"
  },
  "serve": {
    "listen": "127.0.0.1",
    "port": 9191
  }
}`)

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Agent.Provider != "302ai" {
		t.Fatalf("unexpected provider: %s", cfg.Agent.Provider)
	}
	if cfg.Agent.Model != "qwen3-235b-a22b-instruct-2507" {
		t.Fatalf("unexpected normalized model: %s", cfg.Agent.Model)
	}
}

func TestLoad_NormalizesOpenAICodexAliasToOpencodeProvider(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	path := filepath.Join(tmp, ".config", "ninjops", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	data := []byte(`{
  "ninja": {
    "base_url": "https://json.example"
  },
  "agent": {
    "provider": "openai",
    "plan": "default",
    "model": "openai-codex"
  },
  "serve": {
    "listen": "127.0.0.1",
    "port": 9191
  }
}`)

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Agent.Provider != "opencode" {
		t.Fatalf("unexpected provider: %s", cfg.Agent.Provider)
	}
	if cfg.Agent.Model != "gpt-5.3-codex" {
		t.Fatalf("unexpected normalized model: %s", cfg.Agent.Model)
	}
}
