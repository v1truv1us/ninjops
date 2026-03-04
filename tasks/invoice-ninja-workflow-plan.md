# Invoice Ninja Full Workflow Plan

Date: 2026-03-01
Status: Planning

## Execution Status (Current)

- Phase 1 (Foundation): mostly complete
  - Implemented: project/task wrappers, list/show commands, new quote project selection
  - Remaining: `ninja sync` project create/update parity
- Phase 2 (Inline editing/preview): complete in `new quote` workflow (implemented inline in app command path)
- Phase 3 (Invoice workflow): complete for `new invoice`, `convert`, and `edit` command set
- Phase 4 (Task management): complete for `new quote`; partial for `new invoice` task-selection parity
- Phase 5 (polish/docs hardening): in progress

## Goal

Implement a complete CLI-based Invoice Ninja workflow where users can manage the entire client → project → task → quote → invoice lifecycle through ninjops, with **selection or creation at every step**.

## User Requirements

1. **Full workflow via CLI** - no need to switch to Invoice Ninja UI
2. **Selection at each step** - choose from existing entities or create new ones
3. **Inline editing** - view and edit terms/notes directly in CLI before finalizing
4. **Flexible entry points** - start at any point in the workflow based on what already exists

## Workflow Design

### Primary Workflows

#### 1. Full Engagement (New Client)
```
ninjops new quote
  → Select/Create Client
  → Select/Create Project (for client)
  → Select/Create Tasks (for project)
  → Edit Terms/Notes (inline)
  → Preview Quote
  → Create Quote in Invoice Ninja
  → (Optional) Convert to Invoice
```

#### 2. Existing Client, New Project
```
ninjops new quote --client-id <id>
  → Select/Create Project
  → Select/Create Tasks
  → Edit Terms/Notes
  → Preview Quote
  → Create Quote
  → (Optional) Convert to Invoice
```

#### 3. Existing Client & Project, New Quote
```
ninjops new quote --client-id <id> --project-id <id>
  → Select/Create Tasks
  → Edit Terms/Notes
  → Preview Quote
  → Create Quote
  → (Optional) Convert to Invoice
```

#### 4. Quick Invoice (Skip Quote)
```
ninjops new invoice
  → Select Client
  → Select Project (optional)
  → Select Tasks (optional)
  → Edit Terms/Notes
  → Preview Invoice
  → Create Invoice
```

#### 5. Quote → Invoice Conversion
```
ninjops convert <quote-id>
  → Shows quote summary
  → Confirm conversion
  → Creates invoice from quote
```

#### 6. Edit Existing
```
ninjops edit quote <quote-id>
  → Load existing quote
  → Edit terms/notes inline
  → Update in Invoice Ninja

ninjops edit invoice <invoice-id>
  → Load existing invoice
  → Edit terms/notes inline
  → Update in Invoice Ninja
```

## Command Structure

### Core Commands

#### `ninjops new quote`
Interactive quote creation with full workflow.

```
Flags:
  --client-id <id>           Pre-select client
  --client-email <email>     Select client by email
  --client-name <name>       Select client by name
  --project-id <id>          Pre-select project
  --project-name <name>      Create new project with name
  --task-ids <id1,id2>       Pre-select tasks
  --non-interactive          Skip all prompts (for automation)
  --skip-edit                Skip terms/notes editing
  --output <path>            Save QuoteSpec JSON
  --artifacts-dir <dir>      Directory for generated files

Workflow:
  1. Client Selection/Creation
  2. Project Selection/Creation
  3. Task Selection/Creation
  4. QuoteSpec Generation
  5. Terms/Notes Editing (opens $EDITOR or inline prompts)
  6. Preview
  7. Confirm & Create in Invoice Ninja
  8. Offer to convert to invoice
```

#### `ninjops new invoice`
Direct invoice creation (skips quote step).

```
Flags:
  Same as `new quote` plus:
  --from-quote <id>          Convert existing quote to invoice

Workflow:
  1. Client Selection/Creation
  2. Project Selection/Creation (optional)
  3. Task Selection/Creation (optional)
  4. InvoiceSpec Generation
  5. Terms/Notes Editing
  6. Preview
  7. Confirm & Create in Invoice Ninja
```

#### `ninjops edit quote <id>`
Edit existing quote.

