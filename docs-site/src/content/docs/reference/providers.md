---
title: Providers
description: Complete reference of AI providers and providers available in Ninjops
---

import { read } from '../../..//users/johnferguson/Github/ninjops/docs/AGENTS.md' }
# Providers

Ninjops supports multiple AI providers for optional quote enhancement:

 This reference documents the available providers and their usage, configuration, and output schemas.

## Available Providers

### offline (default)
Rule-based provider that works without AI keys. No external dependencies required. deterministic transformations. fast execution. Suitable for baseline improvements.

### openai
OpenAI GPT models
- Requires `NINJOPS_OPENAI_API_KEY` or `NINJOPS_AGENT_PROVIDER_API_KEY` environment variable
- Uses GPT-4o-mini by default
- JSON mode for structured output
- Best for nuanced improvements
- API docs: [OpenAI API documentation](https://platform.openai.com/docs/api-reference)

### anthropic
Anthropic Claude models
- Requires `NINJOPS_ANTHROPIC_API_KEY` or `NINJOPS_AGENT_PROVIDER_API_KEY` environment variable
- Uses Claude 3.5 Sonnet by default
- JSON extraction from response
- best for detailed analysis
- API docs: [Anthropic API documentation](https://docs.anthropic.com/)

### OpenAI-Compatible Providers
Ninjops supports 67+ OpenAI-compatible providers. Configure any provider using:

```bash
export NINJOPS_<PROVIDER_ID>_API_KEY="your-key"
ninjops configure --provider <provider-id>
```

For supported providers, see the [OpenAI-compatible provider list](https://github.com/v1truv1us/ninjops/blob/main).

 OpenAI-compatible providers include providers like OpenAI, Anthropic, Groq, Mistral, Together AI, Perplexity, OpenRouter, Together, LM Studio, and xAI, Fireworks AI, and etc.

## Configuration

Set provider and plan via environment variables or config file.

```bash
# Provider selection
export NINJOPS_AGENT_PROVIDER=offline  # offline, openai, anthropic

# Plan selection
export NINJOPS_AGENT_PLAN=default  # default, codex-pro, opencode-zen, zai-plan
```

### CLI flags
```bash
ninjops assist clarify --input quote.json --provider openai --plan codex-pro
```

## Output Schemas
All roles return JSON output validated against the QuoteSpec schema.

### clarify
```json
{
  "features": [
    {"name": "string", "description": "string", "priority": "string"}
  ],
  "responsibilities": ["string"],
  "questions": ["string"],
  "confidence": 0.0-1.0
}
```

### polish
```json
{
  "polished_sections": {
    "section_name": "polished_content"
  },
  "improvements": ["string"],
  "confidence": 0.0-1.0
}
```

### boundaries
```json
{
  "minor_changes": ["string"],
  "out_of_scope": ["string"],
  "assumptions": ["string"],
  "client_responsibilities": ["string"],
  "confidence": 0.0-1.0
}
```

### line-items
```json
{
  "line_items": [
    {"category": "string", "description": "string", "quantity": number}
  ],
  "notes": ["string"],
  "confidence": 0.0-1.0
}
```

## Safety Rules

1. **Always validate output** - All agent outputs are validated before merging
2. **Deterministic merge** - Changes merged predictably into QuoteSpec
3. **No pricing by default** - line-items doesn't suggest rates
4. **Confidence scoring** - All responses include confidence level
5. **Graceful degradation** - Falls back to offline if AI fails

## Usage Examples
### Basic Usage (Offline)
```bash
ninjops assist clarify --input quote.json
ninjops assist polish --input quote.json --write
ninjops assist boundaries --input quote.json --write
```

### With AI Provider
```bash
export NINJOPS_AGENT_PROVIDER=openai
export NINJOPS_OPENAI_API_KEY=sk-...

ninjops assist clarify --input quote.json --provider openai --plan codex-pro
ninjops assist polish --input quote.json --provider openai --plan codex-pro
ninjops assist boundaries --input quote.json --provider openai --plan codex-pro
```

## HTTP API
See [HTTP API documentation](/ninjops/guides/http-api/) for full details.

## Best Practices
1. Start with offline provider for baseline improvements
2. Use `--dry-run` for preview changes
3. Always validate QuoteSpecs before syncing
4. Review AI suggestions before accepting
5. Use version control to track changes
6. Configure AI providers with environment variables
7. Use confidence scoring to identify low-confidence responses
8. Use JSON output mode for structured responses
9. Start interactive with simple quote creation
10. Create and preview content before syncing
