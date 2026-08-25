#!/usr/bin/env bash
# Local-only seed against builder-api. Does not run in production or Docker builds.
# Requires: compose (or make run) with builder-api healthy, jq, uuidgen.
#
# Workflows call public APIs (no API keys): httpbin.org, open-meteo.com,
# jsonplaceholder.typicode.com. Every non-empty graph uses all five node types
# (start, function, conditional, api, response) with valid config_schema values.
set -euo pipefail

BASE="${BUILDER_URL:-http://localhost:${HTTP_PORT:-8080}}"
EXEC="${EXECUTION_URL:-http://localhost:${EXECUTION_HTTP_PORT:-8081}}"

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

put_graph() {
  local id="$1"
  local body="$2"
  curl -fsS -X PUT "$BASE/v1/workflows/$id" \
    -H 'Content-Type: application/json' \
    -d "$body"
}

put_expect_400() {
  local id="$1"
  local body="$2"
  local label="$3"
  local tmp code
  tmp="$(mktemp)"
  code="$(curl -sS -o "$tmp" -w '%{http_code}' -X PUT "$BASE/v1/workflows/$id" \
    -H 'Content-Type: application/json' \
    -d "$body")"
  if [[ "$code" != "400" ]]; then
    echo "expected 400 for $label, got $code body=$(cat "$tmp")" >&2
    rm -f "$tmp"
    exit 1
  fi
  echo "  ok ($label -> 400): $(jq -r .error "$tmp")"
  rm -f "$tmp"
}

# ---------------------------------------------------------------------------
# High-value order notify (httpbin)
# Start -> Function -> Conditional
#   true  -> POST https://httpbin.org/post  -> 200
#   false -> 202 Accepted (skip notify)
# Public API: httpbin echoes the JSON body (no auth).
# ---------------------------------------------------------------------------
graph_order_notify_json() {
  local name="$1"
  local desc="$2"
  local include_false_edge="${3:-yes}"
  local start fn cond api ok skip
  local e_start e_fn e_true e_api e_false
  start="$(uuid)"
  fn="$(uuid)"
  cond="$(uuid)"
  api="$(uuid)"
  ok="$(uuid)"
  skip="$(uuid)"
  e_start="$(uuid)"
  e_fn="$(uuid)"
  e_true="$(uuid)"
  e_api="$(uuid)"
  e_false="$(uuid)"

  local fn_source
  fn_source='const p = data.payload || {};
const amount = Number(p.amount_cents || 0);
return {
  order_id: p.order_id || null,
  customer_email: p.email || null,
  amount_cents: amount,
  currency: p.currency || "usd",
  high_value: amount >= 10000
};'

  jq -n \
    --arg name "$name" \
    --arg desc "$desc" \
    --arg start "$start" \
    --arg fn "$fn" \
    --arg cond "$cond" \
    --arg api "$api" \
    --arg ok "$ok" \
    --arg skip "$skip" \
    --arg e_start "$e_start" \
    --arg e_fn "$e_fn" \
    --arg e_true "$e_true" \
    --arg e_api "$e_api" \
    --arg e_false "$e_false" \
    --arg fn_source "$fn_source" \
    --argjson edges "$(jq -n \
      --arg e_start "$e_start" --arg start "$start" --arg fn "$fn" \
      --arg e_fn "$e_fn" --arg cond "$cond" \
      --arg e_true "$e_true" --arg api "$api" \
      --arg e_api "$e_api" --arg ok "$ok" \
      --arg e_false "$e_false" --arg skip "$skip" \
      --arg include "$include_false_edge" \
      '
        [
          {id: $e_start, from_node_id: $start, to_node_id: $fn, label: "default"},
          {id: $e_fn, from_node_id: $fn, to_node_id: $cond, label: "default"},
          {id: $e_true, from_node_id: $cond, to_node_id: $api, label: "true"},
          {id: $e_api, from_node_id: $api, to_node_id: $ok, label: "default"}
        ] + (if $include == "yes" then
          [{id: $e_false, from_node_id: $cond, to_node_id: $skip, label: "false"}]
        else [] end)
      ')" \
    '{
      name: $name,
      description: $desc,
      nodes: [
        {
          id: $start,
          node_type: "start",
          name: "Order placed",
          config: {description: "Manual/API trigger with order_id, email, amount_cents"},
          position: {x: 40, y: 120}
        },
        {
          id: $fn,
          node_type: "function",
          name: "Score high-value order",
          config: {runtime: "js", source: $fn_source, timeout_ms: 3000},
          position: {x: 280, y: 120}
        },
        {
          id: $cond,
          node_type: "conditional",
          name: "Amount >= $100?",
          config: {expression: "data.high_value == true"},
          position: {x: 540, y: 120}
        },
        {
          id: $api,
          node_type: "api",
          name: "Notify via httpbin",
          config: {
            method: "POST",
            url: "https://httpbin.org/post",
            headers: {
              "Content-Type": "application/json",
              Accept: "application/json",
              "X-Orchex-Seed": "high-value-order"
            },
            body_template: "{\"event\":\"high_value_order\",\"source\":\"orchex-seed\"}",
            timeout_ms: 20000
          },
          position: {x: 800, y: 40}
        },
        {
          id: $ok,
          node_type: "response",
          name: "Notified",
          config: {
            status_code: 200,
            body_template: "{\"status\":\"notified\",\"channel\":\"httpbin\"}",
            headers: {"Content-Type": "application/json"}
          },
          position: {x: 1060, y: 40}
        },
        {
          id: $skip,
          node_type: "response",
          name: "Skip notify",
          config: {
            status_code: 202,
            body_template: "{\"status\":\"accepted\",\"notified\":false,\"reason\":\"below_threshold\"}",
            headers: {"Content-Type": "application/json"}
          },
          position: {x: 800, y: 220}
        }
      ],
      edges: $edges
    }'
}

