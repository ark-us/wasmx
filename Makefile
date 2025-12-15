#!/usr/bin/make -f

TINYGO_DIR := ./tests/testdata/tinygo
PRECOMPILE_DIR := ./wasmx/x/wasmx/vm/precompiles

# Mapping of tinygo modules to precompile wasm filenames
# Format: module_path:output_filename
TINYGO_TARGETS := \
	wasmx-erc20:32.erc20json_go_0.0.2.wasm \
	wasmx-fsm:28.finite_state_machine.wasm \
	wasmx-raft-lib:2a.raft_library.wasm \
	wasmx-raftp2p-lib:36.raftp2p_library.wasm \
	wasmx-ondemand-single-lib:65.wasmx_ondemand_single_library.wasm \
	wasmx-kayrosp2p-lib:71.kayrosp2p_library.wasm \
	wasmx-kayrosp2p-ondemand-lib:72.kayrosp2p_ondemand_library.wasm \
	wasmx-kayros-verifier:75.kayros_verifier_0.0.1.wasm \
	emailchain:63.wasmx_email_0.0.1.wasm \
	wasmx-oauth2-server:66.wasmx_oauth2_server_0.0.1.wasm \
	wasmx-mcp-registry:67.wasmx_mcp_registry_0.0.1.wasm \
	wasmx-httpserver-registry:68.wasmx_httpserver_registry_0.0.2.wasm \
	wasmx-identity:69.wasmx_identity_0.0.1.wasm \
	wasmx-oauth2-keys:70.wasmx_oauth2_keys_0.0.1.wasm \
	wasmx-erc20x:71.erc20xjson_go_0.0.1.wasm

# wasmx-gov:35.gov_0.0.1.wasm \
# wasmx-gov-continuous:37.gov_cont_0.0.1.wasm \
# wasmx-multichain-registry:4a.multichain_registry_0.0.1.wasm \
# wasmx-staking:30.staking_0.0.1.wasm \
# wasmx-bank:31.bank_0.0.1.wasm \
# wasmx-auth:38.auth_0.0.1.wasm \
# wasmx-slashing:45.slashing_0.0.1.wasm \
# wasmx-distribution:46.distribution_0.0.1.wasm \
# wasmx-lobby:4d.lobby_json_0.0.1.wasm \

.PHONY: tinygo-tidy tinygo


tinygo-tidy:
	@echo "Tidying TinyGo modules..."
	@cd $(TINYGO_DIR) && find . -name go.mod -execdir env GOWORK=off go mod tidy \;

#
# tinygo: Build all TinyGo modules when invoked alone.
# If invoked with additional goals (e.g., `make tinygo wasmx-foo`),
# this target becomes a no-op and the module-specific target handles the build.
tinygo:
	@set -e; \
	if [ -z "$(filter-out tinygo,$(MAKECMDGOALS))" ]; then \
		echo "Tidying TinyGo modules..."; \
		cd $(TINYGO_DIR) && find . -name go.mod -execdir env GOWORK=off go mod tidy \; ; \
		echo "Building TinyGo precompiles..."; \
		cd $(TINYGO_DIR); \
		for pair in $(TINYGO_TARGETS); do \
			mod="$${pair%%:*}"; \
			out="$${pair##*:}"; \
			if [ -f "$$mod/cmd/main.go" ]; then \
				echo "-> $$mod -> $(PRECOMPILE_DIR)/$$out"; \
				cd "$$mod"; \
				env GOWORK=off tinygo build -o "$(abspath $(PRECOMPILE_DIR))/$$out" -no-debug -scheduler=none -gc=leaking -target=wasi ./cmd; \
				cd - >/dev/null; \
			else \
				echo "skipping $$mod (no cmd/main.go)"; \
			fi; \
		done; \
	else \
		echo "tinygo: delegating to module target(s): $(filter-out tinygo,$(MAKECMDGOALS))"; \
	fi

# Derive module list from TINYGO_TARGETS
TINYGO_MODULES := $(foreach pair,$(TINYGO_TARGETS),$(firstword $(subst :, ,$(pair))))

# Helper to resolve output filename for a module from TINYGO_TARGETS
GET_TINYGO_OUT = $(word 2,$(subst :, ,$(filter $(1):%,$(TINYGO_TARGETS))))

.PHONY: $(TINYGO_MODULES)

# Build a specific TinyGo module only
# Usage: make tinygo wasmx-multichain-registry
$(TINYGO_MODULES):
	@set -e; \
	mod="$@"; \
	out='$(call GET_TINYGO_OUT,$@)'; \
	if [ -z "$$out" ]; then echo "Unknown TinyGo module: $$mod"; exit 1; fi; \
	if [ ! -f "$(TINYGO_DIR)/$$mod/cmd/main.go" ]; then echo "No cmd/main.go in $$mod"; exit 1; fi; \
	echo "Tidying $$mod..."; \
	(cd "$(TINYGO_DIR)/$$mod" && env GOWORK=off go mod tidy); \
	echo "Building $$mod -> $(PRECOMPILE_DIR)/$$out"; \
	(cd "$(TINYGO_DIR)/$$mod" && env GOWORK=off tinygo build -o "$(abspath $(PRECOMPILE_DIR))/$$out" -no-debug -scheduler=none -gc=leaking -target=wasi ./cmd); \
	echo "Built $(PRECOMPILE_DIR)/$$out"
