# TFE Demo: Context Menu
# This demo shows the F2 context menu (keyboard alternative to right-click)

Output ../assets/demo-context-menu.gif

Set FontSize 12
Set Width 1200
Set Height 700
Set Theme "CGA"
Set TypingSpeed 100ms

# Launch TFE
Type "../tfe"
Enter
Sleep 2s

# Navigate to a file
Type "jjj"
Sleep 1s

# Open context menu with F2
Type "F2"
Sleep 2s

# Navigate menu
Down
Sleep 500ms
Down
Sleep 500ms
Down
Sleep 500ms

Sleep 1s

# Go back up
Up
Sleep 500ms
Up
Sleep 500ms

Sleep 1s

# Close menu with Escape
Escape
Sleep 1s

# Quit
Type "F10"
Sleep 500ms
