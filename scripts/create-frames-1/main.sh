#!/bin/bash

set -euo pipefail

sourceVid="$1"
outputPath="$(dirname "$sourceVid")"
fps="${3:-2/1}"
framesPath="$outputPath/frames"
pickedFramesPath="$outputPath/picked-frames"

mkdir -p "$framesPath"

rm -rf "$pickedFramesPath" || true
mkdir -p "$pickedFramesPath"

# Export frames to images
ffmpeg -i "$sourceVid" -vf fps="$fps" "$framesPath/img-%04d.jpg" -hide_banner

find "$framesPath" -type f -exec ln -s "{}" "$pickedFramesPath/." ";"
