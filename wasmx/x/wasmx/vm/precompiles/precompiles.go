package precompiles

import (
	_ "embed"

	address "cosmossdk.io/core/address"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

var (
	//go:embed 01.ecrecover.e.wasm
	ecrecovereth []byte

	//go:embed 02.sha2-256.e.wasm
	sha2_256 []byte

	//go:embed 03.ripmd160.e.wasm
	ripmd160 []byte

	//go:embed 04.identity.e.wasm
	identity []byte

	//go:embed 05.modexp.e.wasm
	modexp []byte

	//go:embed 06.ecadd.e.wasm
	ecadd []byte

	//go:embed 07.ecmul.e.wasm
	ecmul []byte

	//go:embed 08.ecpairings.e.wasm
	ecpairings []byte

	//go:embed 09.blake2f.e.wasm
	blake2f []byte

	//go:embed 20.secp384r1.wasm
	secp384r1 []byte

	//go:embed 21.secp384r1_registry.wasm
	secp384r1_registry []byte

	//go:embed 22.secret_sharing.wasm
	secret_sharing []byte

	//go:embed 23.evm_shanghai.wasm
	interpreter_evm_shanghai []byte

	//go:embed 24.alias_eth.wasm
	alias_eth []byte

	//go:embed 26.rustpython.wasm
	rustpython []byte

	//go:embed 27.quickjs.wasm
	quickjs []byte

	//go:embed 28.finite_state_machine.wasm
	state_machine []byte

	//go:embed 29.storage_chain.wasm
	storage_chain []byte

	//go:embed 2a.raft_library.wasm
	raft_library []byte

	//go:embed 36.raftp2p_library.wasm
	raftp2p_library []byte

	//go:embed 2b.tendermint_library.wasm
	tendermint_library []byte

	//go:embed 40.tendermintp2p_library.wasm
	tendermintp2p_library []byte

	//go:embed 2e.ava_snowman_library.wasm
	ava_snowman_library []byte

	//go:embed 30.staking_0.0.1.wasm
	staking_contract []byte

	//go:embed 31.bank_0.0.1.wasm
	bank_contract []byte

	//go:embed 32.erc20json_0.0.1.wasm
	erc20json_contract []byte

	//go:embed 33.derc20json_0.0.1.wasm
	derc20json_contract []byte

	//go:embed 34.hooks_0.0.1.wasm
	hooks_contract []byte

	//go:embed 35.gov_0.0.1.wasm
	gov_contract []byte

	//go:embed 37.gov_cont_0.0.1.wasm
	gov_cont_contract []byte

	//go:embed 38.auth_0.0.1.wasm
	auth_contract []byte

	//go:embed 45.slashing_0.0.1.wasm
	slashing_contract []byte

	//go:embed 46.distribution_0.0.1.wasm
	distribution_contract []byte

	//go:embed 42.chat_0.0.1.wasm
	chat_contract []byte

	//go:embed 44.chat_verifier_0.0.1.wasm
	chat_verifier_contract []byte

	//go:embed 47.time_0.0.1.wasm
	time_contract []byte

	//go:embed 48.level0_0.0.1.wasm
	level0_contract []byte

	//go:embed 4a.multichain_registry_0.0.1.wasm
	multichain_registry []byte

	//go:embed 4b.multichain_registry_local_0.0.1.wasm
	multichain_registry_local []byte

	//go:embed 4c.erc20rollupjson_0.0.1.wasm
	erc20rollupjson []byte

	//go:embed 4d.lobby_json_0.0.1.wasm
	lobbyjson []byte

	//go:embed 4f.metaregistry_json_0.0.1.wasm
	metaregistryjson []byte

	//go:embed 50.tay_interpreter_v0.0.1.wasm
	tay_interpreter []byte

	//go:embed 51.level0_ondemand_0.0.1.wasm
	level0_ondemand_contract []byte

	//go:embed 60.roles_0.0.1.wasm
	roles_contract []byte

	//go:embed 61.wasmx_codes_registry_0.0.1.wasm
	wasmx_codes_registry []byte

	//go:embed 62.wasmx_dtype_0.0.1.wasm
	wasmx_dtype []byte

	//go:embed ff.sys_proxy.wasm
	sys_proxy []byte
)

func GetPrecompileByLabel(addrCodec address.Codec, label string) []byte {
	var wasmbin []byte
	switch label {
	case "ecrecovereth":
		wasmbin = ecrecovereth
	case "sha2-256":
		wasmbin = sha2_256
	case "ripmd160":
		wasmbin = ripmd160
	case "identity":
		wasmbin = identity
	case "modexp":
		wasmbin = modexp
	case "ecadd":
		wasmbin = ecadd
	case "ecmul":
		wasmbin = ecmul
	case "ecpairings":
		wasmbin = ecpairings
	case "blake2f":
		wasmbin = blake2f
	case "secp384r1":
		wasmbin = secp384r1
	case types.SECP384r1_REGISTRY:
		wasmbin = secp384r1_registry
	case "secret_sharing":
		wasmbin = secret_sharing
	case types.INTERPRETER_EVM_SHANGHAI:
		wasmbin = interpreter_evm_shanghai
	case types.ALIAS_ETH:
		wasmbin = alias_eth
	case types.INTERPRETER_PYTHON:
		wasmbin = rustpython
	case types.INTERPRETER_JS:
		wasmbin = quickjs
	case types.INTERPRETER_FSM:
		wasmbin = state_machine
	case types.CONSENSUS_RAFT:
		libaddr := types.AccAddressFromHex(types.ADDR_CONSENSUS_RAFT_LIBRARY)
		libaddrstr, err := addrCodec.BytesToString(libaddr)
		if err != nil {
			panic(err)
		}
		wasmbin = []byte(ConsensusRaftv001(libaddrstr))
	case types.CONSENSUS_RAFTP2P:
		libaddr := types.AccAddressFromHex(types.ADDR_CONSENSUS_RAFTP2P_LIBRARY)
		libaddrstr, err := addrCodec.BytesToString(libaddr)
		if err != nil {
			panic(err)
		}
		wasmbin = []byte(ConsensusRaftP2Pv001(libaddrstr))
	case types.CONSENSUS_TENDERMINT:
		libaddr := types.AccAddressFromHex(types.ADDR_CONSENSUS_TENDERMINT_LIBRARY)
		libaddrstr, err := addrCodec.BytesToString(libaddr)
		if err != nil {
			panic(err)
		}
		wasmbin = []byte(ConsensusTendermintv001(libaddrstr))
	case types.CONSENSUS_TENDERMINTP2P:
		libaddr := types.AccAddressFromHex(types.ADDR_CONSENSUS_TENDERMINTP2P_LIBRARY)
		libaddrstr, err := addrCodec.BytesToString(libaddr)
		if err != nil {
			panic(err)
		}
		wasmbin = []byte(ConsensusTendermintP2Pv001(libaddrstr))
	case types.CONSENSUS_AVA_SNOWMAN:
		libaddr := types.AccAddressFromHex(types.ADDR_CONSENSUS_AVA_SNOWMAN_LIBRARY)
		libaddrstr, err := addrCodec.BytesToString(libaddr)
		if err != nil {
			panic(err)
		}
		wasmbin = []byte(ConsensusAvaSnowmanv001(libaddrstr))
	case types.CONSENSUS_RAFT_LIBRARY:
		wasmbin = raft_library
	case types.CONSENSUS_RAFTP2P_LIBRARY:
		wasmbin = raftp2p_library
	case types.CONSENSUS_TENDERMINT_LIBRARY:
		wasmbin = tendermint_library
	case types.CONSENSUS_TENDERMINTP2P_LIBRARY:
		wasmbin = tendermintp2p_library
	case types.CONSENSUS_AVA_SNOWMAN_LIBRARY:
		wasmbin = ava_snowman_library
	case types.SYS_PROXY:
		wasmbin = sys_proxy
	case types.STORAGE_CHAIN:
		wasmbin = storage_chain
	case types.STAKING_v001:
		wasmbin = staking_contract
	case types.BANK_v001:
		wasmbin = bank_contract
	case types.ERC20_v001:
		wasmbin = erc20json_contract
	case types.DERC20_v001:
		wasmbin = derc20json_contract
	case types.HOOKS_v001:
		wasmbin = hooks_contract
	case types.GOV_v001:
		wasmbin = gov_contract
	case types.GOV_CONT_v001:
		wasmbin = gov_cont_contract
	case types.AUTH_v001:
		wasmbin = auth_contract
	case types.ROLES_v001:
		wasmbin = roles_contract
	case types.SLASHING_v001:
		wasmbin = slashing_contract
	case types.DISTRIBUTION_v001:
		wasmbin = distribution_contract
	case types.CHAT_v001:
		wasmbin = chat_contract
	case types.CHAT_VERIFIER_v001:
		wasmbin = chat_verifier_contract
	case types.TIME_v001:
		wasmbin = time_contract
	case types.CONSENSUS_LEVEL_LIBRARY:
		wasmbin = level0_contract
	case types.LEVEL0_v001:
		libaddr := types.AccAddressFromHex(types.ADDR_LEVEL0_LIBRARY)
		libaddrstr, err := addrCodec.BytesToString(libaddr)
		if err != nil {
			panic(err)
		}
		wasmbin = []byte(Level0P2Pv001(libaddrstr))
	case types.LEVEL0_ONDEMAND_LIBRARY:
		wasmbin = level0_ondemand_contract
	case types.LEVEL0_ONDEMAND_v001:
		libaddr := types.AccAddressFromHex(types.ADDR_LEVEL0_ONDEMAND_LIBRARY)
		libaddrstr, err := addrCodec.BytesToString(libaddr)
		if err != nil {
			panic(err)
		}
		wasmbin = []byte(Level0OnDemand001(libaddrstr))
	case types.MULTICHAIN_REGISTRY_v001:
		wasmbin = multichain_registry
	case types.MULTICHAIN_REGISTRY_LOCAL_v001:
		wasmbin = multichain_registry_local
	case types.ERC20_ROLLUP_v001:
		wasmbin = erc20rollupjson
	case types.LOBBY_LIBRARY:
		wasmbin = lobbyjson
	case types.LOBBY_v001:
		libaddr := types.AccAddressFromHex(types.ADDR_LOBBY_LIBRARY)
		libaddrstr, err := addrCodec.BytesToString(libaddr)
		if err != nil {
			panic(err)
		}
		wasmbin = []byte(LobbyP2Pv001(libaddrstr))
	case types.METAREGISTRY_v001:
		wasmbin = metaregistryjson
	case types.INTERPRETER_TAY:
		wasmbin = tay_interpreter
	case types.STORAGE_CONTRACTS_v001:
		wasmbin = wasmx_codes_registry
	case types.DTYPE_v001:
		wasmbin = wasmx_dtype
	}
	return wasmbin
}
