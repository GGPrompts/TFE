# Multi-Claude Orchestration Pattern

**The Big Idea:** One orchestrator Claude coordinates multiple worker Claude sessions using tmux + Desktop Commander.

## Why This Is Powerful

A single orchestrator Claude can:
- ✅ **See** what other Claude sessions are doing (tmux capture-pane)
- ✅ **Control** other Claude sessions (tmux send-keys)
- ✅ **Monitor** processes those Claudes created (Desktop Commander)
- ✅ **Coordinate** complex multi-session workflows
- ✅ **Synthesize** results from parallel work streams

## How It Works

### The Architecture

```
┌───────────────────────────────────────────────────────────┐
│ Orchestrator Claude Session                              │
│ - Uses Desktop Commander's execute_command                │
│ - Runs tmux commands to control worker sessions           │
│ - Monitors progress and coordinates dependencies          │
└─────────────────┬─────────────────┬───────────────────────┘
                  │                 │
        ┌─────────┴────────┐       │
        ▼                  ▼       ▼
┌─────────────┐   ┌─────────────┐   ┌─────────────┐
│ Worker      │   │ Worker      │   │ Worker      │
│ Claude A    │   │ Claude B    │   │ Claude C    │
│ (Frontend)  │   │ (Backend)   │   │ (Testing)   │
└─────────────┘   └─────────────┘   └─────────────┘
```

### Technical Details

**Desktop Commander Tools Used:**
- `execute_command` - Run tmux commands to control sessions
- `read_file` - Read files created by worker Claudes
- `write_file` - Leave coordination notes

**Tmux Commands the Orchestrator Runs:**

1. **Create worker sessions:**
   ```bash
   tmux new-session -d -s worker-name "claude"
   ```

2. **Send tasks to workers:**
   ```bash
   tmux send-keys -t worker-name "Your task here" C-m
   ```

3. **Monitor worker progress:**
   ```bash
   tmux capture-pane -t worker-name -p -S -100
   ```

4. **List all workers:**
   ```bash
   tmux list-sessions
   tmux list-panes -t session-name
   ```

5. **Auto-compact a worker:**
   ```bash
   auto-compact -t worker-name -g "Next phase of your task"
   ```

## Real-World Use Cases

### Use Case 1: Large Feature Development

**Scenario:** Implement user authentication (UI + API + tests)

**Orchestration:**
```
Orchestrator:
├─ Creates 3 worker Claudes
├─ Assigns tasks:
│  ├─ Frontend Claude: Login UI (ui/login.go)
│  ├─ Backend Claude: Auth API (auth/handler.go)
│  └─ Test Claude: Wait for completion
├─ Monitors progress every 2 minutes
├─ When frontend + backend complete:
│  └─ Coordinates integration
│     └─ Sends API signatures to frontend
│     └─ Tells test Claude to begin
└─ Synthesizes final report
```

**Commands the orchestrator would run:**
```bash
# Create workers
execute_command("tmux new-session -d -s auth-frontend 'claude'")
execute_command("tmux new-session -d -s auth-backend 'claude'")
execute_command("tmux new-session -d -s auth-tests 'claude'")

# Assign tasks
execute_command("tmux send-keys -t auth-frontend 'Create login UI...' C-m")
execute_command("tmux send-keys -t auth-backend 'Create auth API...' C-m")

# Monitor (periodic)
frontend_output = execute_command("tmux capture-pane -t auth-frontend -p")
backend_output = execute_command("tmux capture-pane -t auth-backend -p")

# Coordinate
execute_command("tmux send-keys -t auth-tests 'Test integration...' C-m")
```

---

### Use Case 2: Parallel Research

**Scenario:** Research 4 different UI libraries for TFE

**Orchestration:**
```
Orchestrator:
├─ Creates 4 worker Claudes
├─ Each researches one library:
│  ├─ Worker A: Bubbletea (current)
│  ├─ Worker B: tview
│  ├─ Worker C: termui
│  └─ Worker D: gocui
├─ Each worker creates research doc
├─ Orchestrator reads all 4 docs
└─ Synthesizes comparison and recommendation
```

**Benefit:** Research that would take 40 minutes sequentially completes in ~10 minutes.

---

### Use Case 3: Build Monitoring + Development

**Scenario:** Develop a feature while monitoring build/tests

**Orchestration:**
```
Orchestrator:
├─ Worker Claude A: Implements feature
├─ Monitoring Pane: Runs `go build -o tfe && ./tfe`
├─ Orchestrator watches both:
│  ├─ Sees what Claude is writing
│  ├─ Sees build errors in monitoring pane
│  └─ Coordinates fixes when build breaks
└─ Reports when feature is complete and builds successfully
```

