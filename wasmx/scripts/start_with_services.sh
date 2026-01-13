#!/usr/bin/env bash

# Custom one-validator chain bootstrapper for running the MCP search contract.
# Allows DB/env overrides via flags and auto-discovers code/contract ids.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

WASMX_ROOT_DEFAULT="/Users/user/dev/blockchain/wasmx"
CHAIN_ID_DEFAULT="mythos_7000-14"
OUTPUT_DIR_DEFAULT="$SCRIPT_DIR/testnet"
KEY_NAME_DEFAULT="node0"
KEYRING_BACKEND_DEFAULT="test"
MIN_GAS_PRICES_DEFAULT="10amyt"
FEES_DEFAULT="90000000000amyt"
GAS_DEFAULT="10000000"
RPC_ADDRESS_DEFAULT="tcp://localhost:26657"
RPC_STATUS_URL_DEFAULT="http://localhost:26657"
BROADCAST_MODE_DEFAULT="sync"
OAUTH2_ADDR_DEFAULT="mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqrxpdp7ms"
OAUTH2_KEYS_ADDR_DEFAULT="mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqr2sn852h"
HTTP_REGISTRY_ADDR_DEFAULT="mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqrgrkxhe6"
POST_START_WAIT_DEFAULT=20
POST_TX_WAIT_DEFAULT=6
SCRIPTS_FOLDER_DEFAULT="$SCRIPT_DIR"
LOG_LEVEL_DEFAULT="info"
SUPABASE_JWT_DEFAULT=""
STORAGE_SUPABASE_KEY_ENDPOINT_DEFAULT=""
NOMEN_FILES_DEFAULT="/Users/user/dev/blockchain/nomen/frontend/dist-nomen"
EXPLORER_FILES_DEFAULT="/Users/user/dev/blockchain/wasmxplorer/frontend/dist"

SCRIPTS_FOLDER="${SCRIPTS_FOLDER:-$SCRIPTS_FOLDER_DEFAULT}"
WASMX_ROOT="${WASMX_ROOT:-$WASMX_ROOT_DEFAULT}"
OUTPUT_DIR="${OUTPUT_DIR:-$OUTPUT_DIR_DEFAULT}"
CHAIN_ID="${CHAIN_ID:-$CHAIN_ID_DEFAULT}"
KEY_NAME="${KEY_NAME:-$KEY_NAME_DEFAULT}"
KEYRING_BACKEND="${KEYRING_BACKEND:-$KEYRING_BACKEND_DEFAULT}"
MIN_GAS_PRICES="${MIN_GAS_PRICES:-$MIN_GAS_PRICES_DEFAULT}"
FEES="${FEES:-$FEES_DEFAULT}"
GAS="${GAS:-$GAS_DEFAULT}"
RPC_ADDRESS="${RPC_ADDRESS:-$RPC_ADDRESS_DEFAULT}"
OAUTH2_ADDR="${OAUTH2_ADDR:-$OAUTH2_ADDR_DEFAULT}"
OAUTH2_KEYS_ADDR="${OAUTH2_KEYS_ADDR:-$OAUTH2_KEYS_ADDR_DEFAULT}"
HTTP_REGISTRY_ADDR="${HTTP_REGISTRY_ADDR:-$HTTP_REGISTRY_ADDR_DEFAULT}"
POST_START_WAIT="${POST_START_WAIT:-$POST_START_WAIT_DEFAULT}"
POST_TX_WAIT="${POST_TX_WAIT:-$POST_TX_WAIT_DEFAULT}"
NODE_HOME="${NODE_HOME:-$OUTPUT_DIR/node0/mythosd}"
LOG_LEVEL="${LOG_LEVEL:-$LOG_LEVEL_DEFAULT}"
SUPABASE_JWT="${SUPABASE_JWT:-$SUPABASE_JWT_DEFAULT}"
STORAGE_SUPABASE_KEY_ENDPOINT="${STORAGE_SUPABASE_KEY_ENDPOINT:-$STORAGE_SUPABASE_KEY_ENDPOINT_DEFAULT}"
NOMEN_FILES="${NOMEN_FILES:-$NOMEN_FILES_DEFAULT}"
EXPLORER_FILES="${EXPLORER_FILES:-$EXPLORER_FILES_DEFAULT}"
RECREATE_CONFIG="${RECREATE_CONFIG:-false}"
SIMPLE_STORAGE_WASM_FILE="${SIMPLE_STORAGE_WASM_FILE:-$WASMX_ROOT/tests/testdata/tinygo/simple_storage.wasm}"
SIMPLE_STORAGE_SCHEMA_FILE="${SIMPLE_STORAGE_SCHEMA_FILE:-$WASMX_ROOT/tests/testdata/tinygo/schemas/simple_storage_schema.json}"

