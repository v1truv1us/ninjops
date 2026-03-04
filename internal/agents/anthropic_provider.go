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

	"github.com/ninjops/ninjops/internal/spec"
)

const (
	AnthropicAPIURL     = "https://api.anthropic.com/v1/messages"
	AnthropicModel      = "claude-3-5-sonnet-20241022"
	AnthropicAPIVersion = "2023-06-01"
)

type AnthropicProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	return &AnthropicProvider{
		apiKey: apiKey,
		model:  AnthropicModel,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

func (p *AnthropicProvider) IsAvailable() bool {
	return p.apiKey != ""
}

func (p *AnthropicProvider) Execute(ctx context.Context, req AgentRequest) (*AgentResponse, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("Anthropic API key not configured. Set NINJOPS_ANTHROPIC_API_KEY environment variable")
	}

	builder := NewPromptBuilder(req.Plan, req.Role)
	systemPrompt := builder.BuildSystemPrompt()
	userPrompt := builder.BuildUserPrompt(req.QuoteSpec)
	modelName := strings.TrimSpace(req.Model)
	if modelName == "" {
		modelName = p.model
	}

	apiReq := AnthropicMessagesRequest{
		Model:     modelName,
		MaxTokens: 4096,
		System:    systemPrompt,
		Messages: []AnthropicMessage{
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
	}

	resp, err := p.makeRequest(ctx, apiReq)
	if err != nil {
		return nil, fmt.Errorf("Anthropic API error: %w", err)
	}

	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("no response from Anthropic")
	}

	content := resp.Content[0].Text
	return p.parseResponse(content, req.QuoteSpec, req.Role)
}

func (p *AnthropicProvider) makeRequest(ctx context.Context, req AnthropicMessagesRequest) (*AnthropicMessagesResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", AnthropicAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", AnthropicAPIVersion)

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
		var apiErr AnthropicAPIError
		if err := json.Unmarshal(respBody, &apiErr); err == nil {
			return nil, fmt.Errorf("API error: %s", apiErr.Error.Message)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result AnthropicMessagesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (p *AnthropicProvider) parseResponse(content string, original *spec.QuoteSpec, role Role) (*AgentResponse, error) {
	content = extractJSON(content)

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

func extractJSON(content string) string {
	start := -1
	end := -1

	for i := 0; i < len(content); i++ {
		if content[i] == '{' && start == -1 {
			start = i
		}
		if content[i] == '}' {
			end = i + 1
		}
	}

	if start != -1 && end > start {
		return content[start:end]
	}

	return content
}

type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicMessagesRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []AnthropicMessage `json:"messages"`
}

type AnthropicMessagesResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type AnthropicAPIError struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}
