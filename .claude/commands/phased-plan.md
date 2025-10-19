# Phased Plan: Multi-Phase Execution with Auto-Compacting

You are executing a multi-phase plan where each phase runs in fresh context for maximum accuracy.

**CRITICAL INSIGHT:** Fresh context = better accuracy. Even if context window has room, compacting between phases improves results!

## How This Works

After the user provides a phased plan, you will:

1. **Execute Phase 1** - Work on first phase completely
2. **Auto-compact** - Create summary using Desktop Commander
3. **Execute Phase 2** - Start fresh with Phase 1 summary
4. **Auto-compact** - Cumulative summary of Phases 1-2
5. **Continue** - Each phase gets fresh context

## Command Format

The user will provide phases in one of these formats:

**Format 1: Inline**
```
/phased-plan
Phase 1: Research syntax highlighting libraries for Go
Phase 2: Choose library and implement basic highlighting
Phase 3: Add language detection for code blocks
Phase 4: Integrate with preview rendering
Phase 5: Test and optimize performance
```

**Format 2: With Goal**
```
/phased-plan Implement syntax highlighting in preview
```
Then you ask for the phases.

## Your Execution Process

### **Phase Execution Pattern:**

For each phase:

1. **Announce Phase Start**
   ```
   ═══════════════════════════════════════════════
   📋 PHASE N/TOTAL: [Phase Description]
   ═══════════════════════════════════════════════

   Previous phases summary:
   [Cumulative summary of completed phases]

   Now executing: [Current phase]
   ```

2. **Execute the Phase**
   - Work on the phase completely
   - Show progress as you go
   - Complete all tasks for this phase

3. **Phase Completion**
   ```
   ✅ PHASE N COMPLETE

   What was accomplished:
   - [Key achievements]
   - [Files modified]
   - [Important decisions]
   ```

4. **Git Commit (Optional but Recommended)**
   ```
   📝 Phase complete! Ready to commit?

   Suggested commit message:
   Phase N: [Phase description]

   - [Achievement 1]
   - [Achievement 2]
   - [Files modified]

   Use: auto-compact --commit -g "Phase [N+1]"
   → This will: commit changes + compact + continue
   ```

5. **Auto-Compact (Between Phases)**
   - Use Desktop Commander's `write_file` to save phase summary
   - Create `/tmp/claude-phase-summary.md` with:
     - All completed phases summary
     - Current state
     - Next phase to execute

6. **Instruct User**
   ```
   🔄 Ready for next phase!

   AUTOMATIC COMPACT OPTIONS:

   Option 1 (If in tmux + Git commit):
     Run: auto-compact --commit -g "Continue Phase [N+1]"
     → Commits changes + compacts + continues

   Option 2 (If in tmux, no commit):
     Run: auto-compact -g "Continue Phase [N+1]"
     → Automatic in-session compact

   Option 2 (Manual in session):
     Run: /compact
     → Quick in-place compact

   Option 3 (I'll do it):
     Just say "continue" and I'll proceed
     → I'll load the phase summary and continue

   Option 4 (Fresh instance):
     Run: session-reload
     → Brand new Claude with full context
   ```

### **Phase Summary Format**

```markdown
# Phased Plan Progress - [Date/Time]

## Overall Goal
[The main objective]

## Completed Phases

### Phase 1: [Description]
**Status:** ✅ Complete
**Accomplished:**
- [Achievement 1]
- [Achievement 2]

**Files Modified:**
- `path/file.ext` - What changed

**Key Decisions:**
- [Important choice 1]
- [Important choice 2]

### Phase 2: [Description]
**Status:** ✅ Complete
[Same format as Phase 1]

## Current State
- Code location: [Where we are in codebase]
- Working features: [What works]
- Known issues: [Any problems]

## Next Phase
**Phase [N+1]: [Description]**

**What needs to be done:**
- [Task 1]
- [Task 2]

**Context needed:**
- [Info from previous phases]
- [Files to reference]

## Important Context
- [Key insights that matter for remaining phases]
- [Gotchas or warnings]
- [Performance considerations]
```

### **Special Handling for Auto-Compact**

If the user has `auto-compact` available (tmux workflow):

After completing each phase, you can suggest:
```
✅ Phase N complete!

To continue with maximum accuracy:
  auto-compact -g "Continue Phase [N+1]: [description]"

This will:
  1. Save phase summary
  2. Clear context
  3. Reload with fresh start
  4. Begin Phase [N+1] immediately

[Wait for them to run it, or say "continue" to proceed in current session]
```

## Example Execution

```
User: /phased-plan Implement syntax highlighting
      Phase 1: Research libraries
      Phase 2: Implement basic highlighting
      Phase 3: Add language detection

═══════════════════════════════════════════════════════════════
📋 PHASE 1/3: Research libraries
═══════════════════════════════════════════════════════════════

I'll research Go syntax highlighting libraries...

[Work on research...]

✅ PHASE 1 COMPLETE

What was accomplished:
- Researched 4 libraries: chroma, go-syntax, pygments-go, tree-sitter
- Chroma is best: pure Go, 200+ languages, actively maintained
- Created comparison doc

Files created:
- docs/SYNTAX_HIGHLIGHTING_RESEARCH.md

Recommendation: Use chroma library

🔄 Phase 1 summary saved to /tmp/claude-phase-summary.md

Ready for Phase 2!

COMPACT OPTIONS:
  Tmux: auto-compact -g "Continue Phase 2: Implement basic highlighting"
  Quick: /compact
  Manual: Say "continue"

[User runs auto-compact]

═══════════════════════════════════════════════════════════════
📋 PHASE 2/3: Implement basic highlighting
═══════════════════════════════════════════════════════════════

Previous phases summary:
✅ Phase 1: Researched libraries → Chose chroma

Now implementing basic syntax highlighting with chroma...

[Work on implementation with FRESH CONTEXT...]
```

## Benefits of This Approach

**Why this is better than continuous execution:**

1. **Accuracy** - Each phase starts fresh, reducing compounding errors
2. **Focus** - Clear phase boundaries keep work organized
3. **Recovery** - Can resume from any phase if interrupted
4. **Review** - Natural checkpoints to verify before proceeding
5. **Performance** - Never work with degraded context window

## Phase Planning Tips

**Good phase boundaries:**
- ✅ Research → Implementation → Testing
- ✅ Backend → Frontend → Integration
- ✅ Core feature → UI → Error handling
- ✅ Setup → Development → Optimization

**Poor phase boundaries:**
- ❌ Too granular (every function call)
- ❌ Too broad (entire feature in one phase)
- ❌ Interdependent (can't complete one without another)

**Rule of thumb:** Each phase should be 10-30 minutes of focused work.

## Integration with Auto-Compact

If user is in tmux and has auto-compact available:

**Optimal workflow:**
1. User: `/phased-plan [plan]`
2. You: Execute Phase 1
3. You: Save phase summary
4. You: Suggest `auto-compact -g "Phase 2"`
5. User: Runs auto-compact (10-15 seconds)
6. You: Begin Phase 2 with fresh context
7. Repeat for all phases

**Result:** Maximum accuracy, minimal manual work! 🚀

## Start Execution

When user provides a phased plan, immediately:
1. Confirm you understand all phases
2. Ask if they want phase-by-phase compacting or continuous
3. Begin Phase 1
4. At each phase boundary, provide compact options
5. Continue until all phases complete

Remember: **Fresh context = better accuracy!** Compact between phases for best results!
