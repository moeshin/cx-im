#!/usr/bin/env bash

out='keys.go'

echo 'package config

const (' > "$out"

while read -r l; do
  if [ -z "$l" ]; then
    echo "" >> "$out"
  else
    echo "$l = \"$l\"" >> "$out"
  fi
done < keys.txt

echo ")" >> "$out"

gofmt -w "$out"