# Invoice Ninja Integration

## Overview

Ninjops integrates with Invoice Ninja v5 to sync quotes and invoices.

## Configuration

### Base URL
Default: `https://invoiceninja.fergify.work`

Override via:
```bash
# Environment variable
export NINJOPS_NINJA_BASE_URL=https://your-instance.example.com

# Config file
[ninja]
base_url = "https://your-instance.example.com"

# CLI flag
ninjops ninja test --base-url https://your-instance.example.com
```

### Authentication

Required headers for all API calls:
```
X-API-Token: <your-token>
X-Requested-With: XMLHttpRequest
```

Optional secret header:
```
X-API-Secret: <your-secret>
```

Configure via:
```bash
export NINJOPS_NINJA_API_TOKEN=your-token-here
export NINJOPS_NINJA_API_SECRET=your-secret-here  # optional
```

## API Endpoints Used

### Clients
- `GET /api/v1/clients` - List/search clients
- `GET /api/v1/clients/{id}` - Get client by ID
- `POST /api/v1/clients` - Create client
- `PUT /api/v1/clients/{id}` - Update client

### Quotes
- `GET /api/v1/quotes` - List/search quotes
- `GET /api/v1/quotes/{id}` - Get quote by ID
- `POST /api/v1/quotes` - Create quote
- `PUT /api/v1/quotes/{id}` - Update quote

### Invoices
- `GET /api/v1/invoices` - List/search invoices
- `GET /api/v1/invoices/{id}` - Get invoice by ID
- `POST /api/v1/invoices` - Create invoice
- `PUT /api/v1/invoices/{id}` - Update invoice

## Field Mapping

### QuoteSpec → Invoice Ninja

| QuoteSpec Field | Invoice Ninja Field |
|-----------------|---------------------|
| `metadata.reference` | `custom_value1` |
| Generated `public_notes_text` | `public_notes` |
| Generated `terms_markdown` | `terms` |
| Generated metadata | `private_notes` |
| `pricing.line_items` | `line_items` |
| `pricing.discount` | `discount`, `is_amount_discount` |

### Client Mapping

| QuoteSpec Field | Invoice Ninja Field |
|-----------------|---------------------|
| `client.name` | `name` |
| `client.email` | `email` (search), `contacts[].email` |
| `client.phone` | `phone` |
| `client.address.line1` | `address1` |
| `client.address.city` | `city` |
| `client.address.state` | `state` |
| `client.address.postal_code` | `postal_code` |

## Sync Logic

### Client Sync
1. Search by email (preferred) or name
2. If found: Update client with new info
3. If not found: Create new client
4. Return client ID

### Quote Sync
1. Check local state for quote ID
2. If ID exists: GET quote, verify reference matches
3. If no ID: Search by reference tag
4. If found: Compare fields, update if changed
5. If not found: Create new quote
6. Store quote ID in local state

### Invoice Sync
Same logic as quote sync, but for invoices.

### Update Strategy (GET → PATCH → PUT)

1. **GET** current entity state
   - Captures current field values
   - Ensures entity exists

2. **Build patch** with changed fields
   - Compare local generated vs remote
   - Identify changed fields only

3. **PUT** full payload
   - Submit complete entity
   - Ensures consistency
   - Handles concurrent updates

## Reference Tracking

### Format
```
ninjops:<uuid-v4>
```

Example: `ninjops:550e8400-e29b-41d4-a716-446655440000`

### Storage Locations
1. **QuoteSpec**: `metadata.reference`
2. **Invoice Ninja**: `custom_value1` on quote/invoice
3. **Local state**: `.ninjops/state.json`

### State File Structure
```json
{
  "version": "1.0.0",
  "entries": {
    "550e8400-e29b-41d4-a716-446655440000": {
      "reference_id": "550e8400-e29b-41d4-a716-446655440000",
      "client_id": "abc123",
      "quote_id": "quote456",
      "invoice_id": "invoice789",
      "last_sync_hash": "a1b2c3d4",
      "updated_at": "2024-01-15T10:30:00Z",
      "created_at": "2024-01-15T10:00:00Z"
    }
  }
}
```

## Search Strategies

### By ID (Fastest)
```bash
ninjops ninja pull --quote-id quote123
```

### By Reference
```bash
ninjops ninja pull --ref ninjops:550e8400-e29b-41d4-a716-446655440000
```

### From QuoteSpec
```bash
ninjops ninja pull --input quote.json
```

## Error Handling

### Common Errors

**401 Unauthorized**
- Check API token is set correctly
- Verify token has correct permissions

**404 Not Found**
- Entity doesn't exist
- Reference tag not found

**422 Validation Error**
- Invalid field values
- Missing required fields

### Error Redaction
All errors have tokens redacted:
```
Error: API call failed with token abc1**** 
```

## Pagination

For list operations, pagination is handled automatically:
- Default page size: 15
- Max page size: 100
- Iterates through all pages for search operations
