---
title: "Creative Vision Engine: Make the Printing Press Generate Discrawl-Level CLIs"
type: feat
status: completed
date: 2026-03-26
---

# Creative Vision Engine: Make the Printing Press Generate Discrawl-Level CLIs

## Overview

The printing press generates competent API wrappers. It does not generate creative tools. The difference is discrawl: 12 commands, 539 stars, and the insight that "Discord messages are a searchable knowledge base." That insight didn't come from the OpenAPI spec. It came from understanding what power users actually need.

Right now the press has a creativity pipeline: Phase 0 research -> profiler signals -> vision templates -> Phase 4 hand-written workflows. But the bottleneck is that **Phase 4 workflows are hand-written every single time**. The profiler detects "this API is high-volume" but it can't detect "this API's messages are institutional knowledge." The generator produces `sync.go` and `search.go` from templates, but the 7 workflow commands (stale, orphans, deps, load, triage, overdue, standup) are written from scratch by the LLM running the skill.

This plan upgrades the printing press so the generator itself can produce creative, domain-aware workflow commands - not just API wrappers. The goal: when you run `/printing-press Linear`, Phase 4 shouldn't require hand-writing 7 Go files. The generator should produce them from templates, informed by the profiler's domain classification and the VisionaryPlan's workflow decisions.

## Problem Statement

### The Creativity Ladder (what the press generates today)

| Rung | What It Is | Generated? | Example |
|---|---|---|---|
| 1 | API wrapper commands | Yes (templates) | `issue create --title "..."` |
| 2 | Output formatting | Yes (templates) | `--json`, `--select`, `--csv` |
| 3 | Local persistence | Yes (templates) | `sync`, `search`, `export` |
| 4 | Domain analytics | **No (hand-written)** | `stale --days 30`, `load --team ENG` |
| 5 | Behavioral insights | **No (not even in skill)** | `predict-stale`, `entropy`, `bottleneck` |

The press stops at Rung 3. Rungs 4 and 5 are where discrawl lives. The skill's Phase 4 instructions say "write workflow commands" but the generator has no templates for them and the profiler has no domain archetype system to guide what workflows to generate.

### What the profiler detects today

```go
HighVolume       bool  // >5 paginated endpoints
NeedsSearch      bool  // 3+ list resources, <50% with search
HasRealtime      bool  // webhook/event/callback keywords
HasChronological bool  // audit/log/event/history keywords
OfflineValuable  bool  // >60% GET endpoints
HasLifecycles    bool  // status/state fields with 3+ enums
HasDependencies  bool  // _id fields matching other resources
```

These are **structural signals** - they detect API shape, not domain meaning. The profiler knows "this API has state fields" but doesn't know "this is a project management tool where stale issue detection is the killer feature."

### What the profiler should detect

```go
// Domain archetype (new)
DomainArchetype  string  // "communication", "project-management", "payments", etc.

// Domain-specific signals (new)
HasAssignees     bool    // user assignment patterns (PM tools)
HasDueDates      bool    // temporal deadlines (PM tools)
HasPriority      bool    // priority/severity levels (PM tools)
HasThreading     bool    // nested conversations (communication tools)
HasTransactions  bool    // financial transactions (payment tools)
HasSubscriptions bool    // recurring billing (payment tools)
HasMedia         bool    // file/image/video attachments (content tools)
```

## The Core Concept: Non-Obvious Insight

Every great CLI has a **Non-Obvious Insight** (NOI) - a one-sentence reframe of what the API's data actually IS, not what the API thinks it is. The NOI is what separates a tool from a weapon.

```go
// internal/vision/insight.go

type NonObviousInsight struct {
    // The API thinks it is...
    ObviousFrame    string  // "an issue tracker"

    // But it's actually...
    InsightFrame    string  // "a team behavior observatory"

    // Every [data point] is a signal about...
    SignalSentence  string  // "Every state change, reassignment, stale ticket, and
                            //  carried-over story is a signal about how your team
                            //  actually works vs. how they think they work."

    // Which means the CLI should...
    Implications    []string // ["detect behavioral patterns", "predict staleness",
                             //  "measure workspace entropy", "surface bottlenecks"]
}
```

The NOI drives everything downstream:
- **What the store schema captures** (signals, not just records)
- **What workflow commands do** (surface insights, not just filter data)
- **What the README says** (frame the tool as insight, not wrapper)
- **What the Rung 5 commands detect** (patterns the API never designed for)

### How NOI Gets Generated

The NOI follows a formula that the LLM (or human) fills during Phase 0:

