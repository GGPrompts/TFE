# TFE Test Files

This directory contains sample files for testing TFE's file viewer integrations.

## Available Test Files

### Data Files
- **`sample-data.csv`** - CSV spreadsheet with employee data
  - **Viewer:** VisiData (`sudo apt install visidata`)
  - **Test:** Press F4 to open in spreadsheet viewer

### Configuration Files
- **`config.json`** - JSON configuration
- **`config.yaml`** - YAML configuration
- **`config.toml`** - TOML configuration
  - **Preview:** Already has syntax highlighting built-in
  - **Test:** Preview pane shows colorized syntax

### Database
- **`sample.db`** - SQLite database with users and products tables
  - **Viewer:** harlequin (`pipx install harlequin`)
  - **Test:** Press F4 to explore database interactively

### Binary Files
- **`binary.bin`** - Random binary data (10KB)
  - **Viewer:** hexyl (`cargo install hexyl` or `sudo apt install hexyl`)
  - **Test:** Press F4 to view as hex dump

### Archives
- **`archive.zip`** - ZIP archive with sample files
  - **Preview:** Shows helpful extraction commands
  - **Test:** Shows install hints for archive viewers

## Testing Media Files

For testing video/audio/PDF support, you can:

### PDF Files
```bash
# Download a sample PDF
curl -o examples/test-files/sample.pdf https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf
```
- **Viewer:** timg (`sudo apt install timg`)
- **Fallback:** Opens in browser (F3)

### Audio Files
```bash
# Create a test tone (requires sox)
sox -n examples/test-files/test-audio.mp3 synth 3 sine 440
```
- **Viewer:** mpv (`sudo apt install mpv`)

### Video Files
```bash
# Download a sample video
curl -o examples/test-files/sample-video.mp4 https://download.blender.org/demo/movies/BBB/bbb_sunflower_1080p_30fps_normal.mp4.zip
unzip examples/test-files/sample-video.mp4.zip -d examples/test-files/
```
- **Viewer:** mpv (`sudo apt install mpv`)

## Quick Test Workflow

1. **Build TFE:**
   ```bash
   go build -o tfe .
   ```

2. **Run TFE:**
   ```bash
   ./tfe
   ```

3. **Navigate to test directory:**
   - Navigate to `examples/test-files/`

4. **Test each file type:**
   - Use arrow keys to select a file
   - Press **F4** to open with the appropriate viewer
   - Press **Enter** to preview in TFE's preview pane

## Expected Behavior

### With Viewers Installed
- **CSV** → Opens in VisiData (interactive spreadsheet)
- **SQLite** → Opens in harlequin (database explorer)
- **Binary** → Opens in hexyl (hex viewer with colors)
- **PDF** → Opens in timg (terminal PDF viewer)
- **Video/Audio** → Plays in mpv (media player)

### Without Viewers Installed
- Preview pane shows helpful install instructions
- Includes package manager commands (apt, cargo, pipx)
- Falls back gracefully (CSV → text editor, PDF → browser)

## Demo Tips (for OBS Recording)

1. **Show CSV in VisiData** - Impressive spreadsheet viewer
2. **Show SQLite in harlequin** - Beautiful database TUI
3. **Show binary in hexyl** - Colorful hex dump
4. **Show JSON syntax highlighting** - Already looks great
5. **Show helpful preview messages** - For files without viewers

This demonstrates TFE's smart file type detection and context-aware opening!
