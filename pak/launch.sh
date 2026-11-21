#!/bin/sh
# Cheat Manager - NextUI launcher
PAK_DIR="$(dirname "$0")"
cd "$PAK_DIR" || exit 1
PAK_DIR="$(pwd)"

# libSDL2_gfx is not shipped with NextUI: use the bundled copy.
export LD_LIBRARY_PATH="$PAK_DIR/resources/lib:$LD_LIBRARY_PATH"
# h700 (Anbernic, experimental): libSDL2_ttf is bundled too, as the Pak
# Store does for this platform.
# my355 (Miyoo Flip, experimental): a libSDL2_gfx that needs glibc 2.17
# only (NextUI targets glibc 2.28 there; the default copy needs 2.29).
case "$PLATFORM" in
h700 | my355) export LD_LIBRARY_PATH="$PAK_DIR/resources/lib/$PLATFORM:$LD_LIBRARY_PATH" ;;
esac

LOG_DIR="${LOGS_PATH:-/mnt/SDCARD/.userdata/$PLATFORM/logs}"
mkdir -p "$LOG_DIR"

./cheat-manager 2>>"$LOG_DIR/Cheat Manager.stderr.txt"
sync
