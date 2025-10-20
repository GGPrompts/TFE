#!/bin/bash
# Generate all TFE demo GIFs from VHS tape files
# Usage: ./generate-all.sh

set -e  # Exit on error

echo "🎬 TFE Demo Generator"
echo "===================="
echo ""

# Check if VHS is installed
if ! command -v vhs &> /dev/null; then
    echo "❌ VHS not found!"
    echo "Install with: go install github.com/charmbracelet/vhs@latest"
    echo "Or: sudo snap install vhs"
    exit 1
fi

# Check if gifsicle is installed (optional but recommended)
OPTIMIZE=false
if command -v gifsicle &> /dev/null; then
    OPTIMIZE=true
    echo "✓ gifsicle found - GIFs will be optimized"
else
    echo "⚠️  gifsicle not found - GIFs will not be optimized"
    echo "   Install with: sudo apt install gifsicle"
fi
echo ""

# Create assets directory if it doesn't exist
mkdir -p ../assets

# Count tape files
TAPE_COUNT=$(ls -1 *.tape 2>/dev/null | wc -l)
if [ "$TAPE_COUNT" -eq 0 ]; then
    echo "❌ No .tape files found in demos/"
    exit 1
fi

echo "Found $TAPE_COUNT demo tape files"
echo ""

# Generate each demo
CURRENT=0
for tape in *.tape; do
    CURRENT=$((CURRENT + 1))
    echo "[$CURRENT/$TAPE_COUNT] Generating: $tape"

    # Run VHS
    if vhs "$tape"; then
        echo "  ✓ Generated"
    else
        echo "  ❌ Failed to generate $tape"
        continue
    fi

    # Get output filename from tape
    OUTPUT=$(grep "^Output" "$tape" | head -1 | awk '{print $2}')

    if [ -z "$OUTPUT" ]; then
        echo "  ⚠️  No output file specified in $tape"
        continue
    fi

    # Optimize if gifsicle is available
    if [ "$OPTIMIZE" = true ] && [ -f "$OUTPUT" ]; then
        echo "  🔧 Optimizing..."
        BEFORE=$(du -h "$OUTPUT" | cut -f1)

        # Optimize with gifsicle
        gifsicle -O3 --lossy=80 --colors 256 -o "${OUTPUT}.tmp" "$OUTPUT" 2>/dev/null

        if [ -f "${OUTPUT}.tmp" ]; then
            mv "${OUTPUT}.tmp" "$OUTPUT"
            AFTER=$(du -h "$OUTPUT" | cut -f1)
            echo "  ✓ Optimized: $BEFORE → $AFTER"
        else
            echo "  ⚠️  Optimization failed, keeping original"
        fi
    fi

    echo ""
done

echo "✅ All demos generated!"
echo ""

# Show results
echo "📊 Results:"
echo "==========="
cd ../assets
for gif in demo-*.gif; do
    if [ -f "$gif" ]; then
        SIZE=$(du -h "$gif" | cut -f1)
        printf "  %-30s %8s\n" "$gif" "$SIZE"
    fi
done

echo ""
echo "🎉 Done! GIFs are in assets/"
echo ""
echo "💡 Tips:"
echo "   - Preview: open assets/demo-navigation.gif"
echo "   - Embed in README: ![Demo](assets/demo-navigation.gif)"
echo "   - Re-generate one: vhs demos/01-navigation.tape"
