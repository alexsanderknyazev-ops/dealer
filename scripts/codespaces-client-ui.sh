#!/usr/bin/env bash
# Запускает dev-сервер клиентского UI (Vite) поверх docker-compose-стека.
# Vite-прокси ходит в клиентские гейтвеи через 127.0.0.1:8091/8093 (порты compose).
set -euo pipefail

cd "$(dirname "$0")/../frontend/client"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "ERROR: required command not found: $1" >&2
    exit 1
  }
}

need_cmd npm

[ -d node_modules ] || npm install
npm run dev
