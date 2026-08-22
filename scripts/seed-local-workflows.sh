#!/usr/bin/env bash
# Local-only seed against builder-api. Does not run in production or Docker builds.
# Requires: compose (or make run) with builder-api healthy, jq, uuidgen.
set -euo pipefail

BASE="${BUILDER_URL:-http://localhost:${HTTP_PORT:-8080}}"

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing $1" >&2
    exit 1
  }
}

need curl
need jq
need uuidgen

uuid() {
  uuidgen | tr '[:upper:]' '[:lower:]'
}

create() {
  local name="$1"
  local desc="$2"
  curl -fsS -X POST "$BASE/v1/workflows" \
    -H 'Content-Type: application/json' \
    -d "$(jq -n --arg n "$name" --arg d "$desc" '{name:$n, description:$d}')"
}

put_full() {
  local id="$1"
  local name="$2"
  local desc="$3"
  local start fn cond api ok fail
  local e_start e_fn e_true e_api e_false
  start="$(uuid)"
  fn="$(uuid)"
  cond="$(uuid)"
  api="$(uuid)"
  ok="$(uuid)"
  fail="$(uuid)"
  e_start="$(uuid)"
  e_fn="$(uuid)"
  e_true="$(uuid)"
  e_api="$(uuid)"
  e_false="$(uuid)"
  curl -fsS -X PUT "$BASE/v1/workflows/$id" \
    -H 'Content-Type: application/json' \
    -d "$(jq -n \
      --arg name "$name" \
      --arg desc "$desc" \
      --arg start "$start" \
      --arg fn "$fn" \
      --arg cond "$cond" \
      --arg api "$api" \
      --arg ok "$ok" \
      --arg fail "$fail" \
      --arg e_start "$e_start" \
      --arg e_fn "$e_fn" \
      --arg e_true "$e_true" \
      --arg e_api "$e_api" \
      --arg e_false "$e_false" \
      '{
        name: $name,
        description: $desc,
        nodes: [
          {id: $start, node_type: "start", name: "Start", config: {description: "local seed trigger"}},
          {id: $fn, node_type: "function", name: "Normalize", config: {runtime: "js", source: "return data;"}},
          {id: $cond, node_type: "conditional", name: "OK?", config: {expression: "data.payload.ok == true"}},
          {id: $api, node_type: "api", name: "Notify", config: {method: "GET", url: "https://example.com/health"}},
          {id: $ok, node_type: "response", name: "Success", config: {status_code: 200}},
          {id: $fail, node_type: "response", name: "Rejected", config: {status_code: 400}}
        ],
        edges: [
          {id: $e_start, from_node_id: $start, to_node_id: $fn, label: "default"},
          {id: $e_fn, from_node_id: $fn, to_node_id: $cond, label: "default"},
          {id: $e_true, from_node_id: $cond, to_node_id: $api, label: "true"},
          {id: $e_api, from_node_id: $api, to_node_id: $ok, label: "default"},
          {id: $e_false, from_node_id: $cond, to_node_id: $fail, label: "false"}
        ]
      }')"
}

put_incomplete() {
  local id="$1"
  local name="$2"
  local desc="$3"
  local start fn cond api ok fail
  local e_start e_fn e_true e_api
  start="$(uuid)"
  fn="$(uuid)"
  cond="$(uuid)"
  api="$(uuid)"
  ok="$(uuid)"
  fail="$(uuid)"
  e_start="$(uuid)"
  e_fn="$(uuid)"
  e_true="$(uuid)"
  e_api="$(uuid)"
  curl -fsS -X PUT "$BASE/v1/workflows/$id" \
    -H 'Content-Type: application/json' \
    -d "$(jq -n \
      --arg name "$name" \
      --arg desc "$desc" \
      --arg start "$start" \
      --arg fn "$fn" \
      --arg cond "$cond" \
      --arg api "$api" \
      --arg ok "$ok" \
      --arg fail "$fail" \
      --arg e_start "$e_start" \
      --arg e_fn "$e_fn" \
      --arg e_true "$e_true" \
      --arg e_api "$e_api" \
      '{
        name: $name,
        description: $desc,
        nodes: [
          {id: $start, node_type: "start", name: "Start", config: {description: "local seed trigger"}},
          {id: $fn, node_type: "function", name: "Normalize", config: {runtime: "js", source: "return data;"}},
          {id: $cond, node_type: "conditional", name: "OK?", config: {expression: "data.payload.ok == true"}},
          {id: $api, node_type: "api", name: "Notify", config: {method: "GET", url: "https://example.com/health"}},
          {id: $ok, node_type: "response", name: "Success", config: {status_code: 200}},
          {id: $fail, node_type: "response", name: "Rejected", config: {status_code: 400}}
        ],
        edges: [
          {id: $e_start, from_node_id: $start, to_node_id: $fn, label: "default"},
          {id: $e_fn, from_node_id: $fn, to_node_id: $cond, label: "default"},
          {id: $e_true, from_node_id: $cond, to_node_id: $api, label: "true"},
          {id: $e_api, from_node_id: $api, to_node_id: $ok, label: "default"}
        ]
      }')"
}

echo "seeding against $BASE"

draft_empty="$(create "local-draft-empty" "empty draft" | jq -r .id)"
echo "draft-empty           $draft_empty"

draft_ok="$(create "local-draft-publishable" "all node types, not published" | jq -r .id)"
put_full "$draft_ok" "local-draft-publishable" "all node types, not published" >/dev/null
echo "draft-publishable     $draft_ok"

draft_bad="$(create "local-draft-unpublishable" "all node types, missing false branch" | jq -r .id)"
put_incomplete "$draft_bad" "local-draft-unpublishable" "all node types, missing false branch" >/dev/null
echo "draft-unpublishable   $draft_bad"

published="$(create "local-published" "published graph with all node types" | jq -r .id)"
put_full "$published" "local-published" "published graph with all node types" >/dev/null
curl -fsS -X POST "$BASE/v1/workflows/$published/publish" >/dev/null
echo "published             $published"

archived="$(create "local-archived" "will be archived" | jq -r .id)"
curl -fsS -X DELETE "$BASE/v1/workflows/$archived" >/dev/null
echo "archived              $archived  (omitted from list)"

echo
echo "try:"
echo "  curl $BASE/v1/workflows"
echo "  curl -X POST http://localhost:${EXECUTION_HTTP_PORT:-8081}/v1/runs/$published"
echo "  curl -X POST http://localhost:${EXECUTION_HTTP_PORT:-8081}/v1/runs/$draft_ok"
echo
echo "publish should fail (empty / incomplete):"
echo "  curl -i -X POST $BASE/v1/workflows/$draft_empty/publish"
echo "  curl -i -X POST $BASE/v1/workflows/$draft_bad/publish"
