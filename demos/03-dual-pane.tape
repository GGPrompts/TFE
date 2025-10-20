# TFE Demo: Dual-Pane Mode
# This demo shows split-screen browsing with live preview

Output ../assets/demo-dual-pane.gif

Set FontSize 12
Set Width 1400
Set Height 800
Set Theme "CGA"
Set TypingSpeed 100ms

# Launch TFE
Type "../tfe"
Enter
Sleep 2s

# Navigate to a file
Type "jj"
Sleep 500ms

# Enter dual-pane mode (Space or Tab)
Type " "
Sleep 2s

# Navigate files in left pane (preview updates automatically)
Type "j"
Sleep 800ms
Type "j"
Sleep 800ms
Type "j"
Sleep 800ms

Sleep 1s

# Switch focus to right pane (preview)
Tab
Sleep 1s

# Scroll preview in right pane
Type "jjjjj"
Sleep 1s

# Switch back to left pane
Tab
Sleep 1s

# Navigate more files
Type "jj"
Sleep 1s

# Exit dual-pane mode
Escape
Sleep 1s

# Quit
Type "F10"
Sleep 500ms