usage() {
  cat <<EOF
Usage: $(basename "$0") [options]
  --scripts-folder <dir>      Sets SCRIPTS_FOLDER (default: $SCRIPTS_FOLDER_DEFAULT)
  --wasmx-root <dir>          Root of wasmx checkout (default: $WASMX_ROOT_DEFAULT)
  --output-dir <dir>          Chain output dir (default: script dir)
  --home <dir>                mythosd home (default: <output-dir>/node0/mythosd)
  --chain-id <id>             Chain ID (default: $CHAIN_ID_DEFAULT)
  --min-gas-prices <value>    Minimum gas prices (default: $MIN_GAS_PRICES_DEFAULT)
  --fees <amount>             Fees used for txs (default: $FEES_DEFAULT)
  --gas <amount>              Gas used for txs (default: $GAS_DEFAULT)
  --rpc <addr>                RPC address (default: $RPC_ADDRESS_DEFAULT)
  --rpc-status-url <url>      HTTP status URL for RPC health (default: derived from --rpc)
  --post-start-wait <sec>     Extra wait after RPC is up before first tx (default: $POST_START_WAIT_DEFAULT)
  --post-tx-wait <sec>        Wait after each tx before querying receipt (default: $POST_TX_WAIT_DEFAULT)
  --broadcast-mode <mode>     Tx broadcast mode: async|sync (default: $BROADCAST_MODE_DEFAULT)
  --oauth2-addr <addr>        OAuth2 system contract address
  --http-registry-addr <addr> HTTP registry system contract address
  --simple-storage-wasm-file <path> Simple storage wasm file to upload
  --simple-storage-schema-file <path> Simple storage JSON schema file to attach
  --recreate-config           Delete existing node directory before init-files
  -h, --help                  Show this help

Environment overrides also respected:
  SCRIPTS_FOLDER, WASMX_ROOT, OUTPUT_DIR,
  CHAIN_ID, KEY_NAME, KEYRING_BACKEND, MIN_GAS_PRICES, FEES, GAS, RPC_ADDRESS,
  OAUTH2_ADDR, HTTP_REGISTRY_ADDR, POST_START_WAIT, POST_TX_WAIT, NODE_HOME, LOG_LEVEL,
  SIMPLE_STORAGE_WASM_FILE, SIMPLE_STORAGE_SCHEMA_FILE,
  SUPABASE_JWT, STORAGE_SUPABASE_KEY_ENDPOINT, NOMEN_FILES,
  EXPLORER_FILES,
  RECREATE_CONFIG, MYTHOSD_BIN
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --scripts-folder) SCRIPTS_FOLDER="$2"; shift 2 ;;
    --wasmx-root) WASMX_ROOT="$2"; shift 2 ;;
    --output-dir) OUTPUT_DIR="$2"; shift 2 ;;
    --home) NODE_HOME="$2"; shift 2 ;;
    --log_level) LOG_LEVEL="$2"; shift 2 ;;
    --log_level=*) LOG_LEVEL="${1#*=}"; shift ;;
    --chain-id) CHAIN_ID="$2"; shift 2 ;;
    --min-gas-prices) MIN_GAS_PRICES="$2"; shift 2 ;;
    --fees) FEES="$2"; shift 2 ;;
    --gas) GAS="$2"; shift 2 ;;
    --rpc) RPC_ADDRESS="$2"; shift 2 ;;
    --rpc-status-url) RPC_STATUS_URL="$2"; shift 2 ;;
    --post-start-wait) POST_START_WAIT="$2"; shift 2 ;;
    --post-tx-wait) POST_TX_WAIT="$2"; shift 2 ;;
    --oauth2-addr) OAUTH2_ADDR="$2"; shift 2 ;;
    --http-registry-addr) HTTP_REGISTRY_ADDR="$2"; shift 2 ;;
    --simple-storage-wasm-file) SIMPLE_STORAGE_WASM_FILE="$2"; shift 2 ;;
    --simple-storage-schema-file) SIMPLE_STORAGE_SCHEMA_FILE="$2"; shift 2 ;;
    --recreate-config) RECREATE_CONFIG=true; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage; exit 1 ;;
  esac
