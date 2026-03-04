# Ninjops - Technical Specification

## Architecture Overview

```
cmd/ninjops/main.go
    └── internal/app/        # Cobra commands, CLI UX
        ├── internal/config/ # Configuration loading
        ├── internal/spec/   # QuoteSpec types & validation
        ├── internal/generate/ # Template rendering
        ├── internal/agents/ # AI providers
        └── internal/invoiceninja/ # API client
```

## Configuration

### Precedence (highest to lowest)
1. Command-line flags
2. Environment variables
3. Config file
4. Defaults

### Config File Search Order
1. `./ninjops.toml` (project-local)
2. `~/.config/ninjops/config.toml`
3. `~/.ninjops/config.toml`

### Environment Variables
```
NINJOPS_NINJA_BASE_URL      # Invoice Ninja URL
NINJOPS_NINJA_API_TOKEN     # API token (required for sync)
NINJOPS_NINJA_API_SECRET    # Optional secret header
NINJOPS_AGENT_PROVIDER      # offline | openai | anthropic
NINJOPS_AGENT_PLAN          # default | codex-pro | opencode-zen | zai-plan
NINJOPS_OPENAI_API_KEY      # OpenAI API key
NINJOPS_ANTHROPIC_API_KEY   # Anthropic API key
```

## QuoteSpec Schema

### Version: 1.0.0

```json
{
  "schema_version": "1.0.0",
  "metadata": {
    "reference": "ninjops:<uuid>",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "client": {
    "name": "string (required)",
    "email": "string (required, valid email)",
    "phone": "string (optional)",
    "org_type": "business | church | nonprofit | tax_exempt",
    "address": { ... }
  },
  "project": {
    "name": "string (required)",
    "description": "string (required)",
    "type": "string (required)",
    "timeline": "string (optional)",
    "technologies": ["string"]
  },
  "work": {
    "features": [
      {
        "name": "string (required)",
        "description": "string (required)",
        "priority": "high | medium | low",
        "category": "string"
      }
    ],
    "responsibilities": ["string"],
    "minor_changes": ["string"],
    "out_of_scope": ["string"],
    "assumptions": ["string"]
  },
  "pricing": {
    "currency": "USD (default)",
    "line_items": [
      {
        "description": "string",
        "quantity": "number",
        "rate": "number",
        "amount": "number",
        "category": "string"
      }
    ],
    "discount": { "type", "percentage" | "amount", "description" },
    "recurring": { "type", "amount", "frequency", "description" },
    "deposit": { "percentage" | "amount", "due_date" },
    "payment_terms": "string",
    "total": "number"
  },
  "settings": {
    "tone": "professional | formal | friendly",
    "include_timeline": "boolean",
    "include_pricing": "boolean"
  }
}
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Validation error |
| 3 | Configuration error |
| 4 | API/network error |
| 5 | Not found |

## Security

### Secrets Handling
1. **Never log secrets** - API tokens, API keys redacted in all logs
2. **Env var only for secrets** - Never store in config files
3. **Redact in errors** - Show first 4 chars + `****`
4. **Token redaction** - All error messages checked for token leakage

### Local Server Security
1. **Bind 127.0.0.1 by default** - No external access
2. **No auth by default** - Designed for local/single-user
3. **Optional --listen flag** - For LAN access (user responsibility)

## Invoice Ninja Integration

### API Version
- Invoice Ninja v5
- Base path: `/api/v1`

### Authentication Headers
```
X-API-Token: <token>
X-Requested-With: XMLHttpRequest
X-API-Secret: <secret> (optional)
```

### Sync Strategy
1. **GET** - Fetch current entity state
2. **Patch** - Build update with only changed fields
3. **PUT** - Submit full payload for consistency

### Reference Tracking
- Format: `ninjops:<uuid>`
- Stored in: `custom_value1` field
- Local state: `.ninjops/state.json`

## Testing Strategy

### Unit Tests
- All packages have `*_test.go` files
- Table-driven tests for validation
- Mock interfaces for external dependencies

### Integration Tests
- `httptest` server for Invoice Ninja API mock
- End-to-end sync workflow tests
- Template rendering snapshot tests

### Test Commands
```bash
make test           # Unit tests
make test-integration # Integration tests
make test-coverage  # Coverage report
```

## Template System

### Templates
- `proposal.md.tmpl` - Main proposal document
- `terms.md.tmpl` - Terms and conditions
- `notes.txt.tmpl` - Public notes summary

### Variables
- `.Client.Name`, `.Client.Email`, `.Client.OrgType`
- `.Project.Name`, `.Project.Description`, `.Project.Type`
- `.Work.Features`, `.Work.Responsibilities`
- `.Pricing.LineItems`, `.Pricing.Total`
- `.OrgTypeWording` - Auto-generated based on org type

### Org-Type Wording
- `church` / `tax_exempt` → "tax-exempt religious organization"
- `nonprofit` → "tax-exempt nonprofit organization"
- `business` → No tax-exempt wording

## Error Handling

### Error Types
1. **ValidationError** - Invalid QuoteSpec
2. **ConfigError** - Configuration issues
3. **APIError** - Invoice Ninja API errors
4. **ProviderError** - AI provider errors

### Error Messages
- Clear, actionable messages
- Include field paths for validation
- Redact all sensitive values
- Suggest remediation steps
