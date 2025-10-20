# Test emoji rendering in VHS
Output ../assets/test-emoji.gif

Set FontSize 12
Set Width 1200
Set Height 700
Set Theme "CGA"
Set FontFamily "FiraCode Nerd Font, Noto Color Emoji"
Set TypingSpeed 100ms

# Launch TFE in demo-content directory
Type "cd ../demo-content"
Enter
Sleep 1s
Type "../tfe"
Enter
Sleep 3s

# Navigate
Type "jjj"
Sleep 2s

# Quit
Type "F10"
Sleep 500ms