done

if [[ -z "${RPC_STATUS_URL:-}" || "$RPC_STATUS_URL" == "$RPC_STATUS_URL_DEFAULT" ]]; then
  RPC_STATUS_URL="${RPC_ADDRESS/tcp/http}"
fi

export SCRIPTS_FOLDER
export SUPABASE_JWT
export STORAGE_SUPABASE_KEY_ENDPOINT
export NOMEN_FILES
export EXPLORER_FILES

log() {
  echo "[$(date +"%Y-%m-%dT%H:%M:%S%z")] $*"
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || { echo "Missing dependency: $1" >&2; exit 1; }
}

extract_attr() {
  local json="$1" type="$2" key="$3"
  echo "$json" | jq -r --arg type "$type" --arg key "$key" '
    ([.logs[].events[] | select(.type==$type) | .attributes[] | select(.key==$key) | .value][0] // "")
  '
}

wait_for_rpc() {
  local pid="$1" retries=60
  for ((i=0; i<retries; i++)); do
    if ! kill -0 "$pid" 2>/dev/null; then
      echo "mythosd exited unexpectedly, see $OUTPUT_DIR/mythosd.log" >&2
      exit 1
    fi
    if curl -sf "$RPC_STATUS_URL/status" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done
  echo "Timed out waiting for RPC at $RPC_ADDRESS" >&2
  exit 1
}

post_start_wait() {
  if [[ "$POST_START_WAIT" -gt 0 ]]; then
    log "RPC is up; waiting extra ${POST_START_WAIT}s for first block"
    sleep "$POST_START_WAIT"
  fi
}

query_tx_attr() {
  local txhash="$1" type="$2" key="$3" retries="${4:-30}" sleep_s="${5:-1}"
  local txjson value
  for ((i=0; i<retries; i++)); do
    txjson="$("$MYTHOSD_BIN" query tx "$txhash" --node "$RPC_ADDRESS" --output json 2>/dev/null || true)"
    value="$(echo "$txjson" | jq -r --arg type "$type" --arg key "$key" '
      (
        .tx_response?.logs[]?.events[]?,
        .logs[]?.events[]?,
        .tx_response?.events[]?,
        .events[]?
      )
      | select(.type==$type)
      | .attributes[]?
      | select(.key==$key)
      | .value
      | select(. != null)
      | tostring
      ' | head -n1)"
    if [[ -n "$value" && "$value" != "null" ]]; then
      echo "$value"
      return 0
    fi
    sleep "$sleep_s"
  done
  return 1
}

print_tx_receipt() {
  local txhash="$1"
  if [[ -z "$txhash" || "$txhash" == "null" ]]; then
    return
  fi
  log "Tx receipt for $txhash:"
  "$MYTHOSD_BIN" query tx "$txhash" --node "$RPC_ADDRESS" --output json || true
}

# Decode hex data field from tx receipt and extract JSON response
extract_tx_data_json() {
  local txhash="$1"
  local receipt data_hex decoded
  receipt="$("$MYTHOSD_BIN" query tx "$txhash" --node "$RPC_ADDRESS" --output json 2>/dev/null || true)"
  if [[ -z "$receipt" ]]; then
    return
  fi
  data_hex="$(echo "$receipt" | jq -r '.data // empty')"
  if [[ -z "$data_hex" ]]; then
    return
  fi
  # Decode hex to binary and extract printable ASCII (the JSON part)
  decoded="$(echo "$data_hex" | xxd -r -p 2>/dev/null | strings | grep -o '{.*}' | head -1 || true)"
  echo "$decoded"
}

extract_contract_created_from_receipt() {
  local txhash="$1"
  local receipt
  receipt="$("$MYTHOSD_BIN" query tx "$txhash" --node "$RPC_ADDRESS" --output json 2>/dev/null || true)"
  if [[ -z "$receipt" ]]; then
    return
  fi
  # naive scan: find the key and take the value that follows
  # pattern: "key":"contract_address_created","value":"<addr>"
  local marker="\"key\":\"contract_address_created\",\"value\":\""
  local rest
  rest="${receipt#*$marker}"
  if [[ "$rest" == "$receipt" ]]; then
    return
  fi
  echo "${rest:0:45}"
}

post_tx_wait() {
  if [[ "$POST_TX_WAIT" -gt 0 ]]; then
    sleep "$POST_TX_WAIT"
  fi
}


MYTHOSD_BIN="${MYTHOSD_BIN:-}"
if [[ -z "$MYTHOSD_BIN" ]]; then
  if [[ -x "$WASMX_ROOT/mythos-wazero/build/mythosd" ]]; then
    MYTHOSD_BIN="$WASMX_ROOT/mythos-wazero/build/mythosd"
  elif command -v mythosd >/dev/null 2>&1; then
    MYTHOSD_BIN="$(command -v mythosd)"
  else
    echo "mythosd binary not found; build it under $WASMX_ROOT/mythos-wazero/build or set MYTHOSD_BIN." >&2
    exit 1
  fi
fi

require_cmd jq
require_cmd curl
require_cmd "$MYTHOSD_BIN"

# Clear OUTPUT_DIR to remove any cached state from previous runs
if [[ -d "$OUTPUT_DIR" ]]; then
  log "Removing existing output directory $OUTPUT_DIR to clear cached state"
  rm -rf "$OUTPUT_DIR"
fi

mkdir -p "$OUTPUT_DIR"
mkdir -p "$SCRIPTS_FOLDER"

log "Using mythosd at $MYTHOSD_BIN"
log "Chain home: $NODE_HOME"
log "RPC: $RPC_ADDRESS"

if [[ "$RECREATE_CONFIG" == "true" && -d "$NODE_HOME" ]]; then
  log "Removing existing node directory $NODE_HOME"
  rm -rf "$NODE_HOME"
fi

if [[ ! -f "$NODE_HOME/config/genesis.json" ]]; then
  log "Initializing new single-validator testnet under $OUTPUT_DIR"
  "$MYTHOSD_BIN" testnet init-files \
    --network.initial-chains=ondemand_single \
    --v 1 \
    --output-dir="$OUTPUT_DIR" \
    --minimum-gas-prices="$MIN_GAS_PRICES" \
    --nocors \
    --libp2p \
    --enable-eid=false \
    --chain-id="$CHAIN_ID"
fi

log "Resetting state at $NODE_HOME"
"$MYTHOSD_BIN" tendermint unsafe-reset-all --home="$NODE_HOME"

log "Starting mythosd (logs -> $OUTPUT_DIR/mythosd.log)"
"$MYTHOSD_BIN" start --home="$NODE_HOME" --log_level="$LOG_LEVEL" --log_no_color >"$OUTPUT_DIR/mythosd.log" 2>&1 &
NODE_PID=$!
log "mythosd pid: $NODE_PID"
wait_for_rpc "$NODE_PID"
post_start_wait

tx_flags=(
  --chain-id="$CHAIN_ID"
  --from="$KEY_NAME"
  --keyring-backend="$KEYRING_BACKEND"
  --home="$NODE_HOME"
  --fees="$FEES"
  --gas="$GAS"
  --node "$RPC_ADDRESS"
  --yes
  --output json
)

SENDER_ADDR="${SENDER_ADDR:-$("$MYTHOSD_BIN" keys show "$KEY_NAME" --home "$NODE_HOME" --keyring-backend "$KEYRING_BACKEND" --chain-id="$CHAIN_ID" --address)}"
log "Using sender address: $SENDER_ADDR"

# Export the funder private key for oauth2-keys contract
log "Exporting funder private key for oauth2-keys contract"

# Export with --unsafe --unarmored-hex flags (will prompt for confirmation)
# Pipe "y" to auto-confirm the warning prompt
EXPORT_OUTPUT=$(echo "y" | "$MYTHOSD_BIN" keys export "$KEY_NAME" --keyring-backend "$KEYRING_BACKEND" --home "$NODE_HOME" --unsafe --unarmored-hex 2>&1)

# Extract the 64-character hex private key from the output
FUNDER_PRIVKEY=$(echo "$EXPORT_OUTPUT" | grep -oE '[0-9a-fA-F]{64}' | head -1)

if [[ -z "$FUNDER_PRIVKEY" ]]; then
  echo "Failed to export funder private key" >&2
  echo "Export output was:" >&2
  echo "$EXPORT_OUTPUT" >&2
  exit 1
fi
log "Funder private key exported (length: ${#FUNDER_PRIVKEY} chars)"

# Get the funder address from the same key we exported
log "Getting funder address from key: $KEY_NAME"
FUNDER_ADDR=$("$MYTHOSD_BIN" keys show "$KEY_NAME" --keyring-backend "$KEYRING_BACKEND" --home "$NODE_HOME" --address)
log "Funder address: $FUNDER_ADDR"

# Fund the funder account with a large amount for initializing new user accounts
# Only fund if it's different from the sender (to avoid self-transfer)
if [[ "$FUNDER_ADDR" != "$SENDER_ADDR" ]]; then
  log "Funding funder account with initial balance: 120000000000000000000amyt"
  fund_out="$("$MYTHOSD_BIN" tx cosmosmod bank send "$KEY_NAME" "$FUNDER_ADDR" "120000000000000000000amyt" "${tx_flags[@]}")"
  post_tx_wait
  if [[ "$(echo "$fund_out" | jq -r '.code // 0')" != "0" ]]; then
    echo "Warning: Funder account funding failed" >&2
    echo "$fund_out"
    # Continue anyway
  else
    log "Funder account funded successfully"
  fi
else
  log "Funder address same as sender, skipping funding (validator already has funds)"
fi

log "Storing simple_storage contract wasm"
if [[ ! -f "$SIMPLE_STORAGE_SCHEMA_FILE" ]]; then
  echo "Simple storage schema file not found at $SIMPLE_STORAGE_SCHEMA_FILE" >&2
  exit 1
fi
simple_store_out="$("$MYTHOSD_BIN" tx wasmx store "$SIMPLE_STORAGE_WASM_FILE" --json-schema "$SIMPLE_STORAGE_SCHEMA_FILE" "${tx_flags[@]}")"
simple_store_txhash="$(echo "$simple_store_out" | jq -r '.txhash // empty')"
if [[ -n "$simple_store_txhash" ]]; then
  log "Simple storage store tx hash: $simple_store_txhash"
  post_tx_wait
  print_tx_receipt "$simple_store_txhash"
fi
SIMPLE_STORAGE_CODE_ID="$(extract_attr "$simple_store_out" "store_code" "code_id")"
if [[ -z "$SIMPLE_STORAGE_CODE_ID" || "$SIMPLE_STORAGE_CODE_ID" == "null" ]]; then
  if [[ -n "$simple_store_txhash" ]]; then
    log "Simple storage store response missing code_id; querying tx $simple_store_txhash for code_id"
    SIMPLE_STORAGE_CODE_ID="$(query_tx_attr "$simple_store_txhash" "store_code" "code_id" 60 1 || true)"
  fi
fi
if [[ -z "$SIMPLE_STORAGE_CODE_ID" || "$SIMPLE_STORAGE_CODE_ID" == "null" ]]; then
  echo "Could not determine simple storage code id" >&2
  echo "$simple_store_out" >&2
  exit 1
fi
log "Stored simple_storage contract with code id $SIMPLE_STORAGE_CODE_ID"

log "Instantiating simple_storage contract"
simple_storage_init_payload="{}"
simple_storage_instantiate_out="$("$MYTHOSD_BIN" tx wasmx instantiate "$SIMPLE_STORAGE_CODE_ID" "$simple_storage_init_payload" \
  --label "simple_storage" "${tx_flags[@]}")"
simple_storage_instantiate_txhash="$(echo "$simple_storage_instantiate_out" | jq -r '.txhash // empty')"
if [[ -n "$simple_storage_instantiate_txhash" ]]; then
  log "Simple storage instantiate tx hash: $simple_storage_instantiate_txhash"
  post_tx_wait
  print_tx_receipt "$simple_storage_instantiate_txhash"
fi
SIMPLE_STORAGE_ADDR=""
if [[ -n "$simple_storage_instantiate_txhash" ]]; then
  SIMPLE_STORAGE_ADDR="$(extract_contract_created_from_receipt "$simple_storage_instantiate_txhash")"
fi
if [[ -z "$SIMPLE_STORAGE_ADDR" ]]; then
  echo "Could not determine simple storage contract address" >&2
  echo "$simple_storage_instantiate_out"
  exit 1
fi
log "Simple storage contract instantiated at $SIMPLE_STORAGE_ADDR"

log "Configuring OAuth2 Keys contract with funder private key via $OAUTH2_KEYS_ADDR"
oauth2keys_init_payload="$(jq -n \
  --arg privkey "$FUNDER_PRIVKEY" \
  '{
    init_genesis: {
      funder_priv_key: $privkey,
      init_account_amt: {"amount":"1000000000000000000","denom":"amyt"},
      gas_price: {"amount":"100","denom":"amyt"},
      route_prefix: "/auth"
    }
  }')"
