#!/bin/bash

# script gov,bank,tendermintp2p --wasmxdir --contractsdir

# Get the current working directory's parent
base_dir=$(cd "$(dirname "$0")/../../.." && pwd)

# Default directories
WASMX_PROJECT_ROOT="${base_dir}/wasmx"
CONTRACTS_PROJECT_ROOT="${base_dir}/wasmx-as-contracts"

# Read the first argument as labels
TO_COMPILE=$1
shift # Remove the first argument so we can parse the flags

# Parse flags
while [[ $# -gt 0 ]]; do
  case "$1" in
    --wasmxdir=*)
      WASMX_PROJECT_ROOT="${1#*=}"
      shift
      ;;
    --contractsdir=*)
      CONTRACTS_PROJECT_ROOT="${1#*=}"
      shift
      ;;
    --wasmxdir)
      WASMX_PROJECT_ROOT="$2"
      shift 2
      ;;
    --contractsdir)
      CONTRACTS_PROJECT_ROOT="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

export WASMX_GO_PRECOMPILES="${WASMX_PROJECT_ROOT}/wasmx/x/wasmx/vm/precompiles"
export WASMX_GO_TESTDATA="${WASMX_PROJECT_ROOT}/tests/testdata/wasmx"
export WASMX_GO_TESTDATA_NETWORK="${WASMX_PROJECT_ROOT}/tests/network/testdata/wasmx"
export WASMX_GO_TESTDATA_SQL="${WASMX_PROJECT_ROOT}/tests/vmsql/testdata/as"
export WASMX_GO_TESTDATA_KVDB="${WASMX_PROJECT_ROOT}/tests/vmkv/testdata/as"
export WASMX_GO_TESTDATA_IMAP="${WASMX_PROJECT_ROOT}/tests/vmemail/testdata/as"
export WASMX_GO_TESTDATA_SMTP="${WASMX_GO_TESTDATA_IMAP}"
export WASMX_GO_TESTDATA_EMAIL="${WASMX_GO_TESTDATA_IMAP}"

# precompiles
export WASMX_BLOCKS="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-blocks"
export WASMX_RAFT="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-raft"
export WASMX_RAFT_P2P="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-raft-p2p"
export WASMX_TENDERMINT="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-tendermint"
export WASMX_TENDERMINT_P2P="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-tendermint-p2p"
export WASMX_AVA_SNOWMAN="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-ava-snowman"
export WASMX_FSM="${CONTRACTS_PROJECT_ROOT}/packages/xstate-fsm-as"
export WASMX_STAKING="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-stake"
export WASMX_BANK="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-bank"
export WASMX_ERC20="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-erc20"
export WASMX_DERC20="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-derc20"
export WASMX_ERC20_ROLLUP="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-erc20-rollup"
export WASMX_GOV="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-gov"
export WASMX_GOV_CONTINUOUS="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-gov-continuous"
export WASMX_HOOKS="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-hooks"
export WASMX_AUTH="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-auth"
export WASMX_ROLES="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-roles"
export WASMX_SLASHING="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-slashing"
export WASMX_DISTRIBUTION="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-distribution"
export WASMX_CHAT="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-chat"
export WASMX_CHAT_VERIFIER="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-chat-verifier"
export WASMX_TIME="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-time"
export WASMX_LEVEL0="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-level0"
export WASMX_LEVEL0_ONDEMAND="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-level0-ondemand"
export WASMX_MULTICHAIN="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-multichain-registry"
export WASMX_MULTICHAIN_LOCAL="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-multichain-registry-local"
export WASMX_LOBBY="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-lobby"
export WASMX_METAREGISTRY="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-metaregistry"
export WASMX_PARAMS="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-params"
export WASMX_CODES_REGISTRY="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-codes-registry"
export WASMX_DTYPE="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-dtype"
export WASMX_EMAIL="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-email-prover"

# tests
export WASMX_TESTS_CROSSCHAIN="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-test-crosschain"
export WASMX_TESTS_SIMPLESTORAGE="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-test-simplestorage"
export WASMX_TESTS_SQL="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-test-sql"
export WASMX_TESTS_KVDB="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-test-kvdb"
export WASMX_ERC20_DTYPE="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-erc20-sql"
export WASMX_TESTS_IMAP="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-test-imap"
export WASMX_TESTS_SMTP="${CONTRACTS_PROJECT_ROOT}/packages/wasmx-test-smtp"