```
Flags:
  --field <field>            Edit specific field (terms, public_notes, line_items)
  --editor                   Open in $EDITOR
  --no-preview               Skip preview after edit

Workflow:
  1. Load quote from Invoice Ninja
  2. Display current values
  3. Edit terms/notes (inline or editor)
  4. Preview changes
  5. Confirm & update
```

#### `ninjops edit invoice <id>`
Edit existing invoice.

```
Same as `edit quote` but for invoices.
```

#### `ninjops convert <quote-id>`
Convert quote to invoice.

```
Flags:
  --yes                      Skip confirmation
  --edit                     Edit before converting

Workflow:
  1. Load quote
  2. Show summary
  3. (Optional) Edit terms/notes
  4. Confirm conversion
  5. Create invoice
  6. Show invoice details
```

#### `ninjops list <entity>`
List entities from Invoice Ninja.

```
Entities: clients, projects, tasks, quotes, invoices

Flags:
  --client-id <id>           Filter by client
  --project-id <id>          Filter by project (for tasks)
  --status <status>          Filter by status
  --format <format>          Output: table, json, simple
  --limit <n>                Limit results

Examples:
  ninjops list clients
  ninjops list projects --client-id <id>
  ninjops list tasks --project-id <id>
  ninjops list quotes --status draft
  ninjops list invoices --client-id <id>
```

#### `ninjops show <entity> <id>`
Show entity details.

```
Entities: client, project, task, quote, invoice

Examples:
  ninjops show client <id>
  ninjops show project <id>
  ninjops show quote <id>
  ninjops show invoice <id>
```

### Enhanced Existing Commands

#### `ninjops ninja sync` (enhanced)
Sync QuoteSpec to Invoice Ninja with project support.

```
Enhancements:
  - Create/update project if project_id or project details provided
  - Create/update tasks if task_ids or task details provided
  - Better diffing and preview
  - Support --project-id and --task-ids flags
```

## Implementation Phases

### Phase 1: Project & Task Support (Foundation)
**Duration: 1-2 sprints**

**Goals:**
- Add Invoice Ninja project & task API support
- Enhance `new quote` with project selection
- Add `list` and `show` commands

**Deliverables:**
- [ ] `internal/invoiceninja/projects.go` - Project CRUD operations
- [ ] `internal/invoiceninja/tasks.go` - Task CRUD operations
- [ ] Project & Task models in `internal/invoiceninja/models.go`
- [ ] Tests for project/task operations
- [ ] `internal/app/list.go` - List command implementation
- [ ] `internal/app/show.go` - Show command implementation
- [ ] Tests for list/show commands
- [ ] Update `new quote` to support project selection after client selection
- [ ] Update `ninja sync` to handle project creation/update

**Acceptance Criteria:**
- Can list clients, projects, tasks, quotes, invoices via CLI
- Can show details for any entity
- `new quote` offers project selection after client selection
- `ninja sync` creates/updates projects when provided

### Phase 2: Inline Editing & Preview
**Duration: 1 sprint**

**Goals:**
- Add inline editing of terms/notes
- Add preview functionality
- Improve user experience

**Deliverables:**
- [ ] `internal/app/editor.go` - Inline editing helper (supports $EDITOR and prompts)
- [ ] `internal/app/preview.go` - Preview formatter
- [ ] Enhance `new quote` with editing step
- [ ] Add `--skip-edit` flag
- [ ] Add preview before confirmation
- [ ] Tests for editing functionality
- [ ] Update README with editing workflow

**Acceptance Criteria:**
- `new quote` opens editor or prompts for terms/notes editing
- Preview shows exactly what will be created
- Works with `--non-interactive` mode

### Phase 3: Invoice Workflow
**Duration: 1 sprint**

**Goals:**
- Add direct invoice creation
- Add quote → invoice conversion
- Add invoice editing

**Deliverables:**
- [ ] `internal/app/new_invoice.go` - New invoice command
- [ ] `internal/app/convert.go` - Quote to invoice conversion
- [ ] `internal/app/edit.go` - Edit existing quotes/invoices
- [ ] `internal/invoiceninja/invoices.go` enhancements for updates
- [ ] Tests for invoice commands
- [ ] Update README with invoice workflows

**Acceptance Criteria:**
- Can create invoice directly via `new invoice`
- Can convert quote to invoice via `convert`
- Can edit existing quotes/invoices
- All operations show preview and require confirmation

