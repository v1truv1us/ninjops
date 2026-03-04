---
title: Introduction
description: Learn what Ninjops is and how it can streamline your quote and invoice workflow
---

# Ninjops

Ninjops is a production-ready, test-driven Go CLI for agentic orchestration of Invoice Ninja v5 quote/invoice lifecycle management.

## Who is this for?

Ninjops is designed for:
- **Freelancers** who want to create professional quotes quickly
- **Small agencies** that need consistent quote/invoice workflows
- **Consultants** who manage multiple clients and projects
- **Anyone** using Invoice Ninja v5 who wants to streamline their billing process

## Key Features

### Deterministic-first Generation
Templates and validation work without AI keys. You can create professional quotes using just templates and structured data.

### Optional AI Assistance
Enhance your quotes with AI when you want to. Supports:
- Offline (rule-based, no API keys needed)
- OpenAI
- Anthropic
- 67+ OpenAI-compatible providers

### Full Invoice Ninja v5 Integration
Safe synchronization with a GET → PATCH → PUT pattern that prevents data loss.

### Reference Tracking
Deterministic identification via `ninjops:<uuid>` tags ensures reliable tracking of your quotes and invoices.

### Local State Management
Track sync history in `.ninjops/state.json` - everything stays under your control.

### Clean CLI UX
Features like dry-run, diff, confirm prompts, and JSON output mode make the tool pleasant to use.

## Philosophy

Ninjops follows a **deterministic-first** approach:
- Works perfectly without any AI keys
- AI enhancement is optional and additive
- You maintain full control over your data
- Everything can be validated and previewed before sync

## Next Steps

Ready to get started? Check out the [Installation guide](/getting-started/installation/) or jump straight to the [Quick Start](/getting-started/quick-start/).