**Commands:**
```bash
# Create development session
execute_command("tmux new-session -d -s dev-feature 'claude'")

# Create monitoring pane in same session
execute_command("tmux split-window -t dev-feature 'watch -n 5 go build'")

# Monitor both
dev_output = execute_command("tmux capture-pane -t dev-feature:0 -p")
build_output = execute_command("tmux capture-pane -t dev-feature:1 -p")

# If build fails, tell dev Claude about it
if "error:" in build_output:
    execute_command("tmux send-keys -t dev-feature:0 'Build failed: ...' C-m")
```

---

### Use Case 4: Multi-File Refactoring

**Scenario:** Refactor TFE's keyboard handling across 5 files

**Orchestration:**
```
Orchestrator:
├─ Worker A: update_keyboard.go
├─ Worker B: update_mouse.go
├─ Worker C: context_menu.go
├─ Worker D: dialog.go
├─ Worker E: types.go (shared types)
│
├─ Coordination:
│  ├─ Worker E finishes types first
│  ├─ Orchestrator tells A-D about new types
│  └─ A-D update their files to use new types
│
└─ Verification:
   └─ Test Claude builds and tests all changes
```

---

## Implementation Example

Here's how the orchestrator Claude would actually work:

### Step 1: User Request
```
User: "Implement user profile feature with frontend, backend, and tests"
```

### Step 2: Orchestrator Plans
```
Orchestrator: I'll coordinate 3 Claude sessions:
1. Frontend: UI component
2. Backend: API endpoints
3. Testing: Integration tests

Let me set this up...
```

### Step 3: Orchestrator Creates Workers

Uses Desktop Commander to run:
```bash
tmux new-session -d -s profile-frontend "cd /home/matt/projects/TFE && claude"
tmux new-session -d -s profile-backend "cd /home/matt/projects/TFE && claude"
tmux new-session -d -s profile-tests "cd /home/matt/projects/TFE && claude"
```

### Step 4: Orchestrator Assigns Tasks

```bash
tmux send-keys -t profile-frontend "Create user profile UI component in ui/profile.go. Include fields for name, email, bio. Style with lipgloss." C-m

tmux send-keys -t profile-backend "Create profile API in api/profile.go. Implement GET /api/profile and PUT /api/profile endpoints with JSON." C-m

tmux send-keys -t profile-tests "Wait for instructions. You'll test the integration of frontend UI and backend API." C-m
```

### Step 5: Orchestrator Monitors Progress

Every 2-3 minutes:
```bash
# Check frontend
frontend_status = execute_command("tmux capture-pane -t profile-frontend -p -S -50")

# Check backend
backend_status = execute_command("tmux capture-pane -t profile-backend -p -S -50")

# Report to user
"Frontend: Creating UI components... 50% complete"
"Backend: API endpoints defined, implementing handlers... 70% complete"
```

### Step 6: Orchestrator Coordinates Integration

When both complete:
```bash
# Read the files they created
frontend_code = read_file("ui/profile.go")
backend_code = read_file("api/profile.go")

# Extract API contract from backend
api_endpoints = parse_api_from_code(backend_code)

# Send to test Claude
tmux send-keys -t profile-tests "Frontend complete in ui/profile.go. Backend API at /api/profile (GET and PUT). Write integration tests that verify the UI calls the API correctly." C-m
```

### Step 7: Final Report

```
Orchestrator reports to user:

✅ User Profile Feature Complete

Frontend (profile-frontend):
  - ui/profile.go created (234 lines)
  - Displays name, email, bio fields
  - Styled with lipgloss

Backend (profile-backend):
  - api/profile.go created (189 lines)
  - GET /api/profile - fetch user profile
  - PUT /api/profile - update profile
  - JSON request/response

Testing (profile-tests):
  - tests/profile_integration_test.go (98 lines)
  - ✅ UI renders correctly
  - ✅ API endpoints respond
  - ✅ UI integrates with API
  - All tests passing

Files modified: 3
Total lines: 521
Build status: ✅ Clean build
Test status: ✅ All passing
```

---

## How to Use

### Method 1: Manual Slash Command

1. Start a Claude session (the orchestrator)
2. Use the slash command:
   ```
   /orchestrate-parallel-dev
   ```
3. Describe your multi-part task
4. Claude creates workers, assigns tasks, monitors, coordinates

### Method 2: Direct Prompt

Just ask Claude:
```
"Implement authentication feature using orchestration. Create separate Claude sessions for frontend, backend, and testing. Coordinate their work and synthesize the results."
```

If Claude has Desktop Commander MCP available, it can orchestrate.

---

## Slash Command Usage

```
/orchestrate-parallel-dev
```

Then provide your task:
```
"Implement syntax highlighting for TFE's preview mode.

Tasks:
- Research Go syntax highlighting libraries
- Integrate chosen library with render_preview.go
- Add language detection for file extensions
- Test with multiple file types
- Update documentation
```

