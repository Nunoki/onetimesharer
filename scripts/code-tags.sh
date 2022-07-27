#!/bin/sh
# Finds occurrences of the conventional developer tags left in code (such as TODO, FIXME, etc), in # all files of the project (excluding the vendor directory, and self) and outputs them together with # filenames and line numbers where they appear."

# find . -type f \( -name "*.md" -or -name "*.txt" -or -name "*.go" -or -name "*.sh" \) -not -path './vendor/*' -not -path ${BASH_SOURCE} | xargs grep -n -E 'TODO|FIXME|DOCME|BUG|XXX|HACK|DEPRECATED|REMOVE'

# find all non-binary files and pipe, ref: https://unix.stackexchange.com/a/46290
find . -type f -not -path './vendor/*' -not -path ${BASH_SOURCE} -exec file {} + | \
  awk -F: '/ASCII text/ {print $1}' | \
  xargs grep -n -E 'TODO|FIXME|DOCME|BUG|XXX|HACK|DEPRECATED|REMOVE'
