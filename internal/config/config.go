package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Ninja         NinjaConfig `mapstructure:"ninja"`
	Agent         AgentConfig `mapstructure:"agent"`
	Serve         ServeConfig `mapstructure:"serve"`
	AuthCredsFile string      `mapstructure:"auth_creds_file"`
}

type NinjaConfig struct {
	BaseURL   string `mapstructure:"base_url"`
	APIToken  string `mapstructure:"api_token"`
	APISecret string `mapstructure:"api_secret"`
}

type AgentConfig struct {
	Provider       string `mapstructure:"provider"`
	Plan           string `mapstructure:"plan"`
	Model          string `mapstructure:"model"`
	ProviderAPIKey string `mapstructure:"provider_api_key"`
}

type ServeConfig struct {
	Listen string `mapstructure:"listen"`
	Port   int    `mapstructure:"port"`
}

type ConfigLoader struct {
	v *viper.Viper
}

type authCredsFile struct {
	Ninja struct {
		APIToken  string `json:"api_token"`
		APISecret string `json:"api_secret"`
	} `json:"ninja"`
	Agent struct {
		ProviderAPIKey string `json:"provider_api_key"`
	} `json:"agent"`
}

func NewConfigLoader() *ConfigLoader {
	v := viper.New()
	return &ConfigLoader{v: v}
}

func (cl *ConfigLoader) Load(configPath string) (*Config, error) {
	cl.setDefaults()
	cl.bindEnvVars()

	loadedConfigPath, err := cl.loadConfigFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := cl.v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if err := cl.hydrateFromAuthCreds(&cfg, loadedConfigPath); err != nil {
		return nil, err
	}

	cfg.Agent.Provider, cfg.Agent.Model = ResolveProviderModel(cfg.Agent.Provider, cfg.Agent.Model)

	if err := cl.validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func (cl *ConfigLoader) loadConfigFile(configPath string) (string, error) {
	if configPath != "" {
		if err := cl.readConfigAtPath(configPath); err != nil {
			return "", err
		}
		return configPath, nil
	}

	home := os.Getenv("HOME")
	candidates := []string{
		filepath.Join(home, ".config", "ninjops", "config.json"),
		filepath.Join(home, ".config", "ninjops", "ninjops.jsonc"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			if err := cl.readConfigAtPath(candidate); err != nil {
				return "", err
			}
			return candidate, nil
		}
	}

	cl.setupConfigSearch(configPath)
	if err := cl.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return "", nil
		}
		return "", fmt.Errorf("error reading config file: %w", err)
	}

	return cl.v.ConfigFileUsed(), nil
}

func (cl *ConfigLoader) hydrateFromAuthCreds(cfg *Config, loadedConfigPath string) error {
	authPath := strings.TrimSpace(cfg.AuthCredsFile)
	if authPath == "" {
		return nil
	}

	resolvedPath := authPath
	if !filepath.IsAbs(authPath) {
		if loadedConfigPath != "" {
			resolvedPath = filepath.Join(filepath.Dir(loadedConfigPath), authPath)
		} else {
			absPath, err := filepath.Abs(authPath)
			if err != nil {
				return fmt.Errorf("error resolving auth credentials file path: %w", err)
			}
			resolvedPath = absPath
		}
	}

	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error reading auth credentials file: %w", err)
	}

	var creds authCredsFile
	if err := json.Unmarshal(data, &creds); err != nil {
		return fmt.Errorf("error parsing auth credentials file: %w", err)
	}

	if cfg.Ninja.APIToken == "" {
		cfg.Ninja.APIToken = creds.Ninja.APIToken
	}
	if cfg.Ninja.APISecret == "" {
		cfg.Ninja.APISecret = creds.Ninja.APISecret
	}
	if cfg.Agent.ProviderAPIKey == "" {
		cfg.Agent.ProviderAPIKey = creds.Agent.ProviderAPIKey
	}

	return nil
}

func (cl *ConfigLoader) readConfigAtPath(path string) error {
	ext := strings.ToLower(filepath.Ext(path))

	if ext == ".jsonc" {
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return fmt.Errorf("error reading config file: %w", err)
		}

		cleaned := stripJSONComments(data)
		cl.v.SetConfigType("json")
		if err := cl.v.ReadConfig(bytes.NewReader(cleaned)); err != nil {
			return fmt.Errorf("error reading config file: %w", err)
		}
		return nil
	}

	cl.v.SetConfigFile(path)
	if err := cl.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil
		}
		return fmt.Errorf("error reading config file: %w", err)
	}

	return nil
}

