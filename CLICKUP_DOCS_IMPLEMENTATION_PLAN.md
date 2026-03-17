# ClickUp Docs Support - Complete Implementation Plan

## Objective

Add first-class ClickUp Docs support to this CLI while preserving existing architecture, command UX, and contributor expectations. The proposal is designed to maximize maintainers' acceptance by staying consistent with current patterns in command naming, output style, testing, and generated documentation.

## Scope

This plan covers:

- New command group for Docs: `clickup doc ...`
- Read and write operations supported by public ClickUp v3 Docs API
- Tests for command wiring and validation behavior
- Documentation updates (README + generated reference docs + docs site content)
- Examples updates under `examples/`
- Skill update in `skills/clickup-cli/SKILL.md`

## Important API Constraint

ClickUp public v3 Docs endpoints clearly support:

- List/search Docs
- Fetch Doc
- Create Doc
- Fetch page listing for a Doc
- Fetch page
- Create page
- Edit page

At planning time, public endpoints for deleting/archiving/restoring Docs/Pages are not clearly documented. The implementation must therefore:

- Ship complete support for publicly documented endpoints
- Explicitly document delete/archive/restore as API-limited if no stable endpoint is available during implementation

---

## Phase 0 - Branch and Safety

1. Create feature branch before code changes:
   - `git checkout -b feature/clickup-docs-support`
2. Verify clean workspace:
   - `git status`

---

## Phase 1 - Command Architecture

### 1.1 Add top-level Docs command

Create new package:

- `pkg/cmd/doc/`

Entry point:

- `pkg/cmd/doc/doc.go`

Top-level command shape:

- `clickup doc <command>`

Reasoning:

- Singular naming matches existing command conventions:
  - `task`, `comment`, `status`, `tag`, `field`, `link`
- Better consistency improves UX and acceptance probability

### 1.2 Wire command into root

Update:

- `pkg/cmd/root/root.go`

Add:

- `cmd.AddCommand(doc.NewCmdDoc(f))`

---

## Phase 2 - API and Data Layer (within command package)

Because `go-clickup` does not currently expose v3 Docs endpoints, use the same established approach already used elsewhere in the repo for unsupported endpoints: authenticated raw HTTP through `internal/api.Client.DoRequest()`.

Create helper file:

- `pkg/cmd/doc/api.go`

Include:

- Shared response structs (`docCore`, `docDetail`, `pageRef`, `pageDetail`, pagination wrappers)
- Helper to resolve workspace ID from config
- Shared request function with consistent error handling
- Parent type parser (`SPACE|FOLDER|LIST|EVERYTHING|WORKSPACE` + numeric equivalents)
- URL/query builders for docs endpoints

API base:

- `https://api.clickup.com/api/v3/workspaces/{workspace_id}/...`

---

## Phase 3 - Commands and Flags

### 3.1 Docs commands

Implement:

- `pkg/cmd/doc/list.go` -> `clickup doc list`
- `pkg/cmd/doc/view.go` -> `clickup doc view <doc-id>`
- `pkg/cmd/doc/create.go` -> `clickup doc create --name ...`

Recommended flags:

- `doc list`
  - `--creator`
  - `--deleted`
  - `--archived`
  - `--parent-id`
  - `--parent-type`
  - `--limit`
  - `--cursor`
- `doc create`
  - `--name` (required)
  - `--parent-id`
  - `--parent-type`
  - `--visibility` (`PUBLIC|PRIVATE|PERSONAL|HIDDEN`)
  - `--create-page` (default true)

### 3.2 Page subcommand group

Implement:

- `pkg/cmd/doc/page.go` -> `clickup doc page <command>`
- `pkg/cmd/doc/page_list.go` -> `clickup doc page list <doc-id>`
- `pkg/cmd/doc/page_view.go` -> `clickup doc page view <doc-id> <page-id>`
- `pkg/cmd/doc/page_create.go` -> `clickup doc page create <doc-id> --name ...`
- `pkg/cmd/doc/page_edit.go` -> `clickup doc page edit <doc-id> <page-id> ...`

Recommended flags:

- `doc page list`
  - `--max-depth` (maps to `max_page_depth`)
- `doc page view`
  - `--content-format` (`text/md|text/plain`)
- `doc page create`
  - `--name` (required)
  - `--parent-page-id`
  - `--sub-title`
  - `--content`
  - `--content-format`