```
"[API name] isn't just [obvious thing]. It's [non-obvious thing].
 Every [data point] is a signal about [hidden truth].
 Which means the CLI should [list of implications]."
```

Examples:

| API | Formula |
|---|---|
| Discord | "Discord isn't just a chat app. It's a **searchable knowledge base**. Every message thread is a signal about what your community actually knows. The CLI should archive, index, and surface that knowledge." |
| Linear | "Linear isn't just an issue tracker. It's a **team behavior observatory**. Every state change, reassignment, stale ticket, and carried-over story is a signal about how your team actually works vs. how they think they work." |
| Stripe | "Stripe isn't just a payment processor. It's a **business health monitor**. Every failed charge, churn event, and subscription change is a signal about your product-market fit and customer loyalty." |
| GitHub | "GitHub isn't just a code host. It's an **engineering culture fingerprint**. Every review turnaround, merge pattern, and CI failure rate is a signal about how your team ships and where it breaks." |
| Notion | "Notion isn't just a doc editor. It's a **knowledge decay detector**. Every stale page, orphaned database, and untouched workspace is a signal about what your team has forgotten." |
| Slack | "Slack isn't just messaging. It's an **organizational nervous system**. Every thread response time, emoji reaction pattern, and channel silence is a signal about team health and bottlenecks." |

### Where NOI Lives in the Codebase

The NOI becomes a first-class field in the VisionaryPlan:

```go
type VisionaryPlan struct {
    APIName      string
    Identity     APIIdentity
    Insight      NonObviousInsight  // <-- NEW: the creative soul
    Architecture []ArchitectureDecision
    Features     []FeatureIdea
    Workflows    []WorkflowIdea
}
```

And the NOI flows into:
1. **README generation** - The Long description of the root command uses the InsightFrame
2. **Workflow template selection** - The Implications list maps to specific templates
3. **Scorecard** - A new "Insight" dimension (0-10) measures whether the CLI has an NOI and acts on it
4. **Phase 0 artifact** - The visionary research document must contain the NOI or it fails the phase gate

### The Phase 0 Gate Change

Current Phase 0 gate:
> "Verify: API Identity documented, 3+ usage patterns, tool landscape, workflows, architecture, features"

New Phase 0 gate:
> "Verify ALL of these:
> 1. **Non-Obvious Insight written** - one sentence that reframes what the API actually is
> 2. **3+ implications derived** from the NOI that map to specific commands
> 3. API Identity documented with data profile
> ...rest of existing gates..."

If Phase 0 ends without an NOI, the resulting CLI will be a competent API wrapper. The NOI is what makes it a weapon.

## Proposed Solution

### Four Changes

**Change 0: Non-Obvious Insight System** - Make the NOI a first-class concept in the VisionaryPlan, the Phase 0 gate, the README template, and the scorecard.

**Change 1: Domain Archetype System** - Teach the profiler to classify APIs into domain archetypes, each with a predefined set of workflow templates.

**Change 2: Workflow Template Library** - Create `.go.tmpl` files for domain-specific workflow commands that the generator renders automatically, just like it renders `sync.go.tmpl` today.

**Change 3: Behavioral Insight Templates** - Add Rung 5 commands that go beyond analytics into prediction and pattern recognition. These are driven by the NOI's implications.

---

## Technical Approach

### Change 1: Domain Archetype System

#### 1.1 Extend the Profiler

Add domain archetype detection to `internal/profiler/profiler.go`:

```go
type DomainArchetype string

const (
    ArchetypeCommunication    DomainArchetype = "communication"
    ArchetypeProjectMgmt      DomainArchetype = "project-management"
    ArchetypePayments         DomainArchetype = "payments"
    ArchetypeInfrastructure   DomainArchetype = "infrastructure"
    ArchetypeContent          DomainArchetype = "content"
    ArchetypeCRM              DomainArchetype = "crm"
    ArchetypeDeveloperPlatform DomainArchetype = "developer-platform"
    ArchetypeGeneric          DomainArchetype = "generic"
)

type DomainSignals struct {
    Archetype        DomainArchetype
    HasAssignees     bool   // user_id/assignee/owner fields
    HasDueDates      bool   // due_date/deadline/expires_at fields
    HasPriority      bool   // priority/severity/urgency fields
    HasThreading     bool   // parent_id on message-like resources + reply_to
    HasTransactions  bool   // amount/currency/charge fields
    HasSubscriptions bool   // interval/recurring/plan fields
    HasMedia         bool   // file/attachment/image/video resources
    HasTeams         bool   // team/group/org membership patterns
    HasLabels        bool   // tag/label/category fields
    HasEstimates     bool   // estimate/points/story_points fields
}
```