oauth2keys_init_out="$("$MYTHOSD_BIN" tx wasmx execute "$OAUTH2_KEYS_ADDR" "$oauth2keys_init_payload" "${tx_flags[@]}")"
post_tx_wait
if [[ "$(echo "$oauth2keys_init_out" | jq -r '.code // 0')" != "0" ]]; then
  echo "OAuth2 Keys initialization failed" >&2
  echo "$oauth2keys_init_out"
  exit 1
fi
log "OAuth2 Keys contract configured"

log "Initializing OAuth2 server via $OAUTH2_ADDR"
oauth2_init_payload="$(jq -n \
  --arg supabase_jwt "$SUPABASE_JWT" \
  --arg nomen_files "$NOMEN_FILES" \
  --arg explorer_files "$EXPLORER_FILES" \
  '{
    init_genesis: (
      { supabase_jwt: $supabase_jwt } +
      ([
        (if $nomen_files == "" then empty else { route: "/nomen", folder_path: $nomen_files } end),
        (if $explorer_files == "" then empty else { route: "/explorer", folder_path: $explorer_files } end)
      ] | if length > 0 then { static_routes: . } else {} end)
    )
  }')"
oauth2_init_out="$("$MYTHOSD_BIN" tx wasmx execute "$OAUTH2_ADDR" "$oauth2_init_payload" "${tx_flags[@]}")"
post_tx_wait
if [[ "$(echo "$oauth2_init_out" | jq -r '.code // 0')" != "0" ]]; then
  echo "OAuth2 server init failed" >&2
  echo "$oauth2_init_out"
  exit 1
