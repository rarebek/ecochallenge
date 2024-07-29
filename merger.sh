#!/bin/bash

# Output file
output_file="merged_content.txt"

# Clear the output file if it exists
> "$output_file"

# Find and loop through all files in the current directory and subdirectories
find . -type f | while read -r file; do
  # Get the base name of the file (remove the leading ./)
  base_name=$(basename "$file")

  # Skip go.mod and go.sum files
  if [[ "$base_name" == "go.mod" ]] || [[ "$base_name" == "go.sum" ]]; then
    continue
  fi

  # Get the file type using the file command
  file_type=$(file --mime-type -b "$file")

  # Skip binary and archive files
  if [[ "$file_type" != "text/plain" ]] && [[ "$file_type" != "application/octet-stream" ]]; then
    continue
  fi

  # Print the filename as a header (including the relative path)
  echo "#### $file ####" >> "$output_file"
  # Append the content of the file
  cat "$file" >> "$output_file"
  # Add a newline for separation
  echo -e "\n" >> "$output_file"
done

echo "All relevant files have been merged into $output_file."