The orchestrator will:
1. Break this into parallel work streams
2. Create worker Claude sessions
3. Assign tasks
4. Monitor progress
5. Coordinate dependencies
6. Report results

---

## Benefits

### 1. **Parallelism**
- Tasks that would take 60 minutes sequentially complete in 20 minutes
- Research multiple approaches simultaneously
- Develop frontend + backend + tests in parallel

### 2. **Specialization**
- Each worker Claude focuses on one domain
- Cleaner context (not mixing frontend + backend knowledge)
- Better code quality per domain

### 3. **Coordination**
- Orchestrator ensures API contracts match
- Manages dependencies between components
- Prevents integration issues

### 4. **Monitoring**
- See real-time progress on all work streams
- Catch errors early
- Adjust plans dynamically

### 5. **Context Management**
- Workers can be auto-compacted independently
- Orchestrator maintains high-level view
- Fresh context for each domain

---

## Limitations & Considerations

### When NOT to Use

❌ **Single file changes** - Overhead not worth it
❌ **Tightly coupled work** - Better in one session
❌ **Quick fixes** - Setup time exceeds benefit
❌ **Exploratory work** - Needs human guidance

### Challenges

1. **Coordination overhead** - Orchestrator needs to actively manage
2. **Context sharing** - Workers don't see each other's code automatically
3. **Integration risk** - Components might not fit together perfectly
4. **Complexity** - More moving parts = more to monitor

### Best Practices

1. **Clear boundaries** - Assign independent work streams
2. **Define contracts** - Specify interfaces between components
3. **Monitor frequently** - Check progress every 2-3 minutes
4. **Handle dependencies** - Sequence work appropriately
5. **Test integration** - Dedicated test worker or orchestrator verifies
6. **Auto-compact workers** - Keep context fresh
7. **Clean up** - Kill sessions when done

---

## Demo

Run the demo to see it in action:

```bash
./scripts/demo-orchestration.sh
```

This creates:
- 1 orchestrator session
- 3 worker sessions (frontend, backend, testing)
- Shows task assignment
- Shows progress monitoring
- Shows coordination
- Shows final synthesis

View sessions with:
```bash
tmux attach -t orchestrator-demo
tmux attach -t worker-frontend
tmux attach -t worker-backend
tmux attach -t worker-tests
```

---

## Future Enhancements

### Possible Improvements:

1. **Shared memory** - Coordination file that all workers read
2. **Event bus** - Workers publish events, orchestrator subscribes
3. **Progress API** - Workers report progress via structured format
4. **Auto-recovery** - Restart failed workers
5. **Load balancing** - Distribute work based on worker capacity
6. **Visualization** - Real-time dashboard of all workers

### Advanced Patterns:

**Pipeline orchestration:**
```
Worker A → Worker B → Worker C → Worker D
(Research) (Design)  (Implement) (Test)
```

**Map-reduce:**
```
       ┌─ Worker A ─┐
Input ─┼─ Worker B ─┼─→ Orchestrator synthesizes
       └─ Worker C ─┘
```

**Hierarchical:**
```
       Orchestrator
       ├─ Sub-orchestrator A
       │  ├─ Worker 1
       │  └─ Worker 2
       └─ Sub-orchestrator B
          ├─ Worker 3
          └─ Worker 4
```

---

## Comparison to Other Patterns

| Pattern | Use Case | Benefits | Drawbacks |
|---------|----------|----------|-----------|
| **Single Claude** | Simple tasks | Simple, fast | Context fills quickly |
| **Phased with compact** | Complex multi-phase | Fresh context per phase | Sequential only |
| **Multi-Claude orchestration** | Parallel work streams | Speed, specialization | Coordination overhead |

**When to use each:**
- **Single**: Default for most tasks
- **Phased**: Large features, multiple phases, benefit from fresh context
- **Orchestrated**: Truly parallel work, multiple domains, time-sensitive

---

## Summary

The multi-Claude orchestration pattern enables:

✅ **One orchestrator Claude** that can:
- Create multiple worker Claude sessions
- Assign different tasks to each
- Monitor their progress via tmux capture-pane
- Send coordination messages via tmux send-keys
- Monitor processes workers created via Desktop Commander
- Synthesize results from all workers

✅ **Powered by:**
- Desktop Commander MCP (execute_command tool)
- Tmux (session control and monitoring)
- Your existing auto-compact workflow (for context management)

✅ **Best for:**
- Large features spanning multiple files/domains
- Parallel research tasks
- Frontend + Backend + Testing coordination
- Build monitoring during development
- Multi-file refactoring

**The key insight:** With tmux + Desktop Commander, Claude isn't limited to one session. It can orchestrate an entire team of Claude instances! 🎼