labels=$(echo $TO_COMPILE | tr "," "\n")
for label in $labels
do
    echo $label
    if [[ $label = 'blocks' ]]; then
        echo "building blocks"
        cd $WASMX_BLOCKS && npm run asbuild
        mv -f $WASMX_BLOCKS/build/release.wasm $WASMX_GO_PRECOMPILES/29.storage_chain.wasm
    fi
    if [[ $label = 'raft' ]]; then
        echo "building raft"
        cd $WASMX_RAFT && npm run asbuild
        mv -f $WASMX_RAFT/build/release.wasm $WASMX_GO_PRECOMPILES/2a.raft_library.wasm
    fi
    if [[ $label = 'raftp2p' ]]; then
        echo "building raftp2p"
        cd $WASMX_RAFT_P2P && npm run asbuild
        mv -f $WASMX_RAFT_P2P/build/release.wasm $WASMX_GO_PRECOMPILES/36.raftp2p_library.wasm
    fi
    if [[ $label = 'tendermint' ]]; then
        echo "building tendermint"
        cd $WASMX_TENDERMINT && npm run asbuild
        mv -f $WASMX_TENDERMINT/build/release.wasm $WASMX_GO_PRECOMPILES/2b.tendermint_library.wasm
    fi
    if [[ $label = 'tendermintp2p' ]]; then
        echo "building tendermint-p2p"
        cd $WASMX_TENDERMINT_P2P && npm run asbuild
        mv -f $WASMX_TENDERMINT_P2P/build/release.wasm $WASMX_GO_PRECOMPILES/40.tendermintp2p_library.wasm
    fi
    if [[ $label = 'ava' ]]; then
        echo "building ava"
        cd $WASMX_AVA_SNOWMAN && npm run asbuild
        mv -f $WASMX_AVA_SNOWMAN/build/release.wasm $WASMX_GO_PRECOMPILES/2e.ava_snowman_library.wasm
    fi
    if [[ $label = 'fsm' ]]; then
        echo "building fsm"
        cd $WASMX_FSM && npm run asbuild
        mv -f $WASMX_FSM/build/release.wasm $WASMX_GO_PRECOMPILES/28.finite_state_machine.wasm
    fi
    if [[ $label = 'staking' ]]; then
        echo "building staking"
        cd $WASMX_STAKING && npm run asbuild
        mv -f $WASMX_STAKING/build/release.wasm $WASMX_GO_PRECOMPILES/30.staking_0.0.1.wasm
    fi
    if [[ $label = 'bank' ]]; then
        echo "building bank"
        cd $WASMX_BANK && npm run asbuild
        mv -f $WASMX_BANK/build/release.wasm $WASMX_GO_PRECOMPILES/31.bank_0.0.1.wasm
    fi
    if [[ $label = 'erc20' ]]; then
        echo "building erc20"
        cd $WASMX_ERC20 && npm run asbuild
        mv -f $WASMX_ERC20/build/release.wasm $WASMX_GO_PRECOMPILES/32.erc20json_0.0.1.wasm
    fi
    if [[ $label = 'derc20' ]]; then
        echo "building derc20"
        cd $WASMX_DERC20 && npm run asbuild
        mv -f $WASMX_DERC20/build/release.wasm $WASMX_GO_PRECOMPILES/33.derc20json_0.0.1.wasm
    fi
    if [[ $label = 'erc20-rollup' ]]; then
        echo "building erc20-rollup"
        cd $WASMX_ERC20_ROLLUP && npm run asbuild
        mv -f $WASMX_ERC20_ROLLUP/build/release.wasm $WASMX_GO_PRECOMPILES/4c.erc20rollupjson_0.0.1.wasm
    fi
    if [[ $label = 'gov' ]]; then
        echo "building gov"
        cd $WASMX_GOV && npm run asbuild
        mv -f $WASMX_GOV/build/release.wasm $WASMX_GO_PRECOMPILES/35.gov_0.0.1.wasm
    fi
    if [[ $label = 'hooks' ]]; then
        echo "building hooks"
        cd $WASMX_HOOKS && npm run asbuild
        mv -f $WASMX_HOOKS/build/release.wasm $WASMX_GO_PRECOMPILES/34.hooks_0.0.1.wasm
    fi
    if [[ $label = 'gov2' ]]; then
        echo "building gov2"
        cd $WASMX_GOV_CONTINUOUS && npm run asbuild
        mv -f $WASMX_GOV_CONTINUOUS/build/release.wasm $WASMX_GO_PRECOMPILES/37.gov_cont_0.0.1.wasm
    fi
    if [[ $label = 'auth' ]]; then
        echo "building auth"
        cd $WASMX_AUTH && npm run asbuild
        mv -f $WASMX_AUTH/build/release.wasm $WASMX_GO_PRECOMPILES/38.auth_0.0.1.wasm
    fi
    if [[ $label = 'roles' ]]; then
        echo "building roles"
        cd $WASMX_ROLES && npm run asbuild
        mv -f $WASMX_ROLES/build/release.wasm $WASMX_GO_PRECOMPILES/60.roles_0.0.1.wasm
    fi
    if [[ $label = 'slashing' ]]; then
        echo "building slashing"
        cd $WASMX_SLASHING && npm run asbuild
        mv -f $WASMX_SLASHING/build/release.wasm $WASMX_GO_PRECOMPILES/45.slashing_0.0.1.wasm
    fi
    if [[ $label = 'distribution' ]]; then
        echo "building distribution"
        cd $WASMX_DISTRIBUTION && npm run asbuild
        mv -f $WASMX_DISTRIBUTION/build/release.wasm $WASMX_GO_PRECOMPILES/46.distribution_0.0.1.wasm
    fi
    if [[ $label = 'chat' ]]; then
        echo "building chat"
        cd $WASMX_CHAT && npm run asbuild
        mv -f $WASMX_CHAT/build/release.wasm $WASMX_GO_PRECOMPILES/42.chat_0.0.1.wasm
    fi
    if [[ $label = 'chatverif' ]]; then
        echo "building chat verifier"
        cd $WASMX_CHAT_VERIFIER && npm run asbuild
        mv -f $WASMX_CHAT_VERIFIER/build/release.wasm $WASMX_GO_PRECOMPILES/44.chat_verifier_0.0.1.wasm
    fi
    if [[ $label = 'time' ]]; then
        echo "building time"
        cd $WASMX_TIME && npm run asbuild
        mv -f $WASMX_TIME/build/release.wasm $WASMX_GO_PRECOMPILES/47.time_0.0.1.wasm
    fi
    if [[ $label = 'level0' ]]; then
        echo "building level0"
        cd $WASMX_LEVEL0 && npm run asbuild
        mv -f $WASMX_LEVEL0/build/release.wasm $WASMX_GO_PRECOMPILES/48.level0_0.0.1.wasm
    fi
    if [[ $label = 'level0d' ]]; then
        echo "building leveln0 ondemand"
        cd $WASMX_LEVEL0_ONDEMAND && npm run asbuild
        mv -f $WASMX_LEVEL0_ONDEMAND/build/release.wasm $WASMX_GO_PRECOMPILES/51.level0_ondemand_0.0.1.wasm
    fi
    if [[ $label = 'multichain' ]]; then
        echo "building multichain"
        cd $WASMX_MULTICHAIN && npm run asbuild
        mv -f $WASMX_MULTICHAIN/build/release.wasm $WASMX_GO_PRECOMPILES/4a.multichain_registry_0.0.1.wasm
    fi
    if [[ $label = 'multichain_local' ]]; then
        echo "building multichain_local"
        cd $WASMX_MULTICHAIN_LOCAL && npm run asbuild
        mv -f $WASMX_MULTICHAIN_LOCAL/build/release.wasm $WASMX_GO_PRECOMPILES/4b.multichain_registry_local_0.0.1.wasm
    fi
    if [[ $label = 'lobby' ]]; then
        echo "building lobby"
        cd $WASMX_LOBBY && npm run asbuild
        mv -f $WASMX_LOBBY/build/release.wasm $WASMX_GO_PRECOMPILES/4d.lobby_json_0.0.1.wasm
    fi
    if [[ $label = 'metaregistry' ]]; then
        echo "building metaregistry"
        cd $WASMX_METAREGISTRY && npm run asbuild
        mv -f $WASMX_METAREGISTRY/build/release.wasm $WASMX_GO_PRECOMPILES/4f.metaregistry_json_0.0.1.wasm
    fi
    # if [[ $label = 'params' ]]; then
    #     echo "building params"
    #     cd $WASMX_PARAMS && npm run asbuild
    #     mv -f $WASMX_PARAMS/build/release.wasm $WASMX_GO_PRECOMPILES/50.params_json_0.0.1.wasm
    # fi
    if [[ $label = 'codes' ]]; then
        echo "building metaregistry"
        cd $WASMX_CODES_REGISTRY && npm run asbuild
        mv -f $WASMX_CODES_REGISTRY/build/release.wasm $WASMX_GO_PRECOMPILES/61.wasmx_codes_registry_0.0.1.wasm
    fi

    if [[ $label = 'dtype' ]]; then
        echo "building dtype"
        cd $WASMX_DTYPE && npm run asbuild
        mv -f $WASMX_DTYPE/build/release.wasm $WASMX_GO_PRECOMPILES/62.wasmx_dtype_0.0.1.wasm
    fi

    if [[ $label = 'email' ]]; then
        echo "building email"
        cd $WASMX_EMAIL && npm run asbuild
        mv -f $WASMX_EMAIL/build/release.wasm $WASMX_GO_PRECOMPILES/63.wasmx_email_0.0.1.wasm
    fi

    if [[ $label = 'tests' ]]; then
        echo "building tests crosschain"
        cd $WASMX_TESTS_CROSSCHAIN && npm run asbuild
        mv -f $WASMX_TESTS_CROSSCHAIN/build/release.wasm $WASMX_GO_TESTDATA_NETWORK/crosschain.wasm
        echo "building tests simplestorage"
        cd $WASMX_TESTS_SIMPLESTORAGE && npm run asbuild
        mv -f $WASMX_TESTS_SIMPLESTORAGE/build/release.wasm $WASMX_GO_TESTDATA_NETWORK/simple_storage.wasm

        cd $WASMX_TESTS_SQL && npm run asbuild
        mv -f $WASMX_TESTS_SQL/build/release.wasm $WASMX_GO_TESTDATA_SQL/wasmx_test_sql.wasm

        cd $WASMX_ERC20_DTYPE && npm run asbuild
        mv -f $WASMX_ERC20_DTYPE/build/release.wasm $WASMX_GO_TESTDATA_SQL/wasmx_erc20_sql.wasm

        cd $WASMX_TESTS_KVDB && npm run asbuild
        mv -f $WASMX_TESTS_KVDB/build/release.wasm $WASMX_GO_TESTDATA_KVDB/wasmx_test_kvdb.wasm

        cd $WASMX_TESTS_IMAP && npm run asbuild
        mv -f $WASMX_TESTS_IMAP/build/release.wasm $WASMX_GO_TESTDATA_IMAP/wasmx_test_imap.wasm

        cd $WASMX_TESTS_SMTP && npm run asbuild
        mv -f $WASMX_TESTS_SMTP/build/release.wasm $WASMX_GO_TESTDATA_SMTP/wasmx_test_smtp.wasm
    fi
    if [[ $label = 'imap' ]]; then
        cd $WASMX_TESTS_IMAP && npm run asbuild
        mv -f $WASMX_TESTS_IMAP/build/release.wasm $WASMX_GO_TESTDATA_IMAP/wasmx_test_imap.wasm
    fi
    if [[ $label = 'smtp' ]]; then
        cd $WASMX_TESTS_SMTP && npm run asbuild
        mv -f $WASMX_TESTS_SMTP/build/release.wasm $WASMX_GO_TESTDATA_SMTP/wasmx_test_smtp.wasm
    fi
