// emoji_audit.go - Visual emoji width audit for TFE
// Tests how emojis render in different terminals to detect width quirks
//
// Usage:
//   go run emoji_audit.go                    # Visual output
//   go run emoji_audit.go > results.txt      # Save to file
//
// Run this in multiple terminals (Windows Terminal, WezTerm, Termux, xterm)
// and compare the results to build an emoji width compensation table.

//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-runewidth"
)

func main() {
	// Detect terminal
	termProgram := os.Getenv("TERM_PROGRAM")
	term := os.Getenv("TERM")
	wtSession := os.Getenv("WT_SESSION")
	wezterm := os.Getenv("WEZTERM_EXECUTABLE")

	// Build terminal identification
	var terminalName string
	if wtSession != "" {
		terminalName = "Windows Terminal"
	} else if wezterm != "" || termProgram == "WezTerm" {
		terminalName = "WezTerm"
	} else if termProgram == "iTerm.app" {
		terminalName = "iTerm2"
	} else if strings.Contains(term, "kitty") {
		terminalName = "Kitty"
	} else if term == "xterm" || term == "xterm-256color" {
		terminalName = "xterm"
	} else if strings.Contains(os.Getenv("PREFIX"), "com.termux") {
		terminalName = "Termux"
	} else {
		terminalName = term
	}

	// List of all emojis used in TFE (from menu.go, context_menu.go, etc.)
	emojis := []struct {
		char string
		name string
		file string // Where it's used in TFE
	}{
		{"📁", "Folder", "menu.go, context_menu.go"},
		{"📄", "File", "menu.go, context_menu.go"},
		{"📂", "Open Folder", "context_menu.go"},
		{"📋", "Clipboard", "menu.go"},
		{"📝", "Memo", "menu.go"},
		{"🗑️", "Trash (with variation selector)", "menu.go, context_menu.go"},
		{"🎨", "Art/Palette", "menu.go"},
		{"🌳", "Tree", "menu.go"},
		{"👁️", "Eye (with variation selector)", "menu.go"},
		{"⭐", "Star", "menu.go"},
		{"🔍", "Magnifying Glass", "menu.go"},
		{"🔄", "Refresh", "menu.go"},
		{"⌨️", "Keyboard (with variation selector)", "menu.go"},
		{"ℹ️", "Info (with variation selector)", "menu.go"},
		{"🔗", "Link", "menu.go"},
		{"🌿", "Herb", "context_menu.go"},
		{"🐋", "Whale", "context_menu.go"},
		{"📜", "Scroll", "context_menu.go"},
		{"🖼️", "Picture Frame (with variation selector)", "file_operations.go"},
		{"🗂️", "Folder Dividers (with variation selector)", "context_menu.go"},
		{"✨", "Sparkles", "file_operations.go"},
		{"🌐", "Globe", "render_file_list.go"},
		{"♻️", "Recycle (with variation selector)", "context_menu.go"},
		{"🧹", "Broom", "context_menu.go"},
		{"🚪", "Door", "menu.go"},
		{"⬌", "Left-Right Arrow", "menu.go"},
		{"🔀", "Twisted Arrows", "menu.go"},
		{"🎯", "Direct Hit", "menu.go"},
	}

	// Print header
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                   TFE EMOJI WIDTH AUDIT                       ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("Terminal: %s\n", terminalName)
	fmt.Printf("TERM: %s\n", term)
	fmt.Printf("TERM_PROGRAM: %s\n", termProgram)
	fmt.Printf("WT_SESSION: %s\n", wtSession)
	fmt.Printf("WEZTERM_EXECUTABLE: %s\n", wezterm)
	fmt.Println()

	// Visual alignment test
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("VISUAL ALIGNMENT TEST")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("If the right borders align perfectly, emoji widths are correct.")
	fmt.Println("If borders are jagged (off by 1 space), compensation is needed.")
	fmt.Println()

	// Draw reference ruler
	fmt.Print("    ")
	for i := 0; i < 40; i++ {
		if i%10 == 0 {
			fmt.Printf("%d", i/10)
		} else if i%5 == 0 {
			fmt.Print("+")
		} else {
			fmt.Print("·")
		}
	}
	fmt.Println()

	// Test each emoji with dots showing padding
	testWidth := 38
	for i, emoji := range emojis {
		runeWidth := runewidth.StringWidth(emoji.char)

		// Count variation selectors (U+FE0F) which might affect rendering
		variationSelectors := strings.Count(emoji.char, "\uFE0F")

		// Calculate padding
		padding := testWidth - runeWidth
		if padding < 0 {
			padding = 0
		}

		// Draw test line
		fmt.Printf("%2d. │%s", i+1, emoji.char)
		for j := 0; j < padding; j++ {
			fmt.Print("·")
		}
		fmt.Printf("│ rw=%d", runeWidth)
		if variationSelectors > 0 {
			fmt.Printf(" +%dVS", variationSelectors)
		}
		fmt.Println()
	}

	// Summary table
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("EMOJI WIDTH SUMMARY")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("%-4s %-40s %-8s %-4s\n", "Char", "Name", "Width", "VS")
	fmt.Println("───────────────────────────────────────────────────────────────")

	widthCounts := make(map[int]int)
	variationSelectorCount := 0

	for _, emoji := range emojis {
		width := runewidth.StringWidth(emoji.char)
		vs := strings.Count(emoji.char, "\uFE0F")

		widthCounts[width]++
		if vs > 0 {
			variationSelectorCount++
		}

		fmt.Printf("%-4s %-40s %-8d %-4d\n",
			emoji.char,
			truncate(emoji.name, 40),
			width,
			vs)
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("ANALYSIS")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("Total emojis tested: %d\n", len(emojis))
	fmt.Printf("Emojis with variation selectors: %d\n", variationSelectorCount)
	fmt.Println()
	fmt.Println("Width distribution according to runewidth library:")
	for width, count := range widthCounts {
		fmt.Printf("  Width %d: %d emojis\n", width, count)
	}
	fmt.Println()

	// Instructions
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("INSTRUCTIONS")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("1. Look at the VISUAL ALIGNMENT TEST above")
	fmt.Println("2. Check if all right borders (│) are perfectly aligned")
	fmt.Println()
	fmt.Println("INTERPRETATION:")
	fmt.Println()
	fmt.Println("  ✓ Borders aligned:     Terminal follows Unicode width spec")
	fmt.Println("                         (Windows Terminal is expected baseline)")
	fmt.Println()
	fmt.Println("  ✗ Borders off by 1:    Terminal renders emojis narrower")
	fmt.Println("                         (WezTerm is known to do this)")
	fmt.Println("                         → Needs width compensation (-1)")
	fmt.Println()
	fmt.Println("  ✗ Borders off by -1:   Terminal renders emojis wider")
	fmt.Println("                         (Rare, but possible)")
	fmt.Println("                         → Needs width compensation (+1)")
	fmt.Println()
	fmt.Println("  ? Emojis don't show:   Terminal lacks emoji support")
	fmt.Println("                         (Old xterm, some minimal terminals)")
	fmt.Println("                         → Consider ASCII fallbacks")
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("NEXT STEPS")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("1. Run this script in each terminal:")
	fmt.Println("   - Windows Terminal (baseline)")
	fmt.Println("   - WezTerm")
	fmt.Println("   - Termux")
	fmt.Println("   - xterm")
	fmt.Println()
	fmt.Println("2. Save output to files:")
	fmt.Println("   go run emoji_audit.go > windows_terminal.txt")
	fmt.Println("   go run emoji_audit.go > wezterm.txt")
	fmt.Println("   go run emoji_audit.go > termux.txt")
	fmt.Println("   go run emoji_audit.go > xterm.txt")
	fmt.Println()
	fmt.Println("3. Compare the visual alignment sections")
	fmt.Println()
	fmt.Println("4. Report results to build compensation table")
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
