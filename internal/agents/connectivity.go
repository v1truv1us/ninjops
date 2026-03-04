package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ninjops/ninjops/internal/config"
)

const (
	defaultConnectionTimeout = 10 * time.Second
	modelProbeTimeout        = 5 * time.Second
	openAIBaseURL            = "https://api.openai.com/v1"
	anthropicBaseURL         = "https://api.anthropic.com/v1"
	anthropicVersionHeader   = "2023-06-01"
)

type ProviderConnectionChecker func(ctx context.Context, apiKey string) error

var providerConnectionCheckers = map[string]ProviderConnectionChecker{
	"offline": checkOfflineConnection,
}

func RegisterProviderConnectionChecker(provider string, checker ProviderConnectionChecker) {
	providerConnectionCheckers[provider] = checker
}

func CheckProviderConnection(ctx context.Context, provider string, apiKey string) error {
	return CheckProviderConnectionWithModel(ctx, provider, "", apiKey)
}

func CheckProviderConnectionWithModel(ctx context.Context, provider string, model string, apiKey string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, defaultConnectionTimeout)
	defer cancel()

	trimmedAPIKey := strings.TrimSpace(apiKey)
	trimmedModel := strings.TrimSpace(model)

	switch provider {
	case "offline":
		return checkOfflineConnection(timeoutCtx, trimmedAPIKey)
	case "openai":
		return checkOpenAIConnection(timeoutCtx, trimmedModel, trimmedAPIKey)
	case "anthropic":
		return checkAnthropicConnection(timeoutCtx, trimmedModel, trimmedAPIKey)
	}

	if config.IsOpenAICompatibleProvider(provider) {
		return checkOpenAICompatibleConnection(timeoutCtx, provider, trimmedModel, trimmedAPIKey)
	}

	checker, ok := providerConnectionCheckers[provider]
	if !ok {
		return fmt.Errorf("provider %q connection check is not implemented", provider)
	}

	return checker(timeoutCtx, trimmedAPIKey)
}

func checkOfflineConnection(_ context.Context, _ string) error {
	return nil
}

func checkOpenAIConnection(ctx context.Context, model string, apiKey string) error {
	return checkOpenAICompatibleConnectionAtBaseURL(ctx, "openai", model, apiKey, openAIBaseURL)
}

func checkOpenAICompatibleConnection(ctx context.Context, provider string, model string, apiKey string) error {
	baseURL, err := config.ResolveProviderBaseURL(provider)
	if err != nil {
		return fmt.Errorf("%s connection check setup failed: %w", provider, err)
	}
	return checkOpenAICompatibleConnectionAtBaseURL(ctx, provider, model, apiKey, baseURL)
}

