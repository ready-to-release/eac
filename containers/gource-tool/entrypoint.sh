#!/bin/bash
set -e

# Debug: show output mode
echo "=== Gource Container Starting ==="
echo "GOURCE_OUTPUT_MODE: '${GOURCE_OUTPUT_MODE}'"
echo "GOURCE_DURATION: '${GOURCE_DURATION}'"
echo "================================="

# Parse resolution
WIDTH=$(echo $GOURCE_RESOLUTION | cut -d'x' -f1)
HEIGHT=$(echo $GOURCE_RESOLUTION | cut -d'x' -f2)

# Start Xvfb
Xvfb :99 -screen 0 ${WIDTH}x${HEIGHT}x24 &
export DISPLAY=:99
sleep 2

# Pre-generate git log once (faster than repeated git reads)
cd /visualization/repo
gource --output-custom-log /tmp/gource.log . 2>/dev/null || true

# Get date range from log (timestamps are in first column)
FIRST_TIMESTAMP=$(head -1 /tmp/gource.log | cut -d'|' -f1)
LAST_TIMESTAMP=$(tail -1 /tmp/gource.log | cut -d'|' -f1)
TOTAL_DAYS=$(( (LAST_TIMESTAMP - FIRST_TIMESTAMP) / 86400 ))
COMMIT_COUNT=$(wc -l < /tmp/gource.log)

echo "Commits: ${COMMIT_COUNT}"
echo "Days of history: ${TOTAL_DAYS}"

# Calculate seconds-per-day from target duration and slow factor
TARGET_DURATION="${GOURCE_DURATION:-60}"
SLOW_FACTOR="${GOURCE_SLOW:-1.0}"

if [ "$TOTAL_DAYS" -gt 0 ]; then
    # seconds_per_day = (target_duration * slow_factor) / total_days
    SECONDS_PER_DAY=$(awk "BEGIN {printf \"%.4f\", ($TARGET_DURATION * $SLOW_FACTOR) / $TOTAL_DAYS}")
else
    SECONDS_PER_DAY="1.0"
fi

FINAL_DURATION=$(awk "BEGIN {printf \"%.0f\", $TARGET_DURATION * $SLOW_FACTOR}")
TOTAL_FRAMES=$((FINAL_DURATION * 30))
echo "Target duration: ${TARGET_DURATION}s x ${SLOW_FACTOR} slow = ${FINAL_DURATION}s"
echo "Expected frames: ${TOTAL_FRAMES} (at 30fps)"
echo "Calculated speed: ${SECONDS_PER_DAY} seconds/day"
echo ""

# Check output mode
if [ "$GOURCE_OUTPUT_MODE" = "file" ]; then
    # File output mode - render to video file
    OUTPUT_FORMAT="${GOURCE_OUTPUT_FORMAT:-mp4}"
    OUTPUT_FILENAME="${GOURCE_OUTPUT_FILENAME:-output.mp4}"
    OUTPUT_PATH="/visualization/output/${OUTPUT_FILENAME}"

    echo "Rendering to file: ${OUTPUT_PATH}"
    echo "Format: ${OUTPUT_FORMAT}"
    echo "Resolution: ${GOURCE_RESOLUTION}"
    echo ""

    # Determine FFmpeg codec settings based on format
    if [ "$OUTPUT_FORMAT" = "webm" ]; then
        FFMPEG_OUTPUT_ARGS="-c:v libvpx-vp9 -crf 35 -b:v 0 -cpu-used 4 -deadline realtime -pix_fmt yuv420p"
    else
        FFMPEG_OUTPUT_ARGS="-c:v libx264 -preset ultrafast -crf 25 -pix_fmt yuv420p -movflags +faststart"
    fi

    echo "Starting render pipeline..."
    echo ""

    # Export for progress script
    export TOTAL_FRAMES

    gource /tmp/gource.log \
        --title "$GOURCE_TITLE" \
        --${GOURCE_RESOLUTION} \
        --seconds-per-day ${SECONDS_PER_DAY} \
        --auto-skip-seconds 0.5 \
        --file-idle-time ${GOURCE_FILE_IDLE_TIME:-1} \
        --dir-name-depth 3 \
        --hide mouse,bloom,filenames,progress \
        --highlight-dirs \
        --highlight-users \
        --elasticity 0.005 \
        --user-friction 2.0 \
        --max-user-speed 30 \
        --max-file-lag 0.5 \
        --camera-mode overview \
        --disable-auto-rotate \
        --padding 1.2 \
        --font-size 16 \
        --user-font-size 18 \
        --filename-time 2.0 \
        --output-framerate 30 \
        --stop-at-end \
        --output-ppm-stream - \
        2>/dev/null \
    | ffmpeg -y \
            -f image2pipe -vcodec ppm -r 30 -i - \
            -threads 0 \
            -progress pipe:1 \
            $FFMPEG_OUTPUT_ARGS \
            "${OUTPUT_PATH}" 2>/dev/null \
    | awk -v total="$TOTAL_FRAMES" '
            BEGIN { start_time = systime(); frame = 0; speed = "N/A" }
            /^frame=/ {
                gsub(/frame=/, "", $0);
                frame = int($0)
            }
            /^speed=/ {
                gsub(/speed=/, "", $0);
                speed = $0
            }
            /^progress=/ {
                if (total > 0 && frame > 0) {
                    pct = (frame / total) * 100
                    elapsed = systime() - start_time
                    if (pct > 0) {
                        eta = int((elapsed / pct) * (100 - pct))
                        eta_min = int(eta / 60)
                        eta_sec = eta % 60
                        printf "\rProgress: %d/%d frames (%.1f%%) | Speed: %s | ETA: %dm%02ds    ", frame, total, pct, speed, eta_min, eta_sec
                    }
                }
            }
            END { print "" }
            '

    echo ""
    echo ""

    # Check if video was created successfully
    if [ -s "${OUTPUT_PATH}" ]; then
        echo "Video rendered successfully: ${OUTPUT_PATH}"
        ls -lh "${OUTPUT_PATH}"
    else
        echo "ERROR: Video rendering failed - output file is empty or missing"
        exit 1
    fi
else
    # Streaming mode - start nginx and output to HLS
    nginx &

    exec gource /tmp/gource.log \
        --title "$GOURCE_TITLE" \
        --${GOURCE_RESOLUTION} \
        --seconds-per-day ${SECONDS_PER_DAY} \
        --auto-skip-seconds 1 \
        --file-idle-time ${GOURCE_FILE_IDLE_TIME:-1} \
        --dir-name-depth 3 \
        --hide mouse,bloom,filenames \
        --highlight-dirs \
        --highlight-users \
        --elasticity 0.005 \
        --user-friction 2.0 \
        --max-user-speed 30 \
        --max-file-lag 0.5 \
        --camera-mode overview \
        --disable-auto-rotate \
        --padding 1.2 \
        --font-size 16 \
        --user-font-size 18 \
        --filename-time 2.0 \
        --output-framerate 30 \
        --output-ppm-stream - \
        | ffmpeg -y \
            -f image2pipe -vcodec ppm -r 30 -i - \
            -threads 0 \
            -c:v libx264 -preset ultrafast -tune zerolatency \
            -profile:v baseline -level 3.1 \
            -crf 28 \
            -g 30 -keyint_min 30 \
            -bf 0 \
            -pix_fmt yuv420p \
            -f hls \
            -hls_time 1 \
            -hls_list_size 0 \
            -hls_flags append_list+independent_segments+split_by_time \
            -hls_segment_type mpegts \
            /visualization/hls/stream.m3u8
fi
