# Ninjops Launch Posts

This document contains ready-to-post launch copy for X, LinkedIn, and Dev.to.

## Handle Verification

- Invoice Ninja X handle: `@invoiceninja`
- Verified from Invoice Ninja site metadata (`twitter:site` and `twitter:creator`) and schema links.

---

## X Thread (2 Posts)

### Post 1

```text
Just shipped alpha for ninjops 🚀

A CLI-first @invoiceninja workflow in Go:
client -> project -> tasks -> quote -> invoice

- Interactive + non-interactive modes
- Project-aware sync (create/update)
- AI-assisted terms/notes/proposal text

Looking for early testers: <repo-url>
#InvoiceNinja #golang #cli #opensource
```

### Post 2

```text
Quick commands 👇

ninjops new quote
ninjops new invoice
ninjops ninja sync --input quote.json --mode both --dry-run
ninjops generate --input examples/app_quote.json --format text --out-dir out/

Built for @invoiceninja teams that want a terminal-first workflow.
Feedback welcome 🙌
```

---

## LinkedIn Post

```text
We’ve reached alpha for ninjops, a CLI-first workflow for Invoice Ninja.

What it does:
• End-to-end client/project/task/quote/invoice flow
• Interactive and automation-friendly modes
• Project-aware sync support
• AI-assisted proposal/terms/public-notes generation

Text generation is working and the full test suite is passing.

Now looking for early users to test real-world Invoice Ninja workflows and share feedback.

Repo: <repo-url>
```

Note: On LinkedIn, tag the official company page via autocomplete (`Invoice Ninja`) when posting.

---

## Dev.to Draft

### Title

```text
Building a CLI-First Invoice Ninja Workflow with Go (Alpha Release)
```

### Article

```markdown
I just released **alpha** for `ninjops`, a Go CLI for managing the full Invoice Ninja lifecycle from the terminal.

If your workflow lives in shell, editors, and Git, constantly jumping to a web UI to manage clients, projects, quotes, and invoices creates friction. I wanted a tool that keeps everything in one place and stays automation-friendly.

## What `ninjops` does

At a high level, it supports:

- Client/project/task selection or creation
- Quote generation and creation
- Quote-to-invoice conversion
- Direct invoice creation
- Syncing local QuoteSpec changes back to Invoice Ninja
- AI-assisted content generation (proposal text, terms, public notes)

The core flow is:

`client -> project -> tasks -> quote -> invoice`

## Why CLI-first?

A CLI gives us a few important advantages:

- Fast, keyboard-only workflows
- Repeatable automation in CI/scripts
- Easy diffability for structured quote specs
- Less context switching between systems

It also makes it easier to build deterministic behavior and tests around critical invoice/quote operations.

## Alpha feature set

Current alpha includes:

- **Interactive mode**: guided selection/creation of client, project, and tasks
- **Non-interactive mode**: flags for automation (`--client-id`, `--project-id`, `--task-ids`, etc.)
- **`new invoice` parity** with quote workflow
- **`ninja sync` project support** (create/update project during sync)
- **Text artifact generation** for notes/terms/proposal content
- **Secret separation**: credentials in a dedicated auth credentials file (not mixed into main config)
- **Git safety**: `.gitignore` patterns for env/credential files

## Quick usage

```bash
# Configure
ninjops configure

# Interactive quote workflow
ninjops new quote

# Interactive invoice workflow
ninjops new invoice

# Sync with Invoice Ninja
ninjops ninja sync --input quote.json --mode both

# Generate text artifacts
ninjops generate --input examples/app_quote.json --format text --out-dir out/
```

## Engineering notes

A few implementation details that mattered:

- Added relationship validation so project/client/task IDs stay consistent before API calls.
- Fixed pagination handling to avoid silently truncating larger lists.
- Improved prompt input behavior by sharing buffered input correctly across interactive steps.
- Added tests around project sync and invoice workflow helpers.

## Security and config posture

One design goal was reducing accidental secret leakage:

- Main app config and auth secrets are separated.
- Credential-like files are ignored in Git.
- Local state/config artifacts are excluded from commits by default.

If you’re building internal tooling, this separation is simple but high-impact.

## What “alpha” means here

I consider this **alpha-ready**, meaning:

- Core workflows are implemented and functional
- Tests pass for current behavior
- Real-world usage feedback is now the most valuable next input

What’s next before beta:

- More end-to-end tests with real-world invoice scenarios
- Additional UX polish and error messaging
- Better onboarding defaults/docs for first-time users

## Looking for feedback

If you work with Invoice Ninja and prefer terminal-first workflows, I’d love your feedback on:

- Missing commands/flags
- Pain points in interactive mode
- Sync edge cases
- Output format improvements

Repo: `<repo-url>`
```

---

## Pre-Post Checklist

- Replace `<repo-url>` in all posts.
- Add `@invoiceninja` to X posts.
- Tag the `Invoice Ninja` LinkedIn company page in LinkedIn post.
