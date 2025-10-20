# TFE Demo: Tree View Mode
# This demo shows hierarchical folder exploration

Output ../assets/demo-tree-view.gif

Set FontSize 12
Set Width 1200
Set Height 700
Set Theme "CGA"
Set TypingSpeed 100ms

# Launch TFE
Type "../tfe"
Enter
Sleep 2s

# Switch to tree view (press 3)
Type "3"
Sleep 1.5s

# Navigate to a folder
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms

# Expand folder with right arrow
Right
Sleep 1.5s

# Navigate inside expanded folder
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms

Sleep 1s

# Collapse folder with left arrow
Left
Sleep 1s

# Navigate to another folder
Type "jjj"
Sleep 500ms

# Expand it
Right
Sleep 1.5s

Sleep 1s

# Back to list view
Type "1"
Sleep 1s

# Quit
Type "F10"
Sleep 500ms