func checkOpenAICompatibleConnectionAtBaseURL(ctx context.Context, provider string, model string, apiKey string, baseURL string) error {
	if apiKey == "" {
		return fmt.Errorf("%s API key is required", provider)
	}

	modelsURL := openAICompatibleModelsURL(baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelsURL, nil)
	if err != nil {
		return fmt.Errorf("%s connection check setup failed: %w", provider, err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := (&http.Client{Timeout: defaultConnectionTimeout}).Do(req)
	if err != nil {
		return fmt.Errorf("%s connection check failed: %w", provider, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		probeModel := strings.TrimSpace(model)
		if probeModel == "" {
			probeModel = config.DefaultModelForProvider(provider)
		}
		if err := probeOpenAICompatibleChatCompletion(ctx, provider, probeModel, apiKey, baseURL); err != nil {
			return err
		}
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("%s authentication failed (check API key)", provider)
	}

	if len(body) > 0 {
		return fmt.Errorf("%s connection check failed (HTTP %d): %s", provider, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return fmt.Errorf("%s connection check failed (HTTP %d)", provider, resp.StatusCode)
}

func probeOpenAICompatibleChatCompletion(ctx context.Context, provider string, model string, apiKey string, baseURL string) error {
	probeCtx, cancel := context.WithTimeout(ctx, modelProbeTimeout)
	defer cancel()

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": "ping"},
		},
		"max_tokens":  1,
		"temperature": 0,
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("%s chat probe setup failed: %w", provider, err)
	}

	chatURL := openAICompatibleChatCompletionsURL(baseURL)
	req, err := http.NewRequestWithContext(probeCtx, http.MethodPost, chatURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("%s chat probe setup failed: %w", provider, err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: defaultConnectionTimeout}).Do(req)
	if err != nil {
		return fmt.Errorf("%s chat probe failed: %w", provider, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("%s authentication failed during chat probe (check API key/model access)", provider)
	}
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound {
		if len(body) > 0 {
			return fmt.Errorf("%s model probe failed for %q (check model id/access): %s", provider, model, strings.TrimSpace(string(body)))
		}
		return fmt.Errorf("%s model probe failed for %q (check model id/access)", provider, model)
	}

	if len(body) > 0 {
		return fmt.Errorf("%s chat probe failed (HTTP %d): %s", provider, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return fmt.Errorf("%s chat probe failed (HTTP %d)", provider, resp.StatusCode)
}

func openAICompatibleModelsURL(baseURL string) string {
	return normalizeOpenAICompatibleBaseURL(baseURL) + "/models"
}

func normalizeOpenAICompatibleBaseURL(baseURL string) string {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		trimmed = openAIBaseURL
	}

	trimmed = strings.TrimSuffix(trimmed, "/")
	if strings.HasSuffix(trimmed, "/chat/completions") {
		trimmed = strings.TrimSuffix(trimmed, "/chat/completions")
	}

	return trimmed
}

func checkAnthropicConnection(ctx context.Context, model string, apiKey string) error {
	return checkAnthropicConnectionAtBaseURL(ctx, model, apiKey, anthropicBaseURL)
}

func checkAnthropicConnectionAtBaseURL(ctx context.Context, model string, apiKey string, baseURL string) error {
	if apiKey == "" {
		return fmt.Errorf("anthropic API key is required")
	}

	modelsURL := strings.TrimSuffix(strings.TrimSpace(baseURL), "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelsURL, nil)
	if err != nil {
		return fmt.Errorf("anthropic connection check setup failed: %w", err)
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", anthropicVersionHeader)

	resp, err := (&http.Client{Timeout: defaultConnectionTimeout}).Do(req)
	if err != nil {
		return fmt.Errorf("anthropic connection check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		probeModel := strings.TrimSpace(model)
		if probeModel == "" {
			probeModel = config.DefaultModelForProvider("anthropic")
		}

		if err := probeAnthropicMessages(ctx, probeModel, apiKey, baseURL); err != nil {
			return err
		}
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("anthropic authentication failed (check API key)")
	}

	if len(body) > 0 {
		return fmt.Errorf("anthropic connection check failed (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return fmt.Errorf("anthropic connection check failed (HTTP %d)", resp.StatusCode)
}

func probeAnthropicMessages(ctx context.Context, model string, apiKey string, baseURL string) error {
	probeCtx, cancel := context.WithTimeout(ctx, modelProbeTimeout)
	defer cancel()

	requestBody := map[string]interface{}{
		"model":      model,
		"max_tokens": 1,
		"messages": []map[string]string{
			{"role": "user", "content": "ping"},
		},
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("anthropic message probe setup failed: %w", err)
	}

	messagesURL := strings.TrimSuffix(strings.TrimSpace(baseURL), "/") + "/messages"
	req, err := http.NewRequestWithContext(probeCtx, http.MethodPost, messagesURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("anthropic message probe setup failed: %w", err)
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", anthropicVersionHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: defaultConnectionTimeout}).Do(req)
	if err != nil {
		return fmt.Errorf("anthropic message probe failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("anthropic authentication failed during message probe (check API key/model access)")
	}
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound {
		if len(body) > 0 {
			return fmt.Errorf("anthropic model probe failed for %q (check model id/access): %s", model, strings.TrimSpace(string(body)))
		}
		return fmt.Errorf("anthropic model probe failed for %q (check model id/access)", model)
	}

	if len(body) > 0 {
		return fmt.Errorf("anthropic message probe failed (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return fmt.Errorf("anthropic message probe failed (HTTP %d)", resp.StatusCode)
}
