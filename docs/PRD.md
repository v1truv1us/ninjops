# Ninjops - Product Requirements Document

## Problem Statement

Freelancers and small agencies using Invoice Ninja need a way to quickly generate professional, consistent quotes and invoices with minimal manual effort. Current workflows involve:
- Manually writing proposals and terms for each client
- Inconsistent formatting and terminology across documents
- Time-consuming updates when syncing with Invoice Ninja
- Risk of errors when manually updating invoice fields

## Target Users

1. **Freelance Developers** - Solo developers who need to quickly generate quotes for web/app projects
2. **Small Agencies** - Teams managing multiple client quotes and invoices
3. **Consultants** - Professionals who need consistent, professional documentation

## Goals

1. **Reduce quote generation time** - From hours to minutes
2. **Ensure consistency** - Professional formatting, consistent terminology
3. **Safe Invoice Ninja integration** - Reliable sync without accidental overwrites
4. **Deterministic-first** - Works without AI keys; AI enhancement is optional
5. **Full lifecycle support** - From initial quote to final invoice

## Non-Goals

1. Full Invoice Ninja replacement - ninjops complements, not replaces Invoice Ninja
2. Complex CRM features - Client management happens in Invoice Ninja
3. Payment processing - Handled by Invoice Ninja
4. Multi-tenant/SaaS - Single-user CLI tool

## User Workflows

### Workflow 1: New Quote Creation
1. `ninjops init` - Initialize project
2. Edit `quote_template.json` with project details
3. `ninjops validate --input quote.json` - Validate spec
4. `ninjops generate --input quote.json --format md --out-dir output/` - Generate documents
5. Review generated proposal and terms
6. `ninjops ninja sync --input quote.json --dry-run` - Preview sync
7. `ninjops ninja sync --input quote.json` - Push to Invoice Ninja

### Workflow 2: AI-Assisted Refinement
1. `ninjops new quote > quote.json` - Create new spec
2. Edit basic details
3. `ninjops assist clarify --input quote.json` - Clarify and normalize
4. `ninjops assist boundaries --input quote.json` - Generate scope boundaries
5. `ninjops assist line-items --input quote.json` - Suggest line items
6. `ninjops assist polish --input quote.json` - Improve wording
7. Generate and sync

### Workflow 3: Update Existing Quote
1. `ninjops ninja pull --input quote.json` - Pull current state
2. Edit quote.json with updates
3. `ninjops ninja diff --input quote.json` - Review changes
4. `ninjops ninja sync --input quote.json` - Push updates

## Success Criteria

1. **Quote generation in < 5 minutes** - From spec to Invoice Ninja
2. **Zero data loss** - GET → patch → PUT ensures safe updates
3. **100% test coverage** - All critical paths have tests
4. **Works offline** - Deterministic generation without external APIs
5. **Clear error messages** - Helpful errors with redacted secrets

## Key Features

### Core Features
- QuoteSpec JSON schema with validation
- Template-based proposal/terms generation
- Invoice Ninja v5 API integration
- Reference-based tracking for safe updates
- Local state management

### AI Features (Optional)
- Clarify: Normalize and extract from free text
- Polish: Improve tone and readability
- Boundaries: Generate scope boundaries
- Line Items: Suggest billable items

### CLI Features
- Clean help and documentation
- Dry-run mode for all operations
- Diff display for changes
- JSON output mode for scripting
- Confirmation prompts for safety

## Timeline

### Phase 1: Foundation
- Config system
- QuoteSpec schema
- Template generation
- Validation

### Phase 2: Integration
- Invoice Ninja client
- Sync logic
- State management

### Phase 3: Agents
- Offline provider
- OpenAI provider
- Anthropic provider
- Plans system

### Phase 4: Polish
- CLI UX refinement
- Documentation
- Testing
- CI/CD
