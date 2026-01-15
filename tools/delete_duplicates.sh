#!/bin/bash

# Script to find and remove duplicate words from word list files
# Usage: ./delete_duplicates.sh [file_path]

if [ $# -eq 0 ]; then
    echo "Usage: $0 <translation_file_name>"
    echo "Example: $0 en_us"
    exit 1
fi

FILE="internal/game/words/$1"

if [ ! -f "$FILE" ]; then
    echo "Error: File '$FILE' not found"
    exit 1
fi

echo "Checking for duplicates in: $FILE"
echo ""

# Find and display duplicates
DUPLICATES=$(sort "$FILE" | uniq -d)

if [ -z "$DUPLICATES" ]; then
    echo "✓ No duplicates found!"
    exit 0
fi

echo "Found the following duplicate words:"
echo "$DUPLICATES" | nl
echo ""

# Count duplicates
DUPLICATE_COUNT=$(echo "$DUPLICATES" | wc -l)
echo "Total duplicates: $DUPLICATE_COUNT"
echo ""

# Show line numbers of each duplicate
echo "Line numbers of duplicates:"
while IFS= read -r word; do
    grep -n "^$word$" "$FILE" | cut -d: -f1 | paste -sd ',' - | sed "s/^/  $word: /"
done <<< "$DUPLICATES"
echo ""

# Ask for confirmation before removing
read -p "Do you want to remove duplicates? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    # Create a backup
    BACKUP="${FILE}.backup"
    cp "$FILE" "$BACKUP"
    echo "Backup created: $BACKUP"
    
    # Remove duplicates by keeping only first occurrence
    sort "$FILE" | uniq > "${FILE}.tmp"
    mv "${FILE}.tmp" "$FILE"
    
    echo "✓ Duplicates removed successfully!"
    echo "File has been updated: $FILE"
else
    echo "Cancelled. No changes made."
fi
