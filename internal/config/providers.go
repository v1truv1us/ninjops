package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	ModelAliasOpenAICodex    = "openai-codex"
	ModelResolvedOpenAICodex = "gpt-5-codex"
	SafeFallbackModel        = "gpt-5-codex"
)

var OpenAICompatibleProviderIDs = []string{
	"302ai",
	"abacus",
	"aihubmix",
	"alibaba",
	"alibaba-cn",
	"bailing",
	"baseten",
	"berget",
	"chutes",
	"cloudferro-sherlock",
	"cloudflare-workers-ai",
	"cortecs",
	"deepseek",
	"evroc",
	"fastrouter",
	"fireworks-ai",
	"firmware",
	"friendli",
	"github-copilot",
	"github-models",
	"helicone",
	"huggingface",
	"iflowcn",
	"inception",
	"inference",
	"io-net",
	"jiekou",
	"kilo",
	"kuae-cloud-coding-plan",
	"llama",
	"lmstudio",
	"lucidquery",
	"meganova",
	"moark",
	"modelscope",
	"moonshotai",
	"moonshotai-cn",
	"morph",
	"nano-gpt",
	"nebius",
	"nova",
	"novita-ai",
	"nvidia",
	"ollama-cloud",
	"opencode",
	"opencode-go",
	"ovhcloud",
	"poe",
	"privatemode-ai",
	"qihang-ai",
	"qiniu-ai",
	"requesty",
	"scaleway",
	"siliconflow",
	"siliconflow-cn",
	"stackit",
	"stepfun",
	"submodel",
	"synthetic",
	"upstage",
	"vultr",
	"wandb",
	"xiaomi",
	"zai",
	"zai-coding-plan",
	"zhipuai",
	"zhipuai-coding-plan",
}

