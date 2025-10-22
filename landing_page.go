package main

import (
	"math/rand"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// landing_page.go - 90s Windows Nostalgic Landing Page
// Purpose: Recreate that classic Windows 95/98 aesthetic with starfield screensaver

// Windows 95 color palette (adaptive for light/dark terminals)
var win95Colors = struct {
	WindowChrome    lipgloss.AdaptiveColor // Classic teal window border
	WindowBG        lipgloss.AdaptiveColor // Gray window background
	ButtonFace      lipgloss.AdaptiveColor // Button face color
	ButtonHighlight lipgloss.AdaptiveColor // 3D button highlight (top-left)
	ButtonShadow    lipgloss.AdaptiveColor // 3D button shadow (bottom-right)
	ButtonText      lipgloss.AdaptiveColor // Button text
	TitleBarBG      lipgloss.AdaptiveColor // Title bar background
	TitleBarText    lipgloss.AdaptiveColor // Title bar text
	CheckerDark     lipgloss.AdaptiveColor // Checkered pattern dark
	CheckerLight    lipgloss.AdaptiveColor // Checkered pattern light
	StarColor       lipgloss.AdaptiveColor // Starfield stars/icons
}{
	WindowChrome: lipgloss.AdaptiveColor{
		Light: "#008080", // Teal for light mode
		Dark:  "#00D7D7", // Bright teal for dark mode
	},
	WindowBG: lipgloss.AdaptiveColor{
		Light: "#C0C0C0", // Light gray for light mode
		Dark:  "#404040", // Dark gray for dark mode
	},
	ButtonFace: lipgloss.AdaptiveColor{
		Light: "#C0C0C0", // Light gray
		Dark:  "#606060", // Medium gray
	},
	ButtonHighlight: lipgloss.AdaptiveColor{
		Light: "#FFFFFF", // White highlight
		Dark:  "#909090", // Light gray highlight
	},
	ButtonShadow: lipgloss.AdaptiveColor{
		Light: "#808080", // Dark gray shadow
		Dark:  "#202020", // Very dark gray shadow
	},
	ButtonText: lipgloss.AdaptiveColor{
		Light: "#000000", // Black text
		Dark:  "#FFFFFF", // White text
	},
	TitleBarBG: lipgloss.AdaptiveColor{
		Light: "#000080", // Navy blue
		Dark:  "#0000D7", // Bright blue
	},
	TitleBarText: lipgloss.AdaptiveColor{
		Light: "#FFFFFF", // White
		Dark:  "#FFFFFF", // White
	},
	CheckerDark: lipgloss.AdaptiveColor{
		Light: "#808080", // Dark gray
		Dark:  "#404040", // Darker gray
	},
	CheckerLight: lipgloss.AdaptiveColor{
		Light: "#C0C0C0", // Light gray
		Dark:  "#606060", // Medium gray
	},
	StarColor: lipgloss.AdaptiveColor{
		Light: "#0000FF", // Blue
		Dark:  "#5FD7FF", // Cyan
	},
}

// Star represents a single star/icon in the starfield
type Star struct {
	x, y, z float64 // Position (z is depth)
	icon    string  // Icon character (folder, file, etc.)
}

// Starfield creates a flying starfield effect with file/folder icons
type Starfield struct {
	stars  []Star
	width  int
	height int
	frame  int
}

// NewStarfield creates a new starfield with file/folder icons
func NewStarfield(width, height int) *Starfield {
	sf := &Starfield{
		width:  width,
		height: height,
		stars:  make([]Star, 50), // 50 stars/icons
	}

	// File/folder icons for the starfield with occasional emoji easter eggs
	icons := []string{
		"█", "▓", "▒", "░", "●", "◆", "■", "▲", // Regular icons
		"📁", "📂", "📄", "💾", "🗂️", // File/folder emojis (common)
		"✨", "🌟", "⭐", "💫", // Star emojis (easter eggs - rare)
		"🎮", "🎯", "🎨", "🔥", "💎", "📻", // Fun easter eggs (very rare)
	}

	// Initialize stars at random positions with weighted icon selection
	for i := range sf.stars {
		// Weight the icon selection (90% regular, 10% emoji easter eggs)
		iconIndex := 0
		roll := rand.Float64()
		if roll < 0.75 {
			// 75% chance: Regular ASCII icons (first 8)
			iconIndex = rand.Intn(8)
		} else if roll < 0.90 {
			// 15% chance: File/folder emojis (next 5)
			iconIndex = 8 + rand.Intn(5)
		} else if roll < 0.97 {
			// 7% chance: Star emojis (next 4)
			iconIndex = 13 + rand.Intn(4)
		} else {
			// 3% chance: Rare fun easter eggs (last 6)
			iconIndex = 17 + rand.Intn(6)
		}

		sf.stars[i] = Star{
			x:    (rand.Float64() - 0.5) * 100,
			y:    (rand.Float64() - 0.5) * 100,
			z:    rand.Float64() * 100,
			icon: icons[iconIndex],
		}
	}

	return sf
}

// Update moves stars toward the viewer
func (sf *Starfield) Update() {
	sf.frame++

	for i := range sf.stars {
		star := &sf.stars[i]

		// Move star forward
		star.z -= 1.5

		// Reset star if it passes the viewer
		if star.z <= 0 {
			star.x = (rand.Float64() - 0.5) * 100
			star.y = (rand.Float64() - 0.5) * 100
			star.z = 100
		}
	}
}

// Render generates the starfield as a string
func (sf *Starfield) Render() string {
	// Create canvas
	canvas := make([][]string, sf.height)
	for y := 0; y < sf.height; y++ {
		canvas[y] = make([]string, sf.width)
		for x := 0; x < sf.width; x++ {
			canvas[y][x] = " "
		}
	}

	// Project stars onto 2D screen
	for _, star := range sf.stars {
		// Perspective projection
		scale := 100 / star.z
		screenX := int(star.x*scale) + sf.width/2
		screenY := int(star.y*scale) + sf.height/2

		// Draw star if on screen
		if screenX >= 0 && screenX < sf.width && screenY >= 0 && screenY < sf.height {
			// Use distance to determine brightness/size
			depth := 1.0 - (star.z / 100)

			var char string
			if depth > 0.7 {
				// Close - show icon
				char = lipgloss.NewStyle().
					Foreground(win95Colors.StarColor).
					Bold(true).
					Render(star.icon)
			} else if depth > 0.4 {
				// Medium distance - show smaller icon
				char = lipgloss.NewStyle().
					Foreground(win95Colors.StarColor).
					Render("•")
			} else {
				// Far - show dot
				char = lipgloss.NewStyle().
					Foreground(win95Colors.StarColor).
					Render("·")
			}

			canvas[screenY][screenX] = char
		}
	}

	// Convert canvas to string
	var b strings.Builder
	for y := 0; y < sf.height; y++ {
		for x := 0; x < sf.width; x++ {
			b.WriteString(canvas[y][x])
		}
		if y < sf.height-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// CheckeredBackground creates a classic 90s checkered pattern
type CheckeredBackground struct {
	width  int
	height int
}

// NewCheckeredBackground creates a checkered background
func NewCheckeredBackground(width, height int) *CheckeredBackground {
	return &CheckeredBackground{
		width:  width,
		height: height,
	}
}

// Render generates the checkered pattern
func (cb *CheckeredBackground) Render() string {
	var b strings.Builder

	for y := 0; y < cb.height; y++ {
		for x := 0; x < cb.width; x++ {
			// Checkerboard pattern (4x4 squares)
			isLight := ((x/4)+(y/2))%2 == 0

			var style lipgloss.Style
			if isLight {
				style = lipgloss.NewStyle().Foreground(win95Colors.CheckerLight)
			} else {
				style = lipgloss.NewStyle().Foreground(win95Colors.CheckerDark)
			}

			b.WriteString(style.Render("░"))
		}
		if y < cb.height-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// Windows95Window creates a classic Windows 95 style window
type Windows95Window struct {
	title   string
	content string
	width   int
	height  int
}

// NewWindows95Window creates a classic window
func NewWindows95Window(title, content string, width, height int) string {
	// Title bar
	titleBar := lipgloss.NewStyle().
		Background(win95Colors.TitleBarBG).
		Foreground(win95Colors.TitleBarText).
		Bold(true).
		Width(width - 4). // Account for borders
		Render(" " + title)

	// Close button (just aesthetic)
	closeBtn := lipgloss.NewStyle().
		Background(win95Colors.ButtonFace).
		Foreground(win95Colors.ButtonText).
		Bold(true).
		Render(" X ")

	// Combine title bar and close button
	titleLine := lipgloss.JoinHorizontal(lipgloss.Top, titleBar, closeBtn)

	// Window border (3D effect)
	outerBorder := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(win95Colors.ButtonHighlight).
		BorderBackground(win95Colors.WindowBG)

	innerBorder := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(win95Colors.ButtonShadow).
		BorderBackground(win95Colors.WindowBG).
		Padding(1)

	// Content with background
	contentStyled := lipgloss.NewStyle().
		Background(win95Colors.WindowBG).
		Foreground(win95Colors.ButtonText).
		Width(width - 8).
		Render(content)

	// Assemble window
	windowContent := lipgloss.JoinVertical(
		lipgloss.Left,
		titleLine,
		contentStyled,
	)

	// Double border for 3D effect
	return outerBorder.Render(innerBorder.Render(windowContent))
}

// Button3D creates a 3D raised button (Windows 95 style)
func Button3D(label string, selected bool) string {
	// First, ensure label is padded to consistent width (20 chars total including padding)
	const buttonWidth = 20

	// Center the label text
	labelLen := len(label)
	leftPad := (buttonWidth - labelLen) / 2
	rightPad := buttonWidth - labelLen - leftPad
	paddedLabel := strings.Repeat(" ", leftPad) + label + strings.Repeat(" ", rightPad)

	// Apply colors and style to the padded label
	var borderColor lipgloss.AdaptiveColor
	var content string

	if selected {
		// Selected: bold with shadow border
		borderColor = win95Colors.ButtonShadow
		content = lipgloss.NewStyle().
			Background(win95Colors.ButtonFace).
			Foreground(win95Colors.ButtonText).
			Bold(true).
			Render(paddedLabel)
	} else {
		// Not selected: normal with highlight border
		borderColor = win95Colors.ButtonHighlight
		content = lipgloss.NewStyle().
			Background(win95Colors.ButtonFace).
			Foreground(win95Colors.ButtonText).
			Render(paddedLabel)
	}

	// Apply border to the styled content
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor)

	return style.Render(content)
}

// LandingPage composites all elements for the TFE landing screen
type LandingPage struct {
	starfield   *Starfield
	background  *CheckeredBackground
	width       int
	height      int
	frame       int
	selectedBtn int
	menuItems   []string
	useStarfield bool // Toggle between starfield and checkered background
}

// NewLandingPage creates the landing page
func NewLandingPage(width, height int) *LandingPage {
	return &LandingPage{
		starfield:   NewStarfield(width, height),
		background:  NewCheckeredBackground(width, height),
		width:       width,
		height:      height,
		menuItems: []string{
			"Browse Files",
			"Prompts",
			"Favorites",
			"Trash",
			"Settings",
			"Exit",
		},
		selectedBtn:  0,
		useStarfield: true, // Start with starfield
	}
}

// Update advances animation
func (lp *LandingPage) Update() {
	lp.frame++
	if lp.useStarfield {
		lp.starfield.Update()
	}
}

// Resize updates dimensions
func (lp *LandingPage) Resize(width, height int) {
	lp.width = width
	lp.height = height
	lp.starfield.width = width
	lp.starfield.height = height
	lp.background.width = width
	lp.background.height = height
}

// ToggleBackground switches between starfield and checkered
func (lp *LandingPage) ToggleBackground() {
	lp.useStarfield = !lp.useStarfield
}

// SelectNext moves to next menu item
func (lp *LandingPage) SelectNext() {
	lp.selectedBtn = (lp.selectedBtn + 1) % len(lp.menuItems)
}

// SelectPrev moves to previous menu item
func (lp *LandingPage) SelectPrev() {
	lp.selectedBtn--
	if lp.selectedBtn < 0 {
		lp.selectedBtn = len(lp.menuItems) - 1
	}
}

// GetSelectedItem returns current selection
func (lp *LandingPage) GetSelectedItem() string {
	return lp.menuItems[lp.selectedBtn]
}

// Render creates the complete landing page
func (lp *LandingPage) Render() string {
	// Layer 1: Background (starfield or checkered)
	var bg string
	if lp.useStarfield {
		bg = lp.starfield.Render()
	} else {
		bg = lp.background.Render()
	}
	bgLines := strings.Split(bg, "\n")

	// Layer 2: TFE logo (ASCII art)
	logo := []string{
		"████████╗███████╗███████╗",
		"╚══██╔══╝██╔════╝██╔════╝",
		"   ██║   █████╗  █████╗  ",
		"   ██║   ██╔══╝  ██╔══╝  ",
		"   ██║   ██║     ███████╗",
		"   ╚═╝   ╚═╝     ╚══════╝",
		"",
		"Terminal File Explorer",
	}

	logoStyled := make([]string, len(logo))
	rainbowColors := []lipgloss.AdaptiveColor{
		win95Colors.TitleBarBG,
		win95Colors.WindowChrome,
		win95Colors.StarColor,
	}

	for i, line := range logo {
		color := rainbowColors[(i+lp.frame/10)%len(rainbowColors)]
		logoStyled[i] = lipgloss.NewStyle().
			Foreground(color).
			Bold(true).
			Render(line)
	}

	// Layer 3: Menu window
	var menuContent strings.Builder
	menuContent.WriteString("\n")
	for i, item := range lp.menuItems {
		btn := Button3D(item, i == lp.selectedBtn)
		// Center each button within the window width
		btnLines := strings.Split(btn, "\n")
		for _, btnLine := range btnLines {
			// Calculate visual width properly
			visualWidth := lipgloss.Width(btnLine)
			padding := (34 - visualWidth) / 2  // Window content is ~34 chars wide
			if padding < 0 {
				padding = 0
			}
			menuContent.WriteString(strings.Repeat(" ", padding) + btnLine + "\n")
		}
		menuContent.WriteString("\n")
	}

	window := NewWindows95Window("My Computer", menuContent.String(), 40, len(lp.menuItems)*4+6)
	windowLines := strings.Split(window, "\n")

	// Calculate positions
	logoHeight := len(logoStyled)
	windowHeight := len(windowLines)
	totalHeight := logoHeight + 3 + windowHeight

	logoY := (lp.height - totalHeight) / 2
	if logoY < 0 {
		logoY = 0
	}
	windowY := logoY + logoHeight + 3

	// Composite
	result := make([]string, lp.height)
	for i := 0; i < lp.height; i++ {
		if i < len(bgLines) {
			result[i] = bgLines[i]
		} else {
			result[i] = strings.Repeat(" ", lp.width)
		}
	}

	// Overlay logo
	for i, line := range logoStyled {
		y := logoY + i
		if y >= 0 && y < lp.height {
			logoWidth := lipgloss.Width(line)
			x := (lp.width - logoWidth) / 2
			if x < 0 {
				x = 0
			}
			result[y] = overlayLine(result[y], line, x, lp.width)
		}
	}

	// Overlay window
	for i, line := range windowLines {
		y := windowY + i
		if y >= 0 && y < lp.height {
			windowWidth := lipgloss.Width(line)
			x := (lp.width - windowWidth) / 2
			if x < 0 {
				x = 0
			}
			result[y] = overlayLine(result[y], line, x, lp.width)
		}
	}

	// Footer hint
	hint := lipgloss.NewStyle().
		Foreground(win95Colors.ButtonText).
		Faint(true).
		Render("Press ↑/↓ to navigate • Enter to select • B to toggle background • Q to quit")

	hintY := lp.height - 2
	if hintY >= 0 && hintY < lp.height {
		hintX := (lp.width - lipgloss.Width(hint)) / 2
		if hintX < 0 {
			hintX = 0
		}
		result[hintY] = overlayLine(result[hintY], hint, hintX, lp.width)
	}

	return strings.Join(result, "\n")
}

// overlayLine overlays src onto dst at position x
// Handles ANSI escape codes properly to avoid corruption
func overlayLine(dst, src string, x, maxWidth int) string {
	if x < 0 {
		return dst
	}

	// Get visual widths (ignoring ANSI codes)
	dstWidth := lipgloss.Width(dst)
	srcWidth := lipgloss.Width(src)

	// If overlay is completely off-screen, return original
	if x >= dstWidth {
		return dst
	}

	var result strings.Builder

	// Left part of background (before overlay)
	if x > 0 {
		leftPart := extractVisualChars(dst, 0, x)
		result.WriteString(leftPart)
	}

	// Middle: the overlay itself
	result.WriteString(src)

	// Right part of background (after overlay)
	rightStart := x + srcWidth
	if rightStart < dstWidth {
		rightPart := extractVisualChars(dst, rightStart, dstWidth-rightStart)
		result.WriteString(rightPart)
	}

	// Pad to full width if needed
	currentWidth := lipgloss.Width(result.String())
	if currentWidth < maxWidth {
		result.WriteString(strings.Repeat(" ", maxWidth-currentWidth))
	}

	return result.String()
}

// extractVisualChars extracts count visible characters from position start
// Properly handles ANSI escape codes to avoid showing raw codes
func extractVisualChars(s string, start, count int) string {
	if count <= 0 {
		return ""
	}

	runes := []rune(s)
	visibleCount := 0
	inEscape := false
	startIdx := -1
	endIdx := -1
	escapeStart := -1

	// Track the last complete ANSI sequence before our start position
	var lastEscape strings.Builder

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// Detect start of ANSI escape sequence
		if r == '\x1b' {
			inEscape = true
			escapeStart = i
			continue
		}

		// Inside escape sequence
		if inEscape {
			// Check for end of escape sequence
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == 'm' {
				// If we haven't started extracting yet, save this escape for color inheritance
				if startIdx == -1 && escapeStart >= 0 {
					lastEscape.Reset()
					for j := escapeStart; j <= i; j++ {
						lastEscape.WriteRune(runes[j])
					}
				}

				inEscape = false
				escapeStart = -1
			}
			continue
		}

		// Count visible characters (not in escape sequences)
		if visibleCount == start && startIdx == -1 {
			startIdx = i
		}

		if visibleCount >= start {
			if visibleCount >= start+count {
				endIdx = i
				break
			}
		}

		visibleCount++
	}

	if startIdx == -1 {
		return ""
	}
	if endIdx == -1 {
		endIdx = len(runes)
	}

	// Build result with color inheritance
	var result strings.Builder

	// Prepend the last ANSI code before our slice to maintain colors
	if lastEscape.Len() > 0 {
		result.WriteString(lastEscape.String())
	}

	// Add the actual slice
	result.WriteString(string(runes[startIdx:endIdx]))

	return result.String()
}
