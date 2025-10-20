# TFE Demo: File Operations
# This demo shows creating a directory with F7

Output ../assets/demo-file-ops.gif

Set FontSize 12
Set Width 1200
Set Height 700
Set Theme "CGA"
Set TypingSpeed 100ms

# Launch TFE
Type "../tfe"
Enter
Sleep 2s

# Press F7 to create directory
Type "F7"
Sleep 2s

# Type directory name
Type "demo-folder"
Sleep 1s

# Confirm
Enter
Sleep 2s

# Navigate to the newly created folder
Type "j"
Sleep 400ms

# Enter it
Enter
Sleep 1.5s

# Go back to parent
Left
Sleep 1.5s

# Navigate to the folder again
Type "j"
Sleep 300ms

Sleep 1s

# Quit
Type "F10"
Sleep 500ms
