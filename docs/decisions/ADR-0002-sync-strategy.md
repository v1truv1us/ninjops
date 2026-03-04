# ADR-0002: Sync Strategy and Reference Tracking

## Status
Accepted

## Context
When syncing QuoteSpecs with Invoice Ninja, we need a reliable strategy to:
1. Identify which remote entities correspond to local QuoteSpecs
2. Prevent accidental updates to wrong entities
3. Handle concurrent modifications safely
4. Support both initial creation and subsequent updates

## Decision

### Reference Tag Format
```
ninjops:<uuid-v4>
```
- Stored in QuoteSpec `metadata.reference`
- Stored in Invoice Ninja `custom_value1` field
- UUID provides uniqueness without revealing sensitive info

### Local State Tracking
Store mapping in `.ninjops/state.json`:
```json
{
  "reference_id": "uuid",
  "client_id": "ninja-client-id",
  "quote_id": "ninja-quote-id",
  "invoice_id": "ninja-invoice-id",
  "last_sync_hash": "content-hash",
  "updated_at": "timestamp"
}
```

### Sync Strategy (Priority Order)
1. **Explicit IDs** - If `--quote-id`/`--invoice-id` provided, use directly
2. **Reference tag** - Search by `custom_value1` = reference
3. **Local state** - Use stored IDs from `.ninjops/state.json`
4. **Fuzzy match** - Only with `--allow-fuzzy` flag (disabled by default)

### Update Strategy: GET → PATCH → PUT
1. **GET** current entity state
2. **Build patch** comparing local generated vs remote
3. **PUT** full payload for consistency

## Consequences

### Positive
- Deterministic identification without ambiguity
- Safe updates with change detection
- Clear audit trail via reference tags
- Works offline (local state persists)

### Negative
- Requires `.ninjops/` directory in project
- Reference tag must be preserved in Invoice Ninja

### Risks
- If user deletes `custom_value1` in Invoice Ninja, link is broken (mitigated by local state)

## Alternatives Considered

1. **Name-based matching** - Too unreliable, name collisions
2. **Email-based matching** - Only works for clients, not quotes
3. **Sequential IDs** - Requires server-side generation, not portable
4. **Hash-based IDs** - Changes with content, not stable
