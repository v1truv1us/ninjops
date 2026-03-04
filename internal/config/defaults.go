package config

const (
	DefaultNinjaBaseURL  = "https://invoiceninja.fergify.work"
	DefaultAgentProvider = "offline"
	DefaultAgentPlan     = "default"
	DefaultAgentModel    = SafeFallbackModel
	DefaultServeListen   = "127.0.0.1"
	DefaultServePort     = 8080
)

var (
	ValidProviders = append([]string{"offline", "openai", "anthropic"}, OpenAICompatibleProviderIDs...)
	ValidPlans     = []string{"default", "codex-pro", "opencode-zen", "zai-plan"}
)

func DefaultConfig() *Config {
	return &Config{
		Ninja: NinjaConfig{
			BaseURL: DefaultNinjaBaseURL,
		},
		Agent: AgentConfig{
			Provider:       DefaultAgentProvider,
			Plan:           DefaultAgentPlan,
			Model:          DefaultAgentModel,
			ProviderAPIKey: "",
		},
		Serve: ServeConfig{
			Listen: DefaultServeListen,
			Port:   DefaultServePort,
		},
	}
}