**Detection heuristics** (extend `Profile()` function):

```go
func detectDomainArchetype(s *spec.APISpec) DomainSignals {
    signals := DomainSignals{}

    // Scan all resources and fields for domain keywords
    for _, resource := range s.Resources {
        name := strings.ToLower(resource.Name)

        // Communication signals
        if containsAny(name, "message", "channel", "thread", "conversation", "chat") {
            signals.HasThreading = true
        }

        // PM signals
        if containsAny(name, "issue", "task", "ticket", "story", "epic", "sprint", "cycle") {
            signals.HasAssignees = fieldExists(resource, "assignee", "assigned_to", "owner")
            signals.HasPriority = fieldExists(resource, "priority", "severity", "urgency")
            signals.HasDueDates = fieldExists(resource, "due_date", "deadline", "target_date")
            signals.HasEstimates = fieldExists(resource, "estimate", "story_points", "points")
        }

        // Payment signals
        if containsAny(name, "charge", "payment", "invoice", "subscription", "plan") {
            signals.HasTransactions = true
        }
    }

    // Classify archetype by signal combination
    switch {
    case signals.HasThreading && !signals.HasAssignees:
        signals.Archetype = ArchetypeCommunication
    case signals.HasAssignees && signals.HasPriority:
        signals.Archetype = ArchetypeProjectMgmt
    case signals.HasTransactions:
        signals.Archetype = ArchetypePayments
    // ... more rules
    default:
        signals.Archetype = ArchetypeGeneric
    }

    return signals
}
```

**Files to modify:**
- `internal/profiler/profiler.go` - Add `DomainSignals` struct and `detectDomainArchetype()` function
- `internal/profiler/profiler.go` - Extend `APIProfile` struct to include `Domain DomainSignals`

#### 1.2 Extend VisionaryPlan

Add domain archetype to the vision plan:

**Files to modify:**
- `internal/vision/vision.go` - Add `Domain DomainSignals` to `VisionaryPlan`
- `internal/generator/vision_templates.go` - Use `plan.Domain.Archetype` in `SelectVisionTemplates()`

---

### Change 2: Workflow Template Library

#### 2.1 New Template Directory

Create `internal/generator/templates/workflows/` with per-archetype templates:

```
templates/workflows/
  pm_stale.go.tmpl           # Project management: stale issue detection
  pm_orphans.go.tmpl         # PM: issues missing assignment/project/cycle
  pm_triage.go.tmpl          # PM: unassigned issues by priority
  pm_load.go.tmpl            # PM: workload distribution per assignee
  pm_overdue.go.tmpl         # PM: issues past due date
  pm_standup.go.tmpl         # PM: recent activity summary
  pm_deps.go.tmpl            # PM: cross-team dependency map
  comm_channel_health.go.tmpl # Communication: channel activity analysis
  comm_message_stats.go.tmpl  # Communication: message volume analytics
  comm_audit_report.go.tmpl   # Communication: audit log analysis
  comm_member_report.go.tmpl  # Communication: member composition
  pay_reconcile.go.tmpl       # Payments: transaction reconciliation
  pay_revenue.go.tmpl         # Payments: revenue summary
  pay_churn.go.tmpl           # Payments: subscription churn detection
  infra_health.go.tmpl        # Infrastructure: service health dashboard
  infra_incidents.go.tmpl     # Infrastructure: incident timeline
```

#### 2.2 Template Design: Parameterized SQL

