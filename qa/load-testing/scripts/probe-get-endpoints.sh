#!/usr/bin/env bash
# Wrapper for probe_get.py
exec python3 "$(dirname "$0")/probe_get.py" "$@"
