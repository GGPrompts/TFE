#!/bin/bash
# TFE Project Organization Script
# Cleans up project structure for v1.0.0 public release

set -e  # Exit on error

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧹 TFE Project Organization Script${NC}"
echo "===================================="
echo ""

# Safety check
if [ ! -f "go.mod" ] || [ ! -f "README.md" ]; then
    echo -e "${YELLOW}Error: Not in TFE project root directory${NC}"
    exit 1
fi

echo "This will:"
echo "  1. Move internal docs to docs/internal/"
echo "  2. Move scripts to scripts/"
echo "  3. Move theme files to examples/themes/"
echo "  4. Remove build artifacts"
echo "  5. Remove empty directories"
echo "  6. Update .gitignore"
echo ""
read -p "Continue? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 1
fi

echo ""
echo -e "${BLUE}Step 1: Creating directories${NC}"
mkdir -p docs/internal
mkdir -p examples/themes
echo "  ✅ Created docs/internal/"
echo "  ✅ Created examples/themes/"

echo ""
echo -e "${BLUE}Step 2: Moving internal docs to docs/internal/${NC}"
docs_to_move=(
    "AGENTS.md"
    "BACKLOG.md"
    "PLAN.md"
    "PRE_LAUNCH_PLAN.md"
    "LAUNCH_READINESS.md"
    "COMPREHENSIVE_AUDIT_REPORT.md"
    "PROGRESS_REPORT.md"
    "SECURITY_FIXES_SUMMARY.md"
    "TRASH_FEATURE_SUMMARY.md"
)

for doc in "${docs_to_move[@]}"; do
    if [ -f "$doc" ]; then
        mv "$doc" docs/internal/
        echo "  ✅ Moved $doc"
    else
        echo "  ⏭️  Skipped $doc (not found)"
    fi
done

echo ""
echo -e "${BLUE}Step 3: Moving scripts to scripts/${NC}"
scripts_to_move=(
    "launch-tfe-tmux.sh"
    "test_security_fixes.sh"
)

for script in "${scripts_to_move[@]}"; do
    if [ -f "$script" ]; then
        mv "$script" scripts/
        echo "  ✅ Moved $script"
    else
        echo "  ⏭️  Skipped $script (not found)"
    fi
done

echo ""
echo -e "${BLUE}Step 4: Moving theme files to examples/themes/${NC}"
if [ -f "styles/tfe.json" ]; then
    mv styles/tfe.json examples/themes/
    echo "  ✅ Moved styles/tfe.json to examples/themes/"
else
    echo "  ⏭️  Skipped styles/tfe.json (not found)"
fi

echo ""
echo -e "${BLUE}Step 5: Removing build artifacts${NC}"
artifacts=(
    "coverage.html"
    "coverage.out"
)

for artifact in "${artifacts[@]}"; do
    if [ -f "$artifact" ]; then
        rm "$artifact"
        echo "  ✅ Removed $artifact"
    else
        echo "  ⏭️  Skipped $artifact (not found)"
    fi
done

echo ""
echo -e "${BLUE}Step 6: Removing empty directories${NC}"
empty_dirs=(
    "styles"
    "Screenshots"
)

for dir in "${empty_dirs[@]}"; do
    if [ -d "$dir" ] && [ -z "$(ls -A "$dir")" ]; then
        rmdir "$dir"
        echo "  ✅ Removed empty directory: $dir/"
    else
        echo "  ⏭️  Skipped $dir/ (not empty or doesn't exist)"
    fi
done

echo ""
echo -e "${BLUE}Step 7: Updating .gitignore${NC}"

# Check if .gitignore has the entries we need
if ! grep -q "^coverage.html$" .gitignore 2>/dev/null; then
    echo "" >> .gitignore
    echo "# Test coverage" >> .gitignore
    echo "coverage.out" >> .gitignore
    echo "coverage.html" >> .gitignore
    echo "  ✅ Added coverage files to .gitignore"
else
    echo "  ⏭️  .gitignore already has coverage entries"
fi

if ! grep -q "^tfe$" .gitignore 2>/dev/null; then
    echo "" >> .gitignore
    echo "# Binary" >> .gitignore
    echo "tfe" >> .gitignore
    echo "tfe.exe" >> .gitignore
    echo "  ✅ Added binary to .gitignore"
else
    echo "  ⏭️  .gitignore already has binary entries"
fi

echo ""
echo -e "${BLUE}Step 8: Creating README in legacy demos folder${NC}"
cat > demos/README.md << 'EOF'
# Legacy VHS Demos

**Note:** These VHS tape files (`.tape`) are legacy demos from early development.

**Current demos** are in the `assets/` directory as GIF files, recorded with OBS Studio for better quality and emoji/icon support.

## Files in this directory

- `*.tape` - VHS recording scripts (legacy)
- `generate-all.sh` - Batch generator for VHS demos
- Various markdown guides for VHS workflow

## Why replaced?

VHS demos rendered emojis and Nerd Font icons as boxes, making them unusable for showcasing TFE's visual features. OBS Studio recordings perfectly capture the terminal appearance.

See `demo-content/` for current recording scripts and tools.
EOF
echo "  ✅ Created demos/README.md"

echo ""
echo -e "${GREEN}✨ Organization complete!${NC}"
echo ""
echo "Summary:"
echo "  - Internal docs moved to docs/internal/"
echo "  - Scripts consolidated in scripts/"
echo "  - Theme files in examples/themes/"
echo "  - Build artifacts removed"
echo "  - Empty directories cleaned up"
echo "  - .gitignore updated"
echo ""
echo "Next steps:"
echo "  1. Review changes: git status"
echo "  2. Test build: go build -o tfe"
echo "  3. Commit: git add -A && git commit -m 'chore: Organize project structure for v1.0 release'"
echo ""
echo -e "${BLUE}Project structure is now clean and professional! 🎉${NC}"