### Phase 4: Task Management
**Duration: 1 sprint**

**Goals:**
- Add task selection/creation in workflows
- Link tasks to projects and quotes/invoices

**Deliverables:**
- [ ] Enhance `new quote` with task selection
- [ ] Enhance `new invoice` with task selection
- [ ] Task creation prompts
- [ ] Tests for task integration
- [ ] Update README with task workflow

**Acceptance Criteria:**
- Can select existing tasks during quote/invoice creation
- Can create new tasks during quote/invoice creation
- Tasks are linked to projects and quotes/invoices

### Phase 5: Polish & Documentation
**Duration: 1 sprint**

**Goals:**
- Error handling improvements
- User experience polish
- Comprehensive documentation

**Deliverables:**
- [ ] Comprehensive error messages with suggestions
- [ ] Progress indicators for long operations
- [ ] Color-coded output
- [ ] Complete README rewrite with all workflows
- [ ] Examples for common use cases
- [ ] Integration tests for full workflows
- [ ] Smoke test enhancements

**Acceptance Criteria:**
- Error messages are helpful and actionable
- Documentation covers all workflows with examples
- All smoke tests pass
- Integration tests cover main workflows

## Data Model Enhancements

### QuoteSpec Extensions

```go
type QuoteSpec struct {
    // Existing fields...
    
    // New fields for Invoice Ninja integration
    ClientID    string   `json:"client_id,omitempty"`    // Invoice Ninja client ID
    ProjectID   string   `json:"project_id,omitempty"`   // Invoice Ninja project ID
    TaskIDs     []string `json:"task_ids,omitempty"`     // Invoice Ninja task IDs
}

type ProjectInfo struct {
    // Existing fields...
    
    // New fields for Invoice Ninja project
    ID            string  `json:"id,omitempty"`            // Invoice Ninja project ID
    BudgetedHours float64 `json:"budgeted_hours,omitempty"`
    TaskRate      float64 `json:"task_rate,omitempty"`
    DueDate       string  `json:"due_date,omitempty"`      // YYYY-MM-DD
}
```

### InvoiceSpec (New)

```go
type InvoiceSpec struct {
    // Similar to QuoteSpec but for direct invoices
    Metadata    MetaInfo      `json:"metadata"`
    Client      ClientInfo    `json:"client"`
    ClientID    string        `json:"client_id,omitempty"`
    Project     ProjectInfo   `json:"project"`
    ProjectID   string        `json:"project_id,omitempty"`
    Tasks       []TaskInfo    `json:"tasks"`
    TaskIDs     []string      `json:"task_ids,omitempty"`
    Work        WorkDefinition `json:"work"`
    Pricing     PricingInfo   `json:"pricing"`
    Settings    QuoteSettings `json:"settings"`
}

type TaskInfo struct {
    ID          string  `json:"id,omitempty"`
    Description string  `json:"description"`
    Hours       float64 `json:"hours,omitempty"`
    Rate        float64 `json:"rate,omitempty"`
}
```

## API Integration Enhancements

### Invoice Ninja Client

```go
// internal/invoiceninja/projects.go
type Client struct {
    // Existing...
}

// Project operations
func (c *Client) ListProjects(ctx context.Context, page, perPage int) (*ProjectListResponse, error)
func (c *Client) ListProjectsByClient(ctx context.Context, clientID string, page, perPage int) (*ProjectListResponse, error)
func (c *Client) GetProject(ctx context.Context, id string) (*NinjaProject, error)
func (c *Client) CreateProject(ctx context.Context, req CreateProjectRequest) (*NinjaProject, error)
func (c *Client) UpdateProject(ctx context.Context, id string, req UpdateProjectRequest) (*NinjaProject, error)

// Task operations
func (c *Client) ListTasks(ctx context.Context, page, perPage int) (*TaskListResponse, error)
func (c *Client) ListTasksByProject(ctx context.Context, projectID string, page, perPage int) (*TaskListResponse, error)
func (c *Client) GetTask(ctx context.Context, id string) (*NinjaTask, error)
func (c *Client) CreateTask(ctx context.Context, req CreateTaskRequest) (*NinjaTask, error)
func (c *Client) UpdateTask(ctx context.Context, id string, req UpdateTaskRequest) (*NinjaTask, error)
```