# ---------------------------------------------------------------------------
# Weather gate (Open-Meteo, no API key)
# Start -> Function -> Conditional
#   true  -> GET open-meteo forecast -> 200
#   false -> 400 (coords missing)
# ---------------------------------------------------------------------------
graph_weather_json() {
  local name="$1"
  local desc="$2"
  local start fn cond api ok bad
  local e_start e_fn e_true e_api e_false
  start="$(uuid)"
  fn="$(uuid)"
  cond="$(uuid)"
  api="$(uuid)"
  ok="$(uuid)"
  bad="$(uuid)"
  e_start="$(uuid)"
  e_fn="$(uuid)"
  e_true="$(uuid)"
  e_api="$(uuid)"
  e_false="$(uuid)"

  local fn_source
  fn_source='const p = data.payload || {};
const lat = Number(p.latitude);
const lon = Number(p.longitude);
const valid = Number.isFinite(lat) && Number.isFinite(lon) && lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180;
return {
  latitude: valid ? lat : null,
  longitude: valid ? lon : null,
  place: p.place || null,
  coords_ok: valid
};'

  jq -n \
    --arg name "$name" \
    --arg desc "$desc" \
    --arg start "$start" \
    --arg fn "$fn" \
    --arg cond "$cond" \
    --arg api "$api" \
    --arg ok "$ok" \
    --arg bad "$bad" \
    --arg e_start "$e_start" \
    --arg e_fn "$e_fn" \
    --arg e_true "$e_true" \
    --arg e_api "$e_api" \
    --arg e_false "$e_false" \
    --arg fn_source "$fn_source" \
    '{
      name: $name,
      description: $desc,
      nodes: [
        {
          id: $start,
          node_type: "start",
          name: "Weather request",
          config: {description: "Payload: latitude, longitude, optional place"},
          position: {x: 40, y: 100}
        },
        {
          id: $fn,
          node_type: "function",
          name: "Validate coordinates",
          config: {runtime: "js", source: $fn_source, timeout_ms: 2000},
          position: {x: 280, y: 100}
        },
        {
          id: $cond,
          node_type: "conditional",
          name: "Coords valid?",
          config: {expression: "data.coords_ok == true"},
          position: {x: 520, y: 100}
        },
        {
          id: $api,
          node_type: "api",
          name: "Open-Meteo current weather",
          config: {
            method: "GET",
            url: "https://api.open-meteo.com/v1/forecast",
            headers: {Accept: "application/json"},
            query: {
              latitude: "51.5074",
              longitude: "-0.1278",
              current: "temperature_2m,weather_code",
              timezone: "auto"
            },
            timeout_ms: 15000
          },
          position: {x: 780, y: 20}
        },
        {
          id: $ok,
          node_type: "response",
          name: "Weather OK",
          config: {
            status_code: 200,
            body_template: "{\"status\":\"ok\",\"provider\":\"open-meteo\"}",
            headers: {"Content-Type": "application/json"}
          },
          position: {x: 1040, y: 20}
        },
        {
          id: $bad,
          node_type: "response",
          name: "Invalid coords",
          config: {
            status_code: 400,
            body_template: "{\"error\":\"invalid_coordinates\"}",
            headers: {"Content-Type": "application/json"}
          },
          position: {x: 780, y: 200}
        }
      ],
      edges: [
        {id: $e_start, from_node_id: $start, to_node_id: $fn, label: "default"},
        {id: $e_fn, from_node_id: $fn, to_node_id: $cond, label: "default"},
        {id: $e_true, from_node_id: $cond, to_node_id: $api, label: "true"},
        {id: $e_api, from_node_id: $api, to_node_id: $ok, label: "default"},
        {id: $e_false, from_node_id: $cond, to_node_id: $bad, label: "false"}
      ]
    }'
}

