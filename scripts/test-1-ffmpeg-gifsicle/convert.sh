#!/bin/bash

set -x
set -euo pipefail

vidPath="$1"
sourceVid="$1/vid.mp4"

#avconv -i vid.mp4 -ss 00:15 -t 12 -an -q:v 1 -r 25 -f image2 tmp/img_%05d.jpg

# Export frames to images
#ffmpeg -i "$sourceVid" -vf fps=2/1 "$vidPath/img-%04d.jpg" -hide_banner
# Create gif from files (sort + reverse sort)
#files="$(ls -1 "$vidPath"/*.jpg | sort) $(ls -1 "$vidPath"/*.jpg | sort -r)"
#gm convert $files -loop 0 -coalesce -resize "60%" -quality "80%" "$vidPath"/output.gif
#convert $files -coalesce -loop 0 -fuzz 5% -layers Optimize -quality "80%" "$vidPath"/output.gif


#ffmpeg -i "$sourceVid" -vf fps=2/1 -f gif -pix_fmt rgb24 "$vidPath/img-%04d.gif" -hide_banner
#files="$(ls -1 "$vidPath"/*.gif | sort) $(ls -1 "$vidPath"/*.gif | sort -r)"
#gifsicle -O3 --loop 0 $files > "$vidPath"/output.gif

ffmpeg -i "$sourceVid" -vf "scale=min(iw\,600):-1,fps=2/1" -pix_fmt rgb24 -r 10 -f gif - | gifsicle --optimize=3 --delay=10 --colors 128 > "$vidPath"/output.gif
#gifsicle --optimize=3 --delay=10 --colors 128 "$vidPath"/output1.gif  > "$vidPath"/output1.gif

#ffmpeg -f image2 -pattern_type glob -i "tmp/img-*.jpg" -loop 1 tmp/out.gif
