package main

import "testing"

// TestGetFileIconExtensionLookup verifies extension-based icon lookups against
// the package-level fileIconByExt map (hoisted out of getFileIcon, tfe-989).
func TestGetFileIconExtensionLookup(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"main.go", "🐹"},
		{"script.py", "🐍"},
		{"index.html", "🌐"},
		{"archive.zip", "🗜"},
		{"song.mp3", "🎵"},
		{"unknown.xyz", "📄"}, // unmapped extension falls back to generic document
	}
	for _, tt := range tests {
		got := getFileIcon(fileItem{name: tt.name})
		if got != tt.want {
			t.Errorf("getFileIcon(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

// TestGetFileIconSpecialCases verifies the non-extension branches still take
// priority over the extension map.
func TestGetFileIconSpecialCases(t *testing.T) {
	if got := getFileIcon(fileItem{name: "link", isSymlink: true}); got != "🌀" {
		t.Errorf("symlink icon = %q, want 🌀", got)
	}
	if got := getFileIcon(fileItem{name: "..", isDir: true}); got != "⬆" {
		t.Errorf("parent dir icon = %q, want ⬆", got)
	}
	if got := getFileIcon(fileItem{name: "go.mod"}); got != "🐹" {
		t.Errorf("go.mod icon = %q, want 🐹", got)
	}
}

// TestGetFileTypeExtensionLookup verifies extension-based type lookups against
// the package-level fileTypeByExt map (hoisted out of getFileType, tfe-989).
func TestGetFileTypeExtensionLookup(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"main.go", "Go Source"},
		{"app.tsx", "React (TSX)"},
		{"notes.md", "Markdown"},
		{"photo.jpeg", "JPEG Image"},
		{"lib.so", "Shared Library"},
		{"data.xyz", "xyz File"}, // unmapped extension is echoed back
		{"Procfile", "File"},     // no extension, no special-case name
	}
	for _, tt := range tests {
		got := getFileType(fileItem{name: tt.name})
		if got != tt.want {
			t.Errorf("getFileType(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

// TestGetFileTypeSpecialCases verifies the non-extension branches still take
// priority over the extension map.
func TestGetFileTypeSpecialCases(t *testing.T) {
	if got := getFileType(fileItem{name: "somedir", isDir: true}); got != "Folder" {
		t.Errorf("dir type = %q, want Folder", got)
	}
	if got := getFileType(fileItem{name: "link", isSymlink: true, symlinkTarget: "/tmp"}); got != "Link → /tmp" {
		t.Errorf("symlink type = %q, want Link → /tmp", got)
	}
	if got := getFileType(fileItem{name: "Makefile"}); got != "Makefile" {
		t.Errorf("Makefile type = %q, want Makefile", got)
	}
	if got := getFileType(fileItem{name: "Dockerfile"}); got != "Dockerfile" {
		t.Errorf("Dockerfile type = %q, want Dockerfile", got)
	}
}
