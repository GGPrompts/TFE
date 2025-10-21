# After Auto-Compact: Update Game Controller Button

**Context:** We added a game controller button [🎮] to the toolbar, but it needs to be added in BOTH single-pane and dual-pane headers.

---

## Task: Add [🎮] Button to Both View Modes

### Location 1: Single-Pane Mode Header (view.go)

**Already done!** ✅ The button is at line ~155-157:
```go
// Games launcher button
s.WriteString(homeButtonStyle.Render("[🎮]"))
s.WriteString(" ")
```

**Click handler** at update_mouse.go line ~158-164:
```go
// Game controller button [🎮] (X=35-39: [ + emoji(2) + ] + space)
if msg.X >= 35 && msg.X <= 39 {
    // Open game launcher dialog - show available games
    m.setStatusMessage("Games: Install minesweeper/solitaire from github.com/GGPrompts/TUIClassics", false)
    return m, nil
}
```

---

### Location 2: Dual-Pane Mode Header (render_preview.go) ⚠️ NEEDS UPDATE

**Find the dual-pane toolbar rendering** in `renderDualPane()` function.

Look for toolbar code similar to:
```go
toolbar := lipgloss.JoinHorizontal(
    lipgloss.Top,
    buttonStyle.Render("[🏠] home"),
    " ",
    favoriteButton,
    " ",
    buttonStyle.Render("[>_] command"),
    " ",
    buttonStyle.Render("[🔍] fuzzy search"),
)
```

**Add the game controller button BEFORE the trash button:**

```go
toolbar := lipgloss.JoinHorizontal(
    lipgloss.Top,
    buttonStyle.Render("[🏠] home"),
    " ",
    favoriteButton,
    " ",
    buttonStyle.Render("[>_] command"),
    " ",
    buttonStyle.Render("[🔍] fuzzy search"),
    " ",
    buttonStyle.Render("[🎮] games"),     // ← ADD THIS
    " ",
    buttonStyle.Render("[🗑️] trash"),    // (or whatever trash button looks like)
)
```

**Key points:**
- Add BEFORE trash button
- Include the space separator (" ")
- Use same buttonStyle as other buttons
- Keep the label short: "[🎮] games" or just "[🎮]"

---

### Location 3: Update Click Handler Coordinates (update_mouse.go)

**In dual-pane mode**, the toolbar click coordinates might be DIFFERENT!

**Search for:** "dual-pane" click handling or similar toolbar click code

**If there's separate dual-pane click handling:**
- Add game controller click at appropriate X coordinates
- Adjust trash button coordinates (+5 to account for new button)

**Current single-pane coordinates:**
- Game controller: X=35-39
- Trash: X=40-44

**Dual-pane coordinates will depend on layout!**
- Check actual rendering width
- Test by clicking in dual-pane mode
- Adjust X ranges as needed

---

## Testing Checklist

After making changes:

1. **Build:** `go build -o tfe`

2. **Test Single-Pane Mode:**
   - Open TFE
   - Click [🎮] button
   - Should show: "Games: Install minesweeper/solitaire..."
   - Verify trash button still works

3. **Test Dual-Pane Mode:**
   - Press Space to enter dual-pane
   - Click [🎮] button
   - Should show same message
   - Verify trash button still works

4. **Visual Check:**
   - Toolbar should have: [🏠] [⭐] [👁️] [⬌] [>_] [🔍] [📝] [🎮] [🗑️]
   - All buttons aligned
   - Spacing consistent
   - Works in both single and dual-pane modes

---

## Quick Reference: Button Order

```
[🏠] home
[⭐] favorites
[👁️] view mode
[⬌] pane toggle
[>_] command
[🔍] search
[📝] prompts
[🎮] games     ← NEW!
[🗑️] trash
```

---

## Expected Files to Modify

- `render_preview.go` - Add button to dual-pane toolbar
- `update_mouse.go` - Add dual-pane click handler (if separate)

**Files already updated:**
- ✅ `view.go` - Single-pane toolbar
- ✅ `update_mouse.go` - Single-pane click handler

---

## If You Get Stuck

**Search for existing trash button code:**
```bash
grep -n "trash" render_preview.go
grep -n "🗑️" render_preview.go
grep -n "Trash button" update_mouse.go
```

**The dual-pane toolbar is likely in:**
- `renderDualPane()` function
- Around lines with other toolbar buttons
- Look for `lipgloss.JoinHorizontal` with toolbar items

---

## Quick Copy-Paste Code

**For dual-pane toolbar (add before trash):**
```go
" ",
buttonStyle.Render("[🎮]"),
```

**For dual-pane click handler (adjust X coords as needed):**
```go
// Game controller button [🎮] in dual-pane
if msg.X >= XX && msg.X <= YY {  // Adjust XX and YY based on actual position
    m.setStatusMessage("Games: Install minesweeper/solitaire from github.com/GGPrompts/TUIClassics", false)
    return m, nil
}
```

---

**After completing this, the [🎮] button will work in BOTH single-pane and dual-pane modes!** 🎮✨
