// emoji_audit_targeted.go - Test the specific problematic emojis
// Run in both Windows Terminal and WezTerm to see the difference

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-runewidth"
)

func main() {
	termProgram := os.Getenv("TERM_PROGRAM")
	wtSession := os.Getenv("WT_SESSION")

	var terminalName string
	if wtSession != "" {
		terminalName = "Windows Terminal"
	} else if termProgram == "WezTerm" {
		terminalName = "WezTerm"
	} else {
		terminalName = os.Getenv("TERM")
	}

	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Printf("║  TARGETED TEST - Problem Emojis in %-27s ║\n", terminalName)
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// The four problematic emojis you reported
	problemEmojis := []struct {
		char string
		name string
		where string
	}{
		{"⬆️", "Up Arrow (parent dir)", "file_operations.go:576"},
		{"⚙️", "Gear (config files)", "file_operations.go:629,745-747"},
		{"🗜️", "Clamp (gzip files)", "file_operations.go:721,723-725"},
		{"🖼️", "Picture Frame (png files)", "file_operations.go:728"},
	}

	fmt.Println("THESE ARE THE EMOJIS YOU REPORTED AS CAUSING MISALIGNMENT:")
	fmt.Println()

	// Visual alignment test
	testWidth := 38

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

	for i, emoji := range problemEmojis {
		runeWidth := runewidth.StringWidth(emoji.char)
		variationSelectors := strings.Count(emoji.char, "\uFE0F")

		// Calculate padding
		padding := testWidth - runeWidth
		if padding < 0 {
			padding = 0
		}

		// Draw test line
		fmt.Printf("%d.  │%s", i+1, emoji.char)
		for j := 0; j < padding; j++ {
			fmt.Print("·")
		}
		fmt.Printf("│ rw=%d", runeWidth)
		if variationSelectors > 0 {
			fmt.Printf(" +%dVS", variationSelectors)
		}
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("WHAT TO CHECK:")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("1. Are the right borders (│) perfectly aligned vertically?")
	fmt.Println()
	fmt.Println("2. Count dots VISUALLY in your terminal (not in copied text):")
	fmt.Printf("   - Line 1 (⬆️):  Should have %d dots if aligned\n", testWidth-runewidth.StringWidth("⬆️"))
	fmt.Printf("   - Line 2 (⚙️):  Should have %d dots if aligned\n", testWidth-runewidth.StringWidth("⚙️"))
	fmt.Printf("   - Line 3 (🗜️):  Should have %d dots if aligned\n", testWidth-runewidth.StringWidth("🗜️"))
	fmt.Printf("   - Line 4 (🖼️):  Should have %d dots if aligned\n", testWidth-runewidth.StringWidth("🖼️"))
	fmt.Println()
	fmt.Println("3. If borders are JAGGED (not aligned):")
	fmt.Println("   - Count how many EXTRA dots you see")
	fmt.Println("   - That's the width compensation needed")
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("COMPARISON WITH OTHER VARIATION SELECTOR EMOJIS:")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Compare with other VS emojis from the audit that work correctly
	compareEmojis := []struct {
		char string
		name string
		status string
	}{
		{"🗑️", "Trash", "(from audit, line 6)"},
		{"👁️", "Eye", "(from audit, line 9)"},
		{"⌨️", "Keyboard", "(from audit, line 13)"},
		{"⬆️", "Up Arrow", "(PROBLEM emoji)"},
		{"⚙️", "Gear", "(PROBLEM emoji)"},
		{"🗜️", "Clamp", "(PROBLEM emoji)"},
		{"🖼️", "Picture Frame", "(PROBLEM emoji)"},
	}

	for _, emoji := range compareEmojis {
		runeWidth := runewidth.StringWidth(emoji.char)
		variationSelectors := strings.Count(emoji.char, "\uFE0F")

		padding := testWidth - runeWidth
		if padding < 0 {
			padding = 0
		}

		fmt.Printf("│%s", emoji.char)
		for j := 0; j < padding; j++ {
			fmt.Print("·")
		}
		fmt.Printf("│ rw=%d +%dVS  %s\n", runeWidth, variationSelectors, emoji.status)
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("ANALYSIS:")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	for _, emoji := range problemEmojis {
		runeWidth := runewidth.StringWidth(emoji.char)
		vs := strings.Count(emoji.char, "\uFE0F")

		// Check raw byte representation
		fmt.Printf("%s (%s):\n", emoji.char, emoji.name)
		fmt.Printf("  - Runewidth reports: %d cells\n", runeWidth)
		fmt.Printf("  - Variation selectors: %d (U+FE0F)\n", vs)
		fmt.Printf("  - Byte representation: ")
		for _, r := range emoji.char {
			fmt.Printf("U+%04X ", r)
		}
		fmt.Println()
		fmt.Printf("  - Used in: %s\n", emoji.where)
		fmt.Println()
	}

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("HYPOTHESIS:")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("All four problem emojis have variation selectors (U+FE0F).")
	fmt.Println()
	fmt.Println("Current TFE behavior (file_operations.go:1220):")
	fmt.Println("  visualWidth += variationSelectorCount")
	fmt.Println()
	fmt.Println("This compensates for Windows Terminal rendering them as 2 cells")
	fmt.Println("even though runewidth reports them as 1 cell.")
	fmt.Println()
	fmt.Println("In WezTerm, these specific emojis might render DIFFERENTLY than")
	fmt.Println("other variation selector emojis (like 🗑️, 👁️, ⌨️).")
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
}
