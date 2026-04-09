#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "usage: offline-stt.sh <audio-path>" >&2
  exit 1
fi

audio_path="$1"
transcript_path="${audio_path%.*}.txt"

if [ ! -f "$transcript_path" ]; then
  echo "missing transcript fixture: $transcript_path" >&2
  exit 1
fi

cat "$transcript_path"
