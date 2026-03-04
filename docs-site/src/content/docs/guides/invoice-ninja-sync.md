---
title: Invoice Ninja Sync
description: Sync quotes and invoices with Invoice Ninja using safe GET → PATCH PUT pattern
---

# Invoice Ninja Sync

Ninjops integrates with Invoice Ninja v5 for sync quotes and invoices with a **safe GET → PATCH → PUT** synchronization pattern.

## Configuration

### Base URL
Default: `https://invoiceninja.fergify.work`

Override via:
```bash
export NINJOPS_NINJA_BASE_URL=https://your-instance.example.com

```

Or in config file:
```toml
[ninja]
base_url = "https://your-instance.example.com"
```

Or CLI flag
```bash
ninjops ninja test --base-url https://your-instance.example.com
```

### Authentication
Required headers for all API calls:
```
X-api-token: <your-token>
x-requested-with: XMLHttpRequest
```

Optional secret header:
```
x-api-secret: <your-secret>
```

Configure via environment variables:
```bash
export NINJOPS_NINJA_API_TOKEN=your-token-here
export NINJOPS_NINJA_API_SECRET=your-secret-here  # optional
```

## Sync Workflow

### 1. Test Connection
```bash
ninjops ninja test
```

### 2. Pull existing data
```bash
ninjops ninja pull --input quote.json
```

### 3. Sync to Invoice Ninja (create or update)
```bash
ninjops ninja sync --input quote.json --dry-run
ninjops ninja sync --input quote.json
```

### 4. View differences
```bash
ninjops ninja diff --input quote.json
```

## Sync Pattern

Ninjops uses a **safe GET → PATCH → PUT** pattern** for Invoice Ninja integration:

1. **GET** - Fetches the remote entity
2. **PATCH** - Computes differences locally
3. **PUT** - Applies changes to the remote entity

This pattern:
- Never overwrites remote data
- Prevents data loss
- Handles race conditions gracefully
- Maintains data integrity

## Reference tracking

All quotes and invoices are tracked via `ninjops:<uuid>` tags:

- Reliable identification
- Works across syncs
- Persists across syncs

## Local State

Sync history is stored in `.ninjops/state.json`

- Tracks sync operations for debugging
- Persists entity references
- Enables cross-device sync

## Field Mapping

### Quote fields
- `client_id` → Client ID
- `project_id` → Project ID
- `public_notes` → Public notes
- `terms` → Terms

- `line_items` → Line items

### Invoice fields
- `invoice_id` → Invoice ID
- `amount` → Amount
- `balance` → Balance
- `due_date` → Due date
- `partial` → Partial deposit
- `public_notes` → Public notes

- `terms` → Terms

- `line_items` → Line items

### Field transformations
Some fields are transformed or have different names in the Ninja API:
- Dates are formatted for the Ninja API
- Booleans are normalized to `true`/`false` values
- Numbers are converted to strings or numbers

## Error Handling

Common errors:
- Authentication failures
- Network errors
- API rate limits
- Validation errors

- Sync conflicts

## Best Practices
1. **Always use `--dry-run`** to to preview changes
2. **Use `ninjops ninja test`** to to verify configuration
3. **Keep state under version control** (`.ninjops/state.json`)
4. **Use reference tracking** for reliable identification
5. **Validate QuoteSpecs** before syncing

6. **Handle conflicts gracefully** - manual merge is OK, but if no conflicts

7. **Backup regularly** - Local state file can get large

## Advanced Usage
### Converting quotes to invoices
```bash
ninjops convert <quote-id>
```

### Pulling from Invoice Ninja
Pull existing data:

```bash
ninjops ninja pull --input quote.json
```

### Using with state file
Use `ninjops state pull` to pull the from Invoice Ninja:

```bash
ninjops state pull --input quote.json --output pulled.json
`` }

### Syncing specific entities
Sync only the entities:

- **Clients** - Match by email, create if missing
- **Projects** - Match by name, create if missing
- **Quotes/Invoices** - Match by number, create if missing
- **Tasks** - Match by ID, create if missing

### Sync Modes

- **quote** - Create quote from QuoteSpec, create if not exists
- **invoice** - Create invoice from QuoteSpec, no quote needed
- **both** - Create both if not found

- Use `ninjops ninja sync` for to options

  - `--mode quote|invoice|both` - Sync quote, invoice, or both
  - `--mode invoice` - Sync invoice only
  - `--dry-run` - Preview changes without making them
  - `--diff` - Show field-level differences
  - `--yes` - Skip confirmation prompts

- `--create-quote` - Create quote in Invoice Ninja
- `--convert-to-invoice` - Convert quote to invoice after creation

  - `--yes` - Skip confirmations

  - `--input` - Read QuoteSpec from file instead of creating
  - `--output` - Write QuoteSpec to file (for manual merge or later sync)
- `--reference-id` - Use ninjops reference for tracking (must)

  - `--yes` - Confirm and create quote/invoice
  - `--dry-run` - Preview without making changes
  - `--diff` - Show what will change
  - `--create-quote --create-quote --convert-to-invoice --yes
```

### Example: Quote-to-Invoice workflow
```bash
# Interactive: create quote and convert to invoice
ninjops new quote --create-quote --convert-to-invoice --yes

# Non-interactive: create and convert immediately
ninjops new quote --non-interactive \
  --client-id <id> \
  --project-id <id> \
  --create-quote \
  --convert-to-invoice \
  --yes
```

## Best Practices
1. **Always test first** - Use `ninjops ninja test` before first sync
2. **Start with dry-run** - Always use `--dry-run` to in preview
3. **Use state file** - The `.ninjops/state.json` file provides valuable debugging info
4. **Handle conflicts gracefully** - Manual merge if conflicts occur
5. **Backup state files** - Keep copies of state files for rollback capability
6. **Use reference tracking** - Never remove `ninjops:<uuid>` tags

7. **Validate before syncing** - Run `ninjops validate` to check your QuoteSpec

8. **Review diffs** - Always review the output from `ninjops ninja diff`
9. **Test in staging** - Use a staging environment for Invoice Ninja integration testing
