# ADR-0001: Name and Scope

## Status
Accepted

## Context
We need to name the project and define its scope clearly to ensure focused development and clear user expectations.

## Decision

### Name: `ninjops`
- Short, memorable, unique
- Evokes precision and skill (ninja) + operations (ops)
- Available as npm package, Go module path, and domain

### Scope
Ninjops is a **developer-focused CLI tool** for:
1. Creating and validating QuoteSpec JSON files
2. Generating professional proposals, terms, and notes
3. Optionally enhancing specs with AI assistance
4. Syncing with Invoice Ninja v5 for quote/invoice lifecycle management

### Non-Goals (Explicitly Out of Scope)
- Full Invoice Ninja replacement
- CRM or client management
- Payment processing
- Multi-tenant/SaaS deployment
- Web UI (HTTP API is for future TUI integration, not public web)

## Consequences

### Positive
- Clear, focused scope reduces development complexity
- Developer-focused design fits the target audience
- Deterministic-first approach ensures reliability without AI keys

### Negative
- May not serve non-technical users (requires CLI comfort)
- Limited to Invoice Ninja v5 (no other invoicing platforms)

### Risks
- Name could be confused with DevOps tools (mitigated by clear documentation)

## Alternatives Considered

1. **quotegen** - Too generic, doesn't convey Invoice Ninja integration
2. **invoicer** - Too broad, implies payment processing
3. **ninjabilling** - Misleading, not a billing system
4. **proposal-ninja** - Doesn't convey operations/CLI nature
