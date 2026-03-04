# Task Plan: Fix review findings

- [x] Fix `serve` flag defaults so config is not dereferenced before initialization.
- [x] Implement `new quote --output` file writing behavior with proper errors.
- [x] Add nil `quote_spec` guards in HTTP handlers and defensive validation guard.
- [x] Validate Invoice Ninja action responses and return errors on non-2xx.
- [x] Handle `store.NewStore` errors in app commands.
- [x] Resolve `--ref` behavior in `ninja pull` and `ninja diff`.
- [x] Restore reproducible dependencies by generating `go.sum`.
- [x] Run and verify test suite.

## Results

- `serve` now uses safe config defaults at command construction and has fallback config access in handlers.
- `new quote --output` now creates parent directories and writes output files with wrapped errors.
- HTTP handlers now reject missing `quote_spec` early; spec validation safely returns a descriptive nil error.
- Invoice Ninja action endpoints now fail on non-2xx via shared response-status checks.
- `store.NewStore` initialization errors are now returned in `serve`, `ninja sync`, `ninja pull`, and `ninja diff`.
- `--ref` now works in `ninja pull` and `ninja diff` (including reference override for input-driven lookups), with consistent conflicting-ID validation.
- Added/updated targeted tests for nil validation, handler guard behavior, command validation, and output file writing.
- Full `go test ./...` suite now passes after moving `newNinjaTestCmd` into a non-`*_test.go` app file.

## Task Plan: CLI smoke validation

- [x] Add an automated smoke test that exercises the built CLI binary.
- [x] Validate core command flow (`--help`, `new quote`, `validate`, `generate`) using isolated config.
- [x] Add a dedicated Makefile target to run smoke tests explicitly.
- [x] Run smoke and full test suites to confirm stability.

### Results

- Added `internal/app/cli_smoke_test.go` to build and execute the real CLI binary against an isolated temp workspace.
- Smoke flow now validates `--help`, `ninja --help`, `new quote --output`, `validate --input`, and `generate --input --out-dir`.
- Added `test-smoke`/`smoke` Makefile targets for direct smoke test execution.
- Verified both smoke target and full `go test ./...` pass.

## Task Plan: Configure onboarding workflow

- [x] Add a `ninjops configure` command for onboarding.
- [x] Support saving Invoice Ninja credentials to config file or `.env` file.
- [x] Wire command into root CLI and keep defaults aligned with current config behavior.
- [x] Add tests for non-interactive configure flow and env/config outputs.
- [x] Update docs/help text for onboarding command usage.
- [x] Run smoke + full test suite and verify passing.

### Results

- Added top-level `configure` command with non-interactive flags and default writes to `~/.config/ninjops/config.json` (or `ninjops.jsonc` with `--format jsonc`).
- Config loading now supports `~/.config/ninjops/config.json` and `~/.config/ninjops/ninjops.jsonc` (with `//` and `/* */` comment stripping), while preserving legacy config discovery.
- Added tests for configure output writing and config loader JSON/JSONC paths, and expanded CLI smoke test to validate `configure` execution.
- Updated README quick-start, command docs, and configuration guidance for onboarding via `ninjops configure`.

## Task Plan: Provider login/connection in configure

- [x] Extend `configure` to set up preferred coding provider credentials.
- [x] Add provider connection validation during configure onboarding.
- [x] Persist provider credential in config file and keep env override precedence.
- [x] Add tests for provider credential loading and configure behavior.
- [x] Update docs for provider setup/login/connection workflow.
- [x] Run smoke + full test suite and verify passing.

### Results

- Added `--provider-api-key` and `--skip-provider-test` to `ninjops configure`, with required-key enforcement for non-offline providers.
- Added extensible provider connectivity checks (`offline`, `openai`, `anthropic`) with 10s timeout and clear auth/network failure errors.
- Persisted `agent.provider_api_key` in config and updated runtime key resolution to use provider env var first, then config.
- Added tests covering non-offline key requirement, skip-provider-test path, and env-over-config provider key precedence.
- Verified deterministic smoke path remains offline-based and full test suite passes.

## Task Plan: Auth credentials file integration

