#!/usr/bin/env bash
# Encodes queries via the embed sidecar; usage: ./scripts/diag/search_scores.bash "soc2" "SOC 2"
set -euo pipefail

EMBED_URI="${WEBLENS_EMBED_URI:-http://localhost:5500}"

for q in "$@"; do
  echo "=== query: $q ==="
  curl -s -X POST "$EMBED_URI/encode-text" \
    -H 'Content-Type: application/json' \
    -d "$(jq -nc --arg t "$q" '{text:$t}')" \
  | jq '{plain_len: (.text_features|length), prompted_len: (.image_query_features|length)}' \
    2>/dev/null || echo "  (encode-text returned legacy shape — run Phase 2 first)"
done

echo
echo "Now compare these vectors against stored embeddings with mongosh:"
echo "  db.embeddings.aggregate([{\$vectorSearch:{index:'embeddings_vector',path:'vector',queryVector:<vec>,numCandidates:200,limit:20}},{\$project:{kind:1,snippet:1,score:{\$meta:'vectorSearchScore'}}}])"
