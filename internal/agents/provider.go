package agents

import (
	"context"

	"github.com/ninjops/ninjops/internal/config"
	"github.com/ninjops/ninjops/internal/spec"
)

type Role string

const (
	RoleClarify    Role = "clarify"
	RolePolish     Role = "polish"
	RoleBoundaries Role = "boundaries"
	RoleLineItems  Role = "line-items"
)

type Plan string

const (
	PlanDefault     Plan = "default"
	PlanCodexPro    Plan = "codex-pro"
	PlanOpencodeZen Plan = "opencode-zen"
	PlanZAIPlan     Plan = "zai-plan"
)

type AgentRequest struct {
	Role      Role
	Plan      Plan
	Model     string
	QuoteSpec *spec.QuoteSpec
}

type AgentResponse struct {
	QuoteSpec   *spec.QuoteSpec
	Suggestions []string
	Confidence  float64
	Metadata    map[string]interface{}
}

type Provider interface {
	Name() string
	Execute(ctx context.Context, req AgentRequest) (*AgentResponse, error)
	IsAvailable() bool
}

type ProviderFactory func(apiKey string) Provider

var providers = map[string]ProviderFactory{}

func init() {
	providers["offline"] = func(_ string) Provider { return NewOfflineProvider() }
	providers["openai"] = func(apiKey string) Provider { return NewOpenAIProvider(apiKey) }
	providers["anthropic"] = func(apiKey string) Provider { return NewAnthropicProvider(apiKey) }

	for _, providerID := range config.OpenAICompatibleProviderIDs {
		baseURL, ok := config.OpenAICompatibleBaseURL(providerID)
		if !ok {
			continue
		}

		pid := providerID
		providerBaseURL := baseURL
		defaultModel := config.DefaultModelForProvider(providerID)
		providers[pid] = func(apiKey string) Provider {
			return NewOpenAICompatibleProvider(pid, apiKey, providerBaseURL, defaultModel)
		}
	}
}

func GetProvider(name string, apiKey string) Provider {
	factory, exists := providers[name]
	if !exists {
		return NewOfflineProvider()
	}
	return factory(apiKey)
}

func RegisterProvider(name string, factory ProviderFactory) {
	providers[name] = factory
}

func ValidRoles() []Role {
	return []Role{RoleClarify, RolePolish, RoleBoundaries, RoleLineItems}
}

func ValidPlans() []Plan {
	return []Plan{PlanDefault, PlanCodexPro, PlanOpencodeZen, PlanZAIPlan}
}

func IsValidRole(r string) bool {
	for _, role := range ValidRoles() {
		if string(role) == r {
			return true
		}
	}
	return false
}

func IsValidPlan(p string) bool {
	for _, plan := range ValidPlans() {
		if string(plan) == p {
			return true
		}
	}
	return false
}