# ---------------------------------------------------------------------------
# Demo post to JSONPlaceholder (public fake REST API)
# Start -> Function -> Conditional
#   true  -> POST /posts -> 201-style body from JSONPlaceholder
#   false -> 400
# ---------------------------------------------------------------------------
graph_jsonplaceholder_json() {
  local name="$1"
  local desc="$2"
  local start fn cond api ok bad
  local e_start e_fn e_true e_api e_false
  start="$(uuid)"
  fn="$(uuid)"
  cond="$(uuid)"
  api="$(uuid)"
  ok="$(uuid)"
  bad="$(uuid)"
  e_start="$(uuid)"
  e_fn="$(uuid)"
  e_true="$(uuid)"
  e_api="$(uuid)"
  e_false="$(uuid)"

  local fn_source
  fn_source='const p = data.payload || {};
const title = typeof p.title === "string" ? p.title.trim() : "";
const body = typeof p.body === "string" ? p.body.trim() : "";
return {
  title,
  body,
  user_id: Number(p.user_id) || 1,
  publishable: title.length > 0 && body.length > 0
};'

  jq -n \
    --arg name "$name" \
    --arg desc "$desc" \
    --arg start "$start" \
    --arg fn "$fn" \
    --arg cond "$cond" \
    --arg api "$api" \
    --arg ok "$ok" \
    --arg bad "$bad" \
    --arg e_start "$e_start" \
    --arg e_fn "$e_fn" \
    --arg e_true "$e_true" \
    --arg e_api "$e_api" \
    --arg e_false "$e_false" \
    --arg fn_source "$fn_source" \
    '{
      name: $name,
      description: $desc,
      nodes: [
        {
          id: $start,
          node_type: "start",
          name: "Create post request",
          config: {description: "Payload: title, body, optional user_id"},
          position: {x: 40, y: 100}
        },
        {
          id: $fn,
          node_type: "function",
          name: "Validate post fields",
          config: {runtime: "js", source: $fn_source, timeout_ms: 2000},
          position: {x: 280, y: 100}
        },
        {
          id: $cond,
          node_type: "conditional",
          name: "Title and body set?",
          config: {expression: "data.publishable == true"},
          position: {x: 540, y: 100}
        },
        {
          id: $api,
          node_type: "api",
          name: "POST jsonplaceholder /posts",
          config: {
            method: "POST",
            url: "https://jsonplaceholder.typicode.com/posts",
            headers: {
              "Content-Type": "application/json",
              Accept: "application/json"
            },
            body_template: "{\"title\":\"Orchex seed post\",\"body\":\"Created by local seed workflow\",\"userId\":1}",
            timeout_ms: 15000
          },
          position: {x: 800, y: 20}
        },
        {
          id: $ok,
          node_type: "response",
          name: "Post created",
          config: {
            status_code: 201,
            body_template: "{\"status\":\"created\",\"provider\":\"jsonplaceholder\"}",
            headers: {"Content-Type": "application/json"}
          },
          position: {x: 1060, y: 20}
        },
        {
          id: $bad,
          node_type: "response",
          name: "Validation failed",
          config: {
            status_code: 400,
            body_template: "{\"error\":\"title_and_body_required\"}",
            headers: {"Content-Type": "application/json"}
          },
          position: {x: 800, y: 200}
        }
      ],
      edges: [
        {id: $e_start, from_node_id: $start, to_node_id: $fn, label: "default"},
        {id: $e_fn, from_node_id: $fn, to_node_id: $cond, label: "default"},
        {id: $e_true, from_node_id: $cond, to_node_id: $api, label: "true"},
        {id: $e_api, from_node_id: $api, to_node_id: $ok, label: "default"},
        {id: $e_false, from_node_id: $cond, to_node_id: $bad, label: "false"}
      ]
    }'
}

