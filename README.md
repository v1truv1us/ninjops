# Ninjops

A production-ready, test-driven Go CLI for agentic orchestration of Invoice Ninja v5 quote/invoice lifecycle.

## Features

- **Deterministic-first generation** - Templates + validation work without AI keys
- **Optional AI assistance** - Offline, OpenAI, Anthropic, and 67 OpenAI-compatible providers
- **Full Invoice Ninja v5 integration** - Safe GET → PATCH → PUT sync pattern
- **Reference tracking** - Deterministic identification via `ninjops:<uuid>` tags
- **Local state management** - Track sync history in `.ninjops/state.json`
- **Clean CLI UX** - Dry-run, diff, confirm prompts, JSON output mode

## Installation

### From Source

```bash
git clone https://github.com/v1truv1us/ninjops.git
cd ninjops
make install
```

### Homebrew

```bash
brew tap v1truv1us/tap
brew install v1truv1us/tap/ninjops
```

Notes:
- Use the fully-qualified formula name to avoid conflicts if another tap also ships `ninjops`.
- Upgrade with: `brew upgrade v1truv1us/tap/ninjops`

### Download Binary

Download from [Releases](https://github.com/v1truv1us/ninjops/releases)

### Docker

Requires a running Docker daemon.

```bash
git clone https://github.com/v1truv1us/ninjops.git
cd ninjops
docker build -t ninjops:latest .
docker run --rm -p 8080:8080 ninjops:latest
```

### Verify Installation

```bash
ninjops --version
```

## Quick Start

```bash
# Initialize project
ninjops init

# Configure API credentials in user config
ninjops configure --non-interactive \
  --base-url "https://your-instance.com" \
  --api-token "your-token" \
  --api-secret "your-secret" \
  --provider "offline"

# Edit quote_template.json with your project details

# Validate the spec
ninjops validate --input quote_template.json

# Generate proposal, terms, and notes
ninjops generate --input quote_template.json --format md --out-dir output/

# Interactive end-to-end quote workflow (select/create client/project/tasks)
ninjops new quote

# Convert an existing quote to invoice
ninjops convert <quote-id>

# (Optional) Test Invoice Ninja connection
ninjops ninja test

# (Optional) Sync with Invoice Ninja
ninjops ninja sync --input quote_template.json --dry-run
ninjops ninja sync --input quote_template.json
```

## Commands

### `ninjops init`
Initialize ninjops in the current directory. Creates `.ninjops/` with config and a sample QuoteSpec template.

### `ninjops new quote`
Generate a new QuoteSpec JSON template with a fresh reference ID.

`new quote` workflow behavior:
- Interactive (when stdin is a terminal and `--non-interactive` is not set):
  - Select or create client.
  - Select or create project for the selected client.
  - Select existing tasks (multi-select) and optionally add new tasks.
  - Prompt for missing project details.
  - Generate terms/public notes and optionally edit them inline in the CLI.
  - Show quote preview (client/project/tasks, line items, notes/terms snippets).
  - Optionally create quote in Invoice Ninja and optionally convert to invoice.
- Non-interactive (`--non-interactive`):
  - Never prompts.
  - Uses selectors/flags only and keeps template defaults for missing values.

Client selectors (choose one):
- `--client-id <id>`
- `--client-email <email>`
- `--client-name <name>`

Workflow selectors:
- `--project-id <id>`
- `--task-ids <id1,id2,...>`

Project flags:
- `--project-name <name>`
- `--project-description <description>`
- `--project-type <type>`
- `--project-timeline <timeline>`

Create/convert flags:
- `--create-quote`
- `--convert-to-invoice` (implies `--create-quote`)
- `--yes` (skip confirmations)

Output/artifacts flags:
- `--output <path>` writes QuoteSpec JSON to file.
- When `--output` is set, `new quote` also generates and writes:
  - `terms.md`
  - `public_notes.txt`
- Generated artifact files default to the output file directory, or set `--artifacts-dir <dir>` to override.

Examples:

```bash
# Interactive full workflow
ninjops new quote

# Non-interactive quote prep only
ninjops new quote --non-interactive --client-id <client-id> --project-id <project-id> --output quote.json

# Create quote + convert immediately
ninjops new quote --non-interactive --client-id <client-id> --project-id <project-id> --create-quote --convert-to-invoice --yes
```

### `ninjops new invoice`
Create an invoice by converting a quote, from a QuoteSpec file, or interactively.

Modes:
- `--from-quote <quote-id>`: converts that quote to an invoice
- `--input <quote-spec.json>`: creates invoice from QuoteSpec file
- Interactive (default): select client, project, tasks, and create invoice

Client selectors (choose one):
- `--client-id <id>`: select client by ID
- `--client-email <email>`: find client by email
- `--client-name <name>`: find client by name

Workflow selectors:
- `--project-id <id>`: pre-select project
- `--task-ids <id1,id2,...>`: pre-select tasks

Other flags:
- `--non-interactive`: disable prompts for automation
- `--yes`: skip confirmation prompt

Examples:

```bash
# Convert an existing quote
ninjops new invoice --from-quote <quote-id>

# Create invoice from QuoteSpec file
ninjops new invoice --input quote.json --yes

# Interactive invoice creation (default)
ninjops new invoice

# Non-interactive with selectors
ninjops new invoice --client-id <id> --project-id <id> --task-ids <ids> --non-interactive --yes
```

### `ninjops convert <quote-id>`
Convert a quote to an invoice with preview and confirmation.

```bash
ninjops convert <quote-id>
ninjops convert <quote-id> --yes
```

### `ninjops edit <entity> <id>`
Edit `public_notes` and/or `terms` on existing quotes/invoices from the CLI.

Supported entities:
- `quote`
- `invoice`

Field selection:
- `--field public_notes|terms|both` (default: `both`)

```bash
ninjops edit quote <quote-id> --field terms
ninjops edit invoice <invoice-id> --field both --yes
```

### `ninjops list <entity>`
List Invoice Ninja records for Phase-1 workflow discovery.

Supported entities:
- `clients`
- `projects`
- `tasks`
- `quotes`
- `invoices`

Common flags:
- `--client-id <id>` (supported on `projects`, `tasks`, `quotes`, `invoices`)
- `--project-id <id>` (supported on `tasks`)
- `--limit <n>` (default: `20`)
- `--format table|json|simple` (default: `table`)

Examples:

```bash
ninjops list clients
ninjops list projects --client-id 123 --format simple
ninjops list tasks --client-id 123 --project-id 456 --format json
ninjops list quotes --client-id 123 --limit 10
```

### `ninjops show <entity> <id>`
Show one Invoice Ninja entity as JSON (default output format).

Supported entities:
- `client`
- `project`
- `task`
- `quote`
- `invoice`

Examples:

```bash
ninjops show client 123
ninjops show project 456
ninjops show task 789
ninjops show quote 1011
ninjops show invoice 1213
```

### `ninjops configure`
Write non-secret onboarding config to `~/.config/ninjops/config.json` and secrets to `~/.config/ninjops/auth-creds.json` by default.

Provider onboarding behavior:
- For `offline`, configure does not require a provider API key and does not call provider APIs.
- For non-offline providers (`openai`, `anthropic`, and supported OpenAI-compatible providers), configure requires an API key from one of:
  - `--provider-api-key`
  - provider env var (`NINJOPS_OPENAI_API_KEY` / `NINJOPS_ANTHROPIC_API_KEY` / `NINJOPS_<PROVIDER_ID>_API_KEY` for supported OpenAI-compatible providers)
  - existing config value `agent.provider_api_key`
- Configure runs a provider/model connection check by default for non-offline providers.
- OpenAI and supported OpenAI-compatible providers use a two-step check: `/models` auth sanity + lightweight `/chat/completions` probe on the resolved model.
- Anthropic uses a two-step check: `/models` auth sanity + lightweight `/messages` probe on the resolved model.
- Use `--skip-provider-test` to skip the connectivity check (key is still required).

Model onboarding behavior:
- Configure supports `--model` and persists `agent.model` in main config.
- If `--model` is omitted, configure picks a provider default model from built-in mapping, then falls back to `gpt-5-codex`.
- Alias normalization: `openai-codex` resolves to provider `opencode` with model `gpt-5.3-codex`.

Common options:
- `--format json|jsonc` (default: `json`)
- `--output <path>` to override output location
- `--auth-creds-output <path>` to override auth credentials location
- `--non-interactive` for automation/tests
- `--base-url`, `--api-token`, `--api-secret`, `--provider`, `--plan`, `--model`, `--listen`, `--port`
- `--provider-api-key`, `--skip-provider-test`

### `ninjops validate --input file.json`
Validate a QuoteSpec JSON file against the schema.

### `ninjops generate --input file.json --format md|text|json --out-dir out/`
Generate proposal, terms, and notes from a QuoteSpec.

### `ninjops assist <role> --input file.json`
Get AI assistance with your QuoteSpec:
- `clarify` - Normalize features, extract responsibilities
- `polish` - Improve tone and readability
- `boundaries` - Generate scope boundaries
- `line-items` - Suggest billable line items

### `ninjops ninja test`
Test connection to Invoice Ninja API.

### `ninjops ninja pull --input file.json`
Pull existing quote/invoice from Invoice Ninja.

### `ninjops ninja sync --input file.json`
Sync QuoteSpec with Invoice Ninja (create or update).

Syncs the following entities:
- **Client**: Created if not found by email/name, or reused if matched
- **Project**: Created/updated when `QuoteSpec.Project.Name` is set
- **Quote/Invoice**: Created or updated based on sync mode

Flags:
- `--mode quote|invoice|both` (default: `quote`)
- `--dry-run`: preview changes without making them
- `--diff`: show field-level differences
- `--yes`: skip confirmation prompt

Examples:

```bash
# Sync quote only (default)
ninjops ninja sync --input quote.json

# Sync both quote and invoice
ninjops ninja sync --input quote.json --mode both

# Preview changes without making them
ninjops ninja sync --input quote.json --dry-run --diff
```

### `ninjops ninja diff --input file.json`
Compare local generated content with remote.

### `ninjops serve`
Start local HTTP API server.

## Configuration

### Config File

Recommended: use `ninjops configure` to generate one of:
- `~/.config/ninjops/config.json` (default)
- `~/.config/ninjops/ninjops.jsonc` (`--format jsonc`)

`configure` also writes `~/.config/ninjops/auth-creds.json` by default and stores only secrets there.

You can still use existing project/local config files such as `.ninjops/config.toml`.

Example `config.json`:

```json
{
  "ninja": {
    "base_url": "https://invoiceninja.fergify.work"
  },
  "agent": {
    "provider": "offline",
    "plan": "default",
    "model": "gpt-5-codex"
  },
  "serve": {
    "listen": "127.0.0.1",
    "port": 8080
  },
  "auth_creds_file": "/Users/you/.config/ninjops/auth-creds.json"
}
```

Example `auth-creds.json`:

```json
{
  "ninja": {
    "api_token": "",
    "api_secret": ""
  },
  "agent": {
    "provider_api_key": ""
  }
}
```

Legacy TOML format:

```toml
[ninja]
base_url = "https://invoiceninja.fergify.work"
api_token = ""
api_secret = ""

[agent]
provider = "offline"  # offline, openai, anthropic
plan = "default"       # default, codex-pro, opencode-zen, zai-plan
model = "gpt-5-codex"  # runtime model id (supports alias input in configure)
provider_api_key = ""  # used when provider-specific env var is not set

[serve]
listen = "127.0.0.1"
port = 8080
```

### Environment Variables

```bash
# Invoice Ninja
export NINJOPS_NINJA_BASE_URL="https://your-instance.com"
export NINJOPS_NINJA_API_TOKEN="your-token"
export NINJOPS_NINJA_API_SECRET="your-secret"  # optional

# AI Providers
export NINJOPS_AGENT_PROVIDER="offline"  # offline, openai, anthropic
export NINJOPS_AGENT_PLAN="default"      # default, codex-pro, opencode-zen, zai-plan
export NINJOPS_AGENT_MODEL="gpt-5-codex"
export NINJOPS_AGENT_PROVIDER_API_KEY="" # optional generic provider key override
export NINJOPS_OPENAI_API_KEY="sk-..."
export NINJOPS_ANTHROPIC_API_KEY="sk-ant-..."
export OPENCODE_API_KEY="..."           # preferred opencode auth env
export NINJOPS_OPENCODE_API_KEY="..."    # example openai-compatible provider env var
```

Credential precedence:
- `NINJOPS_NINJA_API_TOKEN` / `NINJOPS_NINJA_API_SECRET` / provider env vars (highest)
- values in main config (`config.json`, `ninjops.jsonc`, or legacy config files)
- values in `auth-creds.json` referenced by `auth_creds_file`

Provider API key precedence (assist + serve) by provider:
- provider-specific env var (`NINJOPS_OPENAI_API_KEY`, `NINJOPS_ANTHROPIC_API_KEY`, or `NINJOPS_<PROVIDER_ID>_API_KEY`)
- opencode special-case env order: `OPENCODE_API_KEY` then `NINJOPS_OPENCODE_API_KEY`
- `agent.provider_api_key` from resolved config (main config or auth-creds file)

Supported OpenAI-compatible provider IDs (current scope):
`302ai`, `abacus`, `aihubmix`, `alibaba`, `alibaba-cn`, `bailing`, `baseten`, `berget`, `chutes`, `cloudferro-sherlock`, `cloudflare-workers-ai`, `cortecs`, `deepseek`, `evroc`, `fastrouter`, `fireworks-ai`, `firmware`, `friendli`, `github-copilot`, `github-models`, `helicone`, `huggingface`, `iflowcn`, `inception`, `inference`, `io-net`, `jiekou`, `kilo`, `kuae-cloud-coding-plan`, `llama`, `lmstudio`, `lucidquery`, `meganova`, `moark`, `modelscope`, `moonshotai`, `moonshotai-cn`, `morph`, `nano-gpt`, `nebius`, `nova`, `novita-ai`, `nvidia`, `ollama-cloud`, `opencode`, `opencode-go`, `ovhcloud`, `poe`, `privatemode-ai`, `qihang-ai`, `qiniu-ai`, `requesty`, `scaleway`, `siliconflow`, `siliconflow-cn`, `stackit`, `stepfun`, `submodel`, `synthetic`, `upstage`, `vultr`, `wandb`, `xiaomi`, `zai`, `zai-coding-plan`, `zhipuai`, `zhipuai-coding-plan`.

## QuoteSpec Schema

```json
{
  "schema_version": "1.0.0",
  "task_ids": ["task-id-1"],
  "client": {
    "id": "client-id",
    "name": "Client Name",
    "email": "client@example.com",
    "org_type": "business"
  },
  "project": {
    "id": "project-id",
    "name": "Project Name",
    "description": "Project description",
    "type": "website",
    "timeline": "4 weeks"
  },
  "work": {
    "features": [
      {"name": "Feature 1", "description": "Description", "priority": "high"}
    ]
  },
  "pricing": {
    "currency": "USD",
    "line_items": [
      {"description": "Item", "quantity": 1, "rate": 1000, "amount": 1000}
    ],
    "total": 1000
  },
  "settings": {
    "tone": "professional"
  }
}
```

## Examples

See [examples/](./examples/) for sample QuoteSpec files:
- `church_quote.json` - Church website with discount
- `biz_quote.json` - Business brochure site
- `app_quote.json` - Full web application

## Documentation

- [PRD](./docs/PRD.md) - Product requirements
- [Technical Spec](./docs/TECH_SPEC.md) - Architecture details
- [API Contract](./docs/API_CONTRACT.md) - JSON schemas
- [Agents](./docs/AGENTS.md) - AI provider documentation
- [Invoice Ninja](./docs/INVOICENINJA.md) - Integration details
- [ADRs](./docs/decisions/) - Architecture decision records

## Development

```bash
# Build
make build

# Test
make test

# Lint
make lint

# Run all checks
make ci
```

### Releases

- Releases are cut from Git tags matching `v*` (example: `v0.1.0`).
- Tag pushes trigger the GitHub Actions release workflow, which runs GoReleaser and publishes:
  - GitHub release artifacts
  - Homebrew formula updates to `v1truv1us/homebrew-tap`
- Homebrew install is available from `v0.1.0+` via `v1truv1us/tap/ninjops`.
- Required repository secret for formula publishing: `HOMEBREW_TAP_GITHUB_TOKEN` (a token with push access to `v1truv1us/homebrew-tap`).

## Security

- **Never log secrets** - API tokens are redacted in all logs
- **Prefer secure local storage** - config files are written with owner-only permissions
- **Localhost by default** - Server binds to 127.0.0.1

See [SECURITY.md](./SECURITY.md) for full security policy.

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.
