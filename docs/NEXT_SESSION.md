# Next Session

## Context
TFE is a terminal file explorer built with Go + Bubbletea. We've just completed a major UX overhaul with F-keys and context menus.

**Current Status:**
- ✅ All core features complete (file browsing, view modes, dual-pane, preview)
- ✅ MC-style command prompt with history
- ✅ External editor integration (Micro/nano)
- ✅ F-key hotkeys (F1-F10, Midnight Commander style) - **Just completed!**
- ✅ Context menu system (right-click + F2) - **Just completed!**
- ✅ Favorites/bookmarks system
- ✅ Markdown rendering with Glamour
- ✅ Line wrapping for all files
- ✅ Text selection in preview mode - **Just completed!**
- ✅ Mouse wheel context menu scrolling - **Just completed!**

**Decision:** TFE is feature-complete as a lightweight file viewer + launcher. Future work focuses on file operations (F7/F8) and polish.

## Goal for Next Session

**Implement File Operations (F7/F8):**

1. **F7: Create Directory**
   - Show input prompt for directory name
   - Create directory in current path
   - Refresh file list after creation
   - Select newly created directory

2. **F8: Delete File/Folder**
   - Show confirmation dialog
   - Support deleting files and empty/non-empty directories
   - Refresh file list after deletion
   - Move cursor to next valid item

3. **Optional Enhancements:**
   - F5: Copy file (vs current clipboard copy path)
   - F6: Move/rename file (vs current favorites filter - could move to different key)
   - Progress indicators for long operations
   - Error handling and user feedback

## Implementation Steps

### Step 1: Add Input Dialog System
Create `dialog.go` for user input dialogs:
- Text input for F7 (directory name)
- Confirmation dialog for F8 (yes/no)
- Render as overlay similar to context menu

### Step 2: Implement F7 (Create Directory)
In `update.go`:
- Detect F7 key press
- Show input dialog for directory name
- Call `os.Mkdir()` with user input
- Refresh file list
- Select new directory

### Step 3: Implement F8 (Delete)
In `update.go`:
- Detect F8 key press
- Show confirmation dialog
- Call `os.Remove()` or `os.RemoveAll()` based on type
- Handle errors gracefully
- Refresh file list

### Step 4: Add to Context Menu
Update `context_menu.go`:
- Add "Delete" option to menu
- Add "New Folder" option for directories

## Files to Create/Modify

1. **dialog.go** (NEW) - Dialog system for user input/confirmation
2. **types.go** - Add dialog state to model
3. **update.go** - F7/F8 handlers, dialog event handling
4. **view.go** - Render dialog overlay
5. **context_menu.go** - Add delete/new folder menu items
6. **file_operations.go** - Helper functions for file ops

## Key Design Decisions

- **Safety first:** Always confirm before deleting
- **Non-recursive by default:** Warn on non-empty directories
- **Clear feedback:** Show success/error messages
- **Keyboard-friendly:** ESC cancels dialogs
- **Keep it simple:** No undo, no trash - direct filesystem ops

## Testing Checklist

After implementation:
- [ ] F7 creates directory successfully
- [ ] F7 handles invalid names gracefully
- [ ] F7 ESC cancels without creating
- [ ] F8 shows confirmation dialog
- [ ] F8 deletes file after confirmation
- [ ] F8 deletes empty directory
- [ ] F8 handles non-empty directory (warn or recursive?)
- [ ] F8 ESC cancels without deleting
- [ ] Context menu "Delete" works
- [ ] Context menu "New Folder" works
- [ ] File list refreshes after operations
- [ ] Cursor moves to sensible location after delete

## Reference Files

Key files for dialogs/overlays:
- **context_menu.go** - Menu overlay system (similar pattern for dialogs)
- **view.go:24-54** - Context menu positioning logic
- **update.go:474-500** - F2 handler (reference for F7/F8)

## Expected Outcome

After this session:
1. Users can create directories with F7
2. Users can delete files/folders with F8
3. All operations have proper confirmation/feedback
4. TFE becomes fully functional for basic file management
5. F7/F8 placeholders are now fully implemented

## Why This Matters

With F7/F8 implemented, TFE becomes:
- Complete Midnight Commander alternative for basic tasks
- No need to drop to shell for mkdir/rm
- Faster workflow for file organization
- True file "management" not just browsing
