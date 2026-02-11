#!/usr/bin/env sh
set -e

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DEST_DIR="$ROOT_DIR/third_party/google/api"

mkdir -p "$DEST_DIR"

curl -fsSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto -o "$DEST_DIR/annotations.proto"
curl -fsSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto -o "$DEST_DIR/http.proto"

echo "Downloaded google api protos to $DEST_DIR"
