# Lessons Learned

## 2026-03-01

- User onboarding expectations: prefer a dedicated `ninjops configure` workflow over ad-hoc environment export instructions.
- For credential setup guidance, prioritize productized CLI onboarding paths (interactive and non-interactive) before manual shell steps.
- Onboarding completeness rule: `configure` must include provider credential setup plus provider connection validation, not only static config writes.
- Secrets persistence rule: store sensitive values in a dedicated credentials file (`auth-creds.json`) referenced by the main config, instead of embedding secrets in primary config JSON/JSONC.
- Provider rollout rule: when adding provider families, include explicit model selection in onboarding (`configure --model`) and persist resolved `agent.model` for runtime use.
- Alias parity rule: if a model alias maps to provider-specific auth/model behavior (for example `openai-codex` -> `opencode`), resolve provider and model together in one shared helper and reuse it in configure, config load, and runtime execution paths.
- Connectivity hardening rule: for providers with stricter runtime auth requirements (for example `opencode`), configure-time checks must include a minimal completion probe against the resolved model, not only `/models` listing.
- Provider validation rule: for any non-offline provider, configure-time connectivity must verify both API auth (`/models`) and a real model execution call using the resolved/default model to catch model-access mismatches early.
- Workflow design rule: when building CLI workflows for complex systems (like Invoice Ninja), design for **selection at every step** - allow users to select existing entities or create new ones at each point in the workflow (client → project → task → quote/invoice).
- Invoice Ninja workflow rule: the canonical flow is Client → Project → Tasks → Quote/Invoice, where each entity can be selected/created independently. Always support both interactive (guided) and non-interactive (automation) modes.
- Entity relationship rule: in Invoice Ninja, Projects belong to Clients, Tasks belong to Projects, and Quotes/Invoices can link to both Clients and Projects. Always maintain these relationships when creating/updating entities.
- User experience rule: for multi-step CLI workflows, always provide: (1) clear prompts with defaults, (2) preview before creation, (3) inline editing capabilities, (4) non-interactive mode for automation, and (5) helpful error messages with suggestions.
- Documentation structure rule: for comprehensive CLI tools, structure docs by workflow (not just commands), include realistic examples, and cover both interactive and non-interactive usage patterns.
- Relationship validation rule: when users pass both `client-id` and `project-id`, validate they belong together before API create calls to prevent avoidable Invoice Ninja `422` errors.
