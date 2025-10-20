# TFE Demo: View Modes
# This demo shows cycling through List, Detail, and Tree views

Output ../assets/demo-view-modes.gif

Set FontSize 12
Set Width 1400
Set Height 800
Set Theme "CGA"
Set TypingSpeed 100ms

# Launch TFE (starts in Detail view by default)
Type "../tfe"
Enter
Sleep 2s

# Show detail view (default)
Sleep 1s

# Switch to list view (press 1)
Type "1"
Sleep 2s

# Navigate a bit
Type "jjj"
Sleep 1s

# Switch to tree view (press 3)
Type "3"
Sleep 2s

# Navigate and expand
Type "jj"
Sleep 500ms
Right
Sleep 1.5s

# Back to detail view (press 2)
Type "2"
Sleep 2s

# Cycle with F9
Type "F9"
Sleep 1.5s

Type "F9"
Sleep 1.5s

Type "F9"
Sleep 1.5s

# Quit
Type "F10"
Sleep 500ms
