# ADR-0003: Deterministic-First with Optional AI

## Status
Accepted

## Context
We need to decide on the role of AI in ninjops:
- AI could greatly enhance output quality
- Not all users want or can use AI (cost, privacy, offline needs)
- System must be reliable and predictable

## Decision

### Deterministic-First Architecture
1. **Core functionality works without AI**
   - Template-based generation
   - Rule-based transformations
   - Full validation
   - Complete Invoice Ninja integration

2. **AI is optional enhancement**
   - Providers: offline, openai, anthropic
   - Default: offline (no API keys required)
   - Graceful degradation if AI fails

3. **Offline provider implements real transformations**
   - Not stubs or placeholders
   - Rule-based clarify, polish, boundaries, line-items
   - Useful output without AI

### Provider Architecture
```go
type Provider interface {
    Name() string
    Execute(ctx, Request) (*Response, error)
    IsAvailable() bool
}
```

### Plan System
Plans modify prompting strategy, not provider:
- `default` - Professional, clear
- `codex-pro` - Data-driven, precise
- `opencode-zen` - Balanced, accessible
- `zai-plan` - Strategic, risk-focused

## Consequences

### Positive
- Works completely offline
- No API key requirements for basic use
- Predictable, testable output
- AI enhances rather than replaces

### Negative
- Offline provider less nuanced than AI
- More code to maintain (two implementations)

### Risks
- Users might expect AI-quality from offline provider (mitigated by clear docs)

## Alternatives Considered

1. **AI-only** - Excludes users without API access, unreliable
2. **No AI at all** - Misses enhancement opportunity
3. **Multiple tools** - Fragmented user experience