- [x] Store secrets in `auth-creds.json` during `ninjops configure`.
- [x] Make primary config reference credentials file and keep non-secret settings separate.
- [x] Load credentials from referenced `auth-creds.json` in config loader with env override precedence.
- [x] Update runtime key/token consumers to use loaded credentials without behavior regressions.
- [x] Add/adjust tests for configure write layout and credential loading.
- [x] Update docs and lessons for credentials-file workflow.
- [x] Run smoke + full test suite and verify passing.

### Results

- `ninjops configure` now writes non-secret settings to main config and writes `ninja.api_token`, `ninja.api_secret`, and `agent.provider_api_key` to `auth-creds.json`.
- Main config now records `auth_creds_file`, defaulting to `~/.config/ninjops/auth-creds.json` (or `auth-creds.json` beside `--output`), with optional `--auth-creds-output` override.
- Config loader now hydrates secrets from referenced auth creds with precedence `env > main config > auth-creds`, preserving backward compatibility with existing main-config secrets.
- Updated configure/config/smoke tests and README docs for the split-config credentials workflow.

## Task Plan: Native provider architecture planning

- [x] Research models.dev provider catalog and adapter metadata.
- [x] Analyze opencode provider adapter architecture and auth/config loading patterns.
- [x] Define target ninjops provider subsystem architecture for scale.
- [x] Draft phased implementation roadmap with acceptance criteria.
- [x] Define adapter mapping strategy to cover all models.dev providers.
- [x] Document provider tiers (MVP, next, long-tail) and testing/CI gates.

### Results

- Confirmed models.dev currently exposes 97 providers and 23 unique adapter npm packages; 67 providers are `@ai-sdk/openai-compatible` and can share one generic adapter family.
- Identified opencode reference patterns to reuse: models.dev-driven registry, auth file split, provider-specific loader hooks, and adapter fallback strategy.
- Produced a right-sized ninjops target architecture (`internal/providers/*`) that decouples provider registry, auth resolution, connectivity probing, and protocol adapters.
- Produced a phased plan (foundation -> tier-1 MVP -> tier-2 -> long-tail -> hardening) with explicit acceptance criteria and CI coverage gates.

## Task Plan: OpenAI-compatible provider rollout

- [x] Expand provider support to the 67 models.dev OpenAI-compatible provider IDs.
- [x] Add a generic OpenAI-compatible runtime adapter and wire provider base URLs.
- [x] Make `configure` support explicit model selection (`--model`) and persist `agent.model`.
- [x] Normalize `openai-codex` model alias and ensure model is used in assist/serve execution.
- [x] Update provider validation, connection checks, and env key resolution for new providers.
- [x] Add/adjust tests and docs for provider/model configuration behavior.
- [x] Run smoke + full test suite and verify passing.

### Results

- Added config-backed provider registry for all 67 current models.dev `@ai-sdk/openai-compatible` provider IDs with base URL mapping and provider defaults.
- Reworked OpenAI transport into a generic OpenAI-compatible adapter and registered all supported OpenAI-compatible providers against it.
- Added `agent.model` to config/defaults/env, added `ninjops configure --model`, and persisted resolved model IDs into main config.
- Added model alias normalization for `openai-codex` (`gpt-5-codex` by default, `gpt-5.3-codex` for `opencode`) and applied model resolution in configure/load/runtime.
- Updated provider API key env resolution to support deterministic provider-specific env vars (`NINJOPS_<PROVIDER_ID>_API_KEY`) while preserving existing OpenAI/Anthropic env names.
- Extended connectivity checks for OpenAI-compatible providers using provider base URLs.
- Added tests for configure model persistence, alias normalization, openai-compatible provider validation, and runtime use of configured model.

## Task Plan: OpenAI-codex alias alignment with opencode

- [x] Make `openai-codex` resolve to opencode auth/model behavior.
- [x] Align provider key env precedence for opencode (`OPENCODE_API_KEY` support).
- [x] Ensure configure persists resolved provider/model when alias is used.
- [x] Ensure assist/serve runtime use opencode provider+model when alias is selected.
- [x] Update tests and docs for alias/provider resolution behavior.
- [x] Run smoke + full test suite and verify passing.

