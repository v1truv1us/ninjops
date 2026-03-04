---
title: API Contract
description: HTTP API request and response conventions for Ninjops
---

# API Contract

Ninjops exposes a lightweight HTTP API for local automation and integrations.

## Conventions

- JSON request/response bodies
- Standard HTTP status codes (`2xx` success, `4xx` client error, `5xx` server error)
- Validation errors return machine-readable details

## Common Endpoints

- `GET /health` - service status
- `POST /validate` - validate QuoteSpec payloads
- `POST /generate` - generate quote artifacts from valid input

## Error Shape

```json
{
  "error": {
    "code": "validation_error",
    "message": "quote_spec.items is required"
  }
}
```

Prefer validating input before sync operations to keep API interactions deterministic and recoverable.