var openAICompatibleProviderBaseURLs = map[string]string{
	"302ai":                  "https://api.302.ai/v1",
	"abacus":                 "https://routellm.abacus.ai/v1",
	"aihubmix":               "https://aihubmix.com/v1",
	"alibaba":                "https://dashscope-intl.aliyuncs.com/compatible-mode/v1",
	"alibaba-cn":             "https://dashscope.aliyuncs.com/compatible-mode/v1",
	"bailing":                "https://api.tbox.cn/api/llm/v1/chat/completions",
	"baseten":                "https://inference.baseten.co/v1",
	"berget":                 "https://api.berget.ai/v1",
	"chutes":                 "https://llm.chutes.ai/v1",
	"cloudferro-sherlock":    "https://api-sherlock.cloudferro.com/openai/v1/",
	"cloudflare-workers-ai":  "https://api.cloudflare.com/client/v4/accounts/${CLOUDFLARE_ACCOUNT_ID}/ai/v1",
	"cortecs":                "https://api.cortecs.ai/v1",
	"deepseek":               "https://api.deepseek.com",
	"evroc":                  "https://models.think.evroc.com/v1",
	"fastrouter":             "https://go.fastrouter.ai/api/v1",
	"fireworks-ai":           "https://api.fireworks.ai/inference/v1/",
	"firmware":               "https://app.firmware.ai/api/v1",
	"friendli":               "https://api.friendli.ai/serverless/v1",
	"github-copilot":         "https://api.githubcopilot.com",
	"github-models":          "https://models.github.ai/inference",
	"helicone":               "https://ai-gateway.helicone.ai/v1",
	"huggingface":            "https://router.huggingface.co/v1",
	"iflowcn":                "https://apis.iflow.cn/v1",
	"inception":              "https://api.inceptionlabs.ai/v1/",
	"inference":              "https://inference.net/v1",
	"io-net":                 "https://api.intelligence.io.solutions/api/v1",
	"jiekou":                 "https://api.jiekou.ai/openai",
	"kilo":                   "https://api.kilo.ai/api/gateway",
	"kuae-cloud-coding-plan": "https://coding-plan-endpoint.kuaecloud.net/v1",
	"llama":                  "https://api.llama.com/compat/v1/",
	"lmstudio":               "http://127.0.0.1:1234/v1",
	"lucidquery":             "https://lucidquery.com/api/v1",
	"meganova":               "https://api.meganova.ai/v1",
	"moark":                  "https://moark.com/v1",
	"modelscope":             "https://api-inference.modelscope.cn/v1",
	"moonshotai":             "https://api.moonshot.ai/v1",
	"moonshotai-cn":          "https://api.moonshot.cn/v1",
	"morph":                  "https://api.morphllm.com/v1",
	"nano-gpt":               "https://nano-gpt.com/api/v1",
	"nebius":                 "https://api.tokenfactory.nebius.com/v1",
	"nova":                   "https://api.nova.amazon.com/v1",
	"novita-ai":              "https://api.novita.ai/openai",
	"nvidia":                 "https://integrate.api.nvidia.com/v1",
	"ollama-cloud":           "https://ollama.com/v1",
	"opencode":               "https://opencode.ai/zen/v1",
	"opencode-go":            "https://opencode.ai/zen/go/v1",
	"ovhcloud":               "https://oai.endpoints.kepler.ai.cloud.ovh.net/v1",
	"poe":                    "https://api.poe.com/v1",
	"privatemode-ai":         "http://localhost:8080/v1",
	"qihang-ai":              "https://api.qhaigc.net/v1",
	"qiniu-ai":               "https://api.qnaigc.com.com/v1",
	"requesty":               "https://router.requesty.ai/v1",
	"scaleway":               "https://api.scaleway.ai/v1",
	"siliconflow":            "https://api.siliconflow.com/v1",
	"siliconflow-cn":         "https://api.siliconflow.cn/v1",
	"stackit":                "https://api.openai-compat.model-serving.eu01.onstackit.cloud/v1",
	"stepfun":                "https://api.stepfun.com/v1",
	"submodel":               "https://llm.submodel.ai/v1",
	"synthetic":              "https://api.synthetic.new/v1",
	"upstage":                "https://api.upstage.ai/v1/solar",
	"vultr":                  "https://api.vultrinference.com/v1",
	"wandb":                  "https://api.inference.wandb.ai/v1",
	"xiaomi":                 "https://api.xiaomimimo.com/v1",
	"zai":                    "https://api.z.ai/api/paas/v4",
	"zai-coding-plan":        "https://api.z.ai/api/coding/paas/v4",
	"zhipuai":                "https://open.bigmodel.cn/api/paas/v4",
	"zhipuai-coding-plan":    "https://open.bigmodel.cn/api/coding/paas/v4",
}