### Models

```go
// internal/invoiceninja/models.go

type NinjaProject struct {
    ID             string    `json:"id"`
    UserID         string    `json:"user_id"`
    ClientID       string    `json:"client_id"`
    Name           string    `json:"name"`
    Description    string    `json:"description"` // or public_notes
    PrivateNotes   string    `json:"private_notes"`
    TaskRate       float64   `json:"task_rate"`
    BudgetedHours  float64   `json:"budgeted_hours"`
    DueDate        string    `json:"due_date"` // YYYY-MM-DD
    Number         string    `json:"number"`
    Color          string    `json:"color"`
    CreatedAt      int64     `json:"created_at"`
    UpdatedAt      int64     `json:"updated_at"`
    IsDeleted      bool      `json:"is_deleted"`
}

type NinjaTask struct {
    ID          string  `json:"id"`
    UserID      string  `json:"user_id"`
    ClientID    string  `json:"client_id"`
    ProjectID   string  `json:"project_id"`
    InvoiceID   string  `json:"invoice_id"`
    Number      string  `json:"number"`
    Description string  `json:"description"`
    Duration    int     `json:"duration"` // seconds
    Rate        float64 `json:"rate"`
    CreatedAt   int64   `json:"created_at"`
    UpdatedAt   int64   `json:"updated_at"`
    IsDeleted   bool    `json:"is_deleted"`
}

type ProjectListResponse struct {
    Data []NinjaProject `json:"data"`
    Meta APIMeta        `json:"meta"`
}

type TaskListResponse struct {
    Data []NinjaTask `json:"data"`
    Meta APIMeta     `json:"meta"`
}

type CreateProjectRequest struct {
    ClientID       string  `json:"client_id"`
    Name           string  `json:"name"`
    Description    string  `json:"description,omitempty"`
    PrivateNotes   string  `json:"private_notes,omitempty"`
    TaskRate       float64 `json:"task_rate,omitempty"`
    BudgetedHours  float64 `json:"budgeted_hours,omitempty"`
    DueDate        string  `json:"due_date,omitempty"`
    Color          string  `json:"color,omitempty"`
}

type CreateTaskRequest struct {
    ClientID    string  `json:"client_id"`
    ProjectID   string  `json:"project_id,omitempty"`
    Description string  `json:"description"`
    Rate        float64 `json:"rate,omitempty"`
}
```

## User Experience Design

### Interactive Flow Example

```
$ ninjops new quote

? Select a client:
  ❯ 1. Acme Corporation (acme@example.com)
    2. Church of Holy Spirit (admin@church.org)
    3. + Create new client
    0. Skip (use template)

→ Selected: Acme Corporation [id=abc123]

? Select a project for Acme Corporation:
  ❯ 1. Website Redesign (Due: 2026-04-15)
    2. Mobile App Development (Due: 2026-05-01)
    3. + Create new project
    0. No project (skip)

→ Selected: Website Redesign [id=proj456]

? Select tasks for Website Redesign:
  ◉ 1. Homepage mockup (12 hours @ $75/hr)
  ◉ 2. Contact form implementation (8 hours @ $75/hr)
  ◉ 3. + Create new task
  ◯ 0. Done selecting

→ Selected 2 tasks

? Project name [Website Redesign]: 
? Project description [Complete redesign...]: 
? Project type [website]: 
? Timeline [6 weeks]: 4 weeks

Generating quote...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PREVIEW: Quote for Acme Corporation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

PROJECT: Website Redesign
CLIENT:  Acme Corporation
TASKS:   2 tasks selected (20 hours total)

TERMS & CONDITIONS:
─────────────────────────────────────────────
Payment is due within 14 days of invoice date.
A deposit of 25% is required to commence work.
[... full terms ...]

PUBLIC NOTES:
─────────────────────────────────────────────
Project: Website Redesign
Client: Acme Corporation
[... full notes ...]

LINE ITEMS:
─────────────────────────────────────────────
Homepage mockup           12 hrs × $75    $900.00
Contact form               8 hrs × $75    $600.00
─────────────────────────────────────────────
Total                                   $1,500.00

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

? Edit terms/notes before creating? (y/N): y

Opening editor...

[Editor opens with terms.md and public_notes.txt]

? Create quote in Invoice Ninja? (Y/n): y

✓ Created quote #QUO-2026-001 [id=quote789]
  Client: Acme Corporation
  Project: Website Redesign
  Tasks: 2 tasks linked
  Amount: $1,500.00

? Convert to invoice now? (y/N): n

Quote created successfully!
View in Invoice Ninja: https://invoiceninja.fergify.work/quotes/quote789

Next steps:
  - Edit:     ninjops edit quote quote789
  - Convert:  ninjops convert quote789
  - View:     ninjops show quote quote789
```

