#!/bin/bash

# Set the directory to store the files
DIR="load_test_files"

# Create the directory if it doesn't exist
mkdir -p "$DIR"
# Generate 5000 1MB files of mp4 format
for i in {1..5000}; do
    dd if=/dev/urandom of="$DIR/file_$i.mp4" bs=1M count=1
done

# Print a message indicating the files have been generated
echo "5000 1MB files of mp4 format have been generated in the $DIR directory"
