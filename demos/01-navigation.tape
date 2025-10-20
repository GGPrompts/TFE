# TFE Demo: Basic Navigation
# This demo shows basic file navigation and cursor movement

Output ../assets/demo-navigation.gif

Set FontSize 12
Set Width 1200
Set Height 700
Set Theme "CGA"
Set TypingSpeed 100ms

# Launch TFE
Type "../tfe"
Enter
Sleep 2s

# Navigate down with arrow keys
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms

Sleep 1s

# Navigate up
Up
Sleep 300ms
Up
Sleep 300ms

Sleep 1s

# Navigate with j/k (vim style)
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "k"
Sleep 300ms

Sleep 2s

# Quit
Type "F10"
Sleep 500ms