func stripJSONComments(data []byte) []byte {
	result := make([]byte, 0, len(data))
	inString := false
	escaped := false

	for i := 0; i < len(data); i++ {
		ch := data[i]

		if inString {
			result = append(result, ch)
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}

		if ch == '"' {
			inString = true
			result = append(result, ch)
			continue
		}

		if ch == '/' && i+1 < len(data) {
			next := data[i+1]

			if next == '/' {
				i += 2
				for i < len(data) && data[i] != '\n' {
					i++
				}
				if i < len(data) && data[i] == '\n' {
					result = append(result, '\n')
				}
				continue
			}

			if next == '*' {
				i += 2
				for i+1 < len(data) {
					if data[i] == '\n' {
						result = append(result, '\n')
					}
					if data[i] == '*' && data[i+1] == '/' {
						i++
						break
					}
					i++
				}
				continue
			}
		}

		result = append(result, ch)
	}

	return result
}

func (cl *ConfigLoader) setDefaults() {
	cl.v.SetDefault("ninja.base_url", "https://invoiceninja.fergify.work")
	cl.v.SetDefault("agent.provider", "offline")
	cl.v.SetDefault("agent.plan", "default")
	cl.v.SetDefault("agent.model", DefaultAgentModel)
	cl.v.SetDefault("serve.listen", "127.0.0.1")
	cl.v.SetDefault("serve.port", 8080)
}

func (cl *ConfigLoader) bindEnvVars() {
	cl.v.SetEnvPrefix("ninjops")
	cl.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	cl.v.AutomaticEnv()

	envBindings := []struct {
		key string
		env string
	}{
		{"ninja.base_url", "NINJOPS_NINJA_BASE_URL"},
		{"ninja.api_token", "NINJOPS_NINJA_API_TOKEN"},
		{"ninja.api_secret", "NINJOPS_NINJA_API_SECRET"},
		{"agent.provider", "NINJOPS_AGENT_PROVIDER"},
		{"agent.plan", "NINJOPS_AGENT_PLAN"},
		{"agent.model", "NINJOPS_AGENT_MODEL"},
		{"agent.provider_api_key", "NINJOPS_AGENT_PROVIDER_API_KEY"},
		{"serve.listen", "NINJOPS_SERVE_LISTEN"},
		{"serve.port", "NINJOPS_SERVE_PORT"},
	}

	for _, binding := range envBindings {
		_ = cl.v.BindEnv(binding.key, binding.env)
	}
}

func (cl *ConfigLoader) setupConfigSearch(configPath string) {
	if configPath != "" {
		cl.v.SetConfigFile(configPath)
		return
	}

	cl.v.SetConfigName("config")
	cl.v.SetConfigType("toml")

	wd, _ := os.Getwd()
	cl.v.AddConfigPath(wd)
	cl.v.AddConfigPath(".")
	cl.v.AddConfigPath(filepath.Join(wd, ".ninjops"))
	cl.v.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".config", "ninjops"))
	cl.v.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".ninjops"))

	extensions := []string{".toml", ".yaml", ".yml", ".json"}
	for _, ext := range extensions {
		if _, err := os.Stat(filepath.Join(wd, "ninjops"+ext)); err == nil {
			cl.v.SetConfigName("ninjops")
			break
		}
	}
}

func (cl *ConfigLoader) validate(cfg *Config) error {
	if !IsValidProvider(cfg.Agent.Provider) {
		return fmt.Errorf("invalid agent provider: %s (valid: %s)", cfg.Agent.Provider, strings.Join(ValidProviders, ", "))
	}

	validPlans := map[string]bool{
		"default":      true,
		"codex-pro":    true,
		"opencode-zen": true,
		"zai-plan":     true,
	}

	if !validPlans[cfg.Agent.Plan] {
		return fmt.Errorf("invalid agent plan: %s (valid: default, codex-pro, opencode-zen, zai-plan)", cfg.Agent.Plan)
	}

	if cfg.Serve.Port < 1 || cfg.Serve.Port > 65535 {
		return fmt.Errorf("invalid serve port: %d (must be 1-65535)", cfg.Serve.Port)
	}

	return nil
}

func (cl *ConfigLoader) SetFlag(key string, value interface{}) {
	cl.v.Set(key, value)
}

func Load(configPath string) (*Config, error) {
	return NewConfigLoader().Load(configPath)
}

func ResolveProviderAPIKey(provider string, configuredKey string) string {
	for _, envVar := range ProviderAPIKeyEnvVars(provider) {
		if v := os.Getenv(envVar); v != "" {
			return v
		}
	}

	return configuredKey
}

func GetAPIKey(provider string) string {
	return ResolveProviderAPIKey(provider, "")
}

func RedactToken(token string) string {
	if len(token) <= 4 {
		return "****"
	}
	return token[:4] + "****"
}

func (c *Config) String() string {
	return fmt.Sprintf(
		"Config{Ninja{BaseURL:%s, APIToken:%s}, Agent{Provider:%s, Plan:%s, Model:%s}, Serve{Listen:%s, Port:%d}}",
		c.Ninja.BaseURL,
		RedactToken(c.Ninja.APIToken),
		c.Agent.Provider,
		c.Agent.Plan,
		c.Agent.Model,
		c.Serve.Listen,
		c.Serve.Port,
	)
}
