# Pre-Push Commit Review Prompt

**Context**: I use `NEXT_SESSION.md` to track what should be accomplished in each coding session. After commits are made but before pushing to remote, I run this prompt in a fresh Claude session to review the work.

---

## The Prompt

```
I need you to review my recent git commits against the plan in NEXT_SESSION.md before I push to remote.

Please:

1. Read `docs/NEXT_SESSION.md` (or `NEXT_SESSION.md` if no docs folder) to understand what was planned for this session

2. Run `git log --oneline -10` to see recent commits

3. For each commit since the last push, run `git show <commit-hash>` and review:
   - Does the commit message clearly describe what changed?
   - Do the code changes align with what NEXT_SESSION.md planned?
   - Are there any obvious code quality issues or bugs?
   - Was the planned work fully completed?

4. Provide a summary in this format:

   📋 Commit Review Summary

   Plan: [Brief summary of what NEXT_SESSION.md said to accomplish]

   Commits:
   ✅ abc1234 - "commit message"
      → [How it relates to plan, any issues]

   ✅ def5678 - "commit message"
      → [How it relates to plan, any issues]

   Overall: [Ready to push / Needs fixes / etc.]

5. After the review, ask if I want you to prepare a fresh NEXT_SESSION.md for the next session
```

---

## Why This Works

- **Fresh perspective**: New Claude session = unbiased review
- **Git history preservation**: Overwriting NEXT_SESSION.md is fine because git keeps the old versions
- **Catch drift**: Finds commits that weren't in the original plan (scope creep or forgotten documentation)
- **Quality gate**: Last check before code goes to remote
- **Continuity**: Ensures next session starts with clear context

## Example Usage

```bash
# After coding session with commits
git status  # Verify commits are ready

# Open fresh Claude Code session
# Paste the prompt above
# Claude reviews commits against NEXT_SESSION.md

# If all good:
git push origin feature-branch

# If issues found:
# Fix them, amend commits, then push
```

---

**Tip**: Keep NEXT_SESSION.md concise (under 400 lines). Archive old sections to docs/archive/ when it gets too large.
