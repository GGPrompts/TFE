#!/bin/bash
# Convert OBS MP4 recordings to optimized GIFs
# Usage: ./convert-to-gif.sh input.mp4 output.gif [quality]
# Quality: high (default), medium, low

INPUT=$1
OUTPUT=$2
QUALITY=${3:-medium}

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

if [ -z "$INPUT" ] || [ -z "$OUTPUT" ]; then
    echo "Usage: ./convert-to-gif.sh input.mp4 output.gif [quality]"
    echo ""
    echo "Quality options:"
    echo "  high   - Best quality, 2-4 MB (fps=15, scale=1280)"
    echo "  medium - Good quality, 1-2 MB (fps=12, scale=1100) [DEFAULT]"
    echo "  low    - Smaller file, 500KB-1MB (fps=10, scale=900)"
    exit 1
fi

if [ ! -f "$INPUT" ]; then
    echo -e "${YELLOW}Error: Input file '$INPUT' not found${NC}"
    exit 1
fi

# Check if ffmpeg is installed
if ! command -v ffmpeg &> /dev/null; then
    echo -e "${YELLOW}Error: ffmpeg is not installed${NC}"
    echo ""
    echo "Install ffmpeg:"
    echo "  Ubuntu/Debian: sudo apt install ffmpeg"
    echo "  macOS:         brew install ffmpeg"
    echo "  Windows:       choco install ffmpeg"
    exit 1
fi

echo -e "${BLUE}🎬 Converting $INPUT to $OUTPUT (quality: $QUALITY)${NC}"
echo ""

# Set parameters based on quality
case "$QUALITY" in
    high)
        FPS=15
        SCALE=1280
        echo "Settings: 15 fps, 1280px width (2-4 MB expected)"
        ;;
    medium)
        FPS=12
        SCALE=1100
        echo "Settings: 12 fps, 1100px width (1-2 MB expected)"
        ;;
    low)
        FPS=10
        SCALE=900
        echo "Settings: 10 fps, 900px width (500KB-1MB expected)"
        ;;
    *)
        echo -e "${YELLOW}Unknown quality '$QUALITY', using 'medium'${NC}"
        FPS=12
        SCALE=1100
        ;;
esac

echo ""
echo "Converting... (this may take 10-30 seconds)"

# Convert with FFmpeg using palette for best color quality
ffmpeg -i "$INPUT" \
  -vf "fps=$FPS,scale=$SCALE:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse" \
  -loop 0 \
  "$OUTPUT" \
  -hide_banner -loglevel error

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✅ Success!${NC}"
    echo ""
    echo "Original MP4: $(ls -lh "$INPUT" | awk '{print $5}')"
    echo "Output GIF:   $(ls -lh "$OUTPUT" | awk '{print $5}')"
    echo ""
    echo "Saved to: $OUTPUT"
else
    echo -e "${YELLOW}❌ Conversion failed${NC}"
    exit 1
fi
