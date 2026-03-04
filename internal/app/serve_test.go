package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ninjops/ninjops/internal/agents"
	"github.com/ninjops/ninjops/internal/config"
	"github.com/ninjops/ninjops/internal/spec"
)

func TestHandleAssist_RejectsMissingQuoteSpec(t *testing.T) {
	prevCfg := cfg
	cfg = config.DefaultConfig()
	t.Cleanup(func() {
		cfg = prevCfg
	})

	req := httptest.NewRequest(http.MethodPost, "/assist/clarify", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	handleAssist(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleNinjaSync_RejectsMissingQuoteSpec(t *testing.T) {
	prevCfg := cfg
	localCfg := config.DefaultConfig()
	localCfg.Ninja.APIToken = "token"
	cfg = localCfg
	t.Cleanup(func() {
		cfg = prevCfg
	})

	req := httptest.NewRequest(http.MethodPost, "/ninja/sync", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	handleNinjaSync(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

type captureProvider struct {
	lastModel string
}

func (p *captureProvider) Name() string {
	return "capture"
}

func (p *captureProvider) IsAvailable() bool {
	return true
}

func (p *captureProvider) Execute(_ context.Context, req agents.AgentRequest) (*agents.AgentResponse, error) {
	p.lastModel = req.Model
	quote := req.QuoteSpec
	if quote == nil {
		quote = spec.NewQuoteSpec()
	}

	return &agents.AgentResponse{
		QuoteSpec:   quote,
		Suggestions: []string{"ok"},
		Confidence:  1,
		Metadata:    map[string]interface{}{},
	}, nil
}

func TestHandleAssist_UsesConfiguredModel(t *testing.T) {
	openAIFake := &captureProvider{}
	opencodeFake := &captureProvider{}
	agents.RegisterProvider("openai", func(_ string) agents.Provider {
		return openAIFake
	})
	agents.RegisterProvider("opencode", func(_ string) agents.Provider {
		return opencodeFake
	})
	t.Cleanup(func() {
		agents.RegisterProvider("openai", func(apiKey string) agents.Provider {
			return agents.NewOpenAIProvider(apiKey)
		})
		agents.RegisterProvider("opencode", func(apiKey string) agents.Provider {
			baseURL, err := config.ResolveProviderBaseURL("opencode")
			if err != nil {
				baseURL = config.ProviderAPIBaseURL("opencode")
			}
			return agents.NewOpenAICompatibleProvider("opencode", apiKey, baseURL, config.DefaultModelForProvider("opencode"))
		})
	})

	prevCfg := cfg
	localCfg := config.DefaultConfig()
	localCfg.Agent.Provider = "openai"
	localCfg.Agent.Model = "openai-codex"
	localCfg.Agent.ProviderAPIKey = "sk-test"
	cfg = localCfg
	t.Cleanup(func() {
		cfg = prevCfg
	})

	payload, err := json.Marshal(map[string]interface{}{
		"quote_spec": map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("failed to marshal request payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/assist/clarify", strings.NewReader(string(payload)))
	w := httptest.NewRecorder()

	handleAssist(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if openAIFake.lastModel != "" {
		t.Fatalf("expected openai provider to be bypassed, got model %q", openAIFake.lastModel)
	}
	if opencodeFake.lastModel != "gpt-5.3-codex" {
		t.Fatalf("expected model %q, got %q", "gpt-5.3-codex", opencodeFake.lastModel)
	}
}
