#!/bin/bash

set -euo pipefail
set -x

echo '
<!DOCTYPE html>
<html>
<body>
<table border="1">
<tr>
    <th rowspan="2">Name</th>
    <th colspan="2">Input</th>
    <th colspan="3">Output</th>
</tr>
<tr>
    <th>Name</th>
    <th>Size</th>
    <th>Timing</th>
    <th>Size</th>
    <th>Image</th>
</tr>
' > output.html

run_test() {
    local image_name start_time end_time
    local test_name=$1
    local input_file=$2
    local output_file=$3
    local output_filepath="$test_name/$output_file"

    image_name="$(basename "$test_name")"

    docker build -t "$image_name" -f "$test_name/Dockerfile" "$test_name"

    start_time=$(date +%s)
    docker run --rm -v "$(pwd)/$test_name":/app -v "$(pwd)/media:/media" "$image_name" /app/main.sh "/media/$input_file" "$output_file"
    end_time=$(date +%s)
    let runtime=$((end_time - start_time))

    input_size="$(du -h "media/$input_file" | awk '{ print $1 }')"
    output_size="$(du -h "$output_filepath" | awk '{ print $1 }')"

    echo "
        <tr>
            <td>$test_name</td>
            <td>$input_file</td>
            <td>$input_size</td>
            <td>${runtime}s</td>
            <td>$output_size</td>
            <td><img width=\"150\" src=\"file://$(pwd)/$output_filepath\"/></td>
        </tr>
    " >> output.html
}

for test_name in test-*; do
    [[ -n "${test_name:-}" ]] || break

    echo "<tr><td colspan=\"100\"><h4>$test_name</h4></td></tr>" >> output.html

    for input_name in media/*; do
        [[ -f "${input_name:-}" ]] || break

        input_filename="$(basename "$input_name")"
        output_filename="${input_filename%.*}"

        run_test "$test_name" "$input_filename" "$output_filename.gif"
    done
done

echo '
</table>
</body>
</html>
' >> output.html