The key insight: workflow commands are **parameterized SQL queries** on the local SQLite database. The template needs:
1. The entity names (from the spec)
2. The field names (from the spec)
3. The join relationships (from the profiler's `HasDependencies`)
4. The state types (from the profiler's `HasLifecycles`)

Example template for `pm_stale.go.tmpl`:

```go
{{/* pm_stale.go.tmpl - Stale issue detection for project management APIs */}}
package cli

import (
    "encoding/json"
    "fmt"
    "time"
    "github.com/spf13/cobra"
)

func newStaleCmd(flags *rootFlags) *cobra.Command {
    var days int
    var team string

    cmd := &cobra.Command{
        Use:   "stale",
        Short: "Find {{.PrimaryEntity.HumanPlural}} with no updates in N days",
        Long: `Queries the local database for {{.PrimaryEntity.HumanPlural}} that have not been
updated in the specified number of days and are not in a terminal state.`,
        Example: `  {{.CLIName}} stale --days 30
  {{.CLIName}} stale --days 14 {{- if .Domain.HasTeams}} --team ENG{{end}}
  {{.CLIName}} stale --days 7 --json`,
        RunE: func(cmd *cobra.Command, args []string) error {
            s, err := flags.openStore()
            if err != nil {
                return err
            }
            defer s.Close()

            cutoff := time.Now().UTC().AddDate(0, 0, -days).Format(time.RFC3339)

            q := `SELECT {{.PrimaryEntity.IdentifierField}}, {{.PrimaryEntity.TitleField}},
                CAST(julianday('now') - julianday({{.PrimaryEntity.UpdatedAtField}}) AS INTEGER) AS days_stale
                {{- if .Domain.HasTeams}}, COALESCE(t.{{.TeamEntity.KeyField}}, '') AS team{{end}}
                {{- if .HasAssignees}}, COALESCE(u.{{.UserEntity.NameField}}, 'unassigned') AS assignee{{end}}
            FROM {{.PrimaryEntity.TableName}} i
            {{- if .HasLifecycleStates}}
            JOIN {{.StateEntity.TableName}} ws ON i.{{.PrimaryEntity.StateFK}} = ws.id
            {{- end}}
            {{- if .Domain.HasTeams}}
            LEFT JOIN {{.TeamEntity.TableName}} t ON i.{{.PrimaryEntity.TeamFK}} = t.id
            {{- end}}
            {{- if .HasAssignees}}
            LEFT JOIN {{.UserEntity.TableName}} u ON i.{{.PrimaryEntity.AssigneeFK}} = u.id
            {{- end}}
            WHERE i.{{.PrimaryEntity.UpdatedAtField}} < ?
            {{- if .HasLifecycleStates}}
            AND ws.{{.StateEntity.TypeField}} NOT IN ({{.TerminalStateTypes}})
            {{- end}}`
            // ... execute query, format output
        },
    }

    cmd.Flags().IntVar(&days, "days", 30, "Number of days without updates")
    {{- if .Domain.HasTeams}}
    cmd.Flags().StringVar(&team, "team", "", "Filter by team key")
    {{- end}}

    return cmd
}
```

**The template variables come from the profiler + spec mapping:**

```go
type WorkflowTemplateContext struct {
    CLIName        string
    Domain         DomainSignals
    PrimaryEntity  EntityMapping  // the main entity (issues, messages, transactions)
    TeamEntity     EntityMapping  // team/guild/org entity
    UserEntity     EntityMapping  // user/member entity
    StateEntity    EntityMapping  // workflow state entity
    // ... field mappings
}

type EntityMapping struct {
    TableName       string  // "issues"
    HumanSingular   string  // "issue"
    HumanPlural     string  // "issues"
    IdentifierField string  // "identifier" or "id"
    TitleField      string  // "title" or "name" or "subject"
    UpdatedAtField  string  // "updated_at"
    StateFK         string  // "state_id"
    TeamFK          string  // "team_id"
    AssigneeFK      string  // "assignee_id"
    KeyField        string  // "key" (for teams)
    NameField       string  // "name" (for users)
    TypeField       string  // "type" (for states)
}
```

#### 2.3 Entity Mapping Engine

New file: `internal/generator/entity_mapper.go`

This maps API spec resources to template variables:

```go
func MapEntities(s *spec.APISpec, domain DomainSignals) WorkflowTemplateContext {
    ctx := WorkflowTemplateContext{Domain: domain}

    // Find the primary entity (highest volume, most fields)
    ctx.PrimaryEntity = findPrimaryEntity(s, domain)

    // Find supporting entities by role
    ctx.TeamEntity = findEntityByRole(s, "team", "guild", "org", "workspace")
    ctx.UserEntity = findEntityByRole(s, "user", "member", "person", "account")
    ctx.StateEntity = findEntityByRole(s, "state", "status", "workflow_state")

    // Map field names
    ctx.PrimaryEntity.TitleField = findField(s, ctx.PrimaryEntity.Name,
        "title", "name", "subject", "summary")
    ctx.PrimaryEntity.UpdatedAtField = findField(s, ctx.PrimaryEntity.Name,
        "updated_at", "updatedAt", "modified_at", "last_modified")
    // ... more field mappings

    return ctx
}
```

**Files to create:**
- `internal/generator/entity_mapper.go` - Entity role detection and field mapping
- `internal/generator/templates/workflows/pm_stale.go.tmpl`
- `internal/generator/templates/workflows/pm_orphans.go.tmpl`
- `internal/generator/templates/workflows/pm_triage.go.tmpl`
- `internal/generator/templates/workflows/pm_load.go.tmpl`
- `internal/generator/templates/workflows/pm_overdue.go.tmpl`
- `internal/generator/templates/workflows/pm_standup.go.tmpl`
- `internal/generator/templates/workflows/pm_deps.go.tmpl`
- `internal/generator/templates/workflows/comm_channel_health.go.tmpl`
- `internal/generator/templates/workflows/comm_message_stats.go.tmpl`
- `internal/generator/templates/workflows/comm_audit_report.go.tmpl`

#### 2.4 Extend the Generator

Modify `internal/generator/generator.go` to render workflow templates:

```go
func (g *Generator) Generate() error {
    // ... existing generation ...

    // NEW: Render domain-specific workflow commands
    if g.profile.Domain.Archetype != ArchetypeGeneric {
        ctx := MapEntities(g.Spec, g.profile.Domain)
        for _, tmpl := range g.workflowTemplates() {
            g.renderTemplate(tmpl, outPath, ctx)
        }
    }
}

func (g *Generator) workflowTemplates() []string {
    switch g.profile.Domain.Archetype {
    case ArchetypeProjectMgmt:
        templates := []string{"pm_stale.go.tmpl", "pm_orphans.go.tmpl", "pm_triage.go.tmpl"}
        if g.profile.Domain.HasAssignees {
            templates = append(templates, "pm_load.go.tmpl", "pm_standup.go.tmpl")
        }
        if g.profile.Domain.HasDueDates {
            templates = append(templates, "pm_overdue.go.tmpl")
        }
        if g.profile.Domain.HasDependencies {
            templates = append(templates, "pm_deps.go.tmpl")
        }
        return templates
    case ArchetypeCommunication:
        return []string{"comm_channel_health.go.tmpl", "comm_message_stats.go.tmpl", "comm_audit_report.go.tmpl"}
    case ArchetypePayments:
        return []string{"pay_reconcile.go.tmpl", "pay_revenue.go.tmpl"}
    default:
        return nil
    }
}
```

**Files to modify:**
- `internal/generator/generator.go` - Add `workflowTemplates()` and workflow rendering loop
- `internal/generator/vision_templates.go` - Register workflow templates in `VisionTemplateSet`

---

### Change 3: Behavioral Insight Templates (Rung 5)

These are the discrawl-level commands that see patterns humans miss. They use the same template system as Change 2 but add predictive/analytical SQL.

#### 3.1 New Templates

```
templates/insights/
  health_score.go.tmpl      # Single-number workspace health (all archetypes)
  bottleneck.go.tmpl         # Where does work get stuck? (PM)
  similar.go.tmpl            # FTS5 duplicate/similarity detection (all with FTS5)
  forecast.go.tmpl           # Completion probability (PM with estimates)
  patterns.go.tmpl           # Per-user/per-label behavioral patterns (all)
  trends.go.tmpl             # Time-series trend detection (all with timestamps)
```

#### 3.2 Example: health_score.go.tmpl

```go
func newHealthCmd(flags *rootFlags) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "health",
        Short: "Composite health score for your workspace (0-100)",
        Long: `Calculates a single health score from multiple signals:
{{- if .HasLifecycleStates}}
- Stale ratio: % of active items with no update in 14+ days
{{- end}}
{{- if .HasAssignees}}
- Orphan ratio: % of items with no assignee
{{- end}}
{{- if .HasDueDates}}
- Overdue ratio: % of items past due date
{{- end}}
- Velocity trend: are you completing more or fewer items this week vs last?
- Entropy: is the workspace getting more or less organized?`,
        RunE: func(cmd *cobra.Command, args []string) error {
            s, err := flags.openStore()
            // ... compute each signal as 0-100 subscore
            // ... weighted average = overall health score
            // ... compare to previous run (store in sync_cursors as "health_history")
            // ... output: "Health: 72/100 (down 4 from last week)"
        },
    }
    return cmd
}
```

#### 3.3 Example: bottleneck.go.tmpl

```go
func newBottleneckCmd(flags *rootFlags) *cobra.Command {
    // For PM archetypes:
    // 1. For each workflow state, compute avg time items spend in that state
    // 2. Identify the state with longest avg dwell time
    // 3. Compare to other states as a ratio
    // Output: "Items spend 4.2x longer in 'In Review' than 'In Progress'.
    //          Your review process is the bottleneck."
    //
    // SQL: Uses completed_at - started_at for cycle time,
    //      groups by state transitions (requires issue_history or diff analysis)
}
```

#### 3.4 Example: similar.go.tmpl

```go
func newSimilarCmd(flags *rootFlags) *cobra.Command {
    // For any archetype with FTS5:
    // 1. For each item, query FTS5 for similar items by title
    // 2. Score similarity using FTS5 rank
    // 3. Group potential duplicates
    // Output: "ENG-456 may be a duplicate of ENG-123 (87% title match)"
    //
    // SQL: SELECT a.identifier, b.identifier,
    //        bm25(issues_fts) as similarity
    //      FROM issues_fts
    //      JOIN issues a ON ...
    //      WHERE issues_fts MATCH a.title
}
```

**Files to create:**
- `internal/generator/templates/insights/health_score.go.tmpl`
- `internal/generator/templates/insights/bottleneck.go.tmpl`
- `internal/generator/templates/insights/similar.go.tmpl`
- `internal/generator/templates/insights/forecast.go.tmpl`
- `internal/generator/templates/insights/patterns.go.tmpl`
- `internal/generator/templates/insights/trends.go.tmpl`

---

### Change 4: Store Schema Generation Upgrade

The current `store.go.tmpl` creates generic tables:

```sql
CREATE TABLE IF NOT EXISTS resources (
    id TEXT PRIMARY KEY,
    resource_type TEXT NOT NULL,
    data JSON NOT NULL  -- generic JSON blob
)
```

This needs to become domain-aware. The Linear CLI already hand-wrote domain tables (issues with 22 columns, proper foreign keys, indexes). The generator should do this automatically.

#### 4.1 Schema Generation from Spec

New file: `internal/generator/schema_builder.go`

```go
func BuildSchema(s *spec.APISpec, domain DomainSignals) []TableDef {
    var tables []TableDef

    for _, resource := range s.Resources {
        gravity := computeDataGravity(resource, s)
        if gravity < 5 {
            continue  // API-only, no local table
        }

        table := TableDef{
            Name: snake(resource.Name),
            Columns: []ColumnDef{
                {Name: "id", Type: "TEXT PRIMARY KEY"},
                {Name: "data", Type: "JSON NOT NULL"},
            },
        }

        // Extract columns from response schema
        for _, field := range resource.ResponseFields {
            if isScalar(field.Type) {
                table.Columns = append(table.Columns, ColumnDef{
                    Name: snake(field.Name),
                    Type: sqlType(field.Type),
                })
            }
            if isFK(field, s) {
                table.ForeignKeys = append(table.ForeignKeys, ...)
                table.Indexes = append(table.Indexes, ...)
            }
        }

        // Add FTS5 if text-heavy
        if countTextFields(resource) >= 2 && gravity >= 8 {
            table.FTS5 = true
            table.FTS5Fields = textFieldNames(resource)
        }

        tables = append(tables, table)
    }

    return tables
}
```

#### 4.2 Updated store.go.tmpl

```go
func (s *Store) migrate() error {
    migrations := []string{
        {{- range .Tables}}
        `CREATE TABLE IF NOT EXISTS {{.Name}} (
            {{- range .Columns}}
            {{.Name}} {{.Type}}{{if not (last .)},{{end}}
            {{- end}}
        )`,
        {{- range .Indexes}}
        `CREATE INDEX IF NOT EXISTS {{.Name}} ON {{.TableName}}({{.Columns}})`,
        {{- end}}
        {{- if .FTS5}}
        `CREATE VIRTUAL TABLE IF NOT EXISTS {{.Name}}_fts USING fts5(
            {{join .FTS5Fields ", "}},
            content='{{.Name}}', content_rowid='rowid',
            tokenize='porter unicode61'
        )`,
        // ... FTS5 triggers
        {{- end}}
        {{- end}}
    }
}
```

**Files to create:**
- `internal/generator/schema_builder.go`

**Files to modify:**
- `internal/generator/templates/store.go.tmpl` - Use `{{range .Tables}}` instead of hardcoded tables

---

## Implementation Phases

### Phase 0: Non-Obvious Insight System (The Soul)

**Effort:** 2 hours

- [ ] Create `internal/vision/insight.go` with `NonObviousInsight` struct
- [ ] Add `Insight NonObviousInsight` field to `VisionaryPlan` struct in `internal/vision/vision.go`
- [ ] Add NOI section to `vision.go` template output (the `printing-press vision` command)
- [ ] Update SKILL.md Phase 0 gate to require NOI before proceeding
- [ ] Update `root.go.tmpl` to use `InsightFrame` in the root command's Long description
- [ ] Update README template to include NOI as the first paragraph of "Why This CLI?"
- [ ] Add "Insight" dimension (0-10) to scorecard in `internal/pipeline/scorecard.go`
- [ ] Create `skills/printing-press/references/noi-examples.md` with 10 example NOIs across different archetypes
- [ ] Test: run `printing-press vision --api Discord` and verify NOI template section appears

**NOI Examples Reference File** (`references/noi-examples.md`):

```markdown
# Non-Obvious Insight Examples

## Formula
"[API] isn't just [obvious]. It's [non-obvious].
Every [data point] is a signal about [hidden truth]."

## Project Management
- Linear: "...team behavior observatory...every state change is a signal"
- Jira: "...process archaeology site...every custom field is scar tissue from a past failure"
- Asana: "...work visibility engine...every completed task is proof of coordination cost"

## Communication
- Discord: "...searchable knowledge base...every thread is institutional memory"
- Slack: "...organizational nervous system...every response time is a health signal"

## Payments
- Stripe: "...business health monitor...every failed charge is a product-market signal"
- Plaid: "...financial behavior fingerprint...every transaction pattern is a life event"

## Infrastructure
- GitHub: "...engineering culture fingerprint...every review turnaround is a team signal"
- Datadog: "...incident prediction engine...every anomaly is a future outage"

## Content
- Notion: "...knowledge decay detector...every stale page is forgotten context"
- Confluence: "...documentation debt ledger...every outdated doc is a lie"
```

### Phase 1: Domain Archetype System (Foundation)

**Effort:** 2-3 hours

- [ ] Add `DomainSignals` struct to `internal/profiler/profiler.go`
- [ ] Implement `detectDomainArchetype()` with keyword heuristics
- [ ] Add domain signals to `APIProfile` struct
- [ ] Add domain to `VisionaryPlan` in `internal/vision/vision.go`
- [ ] Add archetype to scorecard output
- [ ] Test: run profiler on Discord spec, verify `ArchetypeCommunication`
- [ ] Test: run profiler on Linear spec (if available), verify `ArchetypeProjectMgmt`

### Phase 2: Entity Mapping Engine (Plumbing)

**Effort:** 2-3 hours

- [ ] Create `internal/generator/entity_mapper.go`
- [ ] Implement `MapEntities()` - find primary entity, team, user, state by role
- [ ] Implement field mapping (title, updated_at, assignee, etc.)
- [ ] Create `WorkflowTemplateContext` struct
- [ ] Test: map Discord spec entities (guild=team, user=user, message=primary)
- [ ] Test: map a Stripe-like spec (charge=primary, customer=user)

### Phase 3: Workflow Templates (The Product)

**Effort:** 4-6 hours

- [ ] Create `templates/workflows/` directory
- [ ] Write PM templates: `pm_stale.go.tmpl`, `pm_orphans.go.tmpl`, `pm_triage.go.tmpl`, `pm_load.go.tmpl`, `pm_overdue.go.tmpl`, `pm_standup.go.tmpl`, `pm_deps.go.tmpl`
- [ ] Write Communication templates: `comm_channel_health.go.tmpl`, `comm_message_stats.go.tmpl`, `comm_audit_report.go.tmpl`
- [ ] Extend `generator.go` to render workflow templates based on archetype
- [ ] Register workflow commands in `root.go.tmpl`
- [ ] Test: generate a PM CLI, verify 7 workflow commands exist and compile
- [ ] Test: generate a Communication CLI, verify 3 workflow commands exist

### Phase 4: Behavioral Insight Templates (The Magic)

**Effort:** 3-4 hours

- [ ] Create `templates/insights/` directory
- [ ] Write `health_score.go.tmpl` (universal)
- [ ] Write `similar.go.tmpl` (FTS5-dependent)
- [ ] Write `bottleneck.go.tmpl` (PM + lifecycle states)
- [ ] Write `trends.go.tmpl` (any with timestamps)
- [ ] Write `patterns.go.tmpl` (any with assignees)
- [ ] Write `forecast.go.tmpl` (PM + estimates)
- [ ] Extend `generator.go` to render insight templates
- [ ] Test: generate a PM CLI with insights, verify `health`, `similar`, `bottleneck` exist

### Phase 5: Schema Generation Upgrade

**Effort:** 2-3 hours

- [ ] Create `internal/generator/schema_builder.go`
- [ ] Implement `BuildSchema()` with data gravity scoring
- [ ] Implement `computeDataGravity()` heuristic
- [ ] Update `store.go.tmpl` to iterate over generated tables
- [ ] Add FTS5 triggers to template
- [ ] Test: generate store for Discord spec, verify domain-specific tables (not generic JSON blobs)

### Phase 6: Scorecard Update

**Effort:** 1 hour

- [ ] Update scorecard to detect workflow commands by archetype match (not just file naming)
- [ ] Add "Insight" dimension (0-10) for Rung 5 commands
- [ ] Update grade thresholds for 12 dimensions (120 max)

## Acceptance Criteria

- [ ] `printing-press generate --spec linear.yaml` produces 7 PM workflow commands without Phase 4 hand-writing
- [ ] `printing-press generate --spec discord.yaml` produces 3 communication workflow commands
- [ ] Generated workflow commands compile (`go build ./...`)
- [ ] Generated `stale` command for any PM API uses proper entity names from the spec (not hardcoded "issues")
- [ ] `health` command generates for all archetypes and produces a composite score
- [ ] `similar` command generates when FTS5 is enabled
- [ ] Store schema uses domain-specific columns (not generic JSON blobs) for high-gravity entities
- [ ] Scorecard detects generated workflow commands and scores them

## What This Means for the Skill

### Phase 0 changes

The Phase 0 gate adds one mandatory output: the **Non-Obvious Insight**.

Before Phase 0 can complete, the LLM must write:

> "[API name] isn't just [obvious thing]. It's [non-obvious thing].
> Every [data point] is a signal about [hidden truth]."

This sentence becomes the creative DNA of the entire CLI. It informs which workflow templates get rendered, what the README says, and what Rung 5 insight commands get built.

If the LLM can't write an NOI, it means Phase 0 research wasn't deep enough. Go back and read more Reddit/HN posts until the insight emerges.

### Phase 4 changes

Phase 4 changes from:

> "Hand-write 7 workflow commands using the GraphQL schema + Phase 0.5 workflows"

To:

> "Review the 7 generated workflow commands. Customize SQL queries if the domain has unique patterns. Then write 2-3 NOI-driven insight commands that embody the Non-Obvious Insight - these are the commands that make someone say 'I never thought to use [API] this way.'"

Phase 4 shrinks from "write 7 Go files" to "review generated files, write 2-3 creative commands." The LLM's time shifts from mechanical Go coding to creative domain analysis - which is what it's actually good at.

### The NOI drives Phase 4's creative commands

| NOI | Implies these Rung 5 commands |
|---|---|
| "team behavior observatory" | `health` (composite score), `bottleneck` (where work stalls), `patterns` (per-user behavior), `forecast` (sprint completion probability) |
| "searchable knowledge base" | `similar` (duplicate detection), `knowledge-map` (topic clustering), `decay` (stale knowledge detection) |
| "business health monitor" | `pulse` (revenue health), `churn-risk` (at-risk customers), `cohort` (behavioral cohorts) |
| "engineering culture fingerprint" | `culture` (review speed, merge patterns), `bus-factor` (knowledge concentration), `momentum` (velocity trends) |

## System-Wide Impact

### What changes for every future CLI generation

1. **Profiler** now classifies domain archetype automatically
2. **Generator** renders 5-10 workflow + insight commands per archetype
3. **Store** produces domain-specific SQLite tables (not JSON blobs)
4. **Scorecard** measures workflow and insight quality
5. **Phase 4** shifts from "write workflows" to "customize and extend workflows"

### What doesn't change

- Phase 0-1 research still runs (discovery can't be automated)
- Phase 2 spec parsing is unchanged
- Phase 3 audit still runs scorecard
- Phase 4.5 dogfood still validates
- Phase 5 still produces final report

### Risk: Over-generation

If the archetype detection is wrong (classifies a CRM as PM), the generated workflow commands will be wrong. Mitigation: the skill's Phase 3 audit catches this and Phase 4 corrects it.

## Sources

- Discord-cli generated code: `~/cli-printing-press/discord-cli/internal/cli/` (320+ commands, 8 workflow files)
- Linear-cli hand-written code: `~/cli-printing-press/linear-cli/internal/cli/` (7 workflow files, all hand-written)
- Profiler: `~/cli-printing-press/internal/profiler/profiler.go` (signal detection logic)
- Generator: `~/cli-printing-press/internal/generator/generator.go` (template rendering)
- Vision templates: `~/cli-printing-press/internal/generator/vision_templates.go` (feature selection)
- Entity mapper pattern: `~/cli-printing-press/internal/generator/templates/store.go.tmpl` (current generic schema)
- discrawl reference: github.com/discrawl (12 commands, 539 stars - the bar)
