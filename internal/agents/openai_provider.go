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
	"github.com/ninjops/ninjops/internal/spec"
)

const (
	OpenAIBaseURL = "https://api.openai.com/v1"
	DefaultModel  = "gpt-5-codex"
)

type OpenAIProvider struct {
	name       string
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return NewOpenAICompatibleProvider("openai", apiKey, OpenAIBaseURL, DefaultModel)
}

func NewOpenAICompatibleProvider(name string, apiKey string, baseURL string, defaultModel string) *OpenAIProvider {
	if strings.TrimSpace(name) == "" {
		name = "openai-compatible"
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = OpenAIBaseURL
	}
	if strings.TrimSpace(defaultModel) == "" {
		defaultModel = DefaultModel
	}

	return &OpenAIProvider{
		name:    name,
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   defaultModel,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *OpenAIProvider) Name() string {
	return p.name
}

func (p *OpenAIProvider) IsAvailable() bool {
	return p.apiKey != ""
}

func (p *OpenAIProvider) Execute(ctx context.Context, req AgentRequest) (*AgentResponse, error) {
	if !p.IsAvailable() {
		envHint := config.ProviderAPIKeyEnvHint(p.name)
		if envHint == "" {
			envHint = "NINJOPS_AGENT_PROVIDER_API_KEY"
		}
		return nil, fmt.Errorf("%s API key not configured. Set %s environment variable", p.name, envHint)
	}

	builder := NewPromptBuilder(req.Plan, req.Role)
	systemPrompt := builder.BuildSystemPrompt()
	userPrompt := builder.BuildUserPrompt(req.QuoteSpec)

	messages := []OpenAIMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	modelName := strings.TrimSpace(req.Model)
	if modelName == "" {
		modelName = p.model
	}

	apiReq := OpenAIChatCompletionRequest{
		Model:       modelName,
		Messages:    messages,
		Temperature: 0.7,
		ResponseFormat: &OpenAIResponseFormat{
			Type: "json_object",
		},
	}

	resp, err := p.makeRequest(ctx, apiReq)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	content := resp.Choices[0].Message.Content
	return p.parseResponse(content, req.QuoteSpec, req.Role)
}

func (p *OpenAIProvider) makeRequest(ctx context.Context, req OpenAIChatCompletionRequest) (*OpenAIChatCompletionResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	baseURL, err := config.ResolveProviderBaseURL(p.name)
	if err != nil {
		baseURL = p.baseURL
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", openAICompatibleChatCompletionsURL(baseURL), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr OpenAIAPIError
		if err := json.Unmarshal(respBody, &apiErr); err == nil {
			return nil, fmt.Errorf("API error: %s", apiErr.Error.Message)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result OpenAIChatCompletionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func openAICompatibleChatCompletionsURL(baseURL string) string {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		trimmed = OpenAIBaseURL
	}

	trimmed = strings.TrimSuffix(trimmed, "/")
	if strings.HasSuffix(trimmed, "/chat/completions") {
		return trimmed
	}

	return trimmed + "/chat/completions"
}

func (p *OpenAIProvider) parseResponse(content string, original *spec.QuoteSpec, role Role) (*AgentResponse, error) {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	result := *original

	switch role {
	case RoleClarify:
		if features, ok := parsed["features"].([]interface{}); ok {
			result.Work.Features = parseFeatures(features)
		}
		if responsibilities, ok := parsed["responsibilities"].([]interface{}); ok {
			result.Work.Responsibilities = parseStringSlice(responsibilities)
		}

	case RolePolish:
		if polished, ok := parsed["polished_sections"].(map[string]interface{}); ok {
			if desc, ok := polished["description"].(string); ok {
				result.Project.Description = desc
			}
		}

	case RoleBoundaries:
		if minor, ok := parsed["minor_changes"].([]interface{}); ok {
			result.Work.MinorChanges = parseStringSlice(minor)
		}
		if out, ok := parsed["out_of_scope"].([]interface{}); ok {
			result.Work.OutOfScope = parseStringSlice(out)
		}
		if assumptions, ok := parsed["assumptions"].([]interface{}); ok {
			result.Work.Assumptions = parseStringSlice(assumptions)
		}

	case RoleLineItems:
		if items, ok := parsed["line_items"].([]interface{}); ok {
			result.Pricing.LineItems = parseLineItems(items)
		}
	}

	confidence := 0.85
	if c, ok := parsed["confidence"].(float64); ok {
		confidence = c
	}

	return &AgentResponse{
		QuoteSpec:   &result,
		Suggestions: extractSuggestions(parsed),
		Confidence:  confidence,
		Metadata:    parsed,
	}, nil
}

func parseFeatures(features []interface{}) []spec.Feature {
	result := make([]spec.Feature, 0, len(features))
	for _, f := range features {
		if fm, ok := f.(map[string]interface{}); ok {
			feature := spec.Feature{}
			if name, ok := fm["name"].(string); ok {
				feature.Name = name
			}
			if desc, ok := fm["description"].(string); ok {
				feature.Description = desc
			}
			if priority, ok := fm["priority"].(string); ok {
				feature.Priority = priority
			}
			result = append(result, feature)
		}
	}
	return result
}

func parseStringSlice(items []interface{}) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func parseLineItems(items []interface{}) []spec.LineItem {
	result := make([]spec.LineItem, 0, len(items))
	for _, item := range items {
		if im, ok := item.(map[string]interface{}); ok {
			li := spec.LineItem{}
			if desc, ok := im["description"].(string); ok {
				li.Description = desc
			}
			if qty, ok := im["quantity"].(float64); ok {
				li.Quantity = qty
			}
			if cat, ok := im["category"].(string); ok {
				li.Category = cat
			}
			result = append(result, li)
		}
	}
	return result
}

func extractSuggestions(parsed map[string]interface{}) []string {
	suggestions := make([]string, 0)

	if s, ok := parsed["suggestions"].([]interface{}); ok {
		for _, item := range s {
			if str, ok := item.(string); ok {
				suggestions = append(suggestions, str)
			}
		}
	}

	if improvements, ok := parsed["improvements"].([]interface{}); ok {
		for _, item := range improvements {
			if str, ok := item.(string); ok {
				suggestions = append(suggestions, str)
			}
		}
	}

	if notes, ok := parsed["notes"].([]interface{}); ok {
		for _, item := range notes {
			if str, ok := item.(string); ok {
				suggestions = append(suggestions, str)
			}
		}
	}

	return suggestions
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponseFormat struct {
	Type string `json:"type"`
}

type OpenAIChatCompletionRequest struct {
	Model          string                `json:"model"`
	Messages       []OpenAIMessage       `json:"messages"`
	Temperature    float64               `json:"temperature"`
	ResponseFormat *OpenAIResponseFormat `json:"response_format,omitempty"`
}

type OpenAIChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type OpenAIAPIError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}