### Results

- Added shared provider+model resolver so `openai-codex` always canonicalizes to provider `opencode` with model `gpt-5.3-codex` during config load, configure, and runtime assist/serve paths.
- Updated provider key env resolution to check `OPENCODE_API_KEY` first, then `NINJOPS_OPENCODE_API_KEY`, then configured key fallback.
- Updated configure alias persistence tests, config normalization tests, runtime serve alias routing test, and README alias/auth precedence documentation.
- Verified `make smoke` and `go test ./...` both pass.

## Task Plan: Configure connectivity hardening

- [x] Tighten provider connectivity check to validate opencode chat-completions auth (not only `/models`).
- [x] Pass resolved provider/model from `configure` into connectivity validation.
- [x] Add focused tests for stricter opencode check and error behavior.
- [x] Update docs/help text if check semantics changed.
- [x] Run smoke + full test suite and verify passing.

### Results

- Added model-aware connectivity API (`CheckProviderConnectionWithModel`) while preserving `CheckProviderConnection` compatibility.
- Hardened `opencode` connectivity checks to require both `/models` authentication and a minimal `/chat/completions` model probe.
- Updated `configure` to validate with resolved provider+model so onboarding catches chat-auth mismatches early.
- Added focused `httptest` coverage for opencode success and auth-failure probe paths.
- Updated README provider onboarding notes and revalidated with `make smoke` and `go test ./...`.

## Task Plan: Provider/model execution validation

- [x] Expand configure connectivity validation to execute lightweight model calls for provider/model combinations (not only opencode).
- [x] Keep provider-specific behavior safe while ensuring selected model is actually callable.
- [x] Add focused tests proving model-call probes run and fail clearly on auth/model access issues.
- [x] Update docs and lessons for provider/model execution validation semantics.
- [x] Run smoke + full test suite and verify passing.

### Results

- Connectivity checks now execute provider-specific model probes for all non-offline providers after the initial `/models` auth sanity check.
- OpenAI/OpenAI-compatible providers run a minimal `/chat/completions` probe using the resolved model; Anthropic runs a minimal `/messages` probe.
- Compatibility path (`CheckProviderConnection`) now still probes execution by selecting each provider's default model when no model is supplied.
- Added deterministic `httptest` coverage for OpenAI-compatible and Anthropic success/failure probe paths, including model-access failure clarity.
- Revalidated with `make smoke` and `go test ./...`.

## Task Plan: New quote client/project workflow

- [x] Extend `ninjops new quote` to support selecting an Invoice Ninja client.
- [x] Capture project details during quote creation with interactive prompts and non-interactive flags.
- [x] Generate public notes and terms from the created QuoteSpec in the same workflow.
- [x] Add tests for client selection/project capture/artifact generation behavior.
- [x] Update docs for the new `ninjops new quote` onboarding flow.
- [x] Run smoke + full test suite and verify passing.

### Results

- Added new `new quote` flags for client selectors, project fields, `--non-interactive`, and `--artifacts-dir`.
- `new quote --output` now writes QuoteSpec JSON plus generated `terms.md` and `public_notes.txt` artifacts.
- Added deterministic tests for non-interactive project/artifact behavior and client selector resolution + interactive list selection helper.
- Updated README command docs for the expanded `new quote` workflow and flags.

## Task Plan: Full Invoice Ninja Workflow Implementation

**Reference:** `tasks/invoice-ninja-workflow-plan.md` - Complete workflow design and implementation plan

### Phase 1: Project & Task Support (Foundation)

- [x] Add Invoice Ninja project API wrapper (`internal/invoiceninja/projects.go`)
  - ListProjects, GetProject, CreateProject, UpdateProject
  - Project model in models.go
  - Tests for project operations
- [x] Add Invoice Ninja task API wrapper (`internal/invoiceninja/tasks.go`)
  - ListTasks, GetTask, CreateTask, UpdateTask
  - Task model in models.go
  - Tests for task operations
- [x] Implement `ninjops list` command (`internal/app/list.go`)
  - Support: clients, projects, tasks, quotes, invoices
  - Filtering by client_id, project_id, status
  - Output formats: table, json, simple
  - Tests for list command
