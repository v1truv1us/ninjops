---
title: Configuration
description: Configure Ninjops with API credentials, providers, and customization options
---

# Configuration

Ninjops can be configured through multiple sources with a clear precedence hierarchy.

## Configuration Sources

Configuration values are loaded in this order (highest to lowest priority):
1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Config files**
4. **Default values** (lowest priority)

## Quick Setup

The easiest way to configure Ninjops is using the `configure` command:

```bash
ninjops configure --non-interactive \
  --base-url "https://your-instance.com" \
  --api-token "your-token" \
  --api-secret "your-secret" \
  --provider "offline"
```

This creates:
- `~/.config/ninjops/config.json` - Main configuration
- `~/.config/ninjops/auth-creds.json` - Secrets (separate file for security)

## Config File Locations

Ninjops looks for configuration in these locations:

1. **User config** (recommended):
   - `~/.config/ninjops/config.json`
   - `~/.config/ninjops/ninjops.jsonc`

2. **Project-local**:
   - `.ninjops/config.toml` (legacy format)
   - `.ninjops/config.json`

3. **Home directory**:
   - `~/.ninjops/config.toml` (legacy format)

## Configuration File Format

### JSON Format (Recommended)

**`config.json`:**
```json
{
  "ninja": {
    "base_url": "https://invoiceninja.example.com"
  },
  "agent": {
    "provider": "offline",
    "plan": "default",
    "model": "gpt-5-codex"
  },
  "serve": {
    "listen": "127.0.0.1",
    "port": 8080
  },
  "auth_creds_file": "/Users/you/.config/ninjops/auth-creds.json"
}
```

**`auth-creds.json`** (secrets only):
```json
{
  "ninja": {
    "api_token": "",
    "api_secret": ""
  },
  "agent": {
    "provider_api_key": ""
  }
}
```

### Legacy TOML Format

```toml
[ninja]
base_url = "https://invoiceninja.example.com"
api_token = ""
api_secret = ""

[agent]
provider = "offline"
plan = "default"
model = "gpt-5-codex"

[serve]
listen = "127.0.0.1"
port = 8080
```

## Environment Variables

You can override any configuration value with environment variables:

### Invoice Ninja

```bash
export NINJOPS_NINJA_BASE_URL="https://invoiceninja.example.com"
export NINJOPS_NINJA_API_TOKEN="your-token"
export NINJOPS_NINJA_API_SECRET="your-secret"  # optional
```

### AI Providers

```bash
export NINJOPS_AGENT_PROVIDER="offline"  # offline, openai, anthropic
export NINJOPS_AGENT_PLAN="default"      # default, codex-pro, opencode-zen, zai-plan
export NINJOPS_AGENT_MODEL="gpt-5-codex"
export NINJOPS_AGENT_PROVIDER_API_KEY="" # optional generic provider key override
export NINJOPS_OPENAI_API_KEY="sk-..."
export NINJOPS_ANTHROPIC_API_KEY="sk-ant-..."
```

### HTTP Server

```bash
export NINJOPS_SERVE_LISTEN="127.0.0.1"
export NINJOPS_SERVE_PORT="8080"
```

## Credential Precedence

For Invoice Ninja credentials:
1. `NINJOPS_NINJA_API_TOKEN` / `NINJOPS_NINJA_API_SECRET` environment variables
2. Values in main config (`config.json`, `ninjops.jsonc`, or legacy config files)
3. Values in `auth-creds.json` referenced by `auth_creds_file`

For AI provider API keys:
1. Provider-specific env var (e.g., `NINJOPS_OPENAI_API_KEY`, `NINJOPS_ANTHROPIC_API_KEY`)
2. `NINJOPS_AGENT_PROVIDER_API_KEY` (generic override)
3. Values in `auth-creds.json`

## AI Provider Configuration

### Offline Provider (Default)

No API key required. Uses rule-based processing.

```bash
ninjops configure --provider offline
```

### OpenAI

```bash
export NINJOPS_OPENAI_API_KEY="sk-..."
ninjops configure --provider openai
```

### Anthropic

```bash
export NINJOPS_ANTHROPIC_API_KEY="sk-ant-..."
ninjops configure --provider anthropic
```

### OpenAI-Compatible Providers

Ninjops supports 67+ OpenAI-compatible providers. Each has its own environment variable:

```bash
# Example for OpenCode
export NINJOPS_OPENCODE_API_KEY="..."
ninjops configure --provider opencode

# Example for custom provider
export NINJOPS_CUSTOMPROVIDER_API_KEY="..."
ninjops configure --provider customprovider
```

## Security Best Practices

1. **Use `auth-creds.json`** for secrets - it's gitignored by default
2. **Set restrictive permissions** on credential files:
   ```bash
   chmod 600 ~/.config/ninjops/auth-creds.json
   ```
3. **Never commit** API tokens or secrets to version control
4. **Use environment variables** in CI/CD pipelines
5. **Rotate keys** periodically

## Next Steps

- [Creating Quotes](/ninjops/guides/creating-quotes/) - Use your configuration to create quotes
- [AI Assistance](/ninjops/guides/ai-assistance/) - Configure AI providers for enhanced quotes
- [Commands Reference](/ninjops/reference/commands/) - See all available commands and flags
