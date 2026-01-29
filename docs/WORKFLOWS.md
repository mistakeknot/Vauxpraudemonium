# Autarch Workflows

> Step-by-step guides for common tasks across Autarch tools

This guide covers end-user workflows for the Autarch toolset. For developer documentation, see the tool-specific AGENTS.md files.

---

## Quick Start: Which Tool Do I Need?

| I want to... | Use | Command |
|--------------|-----|---------|
| Create a product spec | **Gurgeh** | `./dev gurgeh` |
| Plan and track implementation tasks | **Coldwine** | `./dev coldwine` |
| Research competitors, trends, papers | **Pollard** | `go run ./cmd/pollard` |
| Monitor all my AI agents | **Bigend** | `./dev bigend` |

---

## Workflow 1: Creating a New PRD (Gurgeh)

**Goal:** Create a product requirements document through guided interview.

### Step 1: Initialize Gurgeh

```bash
cd your-project
./dev gurgeh init
```

This creates `.gurgeh/` with:
- `specs/` - Where PRDs are stored
- `research/` - Market research outputs
- `suggestions/` - Staged updates for review
- `config.toml` - Agent configuration

### Step 2: Launch the TUI

```bash
./dev gurgeh
```

### Step 3: Start a New PRD

1. Press `n` to create a new PRD
2. The interview wizard guides you through:
   - **Product vision** - What are you building?
   - **Problem statement** - What problem does it solve?
   - **Target users** - Who is this for?
   - **Critical User Journeys (CUJs)** - Key workflows users will perform
   - **Requirements** - Specific features needed
   - **Files to modify** - Which codebase files will change

### Step 4: Review and Save

1. Review the generated PRD summary
2. Press `Enter` to save
3. PRD is written to `.gurgeh/specs/PRD-001.yaml` (auto-incremented)
4. Git auto-commits the new spec

### Step 5: Iterate

From the TUI:
- `e` - Edit existing PRD
- `r` - Refresh list
- `d` - Delete PRD (with confirmation)
- `?` - Show all keybindings

### Output

```
.gurgeh/
├── specs/
│   └── PRD-001.yaml    # Your new PRD
├── research/           # Market research (if run)
└── briefs/             # Agent briefs (timestamped)
```

---

## Workflow 2: Running Research (Pollard)

**Goal:** Gather competitive landscape, trends, and academic research.

### Step 1: Initialize Pollard

```bash
cd your-project
go run ./cmd/pollard init
```

This creates `.pollard/` with default configuration.

### Step 2: Configure Hunters (Optional)

Edit `.pollard/config.yaml` to customize:

```yaml
# Enable/disable hunters
github-scout:
  enabled: true
  queries:
    - "your search terms"
  min_stars: 50

hackernews:
  enabled: true
  queries:
    - "industry keywords"
  min_points: 30
```

### Step 3: Run a Scan

```bash
# Run all enabled hunters
go run ./cmd/pollard scan

# Run specific hunter
go run ./cmd/pollard scan --hunter github-scout

# Dry run (see what would happen)
go run ./cmd/pollard scan --dry-run
```

### Step 4: Generate Reports

```bash
# Full landscape report
go run ./cmd/pollard report

# Specific report types
go run ./cmd/pollard report --type competitive
go run ./cmd/pollard report --type trends
go run ./cmd/pollard report --type research

# Output to terminal
go run ./cmd/pollard report --stdout
```

### Step 5: View Results

Reports are saved to `.pollard/reports/`. Raw data is in `.pollard/sources/`.

### Pro Tips

- Set `GITHUB_TOKEN` for higher rate limits on GitHub searches
- Use `--hunter openalex` for cross-disciplinary academic research
- Run scans periodically to track changes over time

---

## Workflow 3: Planning Tasks from PRDs (Coldwine)

**Goal:** Break down PRDs into epics, stories, and tasks for implementation.

### Step 1: Ensure PRDs Exist

Coldwine reads from `.gurgeh/specs/`. Create PRDs first with Gurgeh.

### Step 2: Initialize Coldwine

```bash
cd your-project
./dev coldwine init
```

This:
- Creates `.coldwine/` directory
- Scans codebase structure
- Generates initial epic proposals from PRDs

### Step 3: Review Generated Epics

```bash
./dev coldwine
```

In the TUI:
1. Navigate to **Epic Review** tab
2. Review each proposed epic
3. `A` - Accept all proposals
4. `e` - Edit individual epics
5. `d` - Delete unwanted epics
6. `R` - Regenerate proposals (if needed)

### Step 4: Generate Stories and Tasks

After accepting epics:
1. Navigate to **Task Review** tab
2. Review proposed stories and tasks
3. `A` - Accept all
4. `e` - Edit individual items
5. `tab` - Cycle task type (story/task/bug)

### Step 5: Start Working on Tasks

```bash
# From TUI: select a task and press 's' to start

# Or from CLI:
./dev coldwine start <task-id>
```

This:
- Creates a git worktree for isolation
- Sets task status to `in_progress`
- Opens terminal in the worktree

### Step 6: Complete Tasks

```bash
# From TUI: press 'c' to complete

# Or from CLI:
./dev coldwine complete <task-id>
```

### Task States

```
todo → in_progress → review → done
         ↓
       blocked
```

---

## Workflow 4: Monitoring Agents (Bigend)

**Goal:** See all your AI agents across projects in one dashboard.

### Step 1: Configure Scan Roots

Edit `~/.config/bigend/config.toml`:

