#!/bin/sh
set -e

# Read stdin (JSON input from FlowForge engine) to file
cat > /tmp/input.json

# Write user code to temp file
printf '%s' "$CODE" > /tmp/user_code.py

# Execute
exec python3 /tmp/user_code.py
