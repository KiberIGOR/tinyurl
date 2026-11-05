#!/usr/bin/env bash
# Нагрузка hey + снимок heap-профиля pprof.
# Использование: ./scripts/load_and_profile.sh profiles/base.pprof

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${1:-profiles/heap.pprof}"
ADDR="${SERVER_ADDR:-:8099}"
PPROF="${PPROF_ADDR:-localhost:6060}"
BASE_URL="http://127.0.0.1${ADDR}"
STORAGE="${FILE_STORAGE:-/tmp/tinyurl_bench.json}"

mkdir -p "$(dirname "$ROOT/$OUT")"

if ! command -v hey >/dev/null 2>&1; then
  echo "hey not found; install: go install github.com/rakyll/hey@latest"
  exit 1
fi

cd "$ROOT"
go build -o /tmp/tinyurl-shortener ./cmd/shortener

/tmp/tinyurl-shortener -a "$ADDR" -f "$STORAGE" &
PID=$!
trap 'kill $PID 2>/dev/null || true' EXIT

for i in $(seq 1 30); do
  if curl -sf "$BASE_URL/ping" >/dev/null; then
    break
  fi
  sleep 0.2
done

COOKIE=$(curl -si -X POST "$BASE_URL/" \
  -H "Content-Type: text/plain" \
  -d "https://warmup.example.com/" | tr -d '\r' | grep -i '^set-cookie:' | head -1 | sed 's/^[Ss]et-[Cc]ookie: //;s/;.*//')

SHORT=$(curl -s -X POST "$BASE_URL/" \
  -H "Content-Type: text/plain" \
  -H "Cookie: $COOKIE" \
  -d "https://follow.example.com/long/path")

hey -z 15s -c 40 -m POST \
  -H "Content-Type: text/plain" \
  -H "Cookie: $COOKIE" \
  -d "https://load.example.com/page" \
  "$BASE_URL/" >/dev/null

hey -z 10s -c 40 -m GET \
  -H "Cookie: $COOKIE" \
  "$BASE_URL/ping" >/dev/null

if [ -n "$SHORT" ]; then
  ID="${SHORT##*/}"
  hey -z 15s -c 50 -m GET \
    -H "Cookie: $COOKIE" \
    "$BASE_URL/$ID" >/dev/null
fi

curl -sf "http://$PPROF/debug/pprof/heap?gc=1" -o "$ROOT/$OUT"
echo "saved $OUT"
