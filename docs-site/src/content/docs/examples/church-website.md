---
title: Church Website Example
description: Example quote workflow for a church website project
---

# Church Website Example

This example shows a practical QuoteSpec for a church website build with clear phases and deliverables.

## Typical Scope

- Discovery and sitemap planning
- Custom homepage and sermon archive templates
- Contact and prayer request forms
- Content migration and launch support

## Suggested Workflow

1. Start with `ninjops new quote` to scaffold the quote.
2. Refine line items and terms in your QuoteSpec JSON.
3. Run `ninjops validate --input quote.json`.
4. Generate artifacts with `ninjops generate --input quote.json --out-dir dist`.

This flow keeps output predictable and ready for Invoice Ninja sync when approved.
