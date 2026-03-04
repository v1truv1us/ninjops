---
title: AI Assistance
description: Enhance your quotes with optional AI assistance
---

# AI Assistance

Ninjops provides optional AI-powered assistance for improving QuoteSpecs. AI enhancement is completely optional - the deterministic templates work without any AI keys.

## Quick Start

### Check Available Providers

```bash
ninjops configure --provider offline  # No API key needed
```

### Configure OpenAI

```bash
export NINJOPS_OPENAI_API_KEY="sk-..."
ninjops configure --provider openai
```

### Configure Anthropic

```bash
export NINJOPS_ANTHROPIC_API_KEY="sk-ant-..."
ninjops configure --provider anthropic
```

## AI Roles

Use the `assist` command to enhance specific aspects of your QuoteSpec:

### clarify - Normalize Specifications

```bash
ninjops assist clarify --input quote.json
```

What it does:
- Normalizes feature names for consistency
- Extracts implicit requirements
- Identifies missing information
- Cleans up ambiguous descriptions

### polish - Improve Professionalism

```bash
ninjops assist polish --input quote.json
```

What it does:
- Improves tone and readability
- Makes language more professional
- Ensures consistent voice
- Enhances clarity

### boundaries - Define Scope
```bash
ninjops assist boundaries --input quote.json
```

What it does:
- Defines what's included in the project
- Identifies what's out of scope
- Sets clear boundaries
- Manages client expectations

### line-items - Generate Billable Items
```bash
ninjops assist line-items --input quote.json
```

What it does:
- Breaks down features into billable items
- Estimates reasonable quantities
- Groups items by category
- Includes brief descriptions

## Providers

### offline (Default)
Rule-based provider that works without API keys.

**Advantages:**
- No external dependencies
- Deterministic transformations
- Fast execution
- Works offline

**Best for:**
- Baseline improvements
- Offline work
- Privacy-sensitive projects
- Testing and development

### openai
OpenAI GPT models with structured output.

**Requirements:**
- `NINJOPS_OPENAI_API_KEY` environment variable
- Or `NINJOPS_AGENT_PROVIDER_API_KEY` for generic key

**Default model:** `gpt-4o-mini`

**Best for:**
- Nuanced improvements
- Complex clarifications
- Professional tone enhancements

### anthropic
Anthropic Claude models.

**Requirements:**
- `NINJOPS_ANTHROPIC_API_KEY` environment variable
- Or `NINJOPS_AGENT_PROVIDER_API_KEY` for generic key

**Default model:** `claude-3-5-sonnet-20241022`

**Best for:**
- Technical content
- Detailed specifications
- Complex projects

### OpenAI-Compatible Providers

Ninjops supports 67+ OpenAI-compatible providers. Configure any provider using:

```bash
export NINJOPS_<PROVIDER_ID>_API_KEY="your-key"
ninjops configure --provider <provider-id>
```

## Plans

Plans determine the AI's approach to assisting with your QuoteSpec.

### default
Balanced, general-purpose plan.

**Characteristics:**
- Moderate thoroughness
- Balanced approach
- General improvements

### codex-pro
Developer-focused with technical depth.

**Characteristics:**
- Deep technical analysis
- Code-focused improvements
- Developer-centric language

### opencode-zen
Balanced, warm communication.

**Characteristics:**
- Accessible language
- Friendly tone
- Simplicity focus

### zai-plan
Strategic planning with risk mitigation.

**Characteristics:**
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

```

### Configure Command

```bash
ninjops configure --provider openai --plan codex-pro
```

### Skip Connectivity Test

```bash
ninjops configure --provider openai --skip-provider-test
```

## Workflow Example

```bash
# 1. Start with basic QuoteSpec
ninjops validate --input quote.json

# 2. Clarify the specification
ninjops assist clarify --input quote.json --output clarified.json

# 3. Polish the language
ninjops assist polish --input clarified.json --output polished.json

# 4. Define boundaries
ninjops assist boundaries --input polished.json --output with-boundaries.json

# 5. Generate line items
ninjops assist line-items --input with-boundaries.json --output final.json

# 6. Generate documents
ninjops generate --input final.json --format md --out-dir output/
```

## Best Practices

1. **Start offline** - Use the offline provider to establish baseline quality
2. **Iterate** - Run multiple assist roles as needed
3. **Review output** - Always review AI suggestions before accepting
4. **Use version control** - Commit both original and AI-enhanced versions
5. **Test thoroughly** - Validate final QuoteSpec before syncing

## Next Steps

- [Invoice Ninja Sync](/ninjops/guides/invoice-ninja-sync/) - Sync your enhanced quote
- [Examples](/ninjops/examples/church-website/) - See complete examples