- `doc page edit`
  - `--name`
  - `--sub-title`
  - `--content`
  - `--content-format`
  - `--content-edit-mode` (`replace|append|prepend`)

### 3.3 Output conventions

Follow existing CLI behavior:

- `PersistentPreRunE: cmdutil.NeedsAuth(f)`
- JSON flags for read-oriented commands:
  - `--json`, `--jq`, `--template` via `cmdutil.AddJSONFlags`
- Human table output via `internal/tableprinter`
- Optional quick-actions footer where useful
- Errors wrapped with context (`fmt.Errorf("failed to ...: %w", err)`)

---

## Phase 4 - Test Plan

Add tests in:

- `pkg/cmd/doc/doc_test.go`
- `pkg/cmd/doc/list_test.go`
- `pkg/cmd/doc/view_test.go`
- `pkg/cmd/doc/create_test.go`
- `pkg/cmd/doc/page_test.go`
- `pkg/cmd/doc/page_list_test.go`
- `pkg/cmd/doc/page_view_test.go`
- `pkg/cmd/doc/page_create_test.go`
- `pkg/cmd/doc/page_edit_test.go`

Test style should match current repository patterns:

- Command `Use` strings
- Flag existence/defaults
- Required args/flags validation behavior
- Accepted enum values for parent type and edit mode

---

## Phase 5 - Documentation Updates (Required)

### 5.1 README

Update:

- `README.md`

Add:

- Docs feature bullets in "What it does"
- New Docs command entries in command area/table
- One practical usage snippet for Docs/Page workflow

### 5.2 Generated CLI reference docs

Regenerate docs:

- `make docs`

This updates:

- `docs/src/content/docs/reference/*`
- `docs/src/content/docs/commands.md`

### 5.3 Docs category mapping

Update generator categorization:

- `cmd/gen-docs/main.go`

Add `doc` to an appropriate category (recommended: new category `Docs` or grouped with core workflow commands).

### 5.4 Narrative docs site page

Add/update a human guide under:

- `docs/src/content/docs/`

Recommended file:

- `docs/src/content/docs/docs-workflow.md`

Include:

- How to list Docs
- How to inspect page tree
- How to create Doc + first page
- How to append content to a page
- JSON-first examples for automation

---

## Phase 6 - Examples Updates (Required)

Add new examples under:

- `examples/`

Recommended files:

- `examples/clickup-docs-create.yml`
- `examples/clickup-docs-page-edit.yml`

Example scenarios:

1. Create a Doc and initial page
2. Fetch page listing and inspect IDs
3. Append release notes to existing page via `content_edit_mode=append`

---

## Phase 7 - Skill File Update (Required)

Update:

- `skills/clickup-cli/SKILL.md`

Add dedicated section:

- `## Docs`

Include:

- `clickup doc list`
- `clickup doc view`
- `clickup doc create`
- `clickup doc page list`
- `clickup doc page view`
- `clickup doc page create`
- `clickup doc page edit`

Also include:

- JSON-first agent usage examples (`--json`, `--jq`)
- Notes about current API limitations for delete/archive/restore (if still unavailable)

---

## Phase 8 - Validation and QA

Run full quality checks:

- `go test ./...`
- `go vet ./...`
- `go build -o /dev/null ./cmd/clickup`
- `make docs`

Manual smoke checks:

- `clickup doc list --json`
- `clickup doc create --name "Docs API Test"`
- `clickup doc page list <doc-id> --max-depth -1`
- `clickup doc page create <doc-id> --name "Intro" --content "# Intro"`
- `clickup doc page edit <doc-id> <page-id> --content "More content" --content-edit-mode append`

---

## Suggested Rollout Strategy

To reduce review risk, use two PRs:

1. **PR A - Core feature**
   - command implementation + tests
2. **PR B - Content**
   - docs site, README, examples, `SKILL.md`

Alternative: single PR if maintainers prefer one end-to-end feature set.

---

## Acceptance Checklist

- [ ] New `doc` command group implemented and wired
- [ ] Docs/Page create/read/edit flows functional with authenticated API calls
- [ ] Tests added for new commands/flags/validation
- [ ] README updated with Docs commands
- [ ] Generated reference docs refreshed (`make docs`)
- [ ] `cmd/gen-docs/main.go` category mapping updated
- [ ] New examples in `examples/`
- [ ] `skills/clickup-cli/SKILL.md` updated with Docs workflows
- [ ] Quality checks pass (`test`, `vet`, `build`)
- [ ] Any unsupported lifecycle actions documented as API limitations
