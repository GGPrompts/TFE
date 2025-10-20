# TFE Demo: Directory Search
# This demo shows filtering files with / search

Output ../assets/demo-search.gif

Set FontSize 12
Set Width 1200
Set Height 700
Set Theme "CGA"
Set TypingSpeed 100ms

# Launch TFE
Type "../tfe"
Enter
Sleep 2s

# Activate search mode with /
Type "/"
Sleep 1s

# Type search query (searching for .go files)
Type "go"
Sleep 1s

# Results filter in real-time
Sleep 1.5s

# Navigate filtered results
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms

Sleep 1s

# Exit search mode (keeps filter active)
Enter
Sleep 1.5s

# Clear search with Escape
Type "/"
Escape
Sleep 1s

# Try another search (README)
Type "/"
Sleep 500ms
Type "READ"
Sleep 1.5s

# Clear and exit
Escape
Sleep 1s

# Quit
Type "F10"
Sleep 500ms