fi
log "OAuth2 server initialized"

log "Registering OAuth client via $OAUTH2_ADDR"
oauth_payload="$(jq -n '{
  register_oauth_client: {
    name: "MCP Search Client",
    description: "OAuth client for MCP search",
    redirect_uris: ["https://chat.openai.com/aip/*", "https://chatgpt.com/connector_platform_oauth_redirect","http://localhost:3000/callback","http://localhost:8080/callback","http://voorwhwgymjpvrgzskwp.supabase.co/functions/v1/wasmx-oauth2-task/callback","https://voorwhwgymjpvrgzskwp.supabase.co/functions/v1/wasmx-oauth2-task/callback","http://voorwhwgymjpvrgzskwp.supabase.co/*","https://voorwhwgymjpvrgzskwp.supabase.co/*"],
    scopes: ["read", "tools"]
  }
}')"
oauth_out="$("$MYTHOSD_BIN" tx wasmx execute "$OAUTH2_ADDR" "$oauth_payload" "${tx_flags[@]}")"
oauth_txhash="$(echo "$oauth_out" | jq -r '.txhash // empty')"
if [[ -n "$oauth_txhash" ]]; then
  log "OAuth client registration tx hash: $oauth_txhash"
  post_tx_wait
  print_tx_receipt "$oauth_txhash"
