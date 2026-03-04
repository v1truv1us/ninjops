package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ninjops/ninjops/internal/config"
)

func TestCheckProviderConnection_OfflineAlwaysPasses(t *testing.T) {
	err := CheckProviderConnection(context.Background(), "offline", "")
	if err != nil {
		t.Fatalf("expected offline connectivity check to pass, got error: %v", err)
	}
}

func TestCheckProviderConnection_UsesRegisteredChecker(t *testing.T) {
	const providerName = "unit-test-provider"
	called := false

	RegisterProviderConnectionChecker(providerName, func(_ context.Context, apiKey string) error {
		called = true
		if apiKey != "unit-test-key" {
			return fmt.Errorf("unexpected API key: %s", apiKey)
		}
		return nil
	})

	err := CheckProviderConnection(context.Background(), providerName, "unit-test-key")
	if err != nil {
		t.Fatalf("expected registered checker to pass, got error: %v", err)
	}

	if !called {
		t.Fatal("expected registered checker to be called")
	}
}

func TestCheckOpenAICompatibleConnectionAtBaseURL_SuccessCallsModelsAndChatProbe(t *testing.T) {
	t.Helper()

	modelsCalled := false
	chatCalled := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			modelsCalled = true
			if got := r.Header.Get("Authorization"); got != "Bearer sk-test" {
				t.Fatalf("unexpected auth header on /models: %q", got)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"object":"list","data":[]}`))
		case "/v1/chat/completions":
			chatCalled = true
			if got := r.Header.Get("Authorization"); got != "Bearer sk-test" {
				t.Fatalf("unexpected auth header on /chat/completions: %q", got)
			}

			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("failed to decode chat probe payload: %v", err)
			}
			if got := payload["model"]; got != "deepseek-chat" {
				t.Fatalf("unexpected chat probe model: %v", got)
			}
			if got := payload["max_tokens"]; got != float64(1) {
				t.Fatalf("unexpected chat probe max_tokens: %v", got)
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"chatcmpl-test"}`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	err := checkOpenAICompatibleConnectionAtBaseURL(context.Background(), "deepseek", "deepseek-chat", "sk-test", ts.URL+"/v1")
	if err != nil {
		t.Fatalf("expected openai-compatible connectivity check to pass, got error: %v", err)
	}
	if !modelsCalled {
		t.Fatal("expected /models auth probe to be called")
	}
	if !chatCalled {
		t.Fatal("expected /chat/completions probe to be called")
	}
}

func TestCheckOpenAICompatibleConnectionAtBaseURL_ModelsPassButChatAuthFails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"object":"list","data":[]}`))
		case "/v1/chat/completions":
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid key"}`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	err := checkOpenAICompatibleConnectionAtBaseURL(context.Background(), "deepseek", "deepseek-chat", "sk-test", ts.URL+"/v1")
	if err == nil {
		t.Fatal("expected openai-compatible chat auth failure to return error")
	}

	if got := err.Error(); !strings.Contains(got, "deepseek authentication failed during chat probe") {
		t.Fatalf("unexpected error %q", got)
	}
}

func TestCheckOpenAICompatibleConnectionAtBaseURL_EmptyModelUsesProviderDefault(t *testing.T) {
	chatCalled := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"object":"list","data":[]}`))
		case "/v1/chat/completions":
			chatCalled = true

			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("failed to decode chat probe payload: %v", err)
			}
			if got := payload["model"]; got != config.DefaultModelForProvider("openai") {
				t.Fatalf("unexpected default model in chat probe: %v", got)
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"chatcmpl-test"}`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	err := checkOpenAICompatibleConnectionAtBaseURL(context.Background(), "openai", "", "sk-test", ts.URL+"/v1")
	if err != nil {
		t.Fatalf("expected default-model probe to pass, got error: %v", err)
	}
	if !chatCalled {
		t.Fatal("expected /chat/completions probe to be called")
	}
}

func TestCheckAnthropicConnectionAtBaseURL_SuccessCallsModelsAndMessagesProbe(t *testing.T) {
	modelsCalled := false
	messagesCalled := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			modelsCalled = true
			if got := r.Header.Get("x-api-key"); got != "sk-ant-test" {
				t.Fatalf("unexpected x-api-key header on /models: %q", got)
			}
			if got := r.Header.Get("anthropic-version"); got != anthropicVersionHeader {
				t.Fatalf("unexpected anthropic-version header on /models: %q", got)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/v1/messages":
			messagesCalled = true
			if got := r.Header.Get("x-api-key"); got != "sk-ant-test" {
				t.Fatalf("unexpected x-api-key header on /messages: %q", got)
			}
			if got := r.Header.Get("anthropic-version"); got != anthropicVersionHeader {
				t.Fatalf("unexpected anthropic-version header on /messages: %q", got)
			}

			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("failed to decode messages probe payload: %v", err)
			}
			if got := payload["model"]; got != "claude-3-5-sonnet-20241022" {
				t.Fatalf("unexpected messages probe model: %v", got)
			}
			if got := payload["max_tokens"]; got != float64(1) {
				t.Fatalf("unexpected messages probe max_tokens: %v", got)
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"msg_123"}`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	err := checkAnthropicConnectionAtBaseURL(context.Background(), "claude-3-5-sonnet-20241022", "sk-ant-test", ts.URL+"/v1")
	if err != nil {
		t.Fatalf("expected anthropic connectivity check to pass, got error: %v", err)
	}
	if !modelsCalled {
		t.Fatal("expected /models auth probe to be called")
	}
	if !messagesCalled {
		t.Fatal("expected /messages probe to be called")
	}
}

func TestCheckAnthropicConnectionAtBaseURL_ModelsPassButMessagesProbeFailsModelAccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/v1/messages":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"model not found"}`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	err := checkAnthropicConnectionAtBaseURL(context.Background(), "claude-not-allowed", "sk-ant-test", ts.URL+"/v1")
	if err == nil {
		t.Fatal("expected anthropic messages model access failure to return error")
	}

	if got := err.Error(); !strings.Contains(got, `anthropic model probe failed for "claude-not-allowed"`) {
		t.Fatalf("unexpected error %q", got)
	}
}
