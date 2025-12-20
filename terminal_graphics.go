package main

// Module: terminal_graphics.go
// Purpose: Terminal graphics protocol support for HD image rendering
// Responsibilities:
// - Terminal protocol detection (Kitty, iTerm2, Sixel)
// - Image encoding using rasterm library
// - Image scaling and dimension calculations

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"

	"github.com/BourgeoisBear/rasterm"
	_ "golang.org/x/image/webp"
)

// TerminalProtocol represents supported terminal graphics protocols
type TerminalProtocol int

const (
	ProtocolNone TerminalProtocol = iota
	ProtocolKitty
	ProtocolITerm2
	ProtocolSixel
)

// detectTerminalProtocol determines which graphics protocol the current terminal supports
func detectTerminalProtocol() TerminalProtocol {
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")

	// Manual override via environment variable
	// Useful for cases where auto-detection fails (e.g., WSL)
	// Valid values: "kitty", "iterm2", "sixel", "none"
	if override := os.Getenv("TFE_TERMINAL_PROTOCOL"); override != "" {
		switch strings.ToLower(override) {
		case "kitty":
			return ProtocolKitty
		case "iterm2":
			return ProtocolITerm2
		case "sixel":
			return ProtocolSixel
		case "none":
			return ProtocolNone
		}
	}

	// Check for Kitty terminal
	if strings.Contains(term, "kitty") || os.Getenv("KITTY_WINDOW_ID") != "" {
		return ProtocolKitty
	}

	// Check for WezTerm (supports Kitty protocol)
	// Standard detection (works on native Linux/macOS)
	if os.Getenv("WEZTERM_EXECUTABLE") != "" || os.Getenv("TERM_PROGRAM") == "WezTerm" {
		return ProtocolKitty
	}

	// WSL-specific WezTerm detection
	// WezTerm environment variables don't pass through to WSL
	// IMPORTANT: Kitty protocol does NOT work in WezTerm on WSL/Windows
	// Even though WezTerm supports Kitty protocol natively on Linux,
	// the protocol fails when WezTerm is running on Windows (WSL environment)
	// Return ProtocolNone to show fallback viewer options instead
	if isWSL() && isWezTermInPath() {
		return ProtocolNone  // Kitty protocol doesn't work in WSL
	}

	// Check for iTerm2
	if termProgram == "iTerm.app" {
		return ProtocolITerm2
	}

	// Check for Sixel support (xterm, mlterm, foot, wezterm)
	// Note: Sixel detection via terminal query would be better, but requires async handling
	if strings.Contains(term, "xterm") || strings.Contains(term, "mlterm") ||
		strings.Contains(term, "foot") || strings.Contains(term, "sixel") {
		// For now, prefer no sixel since we have better protocols
		// Can enable if Kitty/iTerm2 detection fails
		return ProtocolNone
	}

	return ProtocolNone
}

// isWezTermInPath checks if WezTerm is available in the PATH
// This is useful for WSL where WezTerm is installed on Windows
func isWezTermInPath() bool {
	// Check common Windows drive letters for WezTerm in WSL
	// Users may have Windows installed on drives other than C:
	driveLetters := []string{"c", "d", "e"}
	for _, drive := range driveLetters {
		paths := []string{
			fmt.Sprintf("/mnt/%s/Program Files/WezTerm/wezterm.exe", drive),
			fmt.Sprintf("/mnt/%s/Program Files (x86)/WezTerm/wezterm.exe", drive),
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return true
			}
		}
	}

	// Also check if wezterm.exe is in PATH
	pathEnv := os.Getenv("PATH")
	for _, dir := range strings.Split(pathEnv, ":") {
		if strings.Contains(strings.ToLower(dir), "wezterm") {
			return true
		}
	}

	return false
}

// renderImageWithProtocol attempts to render an image using terminal graphics protocols
// Returns the rendered string and a boolean indicating success
func renderImageWithProtocol(imagePath string, maxWidth, maxHeight int) (string, bool) {
	protocol := detectTerminalProtocol()

	// If no protocol support, return false to fall back to ASCII
	if protocol == ProtocolNone {
		return "", false
	}

	// Load the image
	img, err := loadImageFile(imagePath)
	if err != nil {
		return "", false
	}

	// Scale image to fit dimensions while maintaining aspect ratio
	scaledImg := scaleImage(img, maxWidth, maxHeight)

	// Encode based on protocol
	switch protocol {
	case ProtocolKitty:
		encoded, err := encodeKittyImage(scaledImg)
		if err != nil {
			return "", false
		}
		return encoded, true

	case ProtocolITerm2:
		encoded, err := encodeITerm2Image(scaledImg)
		if err != nil {
			return "", false
		}
		return encoded, true

	case ProtocolSixel:
		encoded, err := encodeSixelImage(scaledImg)
		if err != nil {
			return "", false
		}
		return encoded, true
	}

	return "", false
}

// loadImageFile loads an image from the filesystem
func loadImageFile(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// scaleImage scales an image to fit within maxWidth x maxHeight while maintaining aspect ratio
func scaleImage(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// Terminal cells are roughly 2:1 height:width ratio
	// Adjust maxHeight to account for cell aspect ratio
	maxHeight = maxHeight * 2

	// Calculate scaling factor
	scaleX := float64(maxWidth) / float64(imgWidth)
	scaleY := float64(maxHeight) / float64(imgHeight)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// If image is already smaller, don't upscale
	if scale >= 1.0 {
		return img
	}

	newWidth := int(float64(imgWidth) * scale)
	newHeight := int(float64(imgHeight) * scale)

	// Create new scaled image
	scaled := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Simple nearest-neighbor scaling
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := int(float64(x) / scale)
			srcY := int(float64(y) / scale)
			scaled.Set(x, y, img.At(srcX, srcY))
		}
	}

	return scaled
}

// encodeKittyImage encodes an image using the Kitty graphics protocol
func encodeKittyImage(img image.Image) (string, error) {
	var buf strings.Builder
	opts := rasterm.KittyImgOpts{}
	err := rasterm.KittyWriteImage(&buf, img, opts)
	if err != nil {
		return "", fmt.Errorf("kitty encoding failed: %w", err)
	}
	return buf.String(), nil
}

// encodeITerm2Image encodes an image using the iTerm2 inline images protocol
func encodeITerm2Image(img image.Image) (string, error) {
	var buf strings.Builder
	err := rasterm.ItermWriteImage(&buf, img)
	if err != nil {
		return "", fmt.Errorf("iterm2 encoding failed: %w", err)
	}
	return buf.String(), nil
}

// encodeSixelImage encodes an image using the Sixel protocol
func encodeSixelImage(img image.Image) (string, error) {
	// Convert to paletted image for Sixel
	bounds := img.Bounds()
	paletted := image.NewPaletted(bounds, nil)

	// Copy image to paletted format
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			paletted.Set(x, y, img.At(x, y))
		}
	}

	var buf strings.Builder
	err := rasterm.SixelWriteImage(&buf, paletted)
	if err != nil {
		return "", fmt.Errorf("sixel encoding failed: %w", err)
	}
	return buf.String(), nil
}

// getProtocolName returns a human-readable name for the detected protocol
func getProtocolName() string {
	protocol := detectTerminalProtocol()
	switch protocol {
	case ProtocolKitty:
		return "Kitty"
	case ProtocolITerm2:
		return "iTerm2"
	case ProtocolSixel:
		return "Sixel"
	default:
		return "None"
	}
}
