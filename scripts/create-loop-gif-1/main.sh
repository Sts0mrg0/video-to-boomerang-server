#!/bin/bash

set -x
set -euo pipefail

output="$1"
outputPath="$(dirname "$output")"
framesPath="$outputPath/picked-frames"

mkdir -p "$framesPath"

# Create gif from files (sort + reverse sort)
files="$(ls -1 "$framesPath"/*.jpg | sort) $(ls -1 "$framesPath"/*.jpg | sort -r)"
gm convert $files -loop 0 -coalesce -resize "60%" -quality "80%" "$output"
