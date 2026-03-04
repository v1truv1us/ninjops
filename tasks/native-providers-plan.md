# Native Provider Support Plan

Date: 2026-03-01

## Goal

Add native provider support to ninjops with coverage for all providers in `models.dev`, using opencode's adapter strategy as the reference pattern.

## Current State

- Runtime supports three providers only: `offline`, `openai`, `anthropic`.
- Provider validation is mostly static and not registry-driven.
- `configure` now writes split files:
  - Main config: non-secrets
  - `auth-creds.json`: provider and Invoice Ninja secrets
- Provider connectivity checks exist but are hardcoded.

## External Findings

### models.dev snapshot

- Provider count: 97
- Unique npm adapter packages: 23
- Provider distribution by npm:
  - `@ai-sdk/openai-compatible`: 67 providers
  - `@ai-sdk/anthropic`: 7
  - `@ai-sdk/openai`: 2
  - `@ai-sdk/azure`: 2
  - 12 other npm packages with 1 provider each

### opencode patterns to emulate

- models.dev as source-of-truth provider/model catalog
- Registry merge pipeline: models.dev + config overrides + auth/env
- Adapter loader split:
  - Generic bundled adapters
  - Provider-specific custom loaders/hooks for auth, headers, model resolution
- Separate auth store from config (`auth.json`) and env precedence

## Target Architecture (ninjops)

Introduce a dedicated provider subsystem:

- `internal/providers/registry`
  - Provider metadata, adapter type, auth schema, capability flags
- `internal/providers/catalog`
  - models.dev fetch/cache (`provider-catalog.json`) + normalization
- `internal/providers/auth`
  - Resolve credentials from flags/env/config/auth-creds
- `internal/providers/adapters`
  - Protocol families (not one adapter per provider)
- `internal/providers/connectivity`
  - Probe framework by provider/adaptor type
- `internal/providers/runtime`
  - Client factory + request execution abstraction

`internal/agents` stays responsible for prompting and response shaping; it delegates transport/provider behavior to `internal/providers`.

## Adapter Families and Coverage Strategy

Use adapter families, not 97 bespoke adapters.

### Family mapping

- `openai_compat`:
  - covers all providers with npm `@ai-sdk/openai-compatible`
  - expected immediate coverage: 67 providers
- `anthropic_messages`:
  - npm `@ai-sdk/anthropic`
  - expected coverage: 7 providers
- `openai_native`:
  - npm `@ai-sdk/openai`
  - expected coverage: 2 providers
- `google_genai`:
  - npm `@ai-sdk/google`
  - expected coverage: 1 provider
- `google_vertex` and `google_vertex_anthropic`:
  - npm `@ai-sdk/google-vertex`, `@ai-sdk/google-vertex/anthropic`
  - expected coverage: 2 providers
- `azure_openai`:
  - npm `@ai-sdk/azure`
  - expected coverage: 2 providers
- `bedrock`:
  - npm `@ai-sdk/amazon-bedrock`
  - expected coverage: 1 provider
- `provider_specific` (fallback explicit adapters):
  - `@openrouter/ai-sdk-provider`, `@gitlab/gitlab-ai-provider`, `ai-gateway-provider`, `venice-ai-sdk-provider`, `@jerome-benoit/sap-ai-provider-v2`, and any package requiring non-standard auth/workflow

## Coverage Rule (for "all models.dev providers")

All providers are considered covered when:

1. Provider has a registry entry loaded from models.dev.
2. Registry resolves a valid adapter family or explicit adapter.
3. Required credential fields are known.
4. Connectivity probe is declared.
5. At least one smoke test path exists (mocked or live-tier policy).

Add CI check that fails if any models.dev provider has no adapter mapping.

## Config and Auth Schema

Keep current files and extend them.

### Main config (`config.json`/`ninjops.jsonc`)

- `agent.provider`
- `agent.model` (new)
- `agent.provider_options` (new)
- `auth_creds_file` (existing)

### Auth creds (`auth-creds.json`)

Keep legacy + add provider-scoped credentials:

```json
{
  "agent": {
    "provider_api_key": "legacy-default",
    "providers": {
      "openai": {"api_key": "..."},
      "anthropic": {"api_key": "..."},
      "amazon-bedrock": {
        "aws_access_key_id": "...",
        "aws_secret_access_key": "...",
        "aws_region": "us-east-1"
      }
    }
  },
  "ninja": {
    "api_token": "...",
    "api_secret": "..."
  }
}
```

Credential precedence:

- flags > env > main config > auth-creds > defaults

## Phased Roadmap

| Phase | Scope | Deliverables | Acceptance Criteria |
| --- | --- | --- | --- |
| 0 | Foundation | `internal/providers` skeleton, registry interfaces, integration shim | Existing tests pass; no behavior break for offline/openai/anthropic |
| 1 | MVP tier-1 | Adapter families: openai_compat, openai_native, anthropic_messages, azure_openai, bedrock; config/auth resolver v2; `providers list` and `providers test` commands | Tier-1 providers configurable and testable end-to-end |
| 2 | Catalog sync | models.dev cache + automatic mapping from `npm` field to adapter family | CI coverage check green for all providers with known mapping |
| 3 | Tier-2 and long-tail | Explicit adapters and hooks for remaining provider-specific packages; improved per-provider options | All models.dev providers map to adapter + probe; docs complete |
| 4 | Hardening | retries, normalized errors, nightly smoke matrix | Stable nightly pass rate and actionable diagnostics |

## Tier Recommendations

- Tier-1 (first 2-3 sprints):
  - `offline`, `openai`, `anthropic`, `google`, `azure`, `amazon-bedrock`, `openrouter`, `groq`
- Tier-2 (next):
  - `mistral`, `cohere`, `xai`, `perplexity`, `togetherai`, `deepinfra`, `cloudflare-*`
- Long-tail:
  - remaining models.dev providers primarily via `openai_compat`

## Testing and CI

- Unit tests:
  - adapter selection by provider metadata
  - credential precedence
  - option validation per family
- Contract tests (mock HTTP):
  - one suite per adapter family
- Smoke tests:
  - local deterministic smoke remains offline
  - live nightly smoke for selected tier-1 providers with CI secrets
- Coverage gate:
  - fail build if any models.dev provider has no adapter mapping

## Migration Notes

- Keep current `internal/agents` public behavior stable during migration.
- Additive config changes only; keep legacy `agent.provider_api_key` fallback.
- Move hardcoded provider validation out of config package and into registry-backed checks.

## Risks and Mitigation

- Provider sprawl: mitigate with family-based adapters and metadata-driven mapping.
- Complex auth variants (AWS/Azure/Vertex): encode auth schemas and targeted probes.
- Upstream API drift: cache versioning + fallback + explicit error normalization.
- Breaking existing users: additive migration + fallback + deprecation warnings.
