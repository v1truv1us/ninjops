# API Contract

## QuoteSpec Input/Output

### Input: QuoteSpec JSON

```json
{
  "schema_version": "1.0.0",
  "metadata": {
    "reference": "ninjops:550e8400-e29b-41d4-a716-446655440000",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  },
  "client": {
    "name": "Acme Corporation",
    "email": "contact@acme.com",
    "phone": "555-123-4567",
    "org_type": "business",
    "address": {
      "line1": "123 Business Ave",
      "city": "San Francisco",
      "state": "CA",
      "postal_code": "94105",
      "country": "US"
    }
  },
  "project": {
    "name": "E-commerce Platform",
    "description": "Modern e-commerce platform with inventory management",
    "type": "web application",
    "timeline": "8-10 weeks",
    "deadline": "2024-04-01",
    "technologies": ["React", "Node.js", "PostgreSQL", "Stripe"]
  },
  "work": {
    "features": [
      {
        "name": "Product Catalog",
        "description": "Full product catalog with categories, variants, and search",
        "priority": "high",
        "category": "Development"
      },
      {
        "name": "Shopping Cart",
        "description": "Persistent shopping cart with saved items",
        "priority": "high",
        "category": "Development"
      },
      {
        "name": "Checkout Flow",
        "description": "Multi-step checkout with Stripe integration",
        "priority": "high",
        "category": "Development"
      }
    ],
    "responsibilities": [
      "Full implementation of all features",
      "Responsive design for all screen sizes",
      "Integration with Stripe for payments",
      "Basic SEO optimization"
    ],
    "minor_changes": [
      "Minor text changes",
      "Color adjustments",
      "Small layout tweaks"
    ],
    "out_of_scope": [
      "Mobile app development",
      "Content writing",
      "Marketing services"
    ],
    "assumptions": [
      "Client will provide product data",
      "Client will provide feedback within 3 business days"
    ]
  },
  "pricing": {
    "total": 15000,
    "currency": "USD",
    "line_items": [
      {
        "description": "Product Catalog Development",
        "quantity": 1,
        "rate": 4000,
        "amount": 4000,
        "category": "Development"
      },
      {
        "description": "Shopping Cart & Checkout",
        "quantity": 1,
        "rate": 5000,
        "amount": 5000,
        "category": "Development"
      },
      {
        "description": "Admin Dashboard",
        "quantity": 1,
        "rate": 3000,
        "amount": 3000,
        "category": "Development"
      },
      {
        "description": "Testing & QA",
        "quantity": 1,
        "rate": 2000,
        "amount": 2000,
        "category": "Quality Assurance"
      },
      {
        "description": "Project Management",
        "quantity": 1,
        "rate": 1000,
        "amount": 1000,
        "category": "Management"
      }
    ],
    "deposit": {
      "percentage": 30,
      "due_date": "Upon acceptance"
    },
    "payment_terms": "Net 15",
    "recurring": {
      "type": "hosting",
      "amount": 100,
      "frequency": "monthly",
      "description": "Managed hosting and maintenance"
    }
  },
  "settings": {
    "tone": "professional",
    "include_timeline": true,
    "include_pricing": true
  }
}
```

### Output: GeneratedArtifacts

```json
{
  "proposal_markdown": "# Project Name\n\n## Overview\n...",
  "terms_markdown": "# Terms & Conditions\n\n## Agreement\n...",
  "public_notes_text": "Project: Project Name\nClient: Client Name\n...",
  "meta": {
    "generated_at": "2024-01-15T10:35:00Z",
    "template_version": "1.0.0",
    "hash": "a1b2c3d4e5f6g7h8"
  }
}
```

## HTTP API Endpoints

### POST /generate

Generate artifacts from QuoteSpec.

**Request:**
```json
{
  "schema_version": "1.0.0",
  "client": { ... },
  "project": { ... },
  ...
}
```

**Response:**
```json
{
  "proposal_markdown": "...",
  "terms_markdown": "...",
  "public_notes_text": "...",
  "meta": { ... }
}
```

### POST /assist/{role}

Get AI assistance for a QuoteSpec.

**Path Parameters:**
- `role`: clarify | polish | boundaries | line-items

**Request:**
```json
{
  "quote_spec": { ... },
  "provider": "offline",
  "plan": "default"
}
```

**Response:**
```json
{
  "quote_spec": { /* Updated QuoteSpec */ },
  "suggestions": ["Suggestion 1", "Suggestion 2"],
  "confidence": 0.85,
  "metadata": { ... }
}
```

### POST /ninja/sync

Sync QuoteSpec with Invoice Ninja.

**Request:**
```json
{
  "quote_spec": { ... },
  "mode": "quote",
  "dry_run": false
}
```

**Response:**
```json
{
  "client_id": "abc123",
  "client_created": true,
  "quote_id": "quote456",
  "quote_created": true,
  "quote_updated": false,
  "diffs": [
    {
      "field": "public_notes",
      "old_value": "...",
      "new_value": "...",
      "type": "~"
    }
  ]
}
```

### GET /health

Health check endpoint.

**Response:**
```json
{
  "status": "ok"
}
```

## Versioning

- Schema version in QuoteSpec: `schema_version` field
- API versioning: URL path (future: `/api/v2/...`)
- Backward compatibility: New fields optional, old fields required

## Error Responses

### Validation Error (422)
```json
{
  "error": "validation_failed",
  "details": [
    {"field": "client.email", "message": "invalid email format"},
    {"field": "project.name", "message": "project name is required"}
  ]
}
```

### API Error (500/502/503)
```json
{
  "error": "api_error",
  "message": "Failed to connect to Invoice Ninja",
  "details": "Connection refused"
}
```

### Not Found (404)
```json
{
  "error": "not_found",
  "message": "Quote not found"
}
```