- [x] Implement `ninjops show` command (`internal/app/show.go`)
  - Show entity details
  - Tests for show command
- [x] Enhance `new quote` with project selection
  - After client selection, fetch and display projects
  - Allow selection or creation of new project
  - Map selected project to QuoteSpec
  - Tests for project selection flow
- [ ] Enhance `ninja sync` with project support
  - Create/update project when provided
  - Link quote to project_id
  - Tests for project sync
- [x] Update README with list/show commands and project selection

### Phase 2: Inline Editing & Preview

- [x] Implement inline editing helper (`internal/app/editor.go`)
  - Support $EDITOR
  - Fallback to prompts
  - Tests for editing
- [x] Implement preview formatter (`internal/app/preview.go`)
  - Format quotes/invoices for preview
  - Tests for preview
- [x] Enhance `new quote` with editing step
  - Add terms/notes editing before creation
  - Add --skip-edit flag
  - Add preview before confirmation
  - Tests for editing flow
- [x] Update README with editing workflow

### Phase 3: Invoice Workflow

- [x] Implement `ninjops new invoice` command
  - Similar to new quote but creates invoice directly
  - Support --from-quote flag
  - Tests for invoice creation
- [x] Implement `ninjops convert` command
  - Convert quote to invoice
  - Support --edit flag
  - Tests for conversion
- [x] Implement `ninjops edit` command
  - Edit existing quotes
  - Edit existing invoices
  - Tests for edit command
- [x] Update README with invoice workflows

### Phase 4: Task Management

- [x] Enhance `new quote` with task selection
  - After project selection, show tasks
  - Allow selection or creation
  - Link tasks to quote
  - Tests for task selection
- [ ] Enhance `new invoice` with task selection
  - Same as quote but for invoices
  - Tests for task selection in invoices
- [x] Update README with task workflow

### Phase 5: Polish & Documentation

- [ ] Improve error messages with suggestions
- [ ] Add progress indicators
- [ ] Add color-coded output
- [ ] Complete README rewrite
- [ ] Create examples documentation
- [ ] Add integration tests for full workflows
- [ ] Enhance smoke tests
- [ ] Update all documentation

### Results (Current Execution)

- Delivered full quote workflow path in CLI: select/create client, select/create project, select existing tasks or add new tasks, edit terms/public notes inline, preview output, create quote, and optional quote->invoice conversion.
- Added lifecycle commands for existing records: `ninjops convert`, `ninjops edit`, and `ninjops new invoice` (`--from-quote` or `--input`).
- Added discovery commands: `ninjops list` and `ninjops show` for clients/projects/tasks/quotes/invoices.
- Expanded Invoice Ninja API wrappers and models for projects/tasks and `project_id` linkage in quote/invoice payloads.
- Extended QuoteSpec with optional linkage IDs (`client.id`, `project.id`, `task_ids`) while keeping backward compatibility.
- Validated with `make smoke` and `go test ./...` (passing), plus live command checks against configured Invoice Ninja instance.

## Task Plan: Phase 1 CLI entity discovery (`list`/`show`)

- [x] Add/extend Invoice Ninja wrappers needed by list/show (`tasks`, `projects` get/list shapes).
- [x] Implement `ninjops list <entity>` with filters, formatters, and validation.
- [x] Implement `ninjops show <entity> <id>` with arg validation and JSON output.
- [x] Wire new top-level commands in root.
- [x] Add deterministic offline tests for argument validation and formatting/filter helpers.
- [x] Update README docs with `list`/`show` examples.
- [x] Update Phase 1 tracker checkboxes/results for completed list/show work.
- [x] Run `go test ./internal/app` and capture result.

### Results

- Added `ninjops list` with entity-scoped filter validation, format selection (`table|json|simple`), and concise renderers.
- Added `ninjops show` for singular entity lookup (`client|project|task|quote|invoice`) with JSON output by default.
- Wired `list`/`show` into root command registration and added missing Invoice Ninja `tasks` API wrapper (`ListTasks`, `GetTask`).
- Added deterministic tests for list/show argument validation, missing-token behavior, list filtering helper behavior, and formatter output.
- Updated README with usage docs and examples for both new commands.
- Verified `go test ./internal/app` passes.

