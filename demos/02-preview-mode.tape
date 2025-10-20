# TFE Demo: Preview Mode
# This demo shows entering preview mode and scrolling through a file

Output ../assets/demo-preview.gif

Set FontSize 12
Set Width 1200
Set Height 700
Set Theme "CGA"
Set TypingSpeed 100ms

# Launch TFE
Type "../tfe"
Enter
Sleep 2s

# Navigate to a file (README.md should be visible)
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms

Sleep 1s

# Enter preview mode
Enter
Sleep 2s

# Scroll down in preview
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms

Sleep 1s

# Scroll up
Type "k"
Sleep 400ms
Type "k"
Sleep 400ms

Sleep 1s

# Page down
PageDown
Sleep 1s

# Exit preview
Escape
Sleep 1s

# Quit
Type "F10"
Sleep 500ms
