# Agents System

## Overview

Ninjops provides an optional AI-powered assistance layer for improving QuoteSpecs. The system supports multiple providers and plans.

## Roles

### clarify
Normalize and clarify the project specification:
- Normalize feature names for consistency
- Extract implicit requirements from descriptions
- Identify missing information
- Ensure required sections are present
- Standardize terminology

### polish
Improve the professionalism and readability:
- Improve tone consistency
- Tighten verbose sentences
- Ensure professional language
- Add helpful section headings
- Standardize formatting

### boundaries
Define project scope boundaries:
- Generate appropriate minor_changes items
- Create out_of_scope items
- List reasonable assumptions
- Identify potential risks
- Suggest client responsibilities

### line-items
Suggest billable line items:
- Break down features into billable items
- Estimate reasonable quantities
- Group items by category
- Include brief descriptions

## Providers

### offline (default)
Rule-based provider that works without API keys:
- No external dependencies
- Deterministic transformations
- Fast execution
- Suitable for baseline improvements

### openai
OpenAI GPT models:
- Requires `NINJOPS_OPENAI_API_KEY`
- Uses GPT-4o-mini by default
- JSON mode for structured output
- Best for nuanced improvements

### anthropic
Anthropic Claude models:
- Requires `NINJOPS_ANTHROPIC_API_KEY`
- Uses Claude 3.5 Sonnet by default
- JSON extraction from response
- Best for detailed analysis

## Plans

Plans modify the system prompt and approach for different use cases.

### default
Professional, clear proposals tailored to client needs.

### codex-pro
Data-driven, precise proposals with detailed specifications:
- Focus on measurable outcomes
- Technical precision
- Detailed specifications
- Clear acceptance criteria

### opencode-zen
Balanced, warm communication making complex topics accessible:
- Accessible language
- Balanced approach
- Developer-friendly tone
- Simplicity focus

### zai-plan
Strategic planning with risk mitigation and milestone focus:
- Risk assessment
- Milestone planning
- Resource optimization
- Contingency plans

## Configuration

### Environment Variables
```bash
# Provider selection
export NINJOPS_AGENT_PROVIDER=offline  # offline, openai, anthropic

# Plan selection
export NINJOPS_AGENT_PLAN=default  # default, codex-pro, opencode-zen, zai-plan

# API Keys (optional, required for non-offline providers)
export NINJOPS_OPENAI_API_KEY=sk-...
export NINJOPS_ANTHROPIC_API_KEY=sk-ant-...
```

### Config File
```toml
[agent]
provider = "offline"
plan = "default"
```

### CLI Flags
```bash
ninjops assist clarify --input quote.json --provider openai --plan codex-pro
```

## Output Schemas

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
ninjops assist line-items --input quote.json --write
```

### With AI Provider
```bash
export NINJOPS_AGENT_PROVIDER=openai
export NINJOPS_OPENAI_API_KEY=sk-...

ninjops assist clarify --input quote.json --plan codex-pro
ninjops assist polish --input quote.json --write
```

### HTTP API
```bash
curl -X POST http://localhost:8080/assist/clarify \
  -H "Content-Type: application/json" \
  -d '{"quote_spec": {...}, "provider": "openai", "plan": "codex-pro"}'
```