# Broken API config (all node types present; API config empty -> 400).
graph_bad_api_config_json() {
  local name="$1"
  local start fn cond api ok skip
  local e_start e_fn e_true e_api e_false
  start="$(uuid)"
  fn="$(uuid)"
  cond="$(uuid)"
  api="$(uuid)"
  ok="$(uuid)"
  skip="$(uuid)"
  e_start="$(uuid)"
  e_fn="$(uuid)"
  e_true="$(uuid)"
  e_api="$(uuid)"
  e_false="$(uuid)"
  jq -n \
    --arg name "$name" \
    --arg start "$start" --arg fn "$fn" --arg cond "$cond" \
    --arg api "$api" --arg ok "$ok" --arg skip "$skip" \
    --arg e_start "$e_start" --arg e_fn "$e_fn" --arg e_true "$e_true" \
    --arg e_api "$e_api" --arg e_false "$e_false" \
    '{
      name: $name,
      description: "WIP webhook forwarder — API node missing method/url on purpose",
      nodes: [
        {id: $start, node_type: "start", name: "Inbound webhook", config: {description: "test trigger"}},
        {
          id: $fn,
          node_type: "function",
          name: "Pass through",
          config: {runtime: "js", source: "return data.payload || {};", timeout_ms: 1000}
        },
        {
          id: $cond,
          node_type: "conditional",
          name: "Always true?",
          config: {expression: "true"}
        },
        {id: $api, node_type: "api", name: "Forward to partner", config: {}},
        {id: $ok, node_type: "response", name: "OK", config: {status_code: 200}},
        {id: $skip, node_type: "response", name: "Skipped", config: {status_code: 204}}
      ],
      edges: [
        {id: $e_start, from_node_id: $start, to_node_id: $fn, label: "default"},
        {id: $e_fn, from_node_id: $fn, to_node_id: $cond, label: "default"},
        {id: $e_true, from_node_id: $cond, to_node_id: $api, label: "true"},
        {id: $e_api, from_node_id: $api, to_node_id: $ok, label: "default"},
        {id: $e_false, from_node_id: $cond, to_node_id: $skip, label: "false"}
      ]
    }'
}

echo "seeding public-API workflows against $BASE"
echo

w_empty="$(create \
  "Untitled automation" \
  "Blank draft — no nodes yet." | jq -r .id)"
echo "untitled (empty)                 $w_empty"

w_order="$(create \
  "High-value order -> httpbin notify" \
  "If amount_cents >= 10000, POST to https://httpbin.org/post; else 202 without calling out." | jq -r .id)"
put_graph "$w_order" "$(graph_order_notify_json \
  "High-value order -> httpbin notify" \
  "If amount_cents >= 10000, POST to https://httpbin.org/post; else 202 without calling out." \
  yes)" >/dev/null
echo "order notify (draft)             $w_order"

w_weather="$(create \
  "London weather via Open-Meteo" \
  "Validate lat/lon then GET https://api.open-meteo.com/v1/forecast (no API key)." | jq -r .id)"
put_graph "$w_weather" "$(graph_weather_json \
  "London weather via Open-Meteo" \
  "Validate lat/lon then GET https://api.open-meteo.com/v1/forecast (no API key).")" >/dev/null
echo "weather (draft)                  $w_weather"

w_posts="$(create \
  "Create demo post (JSONPlaceholder)" \
  "Validate title/body then POST https://jsonplaceholder.typicode.com/posts." | jq -r .id)"
put_graph "$w_posts" "$(graph_jsonplaceholder_json \
  "Create demo post (JSONPlaceholder)" \
  "Validate title/body then POST https://jsonplaceholder.typicode.com/posts.")" >/dev/null
echo "jsonplaceholder (draft)          $w_posts"

w_incomplete="$(create \
  "High-value order (missing low-value path)" \
  "Same httpbin flow but Conditional has no false edge — publish must fail." | jq -r .id)"
put_graph "$w_incomplete" "$(graph_order_notify_json \
  "High-value order (missing low-value path)" \
  "Same httpbin flow but Conditional has no false edge — publish must fail." \
  no)" >/dev/null
echo "order notify (incomplete graph)  $w_incomplete"

w_bad_cfg="$(create \
  "Forward webhook to partner" \
  "All node types present; API config empty to prove config_schema rejection." | jq -r .id)"
echo "checking config_schema ..."
put_expect_400 "$w_bad_cfg" "$(graph_bad_api_config_json "Forward webhook to partner")" "empty partner API config"
echo "partner forward (bad cfg)        $w_bad_cfg  (left empty; bad PUT rejected)"