fi
if [[ "$(echo "$oauth_out" | jq -r '.code // 0')" != "0" ]]; then
  echo "OAuth client registration failed" >&2
  echo "$oauth_out"
  exit 1
fi

# Extract and display OAuth client credentials
OAUTH_CLIENT_JSON=""
if [[ -n "$oauth_txhash" ]]; then
  OAUTH_CLIENT_JSON="$(extract_tx_data_json "$oauth_txhash")"
fi
if [[ -n "$OAUTH_CLIENT_JSON" ]]; then
  OAUTH_CLIENT_ID="$(echo "$OAUTH_CLIENT_JSON" | jq -r '.client_id // empty')"
  OAUTH_CLIENT_SECRET="$(echo "$OAUTH_CLIENT_JSON" | jq -r '.client_secret // empty')"
  log "OAuth client registered successfully:"
  log "  client_id:     $OAUTH_CLIENT_ID"
  log "  client_secret: $OAUTH_CLIENT_SECRET"
else
  log "OAuth client registered; could not extract credentials from tx response."
fi

log "Registering test user via $OAUTH2_ADDR"
user_payload="$(jq -n '{
  register_user: {
    email: "test@mail.provable.dev",
    password: "123456789",
    username: "test"
  }
}')"
user_out="$("$MYTHOSD_BIN" tx wasmx execute "$OAUTH2_ADDR" "$user_payload" "${tx_flags[@]}")"
post_tx_wait
if [[ "$(echo "$user_out" | jq -r '.code // 0')" != "0" ]]; then
  echo "User registration failed" >&2
  echo "$user_out"
  exit 1