```toml
[discovery]
scan_roots = ["~/projects", "/path/to/more/projects"]

[server]
port = 8099
```

### Step 2: Start Bigend

```bash
# Web dashboard
./dev bigend

# Or TUI mode
./dev bigend --tui
```

### Step 3: Access Dashboard

**Web:** Open http://localhost:8099

**TUI:** Navigate with `j`/`k`, press `Enter` to view details.

### What You'll See

- **Projects** - All discovered projects with Gurgeh/Coldwine/Pollard status
- **Agents** - Registered AI agents (via Intermute)
- **Sessions** - Active tmux sessions with agent detection
- **Activity** - Recent events across all projects

### Agent States

| State | Meaning |
|-------|---------|
| `working` | Actively processing |
| `waiting` | Waiting for user input |
| `blocked` | Error or stuck |
| `stalled` | No activity for extended period |
| `done` | Task completed |

### Session Actions (TUI)

- `Enter` - View session details
- `a` - Attach to session (opens tmux)
- `k` - Kill session (with confirmation)
- `r` - Refresh

---

## Workflow 5: Setting Up a New Project

**Goal:** Initialize all Autarch tools for a new project.

### Full Setup

```bash
cd your-new-project

# 1. Initialize Gurgeh for PRDs
./dev gurgeh init

# 2. Initialize Pollard for research
go run ./cmd/pollard init

# 3. Create your first PRD
./dev gurgeh
# Press 'n', complete interview

# 4. Run initial research
go run ./cmd/pollard scan

# 5. Initialize Coldwine for task planning
./dev coldwine init

# 6. Start monitoring (in separate terminal)
./dev bigend
```

### Minimal Setup

If you only need one tool:

```bash
# Just PRDs
./dev gurgeh init

# Just research
go run ./cmd/pollard init

# Just task tracking
./dev coldwine init
```

---

## Workflow 6: Research-Driven PRD Creation

**Goal:** Use research to inform PRD creation.

### Step 1: Run Research First

```bash
go run ./cmd/pollard init
go run ./cmd/pollard scan --hunter github-scout --hunter hackernews
go run ./cmd/pollard report --type landscape
```

### Step 2: Review Research

Read the generated report:
```bash
cat .pollard/reports/landscape-*.md
```

### Step 3: Create PRD with Research Context

```bash
./dev gurgeh
```

When creating the PRD:
- Reference competitive findings
- Include discovered user pain points
- Link to relevant research in requirements

### Step 4: Link Research to PRD (Programmatic)

Research insights can be linked to PRD features:

```bash
# From Coldwine, insights are shown alongside tasks
./dev coldwine
# View task details to see related research
```

---

## Workflow 7: Agent Coordination (Multi-Agent)

**Goal:** Have multiple AI agents work on different tasks simultaneously.

### Step 1: Set Up Intermute

Ensure Intermute is running (see Intermute docs).

Set environment:
```bash
export INTERMUTE_URL=http://localhost:8080
export INTERMUTE_PROJECT=my-project
```

### Step 2: Start Multiple Agents

```bash
# Terminal 1: Start Bigend to monitor
./dev bigend

# Terminal 2: Agent 1 works on Task A
./dev coldwine start task-001
claude  # or codex

# Terminal 3: Agent 2 works on Task B
./dev coldwine start task-002
claude  # or codex
```

### Step 3: Monitor Coordination

In Bigend:
- See both agents in the Agents tab
- View message threads between agents
- Check file reservations (prevents conflicts)

### Step 4: Handle Conflicts

If agents need the same file:
1. Bigend shows reservation conflicts
2. First agent has priority
3. Second agent waits or works on different area
4. Reservations auto-expire when task completes

---

## Common Issues

### "No projects found" in Bigend

- Check `scan_roots` in config
- Ensure projects have `.gurgeh/`, `.coldwine/`, or `.pollard/` directories
- Run `./dev bigend --scan-root /your/path` to test

### "Rate limited" in Pollard

- Set `GITHUB_TOKEN` for GitHub searches
- Wait and retry (rate limits reset hourly)
- Use `--dry-run` to check before running

### Tasks not showing in Coldwine

- Ensure `.gurgeh/specs/` has PRD files
- Run `./dev coldwine init` to regenerate
- Check `.coldwine/specs/` for generated content

### Agents not detected in Bigend

- Session names should contain "claude" or "codex"
- Check tmux session list: `tmux ls`
- Verify Intermute is running if using agent coordination

---

## Keyboard Shortcuts Cheat Sheet

### Universal

| Key | Action |
|-----|--------|
| `?` | Show help |
| `q` | Quit |
| `j`/`k` | Navigate down/up |
| `Enter` | Select/confirm |
| `Esc` | Cancel/back |

### Gurgeh

| Key | Action |
|-----|--------|
| `n` | New PRD |
| `e` | Edit PRD |
| `d` | Delete PRD |
| `r` | Refresh |

### Coldwine

| Key | Action |
|-----|--------|
| `A` | Accept all proposals |
| `s` | Start task |
| `c` | Complete task |
| `g` | Toggle grouped view |

### Bigend (TUI)

| Key | Action |
|-----|--------|
| `a` | Attach to session |
| `k` | Kill session |
| `Tab` / `Ctrl+Left/Right` | Switch tabs |

---

## Next Steps

- **Developer docs:** See `AGENTS.md` in each tool's docs folder
- **Architecture:** See `docs/ARCHITECTURE.md`
- **Quick reference:** See `docs/QUICK_REFERENCE.md`
- **Integration details:** See `docs/INTEGRATION.md`
