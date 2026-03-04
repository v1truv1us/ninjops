---
title: QuoteSpec Schema
description: Field-level guide for QuoteSpec files used by Ninjops
---

# QuoteSpec Schema

`QuoteSpec` is the canonical input format for generating and syncing quotes.

## Core Sections

- `client`: Who the quote is for (`name`, optional `id`, contact fields)
- `project`: Optional project metadata (`name`, optional `id`, summary)
- `items`: Line items used to calculate totals
- `terms`: Payment terms and legal language
- `public_notes`: Client-visible notes shown on generated documents

## Minimal Example

```json
{
  "client": { "name": "Grace Community Church" },
  "project": { "name": "Website Redesign" },
  "items": [
    { "name": "Design and Build", "quantity": 1, "cost": 3500 }
  ],
  "terms": "Net 15",
  "public_notes": "Thank you for the opportunity to work together."
}
```

Use `ninjops validate --input <file>` to verify schema correctness before generation or sync.
