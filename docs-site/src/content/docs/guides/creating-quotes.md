---
title: Creating Quotes
description: Step-by-step guide to creating professional quotes with Ninjops
---

# Creating Quotes

This guide walks you through the complete quote creation workflow using Ninjops.

## Overview

Ninjops provides two main approaches to creating quotes:

1. **Interactive workflow** - Guided, step-by-step quote creation
2. **Non-interactive workflow** - Scriptable, automation-friendly quote creation

## Interactive Workflow

The interactive workflow is perfect for creating quotes manually with full control over the process.

### Basic Interactive Quote

```bash
ninjops new quote
```

This launches an interactive workflow that guides you through:

### Step 1: Client Selection

Choose from existing clients or create a new one:
- Select by client ID
- Select by client email
- Select by client name
- Or create a new client

### Step 2: Project Setup

Select or create a project for the selected client:
- Choose from existing projects
- Create a new project with name, description, type, and timeline

### Step 3: Task Selection

Select tasks for the quote:
- Multi-select existing tasks
- Add new tasks as needed
- Configure task details

### Step 4: Generate Documents

Ninjops automatically generates:
- Professional proposal text
- Terms and conditions
- Public notes

You can optionally edit these inline in the CLI.

### Step 5: Preview and Confirm

Review the complete quote:
- Client and project details
- Line items with pricing
- Generated terms and notes snippets

### Step 6: Create and Optionally Convert

Choose to:
- Create quote in Invoice Ninja
- Create and immediately convert to invoice
- Save as QuoteSpec JSON for later

## Non-Interactive Workflow

For automation or scripting, use non-interactive mode with selectors.

### Create QuoteSpec JSON

```bash
ninjops new quote --non-interactive \
  --client-id <client-id> \
  --project-id <project-id> \
  --output quote.json
```

### Create Quote in Invoice Ninja

```bash
ninjops new quote --non-interactive \
  --client-id <client-id> \
  --project-id <project-id> \
  --create-quote
```

### Create and Convert to Invoice

```bash
ninjops new quote --non-interactive \
  --client-id <client-id> \
  --project-id <project-id> \
  --create-quote \
  --convert-to-invoice \
  --yes  # Skip confirmations
```

## Client Selectors

Choose one method to identify the client:

```bash
--client-id <id>        # By Invoice Ninja ID
--client-email <email>  # By email address
--client-name <name>    # By client name
```

## Workflow Selectors

```bash
--project-id <id>           # Project ID
--task-ids <id1,id2,...>  # Comma-separated task IDs
```

## Project Flags

When creating a new project:

```bash
--project-name <name>
--project-description <description>
--project-type <type>          # e.g., "website", "webapp", "ecommerce"
--project-timeline <timeline>  # e.g., "2-3 weeks", "1 month"
```

## Output Flags

```bash
--output <path>              # Write QuoteSpec JSON to file
--artifacts-dir <dir>        # Override artifacts directory
```

When `--output` is set, Ninjops also generates:
- `terms.md`
- `public_notes.txt`

## Examples

### Quick Quote Creation

```bash
# Interactive, create quote and convert
ninjops new quote
```

### Automated Quote Prep

```bash
# Non-interactive, save for review
ninjops new quote --non-interactive \
  --client-id 123 \
  --project-id 456 \
  --output quotes/client-a-quote.json
```

### End-to-End Automation

```bash
# Create quote and immediately convert to invoice
ninjops new quote --non-interactive \
  --client-id 123 \
  --project-id 456 \
  --task-ids 789,790 \
  --create-quote \
  --convert-to-invoice \
  --yes
```

## Best Practices

1. **Start interactive** - Use interactive mode to learn the workflow
2. **Use `--output`** - Save QuoteSpec JSON for review before creating
3. **Preview with `--dry-run`** - Always preview changes before syncing
4. **Validate first** - Run `ninjops validate` on your QuoteSpec before syncing

## Next Steps

- [AI Assistance](/guides/ai-assistance/) - Enhance your quotes with AI
- [Invoice Ninja Sync](/guides/invoice-ninja-sync/) - Sync quotes to Invoice Ninja
- [Examples](/examples/church-website/) - See complete quote examples
