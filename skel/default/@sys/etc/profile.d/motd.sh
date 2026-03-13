#!/bin/bash
# Display MOTD with glow if available, otherwise fallback to cat
if command -v glow &> /dev/null; then
    glow -p /etc/motd.md 2>/dev/null
else
    cat /etc/motd.md 2>/dev/null
fi