done

if [[ $TO_COMPILE = '' ]]; then
	echo "building all contracts..."
    cd $WASMX_BLOCKS && npm run asbuild
    cd $WASMX_RAFT && npm run asbuild
    cd $WASMX_RAFT_P2P && npm run asbuild
    cd $WASMX_TENDERMINT && npm run asbuild
    cd $WASMX_TENDERMINT_P2P && npm run asbuild
    cd $WASMX_AVA_SNOWMAN && npm run asbuild
    cd $WASMX_FSM && npm run asbuild
    cd $WASMX_STAKING && npm run asbuild
    cd $WASMX_BANK && npm run asbuild
    cd $WASMX_ERC20 && npm run asbuild
    cd $WASMX_DERC20 && npm run asbuild
    cd $WASMX_ERC20_ROLLUP && npm run asbuild
    cd $WASMX_GOV && npm run asbuild
    cd $WASMX_HOOKS && npm run asbuild
    cd $WASMX_GOV_CONTINUOUS && npm run asbuild
    cd $WASMX_AUTH && npm run asbuild
    cd $WASMX_ROLES && npm run asbuild
    cd $WASMX_SLASHING && npm run asbuild
    cd $WASMX_DISTRIBUTION && npm run asbuild
    cd $WASMX_CHAT && npm run asbuild
    cd $WASMX_CHAT_VERIFIER && npm run asbuild
    cd $WASMX_TIME && npm run asbuild
    cd $WASMX_LEVEL0 && npm run asbuild
    cd $WASMX_LEVEL0_ONDEMAND && npm run asbuild
    cd $WASMX_MULTICHAIN && npm run asbuild
    cd $WASMX_MULTICHAIN_LOCAL && npm run asbuild
    cd $WASMX_TESTS_CROSSCHAIN && npm run asbuild
    cd $WASMX_TESTS_SIMPLESTORAGE && npm run asbuild
    cd $WASMX_TESTS_SQL && npm run asbuild
    cd $WASMX_TESTS_KVDB && npm run asbuild
    cd $WASMX_LOBBY && npm run asbuild
    cd $WASMX_METAREGISTRY && npm run asbuild
    cd $WASMX_CODES_REGISTRY && npm run asbuild
    cd $WASMX_DTYPE && npm run asbuild
    cd $WASMX_EMAIL && npm run asbuild
    cd $WASMX_ERC20_DTYPE && npm run asbuild
    cd $WASMX_TESTS_IMAP && npm run asbuild
    cd $WASMX_TESTS_SMTP && npm run asbuild
    # cd $WASMX_PARAMS && npm run asbuild

    mv -f $WASMX_FSM/build/release.wasm $WASMX_GO_PRECOMPILES/28.finite_state_machine.wasm
    mv -f $WASMX_BLOCKS/build/release.wasm $WASMX_GO_PRECOMPILES/29.storage_chain.wasm
    mv -f $WASMX_RAFT/build/release.wasm $WASMX_GO_PRECOMPILES/2a.raft_library.wasm
    mv -f $WASMX_TENDERMINT/build/release.wasm $WASMX_GO_PRECOMPILES/2b.tendermint_library.wasm
    mv -f $WASMX_TENDERMINT_P2P/build/release.wasm $WASMX_GO_PRECOMPILES/40.tendermintp2p_library.wasm
    mv -f $WASMX_AVA_SNOWMAN/build/release.wasm $WASMX_GO_PRECOMPILES/2e.ava_snowman_library.wasm
    mv -f $WASMX_STAKING/build/release.wasm $WASMX_GO_PRECOMPILES/30.staking_0.0.1.wasm
    mv -f $WASMX_BANK/build/release.wasm $WASMX_GO_PRECOMPILES/31.bank_0.0.1.wasm
    mv -f $WASMX_ERC20/build/release.wasm $WASMX_GO_PRECOMPILES/32.erc20json_0.0.1.wasm
    mv -f $WASMX_DERC20/build/release.wasm $WASMX_GO_PRECOMPILES/33.derc20json_0.0.1.wasm
    mv -f $WASMX_HOOKS/build/release.wasm $WASMX_GO_PRECOMPILES/34.hooks_0.0.1.wasm
    mv -f $WASMX_GOV/build/release.wasm $WASMX_GO_PRECOMPILES/35.gov_0.0.1.wasm
    mv -f $WASMX_RAFT_P2P/build/release.wasm $WASMX_GO_PRECOMPILES/36.raftp2p_library.wasm
    mv -f $WASMX_GOV_CONTINUOUS/build/release.wasm $WASMX_GO_PRECOMPILES/37.gov_cont_0.0.1.wasm
    mv -f $WASMX_AUTH/build/release.wasm $WASMX_GO_PRECOMPILES/38.auth_0.0.1.wasm
    mv -f $WASMX_ROLES/build/release.wasm $WASMX_GO_PRECOMPILES/60.roles_0.0.1.wasm
    mv -f $WASMX_SLASHING/build/release.wasm $WASMX_GO_PRECOMPILES/45.slashing_0.0.1.wasm
    mv -f $WASMX_DISTRIBUTION/build/release.wasm $WASMX_GO_PRECOMPILES/46.distribution_0.0.1.wasm
    mv -f $WASMX_CHAT/build/release.wasm $WASMX_GO_PRECOMPILES/42.chat_0.0.1.wasm
    mv -f $WASMX_CHAT_VERIFIER/build/release.wasm $WASMX_GO_PRECOMPILES/44.chat_verifier_0.0.1.wasm
    mv -f $WASMX_TIME/build/release.wasm $WASMX_GO_PRECOMPILES/47.time_0.0.1.wasm
    mv -f $WASMX_LEVEL0/build/release.wasm $WASMX_GO_PRECOMPILES/48.level0_0.0.1.wasm
    mv -f $WASMX_MULTICHAIN/build/release.wasm $WASMX_GO_PRECOMPILES/4a.multichain_registry_0.0.1.wasm
    mv -f $WASMX_MULTICHAIN_LOCAL/build/release.wasm $WASMX_GO_PRECOMPILES/4b.multichain_registry_local_0.0.1.wasm
    mv -f $WASMX_ERC20_ROLLUP/build/release.wasm $WASMX_GO_PRECOMPILES/4c.erc20rollupjson_0.0.1.wasm
    mv -f $WASMX_LOBBY/build/release.wasm $WASMX_GO_PRECOMPILES/4d.lobby_json_0.0.1.wasm
    mv -f $WASMX_METAREGISTRY/build/release.wasm $WASMX_GO_PRECOMPILES/4f.metaregistry_json_0.0.1.wasm
    # mv -f $WASMX_PARAMS/build/release.wasm $WASMX_GO_PRECOMPILES/50.params_json_0.0.1.wasm
    mv -f $WASMX_LEVEL0_ONDEMAND/build/release.wasm $WASMX_GO_PRECOMPILES/51.level0_ondemand_0.0.1.wasm
    mv -f $WASMX_CODES_REGISTRY/build/release.wasm $WASMX_GO_PRECOMPILES/61.wasmx_codes_registry_0.0.1.wasm
    mv -f $WASMX_DTYPE/build/release.wasm $WASMX_GO_PRECOMPILES/62.wasmx_dtype_0.0.1.wasm
    mv -f $WASMX_EMAIL/build/release.wasm $WASMX_GO_PRECOMPILES/63.wasmx_email_0.0.1.wasm

    # tests
    mv -f $WASMX_TESTS_CROSSCHAIN/build/release.wasm $WASMX_GO_TESTDATA_NETWORK/crosschain.wasm
    mv -f $WASMX_TESTS_SIMPLESTORAGE/build/release.wasm $WASMX_GO_TESTDATA_NETWORK/simple_storage.wasm

    mv -f $WASMX_TESTS_SQL/build/release.wasm $WASMX_GO_TESTDATA_SQL/wasmx_test_sql.wasm
    mv -f $WASMX_TESTS_KVDB/build/release.wasm $WASMX_GO_TESTDATA_KVDB/wasmx_test_kvdb.wasm
    mv -f $WASMX_ERC20_DTYPE/build/release.wasm $WASMX_GO_TESTDATA_SQL/wasmx_erc20_sql.wasm
    mv -f $WASMX_TESTS_IMAP/build/release.wasm $WASMX_GO_TESTDATA_IMAP/wasmx_test_imap.wasm
     mv -f $WASMX_TESTS_SMTP/build/release.wasm $WASMX_GO_TESTDATA_SMTP/wasmx_test_smtp.wasm
fi
