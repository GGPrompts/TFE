---
name: Commit Review Workflow
description: Review git commits against NEXT_SESSION.md plan before pushing to remote
category: Development Workflow
tags: [git, code-review, session-management, quality-assurance]
variables:
  - name: branch_name
    description: Current git branch name
    default: main
---

# Commit Review & Next Session Planning

I use a workflow where `NEXT_SESSION.md` (or `docs/NEXT_SESSION.md`) contains the plan for what should be accomplished in the current session. After coding, I want you to:

## Review Process

1. **Read the Current Plan**:
   - Read `NEXT_SESSION.md` (or `docs/NEXT_SESSION.md`) to understand what was supposed to be accomplished

2. **Review Recent Commits**:
   - Run `git log --oneline -10` to see recent commits
   - For each commit since the last push, run `git show <commit-hash>` to review:
     - **Commit message quality**: Is it clear and descriptive?
     - **Code changes alignment**: Do the changes match what NEXT_SESSION.md planned?
     - **Code quality**: Any obvious bugs, style issues, or improvements needed?
     - **Completeness**: Was the planned task fully implemented?

3. **Provide Feedback**:
   - List each commit with a ✅ (good) or ⚠️ (needs attention) indicator
   - Highlight any discrepancies between the plan and actual implementation
   - Note any commits that seem unrelated to the NEXT_SESSION.md plan
   - Suggest improvements or fixes if needed

4. **Ask About Next Steps**:
   - After the review, ask: "Would you like me to prepare the NEXT_SESSION.md for the next session?"

## Example Output Format

```
📋 NEXT_SESSION.md Review for Branch: {{branch_name}}

Plan Summary:
- Fix dropdown menu performance lag
- Add emoji support to menus
- Improve context menu alignment

Commits Review:
✅ abc1234 - "fix: Cache tool availability at startup"
   → Matches planned performance fix
   → Clean implementation, no issues found

✅ def5678 - "feat: Restore emojis to dropdown menus"
   → Matches planned emoji support
   → All menus updated consistently

⚠️ ghi9012 - "refactor: Update menu rendering"
   → Not mentioned in NEXT_SESSION.md plan
   → Changes look good but should update plan to reflect this work

Overall: 3 commits ready to push. One minor note about documenting unplanned work.

Would you like me to prepare the NEXT_SESSION.md for the next session?
```

## Workflow Notes

- **Overwrite vs Update**: I typically OVERWRITE `NEXT_SESSION.md` each session (not append), because git history preserves the old versions
- **When to Update Instead**: Only UPDATE (append) if the current session's work is incomplete and needs to continue
- **Git History Benefit**: Each commit in git history has a corresponding NEXT_SESSION.md that shows the intent
- **Fresh Session**: Use a fresh Claude session for reviews to get unbiased perspective

## Follow-Up Actions

After review, you can:
1. Prepare a new NEXT_SESSION.md with upcoming tasks
2. Create a summary for PR description (if pushing to feature branch)
3. Suggest additional tests or documentation needed
4. Recommend refactoring opportunities discovered during review
