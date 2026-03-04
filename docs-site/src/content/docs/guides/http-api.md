---
title: HTTP API
description: Use Ninjops HTTP API for automation and integration
---

# HTTP API

Ninjops includes a local HTTP API server for automation and integration.

## Starting the Server

```bash
ninjops serve
```

The server runs on `127.0.0.1` by default and accepts JSON requests at `/assist`, and.

## API Endpoints

### POST /assist
Use the AI assistance endpoints:
- `POST /assist/clarify` - Clarify QuoteSpec
- `POST /assist/polish` - Polish QuoteSpec
- `POST /assist/boundaries` - Add scope boundaries
- `POST /assist/line-items` - Suggest line items

- `POST /generate` - Generate documents
- `POST /validate` - Validate QuoteSpec
- `POST /ninja/sync` - Sync to Invoice Ninja

- `POST /ninja/pull` - Pull from Invoice Ninja
- `POST /convert` - Convert quote to invoice

### Quote/Invoice Management
- `GET /quotes` - List quotes
- `GET /quotes/{id}` - Get specific quote
- `POST /quotes` - Create quote
- `PUT /quotes/{id}` - Update quote
- `DELETE /quotes/{id}` - Delete quote

- `GET /quotes/{id}?include=client,tasks` - Get quote with relations
### Project Management
- `GET /projects` - List projects
- `GET /projects/{id}` - Get specific project
- `POST /projects` - Create project
- `PUT /projects/{id}` - Update project
### Task Management
- `GET /tasks` - List tasks
- `GET /tasks/{id}` - Get specific task
- `POST /tasks` - Create task
    - `PUT /tasks/{id}` - Update task
- `DELETE /tasks/{id}` - Delete task
### Invoice Management
- `GET /invoices` - List invoices
- `GET /invoices/{id}` - Get specific invoice
- `POST /invoices` - Create invoice
    - `PUT /invoices/{id}` - Update invoice
- `DELETE /invoices/{id}` - Delete invoice
- `GET /invoices/{id}?include=client,tasks` - Get invoice with relations

## Request Format

All requests use JSON content type.

### Example: Create quote with assist
```bash
curl -X POST http://localhost:8080/assist/clarify \
  -H "Content-Type: application/json" \
  -d @request.body: { ... }' \
  ...
```

**Response:** QuoteSpec with enhancements
```json
{
  "schema_version": "1.0.0",
  "metadata": { ... },
  "client": { ... },
  "project": { ... },
  "work": {
    "features": [...],
    "responsibilities": [...],
    "assumptions": [...]
  },
  "pricing": {
    "total": 3500,
    "currency": "USD",
    "line_items": [...]
  },
  "settings": { ... }
}
```

**Error Response:**
```json
{
  "error": "Error message",
  "code": "ERROR_CODE"
}
```

## Configuration

### Environment Variables
```bash
export NINJOPS_SERVE_LISTEN=127.0.0.1  # default
export NINJOPS_SERVE_PORT=8080  # default
```

### Command-line Flags
```bash
ninjops serve --listen 0.0.0.0 --port 3000
```

## Example Usage

### Start server
```bash
ninjops serve
```

Server runs at `http://127.0.0.1:8080`

### Assist via API
```bash
curl -X POST http://localhost:8080/assist/clarify \
  -H "Content-Type: application/json" \
  -d @request.body: { ... }'
```

### Generate documents
```bash
curl -X POST http://localhost:8080/generate \
  -H "Content-Type: application/json" \
  -d @request.body: { ... }'
```

## Best Practices
1. **Use in development** - Run locally during development
2. **Secure in production** - Use HTTPS and proper authentication
3. **Monitor resources** - Set appropriate resource limits
4. **Use health checks** - Implement health check endpoints

## Next steps
- [AI Assistance](/guides/ai-assistance/) - Enhance QuoteSpecs with AI
- [Invoice Ninja Sync](/guides/invoice-ninja-sync/) - Sync with Invoice Ninja