## Task Plan: Phase-1 data-layer foundations for full Invoice Ninja workflow

- [x] Expand project API wrapper with CRUD + deterministic find helpers.
- [x] Add task API wrapper with CRUD + deterministic find helpers.
- [x] Extend Invoice Ninja models and quote/invoice payloads for project linking.
- [x] Extend QuoteSpec linkage IDs with backward-compatible schema changes.
- [x] Add focused unit tests for helper filtering and request payload mapping.
- [x] Run `go test ./internal/invoiceninja ./internal/spec` and capture results.

### Results

- Added full project and task wrappers (list/get/create/update + client/project-scoped find helpers) with deterministic local filtering on response data.
- Extended Invoice Ninja models with task types plus project linkage (`project_id`) on quote/invoice entities and payloads.
- Wired quote/invoice build helpers to include `project_id` from `QuoteSpec.Project.ID`, with fallback to existing entity project when updating.
- Added optional linkage IDs in `QuoteSpec` (`client.id`, `project.id`, `task_ids`) while preserving compatibility for existing JSON fixtures.
- Added deterministic unit tests for project/task filter helpers, name/description matching helpers, and quote/invoice request project-id mapping.
- Verified with `go test ./internal/invoiceninja ./internal/spec` (both packages passing).

## Task Plan: Targeted code-review bug fixes (new/list workflows)

- [x] Fix `new quote` IO stream/TTY handling to use command IO sources.
- [x] Eliminate buffered-reader rewrapping across quote workflow prompts.
- [x] Add task/client/project relationship validation + derivation for `--task-ids`.
- [x] Return explicit errors for API fetch failures in optional selection paths.
- [x] Correct list filtering/pagination behavior using Invoice Ninja find helpers.
- [x] Add/adjust focused tests in `internal/app` and `internal/invoiceninja`.
- [x] Run required targeted + full test suites.

### Results

- `new quote` now consistently uses `cmd.InOrStdin()` / `cmd.OutOrStdout()` for prompts/output, and interactive detection now checks the command input stream.
- Added `asBufferedReader` + shared-reader threading through interactive helpers to prevent dropped pre-seeded/piped input across chained prompts.
- Added task relationship validation/derivation in quote enrichment: validates `--task-ids` against selected client/project, derives IDs when unset, and errors clearly on conflicts.
- Optional selection paths now return explicit API errors (client/project/task list fetch helpers no longer silently swallow failures).
- `list` now uses Invoice Ninja find helpers for filtered entities and paginates tasks robustly (including `FindTasksByClient` path and `--limit`-bounded aggregation).
- Added focused app tests (reader threading, relationship validation/derivation, conflict/error propagation, list helper usage + pagination) and invoiceninja tests (`FindTasksByClient` and client filter helper).
- Verified with `go test ./internal/app ./internal/invoiceninja` and `go test ./...` (both passing).

## Task Plan: Homebrew install + release automation

- [x] Add GoReleaser config for cross-platform archives and Homebrew tap publishing.
- [x] Add tag-triggered GitHub Actions release workflow to run GoReleaser.
- [x] Update README install/release docs for Homebrew usage and tag-based releases.
- [x] Document required `HOMEBREW_TAP_GITHUB_TOKEN` secret for release workflow.
- [x] Run `go test ./...` and run cheap GoReleaser validation if available.

### Results

- Added `.goreleaser.yml` with multi-OS/multi-arch builds (`linux`, `darwin`, `windows`; `amd64`, `arm64`), archives, checksums, and Homebrew formula publishing to `v1truv1us/homebrew-tap`.
- Added `.github/workflows/release.yml` to trigger on tag pushes `v*` and run GoReleaser with `GITHUB_TOKEN` plus `HOMEBREW_TAP_GITHUB_TOKEN`.
- Updated `README.md` installation docs with Homebrew tap/install commands, corrected source/release URLs to `v1truv1us/ninjops`, and documented tag-based release flow + required secret.
- Verified with `go test ./...` and `goreleaser check`.
