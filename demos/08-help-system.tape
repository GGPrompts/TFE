# TFE Demo: Smart Help System (F1)
# This demo shows context-aware help navigation

Output ../assets/demo-help.gif

Set FontSize 12
Set Width 1200
Set Height 700
Set Theme "CGA"
Set TypingSpeed 100ms

# Launch TFE
Type "../tfe"
Enter
Sleep 2s

# Press F1 for help (opens HOTKEYS.md)
Type "F1"
Sleep 2s

# Scroll through help
Type "jjjjj"
Sleep 1s

PageDown
Sleep 1s

PageDown
Sleep 1s

# Exit help
Escape
Sleep 1s

# Navigate to a file and preview it
Type "jj"
Sleep 500ms
Enter
Sleep 2s

# Press F1 from preview mode (jumps to "Preview & Full-Screen Mode" section)
Type "F1"
Sleep 2s

# Scroll help
Type "jjjjj"
Sleep 1.5s

# Exit help
Escape
Sleep 1s

# Quit
Type "F10"
Sleep 500ms