### Non-Interactive Mode

```
$ ninjops new quote \
    --client-id abc123 \
    --project-id proj456 \
    --task-ids task1,task2 \
    --project-timeline "4 weeks" \
    --skip-edit \
    --non-interactive \
    --output quote.json

✓ Created QuoteSpec: quote.json
✓ Created quote #QUO-2026-001 [id=quote789]
✓ Generated artifacts:
  - quote.json
  - terms.md
  - public_notes.txt
```

## Testing Strategy

### Unit Tests
- [ ] Project CRUD operations
- [ ] Task CRUD operations
- [ ] List/show commands
- [ ] Inline editing helpers
- [ ] Preview formatting
- [ ] Quote → Invoice conversion logic

### Integration Tests
- [ ] Full workflow: client → project → tasks → quote → invoice
- [ ] Edit existing quote/invoice
- [ ] Convert quote to invoice
- [ ] Non-interactive mode workflows

### Smoke Tests
- [ ] CLI smoke test includes project selection
- [ ] List/show commands work
- [ ] Edit commands work
- [ ] Convert command works

## Documentation Updates

### README.md Structure

```markdown
# Ninjops

## Overview
- What is ninjops
- Key features
- Installation

## Quick Start
- Configure
- First quote
- First invoice

## Workflows

### Creating a New Quote
- Interactive mode
- Non-interactive mode
- With existing client/project
- With tasks

### Creating an Invoice Directly
- When to skip quote
- How to create invoice

### Converting Quote to Invoice
- Automatic conversion
- Edit before converting

### Editing Existing Documents
- Edit quote
- Edit invoice

### Managing Entities
- List clients/projects/tasks
- Show entity details

## Commands Reference
- `ninjops new quote`
- `ninjops new invoice`
- `ninjops convert`
- `ninjops edit`
- `ninjops list`
- `ninjops show`
- `ninjops configure`
- `ninjops validate`
- `ninjops generate`
- `ninjops assist`
- `ninjops ninja *`

## Configuration
- Config file location
- Auth credentials
- Provider setup
- Model selection

## Examples
- New client engagement
- Existing client, new project
- Quick invoice
- Recurring work

## Troubleshooting
- Common issues
- Error messages
- Debug mode
```

### Examples Document

Create `docs/examples.md` with detailed walkthroughs of common scenarios.

## Risks and Mitigation

### Risk: Complex CLI Flow
**Mitigation:**
- Clear prompts with defaults
- Skip options at every step
- Non-interactive mode for automation
- Good error messages

### Risk: API Changes
**Mitigation:**
- Version API calls
- Graceful degradation
- Clear error messages when API unavailable

### Risk: User Data Loss
**Mitigation:**
- Always show preview before creating
- Confirmation prompts for destructive actions
- Dry-run mode for testing

### Risk: Performance with Many Entities
**Mitigation:**
- Pagination in list commands
- Caching of entity lists
- Search/filter options

## Success Metrics

- [ ] Can create quote from scratch in < 2 minutes (interactive)
- [ ] Can create invoice in < 1 minute (with existing client/project)
- [ ] All entity types are accessible via CLI
- [ ] Zero data loss scenarios
- [ ] Clear error messages for all failure modes
- [ ] Full test coverage for workflows
- [ ] Documentation covers all use cases

## Timeline

- **Week 1-2**: Phase 1 (Project & Task Support)
- **Week 3**: Phase 2 (Inline Editing)
- **Week 4**: Phase 3 (Invoice Workflow)
- **Week 5**: Phase 4 (Task Management)
- **Week 6**: Phase 5 (Polish & Documentation)

**Total: 6 weeks (1.5 months)**

## Next Steps

1. Review and approve this plan
2. Update `tasks/todo.md` with Phase 1 tasks
3. Begin implementation of Phase 1
4. Document lessons learned as we progress
