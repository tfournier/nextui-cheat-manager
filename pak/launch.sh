#!/bin/sh
# Cheat Manager - NextUI launcher
PAK_DIR="$(dirname "$0")"
cd "$PAK_DIR" || exit 1

# libSDL2_gfx is not shipped with NextUI: use the bundled copy.
export LD_LIBRARY_PATH="$PAK_DIR/resources/lib:$LD_LIBRARY_PATH"

LOG_DIR="${LOGS_PATH:-/mnt/SDCARD/.userdata/$PLATFORM/logs}"
mkdir -p "$LOG_DIR"

./cheat-manager 2>>"$LOG_DIR/Cheat Manager.stderr.txt"
sync
