---
title: Quick Start
description: Get up and running with Ninjops in minutes
---

# Quick Start

This guide will walk you through creating your first quote with Ninjops.

## Prerequisites

- Ninjops installed (see [Installation](/ninjops/getting-started/installation/))
- An Invoice Ninja v5 instance (optional for local-only workflows)

## Step 1: Initialize Your Project

Initialize Ninjops in your project directory:

```bash
ninjops init
```

This creates a `.ninjops/` directory containing:
- Configuration file
- Sample QuoteSpec template (`quote_template.json`)

## Step 2: Configure Credentials

Set up your Invoice Ninja API credentials:

```bash
ninjops configure --non-interactive \
  --base-url "https://your-instance.com" \
  --api-token "your-token" \
  --api-secret "your-secret" \
  --provider "offline"
```

**Note:** The `--provider "offline"` setting means you'll use templates without AI enhancement. You can change this later if you want AI features.

## Step 3: Create Your First Quote

### Interactive Mode (Recommended)

The easiest way to create a quote is using the interactive workflow:

```bash
ninjops new quote
```

This will guide you through:
1. Selecting or creating a client
2. Selecting or creating a project
3. Selecting tasks (or adding new ones)
4. Generating terms and public notes
5. Previewing the quote
6. Optionally creating the quote in Invoice Ninja

### Non-Interactive Mode

For automation or scripting, use non-interactive mode:

```bash
# Create a QuoteSpec JSON file
ninjops new quote --non-interactive \
  --client-id <client-id> \
  --project-id <project-id> \
  --output my-quote.json
```

## Step 4: Validate Your QuoteSpec

Before syncing, validate your QuoteSpec:

```bash
ninjops validate --input my-quote.json
```

This checks that your QuoteSpec conforms to the schema and catches common errors.

## Step 5: Generate Documents

Generate professional documents from your QuoteSpec:

```bash
ninjops generate --input my-quote.json --format md --out-dir output/
```

This creates:
- `proposal.md` - Full proposal document
- `terms.md` - Terms and conditions
- `public_notes.txt` - Public notes for the quote

## Step 6: Sync with Invoice Ninja

### Test First

Always test with a dry-run:

```bash
ninjops ninja sync --input my-quote.json --dry-run
```

This shows what would happen without making any changes.

### Sync for Real

When ready, sync to Invoice Ninja:

```bash
ninjops ninja sync --input my-quote.json
```

## Next Steps

Now that you've created your first quote, explore these topics:

- [Configuration](/ninjops/getting-started/configuration/) - Learn about all configuration options
- [Creating Quotes](/ninjops/guides/creating-quotes/) - Deep dive into quote creation
- [AI Assistance](/ninjops/guides/ai-assistance/) - Enhance your quotes with AI
- [Invoice Ninja Sync](/ninjops/guides/invoice-ninja-sync/) - Advanced sync workflows