var providerDefaultModels = map[string]string{
	"openai":                 "gpt-5-codex",
	"anthropic":              "claude-3-5-sonnet-20241022",
	"302ai":                  "qwen3-235b-a22b-instruct-2507",
	"abacus":                 "gemini-2.0-pro-exp-02-05",
	"aihubmix":               "qwen3-235b-a22b-instruct-2507",
	"alibaba":                "qwen-vl-plus",
	"alibaba-cn":             "qwen-vl-plus",
	"bailing":                "Ring-1T",
	"baseten":                "zai-org/GLM-4.6",
	"berget":                 "zai-org/GLM-4.7",
	"chutes":                 "zai-org/GLM-4.7-FP8",
	"cloudferro-sherlock":    "speakleash/Bielik-11B-v2.6-Instruct",
	"cloudflare-workers-ai":  "@cf/ibm-granite/granite-4.0-h-micro",
	"cortecs":                "kimi-k2-instruct",
	"deepseek":               "deepseek-reasoner",
	"evroc":                  "nvidia/Llama-3.3-70B-Instruct-FP8",
	"fastrouter":             "deepseek-ai/deepseek-r1-distill-llama-70b",
	"fireworks-ai":           "accounts/fireworks/models/kimi-k2-instruct",
	"firmware":               "claude-opus-4-6",
	"friendli":               "zai-org/GLM-4.7",
	"github-copilot":         "gpt-5.1-codex-max",
	"github-models":          "ai21-labs/ai21-jamba-1.5-mini",
	"helicone":               "claude-4.5-haiku",
	"huggingface":            "zai-org/GLM-4.7-Flash",
	"iflowcn":                "kimi-k2",
	"inception":              "mercury",
	"inference":              "mistral/mistral-nemo-12b-instruct",
	"io-net":                 "zai-org/GLM-4.6",
	"jiekou":                 "gpt-5-codex",
	"kilo":                   "prime-intellect/intellect-3",
	"kuae-cloud-coding-plan": "GLM-4.7",
	"llama":                  "cerebras-llama-4-maverick-17b-128e-instruct",
	"lmstudio":               "qwen/qwen3-30b-a3b-2507",
	"lucidquery":             "lucidquery-nexus-coder",
	"meganova":               "zai-org/GLM-4.6",
	"moark":                  "GLM-4.7",
	"modelscope":             "Qwen/Qwen3-30B-A3B-Instruct-2507",
	"moonshotai":             "kimi-k2-0905-preview",
	"moonshotai-cn":          "kimi-k2-0711-preview",
	"morph":                  "auto",
	"nano-gpt":               "zai-org/glm-5-original:thinking",
	"nebius":                 "zai-org/glm-4.7-fp8",
	"nova":                   "nova-2-lite-v1",
	"novita-ai":              "zai-org/glm-5",
	"nvidia":                 "nvidia/llama-3.1-nemotron-70b-instruct",
	"ollama-cloud":           "glm-5",
	"opencode":               "gpt-5.3-codex",
	"opencode-go":            "glm-5",
	"ovhcloud":               "meta-llama-3_3-70b-instruct",
	"poe":                    "stabilityai/stablediffusionxl",
	"privatemode-ai":         "gemma-3-27b",
	"qihang-ai":              "claude-opus-4-5-20251101",
	"qiniu-ai":               "claude-4.5-haiku",
	"requesty":               "google/gemini-2.5-flash",
	"scaleway":               "voxtral-small-24b-2507",
	"siliconflow":            "nex-agi/DeepSeek-V3.1-Nex-N1",
	"siliconflow-cn":         "zai-org/GLM-4.6V",
	"stackit":                "intfloat/e5-mistral-7b-instruct",
	"stepfun":                "step-3.5-flash",
	"submodel":               "zai-org/GLM-4.5-Air",
	"synthetic":              "hf:MiniMaxAI/MiniMax-M2",
	"upstage":                "solar-pro2",
	"vultr":                  "kimi-k2-instruct",
	"wandb":                  "microsoft/Phi-4-mini-instruct",
	"xiaomi":                 "mimo-v2-flash",
	"zai":                    "glm-5",
	"zai-coding-plan":        "glm-5",
	"zhipuai":                "glm-5",
	"zhipuai-coding-plan":    "glm-5",
}

var providerAliasModelOverrides = map[string]string{
	"openai":         ModelResolvedOpenAICodex,
	"opencode":       "gpt-5.3-codex",
	"github-copilot": "gpt-5.1-codex-max",
	"jiekou":         ModelResolvedOpenAICodex,
}

var modelAliasProviderOverrides = map[string]string{
	ModelAliasOpenAICodex: "opencode",
}

var providerIDSanitizer = regexp.MustCompile(`[^A-Z0-9]+`)

func IsValidProvider(provider string) bool {
	if provider == "offline" || provider == "openai" || provider == "anthropic" {
		return true
	}
	return IsOpenAICompatibleProvider(provider)
}

func IsOpenAICompatibleProvider(provider string) bool {
	_, ok := openAICompatibleProviderBaseURLs[provider]
	return ok
}