w_live="$(create \
  "High-value order -> httpbin notify" \
  "Published path: real POST to httpbin.org when order is high-value." | jq -r .id)"
put_graph "$w_live" "$(graph_order_notify_json \
  "High-value order -> httpbin notify" \
  "Published path: real POST to httpbin.org when order is high-value." \
  yes)" >/dev/null
curl -fsS -X POST "$BASE/v1/workflows/$w_live/publish" >/dev/null
echo "order notify (published)         $w_live"

w_archived="$(create \
  "Old httpbin ping (retired)" \
  "Retired smoke workflow against httpbin — archived." | jq -r .id)"
curl -fsS -X DELETE "$BASE/v1/workflows/$w_archived" >/dev/null
echo "httpbin ping (archived)          $w_archived  (omitted from list)"

echo
echo "-- builder curls --"
echo
echo "# list"
echo "curl -sS $BASE/v1/workflows | jq '.items[] | {id, name, status}'"
echo
echo "# order notify (httpbin) configs"
echo "curl -sS $BASE/v1/workflows/$w_order | jq '.graph.nodes[] | {name, node_type, config}'"
echo
echo "# weather (open-meteo) configs"
echo "curl -sS $BASE/v1/workflows/$w_weather | jq '.graph.nodes[] | {name, node_type, config}'"
echo
echo "# jsonplaceholder configs"
echo "curl -sS $BASE/v1/workflows/$w_posts | jq '.graph.nodes[] | {name, node_type, config}'"
echo
echo "# published order notify"
echo "curl -sS '$BASE/v1/workflows/$w_live?version=published' | jq '.graph.nodes[] | {name, node_type, config}'"
echo
echo "# re-save weather graph"
weather_body="$(graph_weather_json \
  "London weather via Open-Meteo" \
  "Validate lat/lon then GET https://api.open-meteo.com/v1/forecast (no API key).")"
printf '%s\n' \
  "curl -sS -X PUT $BASE/v1/workflows/$w_weather \\" \
  "  -H 'Content-Type: application/json' \\" \
  "  -d $(jq -c -n --argjson b "$weather_body" '$b | @json') | jq '.graph.nodes[] | {name, config}'"
echo
echo "# bad API config must 400"
bad_body="$(graph_bad_api_config_json "Forward webhook to partner")"
printf '%s\n' \
  "curl -i -sS -X PUT $BASE/v1/workflows/$w_bad_cfg \\" \
  "  -H 'Content-Type: application/json' \\" \
  "  -d $(jq -c -n --argjson b "$bad_body" '$b | @json')"
echo
echo "# publish order draft"
echo "curl -i -sS -X POST $BASE/v1/workflows/$w_order/publish"
echo
echo "# publish should fail: empty / missing false branch"
echo "curl -i -sS -X POST $BASE/v1/workflows/$w_empty/publish"
echo "curl -i -sS -X POST $BASE/v1/workflows/$w_incomplete/publish"
echo
echo "# archive bad-config draft"
echo "curl -i -sS -X DELETE $BASE/v1/workflows/$w_bad_cfg"
echo
echo "-- sample run payloads (execution, when wired) --"
order_hi="$(jq -c -n '{payload:{order_id:"ord_9001",email:"buyer@example.com",amount_cents:14999,currency:"usd"}}')"
order_lo="$(jq -c -n '{payload:{order_id:"ord_22",email:"buyer@example.com",amount_cents:2500,currency:"usd"}}')"
weather_payload="$(jq -c -n '{payload:{latitude:51.5074,longitude:-0.1278,place:"London"}}')"
post_payload="$(jq -c -n '{payload:{title:"Hello Orchex",body:"Public API seed post",user_id:1}}')"
echo "curl -sS -X POST $EXEC/v1/runs/$w_live -H 'Content-Type: application/json' -d $(jq -n --arg p "$order_hi" '$p|@json')"
echo "curl -sS -X POST $EXEC/v1/runs/$w_order -H 'Content-Type: application/json' -d $(jq -n --arg p "$order_lo" '$p|@json')"
echo "curl -sS -X POST $EXEC/v1/runs/$w_weather -H 'Content-Type: application/json' -d $(jq -n --arg p "$weather_payload" '$p|@json')"
echo "curl -sS -X POST $EXEC/v1/runs/$w_posts -H 'Content-Type: application/json' -d $(jq -n --arg p "$post_payload" '$p|@json')"