fi

log "Initializing HTTP registry with gas price via $HTTP_REGISTRY_ADDR"
http_init_payload="$(jq -n '{
  init_genesis: {
    gas_price: {"amount": "100", "denom": "amyt"}
  }
}')"
http_init_out="$("$MYTHOSD_BIN" tx wasmx execute "$HTTP_REGISTRY_ADDR" "$http_init_payload" "${tx_flags[@]}")"
post_tx_wait
if [[ "$(echo "$http_init_out" | jq -r '.code // 0')" != "0" ]]; then
  echo "HTTP registry init failed" >&2
  echo "$http_init_out"
  exit 1
fi
log "HTTP registry initialized with gas price"

log "Starting HTTP server via $HTTP_REGISTRY_ADDR"
http_payload="$(jq -n '{
  start_web_server: {
    config: {
      enable_oauth: true,
      address: "0.0.0.0:8080",
      cors_allowed_origins: ["*"],
      cors_allowed_methods: [],
      cors_allowed_headers: [],
      max_open_connections: 1000,
      request_body_max_size: 1000000000
    }
  }
}')"
http_out="$("$MYTHOSD_BIN" tx wasmx execute "$HTTP_REGISTRY_ADDR" "$http_payload" "${tx_flags[@]}")"
post_tx_wait
if [[ "$(echo "$http_out" | jq -r '.code // 0')" != "0" ]]; then
  echo "HTTP server start failed" >&2
  echo "$http_out"
  exit 1
fi

cat <<EOF
Chain ready.
  RPC:                 $RPC_ADDRESS
  mythosd pid:         $NODE_PID (logs: $OUTPUT_DIR/mythosd.log)

  OAuth2 addr:         $OAUTH2_ADDR
  OAuth2 client_id:    ${OAUTH_CLIENT_ID:-<not extracted>}
  OAuth2 client_secret: ${OAUTH_CLIENT_SECRET:-<not extracted>}
  HTTP registry:       $HTTP_REGISTRY_ADDR

EOF
RPC_STATUS_URL="${RPC_STATUS_URL:-$RPC_STATUS_URL_DEFAULT}"