func OpenAICompatibleBaseURL(provider string) (string, bool) {
	v, ok := openAICompatibleProviderBaseURLs[provider]
	return v, ok
}

func ProviderAPIBaseURL(provider string) string {
	if provider == "openai" {
		return "https://api.openai.com/v1"
	}
	if provider == "anthropic" {
		return "https://api.anthropic.com/v1"
	}
	if v, ok := openAICompatibleProviderBaseURLs[provider]; ok {
		return v
	}
	return ""
}

func DefaultModelForProvider(provider string) string {
	if model := strings.TrimSpace(providerDefaultModels[provider]); model != "" {
		return model
	}
	return SafeFallbackModel
}

func NormalizeModelAlias(provider string, model string) string {
	trimmed := strings.TrimSpace(model)
	if trimmed == "" {
		return ""
	}

	if !strings.EqualFold(trimmed, ModelAliasOpenAICodex) {
		return trimmed
	}

	if resolved := providerAliasModelOverrides[provider]; resolved != "" {
		return resolved
	}

	return ModelResolvedOpenAICodex
}

func ResolveModel(provider string, selected string) string {
	model := NormalizeModelAlias(provider, selected)
	if model != "" {
		return model
	}
	return DefaultModelForProvider(provider)
}

func ResolveProviderModel(selectedProvider string, selectedModel string) (string, string) {
	provider := strings.TrimSpace(selectedProvider)
	if provider == "" {
		provider = DefaultAgentProvider
	}

	model := strings.TrimSpace(selectedModel)
	if model != "" {
		if overrideProvider := modelAliasProviderOverrides[strings.ToLower(model)]; overrideProvider != "" {
			provider = overrideProvider
		}
	}

	return provider, ResolveModel(provider, model)
}

func ProviderAPIKeyEnvVars(provider string) []string {
	switch provider {
	case "openai":
		return []string{"NINJOPS_OPENAI_API_KEY"}
	case "anthropic":
		return []string{"NINJOPS_ANTHROPIC_API_KEY"}
	case "opencode":
		return []string{"OPENCODE_API_KEY", "NINJOPS_OPENCODE_API_KEY"}
	}

	if !IsOpenAICompatibleProvider(provider) {
		return nil
	}

	upper := strings.ToUpper(strings.TrimSpace(provider))
	if upper == "" {
		return nil
	}

	sanitized := providerIDSanitizer.ReplaceAllString(upper, "_")
	sanitized = strings.Trim(sanitized, "_")
	if sanitized == "" {
		return nil
	}

	return []string{fmt.Sprintf("NINJOPS_%s_API_KEY", sanitized)}
}

func ProviderAPIKeyEnvVar(provider string) string {
	envVars := ProviderAPIKeyEnvVars(provider)
	if len(envVars) == 0 {
		return ""
	}

	return envVars[0]
}

func ProviderAPIKeyEnvHint(provider string) string {
	envVars := ProviderAPIKeyEnvVars(provider)
	if len(envVars) == 0 {
		return ""
	}

	return strings.Join(envVars, " or ")
}

func ResolveProviderBaseURL(provider string) (string, error) {
	baseURL := strings.TrimSpace(ProviderAPIBaseURL(provider))
	if baseURL == "" {
		return "", fmt.Errorf("provider %q does not have a configured base URL", provider)
	}

	start := strings.Index(baseURL, "${")
	for start >= 0 {
		end := strings.Index(baseURL[start:], "}")
		if end < 0 {
			break
		}

		end += start
		placeholder := baseURL[start : end+1]
		envVar := strings.TrimSuffix(strings.TrimPrefix(placeholder, "${"), "}")
		envValue := os.Getenv(envVar)
		if envValue == "" {
			return "", fmt.Errorf("provider %q requires %s to resolve base URL", provider, envVar)
		}

		baseURL = strings.Replace(baseURL, placeholder, envValue, 1)
		start = strings.Index(baseURL, "${")
	}

	return baseURL, nil
}
