#!/usr/bin/make -f

TINYGO_DIR := ./tests/testdata/tinygo
PRECOMPILE_DIR := ./wasmx/x/wasmx/vm/precompiles

# Mapping of tinygo modules to precompile wasm filenames
# Format: module_path:output_filename
TINYGO_TARGETS := \
	wasmx-fsm:28.finite_state_machine.wasm

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
	(cd "$(TINYGO_DIR)/$$mod" && env GOWORK=off go mod tidy) 2>/dev/null; \
	echo "Building $$mod -> $(PRECOMPILE_DIR)/$$out"; \
	(cd "$(TINYGO_DIR)/$$mod" && env GOWORK=off tinygo build -o "$(abspath $(PRECOMPILE_DIR))/$$out" -no-debug -scheduler=none -gc=leaking -target=wasi ./cmd) 2>/dev/null; \
	echo "Built $(PRECOMPILE_DIR)/$$out"
