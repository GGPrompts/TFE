# TFE Demo: Complete Workflow
# This demo showcases multiple features in a realistic workflow

Output ../assets/demo-complete.gif

Set FontSize 12
Set Width 1400
Set Height 800
Set Theme "CGA"
Set TypingSpeed 100ms

# Launch TFE
Type "../tfe"
Enter
Sleep 2s

# Basic navigation
Type "jjjj"
Sleep 1s

# Switch to tree view
Type "3"
Sleep 1.5s

# Expand folder
Type "jj"
Sleep 500ms
Right
Sleep 1.5s

# Navigate inside
Type "jj"
Sleep 1s

# Enter dual-pane mode
Type " "
Sleep 2s

# Navigate files (preview updates)
Type "jjj"
Sleep 1.5s

# Switch to preview pane
Tab
Sleep 1s

# Scroll preview
Type "jjjjj"
Sleep 1s

# Back to file list
Tab
Sleep 500ms

# Exit dual-pane
Escape
Sleep 1s

# Preview a file
Enter
Sleep 2s

# Scroll
PageDown
Sleep 1s

# Exit preview
Escape
Sleep 1s

# Quit
Type "F10"
Sleep 500ms
