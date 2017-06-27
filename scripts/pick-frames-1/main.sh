#!/bin/bash

set -x
set -euo pipefail

outputPath="$1"
frames="${*:2}"
framesPath="$outputPath/frames"
pickedFramesPath="$outputPath/picked-frames"

rm -rf "${pickedFramesPath}"
mkdir -p "$pickedFramesPath"

for frame in $frames; do
    frame_file="$framesPath/$frame"
    [[ -f "${frame_file:-}" ]] || break
    ln -s "$frame_file" "$pickedFramesPath/$frame"
done

ls "$pickedFramesPath/"
