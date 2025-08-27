package lib

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	authlib "github.com/loredanacirstea/wasmx-auth/lib"
	banklib "github.com/loredanacirstea/wasmx-bank/lib"
	consensus "github.com/loredanacirstea/wasmx-consensus"
	distributionlib "github.com/loredanacirstea/wasmx-distribution/lib"
	wasmx "github.com/loredanacirstea/wasmx-env"
	xchain "github.com/loredanacirstea/wasmx-env-crosschain"
	envutils "github.com/loredanacirstea/wasmx-env-utils"
	govmod "github.com/loredanacirstea/wasmx-gov/gov"
	slashinglib "github.com/loredanacirstea/wasmx-slashing/lib"
	stakinglib "github.com/loredanacirstea/wasmx-staking/lib"
	wasmxwasmx "github.com/loredanacirstea/wasmx-wasmx"
)

// External entrypoints — return []byte to mirror AS ArrayBuffer

func InitSubChain(req InitSubChainRequest) []byte {
	LoggerInfo("initializing subchain", []string{"subchain_id", req.ChainID})
	lastLevel := GetLevelLast()
	chaindata := GetChainData(req.ChainID)
	if chaindata == nil {
		Revert("subchain not registered: " + req.ChainID)
		return []byte{}
	}
	params := GetParams()
	maxNewUpperLevels := lastLevel + 1 - chaindata.Level
	initSubChainInternalWithAuth(*chaindata, params.MinValidatorsCount, maxNewUpperLevels)
	LoggerInfo("initialized subchain", []string{"subchain_id", req.ChainID})
	return []byte{}
}

func RegisterDefaultSubChain(req RegisterDefaultSubChainRequest) []byte {
	if !passCheckEIDActive(wasmx.GetCaller()) {
		Revert("unauthorized: no eID active")
	}
	LoggerInfo("start registering new default subchain", []string{"chain_base_name", req.ChainBaseName})
	chainId := registerDefaultChainId(req.ChainBaseName, int32(req.LevelIndex))
	_ = registerDefaultSubChainInternal(GetParams(), chainId.Full, req, INITIAL_LEVEL, map[wasmx.Bech32String][]wasmx.ContractStorage{})
	LoggerInfo("registered new default subchain", []string{"chain_base_name", req.ChainBaseName})
	return []byte{}
}

func RegisterSubChain(req RegisterSubChainRequest) []byte {
	if !passCheckEIDActive(wasmx.GetCaller()) {
		Revert("unauthorized: no eID active")
	}
	initialLevel := int32(1)
	_ = registerSubChainInternal(req.Data, req.GenTxs, req.InitialBalance, initialLevel)
	return []byte{}
}

func RegisterSubChainValidator(req RegisterSubChainValidatorRequest) []byte {
	if !passCheckEIDActive(wasmx.GetCaller()) {
		Revert("unauthorized: no eID active")
	}
	LoggerInfo("start registering new subchain validator", []string{"subchain_id", req.ChainID})
	registerSubChainValidatorInternal(req.ChainID, req.GenTx)
	return []byte{}
}

func RemoveSubChain(req RemoveSubChainRequest) []byte {
	removeSubChain(req.ChainID) // TODO: authorization + not initialized
	LoggerInfo("removed subchain temporary data", []string{"subchain_id", req.ChainID})
	return []byte{}
}

// Queries
func GetSubChains(_ QueryGetSubChainsRequest) []byte {
	ids := GetChainIds()
	out := []consensus.InitSubChainDeterministicRequest{}
	for _, id := range ids {
		chain := GetChainData(id)
		if chain != nil && chain.Initialized {
			out = append(out, chain.Data)
		}
	}
	bz, _ := json.Marshal(&out)
	return bz
}

func GetSubChainsByIds(req QueryGetSubChainsByIdsRequest) []byte {
	out := []consensus.InitSubChainDeterministicRequest{}
	for _, id := range req.Ids {
		chain := GetChainData(id)
		if chain != nil && chain.Initialized {
			out = append(out, chain.Data)
		}
	}
	bz, _ := json.Marshal(&out)
	return bz
}

func GetSubChainIds(_ QueryGetSubChainIdsRequest) []byte {
	ids := GetChainIds()
	bz, _ := json.Marshal(&ids)
	return bz
}

func GetSubChainById(req QueryGetSubChainRequest) []byte {
	chain := GetChainData(req.ChainID)
	if chain == nil {
		return []byte{}
	}
	bz, _ := json.Marshal(&chain.Data)
	return bz
}

func GetSubChainConfigById(req QueryGetSubChainRequest) []byte {
	c := subChainConfigById(req.ChainID)
	if c == nil {
		return []byte{}
	}
	bz, _ := json.Marshal(c)
	return bz
}

func GetSubChainConfigByIds(req QuerySubChainConfigByIdsRequest) []byte {
	out := []consensus.ChainConfig{}
	for _, id := range req.Ids {
		c := subChainConfigById(id)
		if c != nil {
			out = append(out, *c)
		}
	}
	bz, _ := json.Marshal(&out)
	return bz
}

func GetSubChainIdsByLevel(req QueryGetSubChainIdsByLevelRequest) []byte {
	ids := GetLevelChainIds(req.Level)
	bz, _ := json.Marshal(&ids)
	return bz
}

func GetCurrentLevel(_ QueryGetCurrentLevelRequest) []byte {
	level := GetCurrentLevel()
	bz, _ := json.Marshal(&QueryGetCurrentLevelResponse{Level: level})
	return bz
}

func GetSubChainIdsByValidator(req QueryGetSubChainIdsByValidatorRequest) []byte {
	ids := GetValidatorChains(string(req.ValidatorAddress))
	bz, _ := json.Marshal(&ids)
	return bz
}

func GetValidatorsByChainId(req QueryGetValidatorsByChainIdRequest) []byte {
	gentxs := GetChainValidators(req.ChainID)
	// JSON encodes []byte as base64 strings, which matches AS encoding
	bz, _ := json.Marshal(&gentxs)
	return bz
}

func GetValidatorAddressesByChainId(req QueryValidatorAddressesByChainIdRequest) []byte {
	addrs := GetChainValidatorAddresses(req.ChainID)
	bz, _ := json.Marshal(&addrs)
	return bz
}

func ConvertAddressByChainId(req QueryConvertAddressByChainIdRequest) []byte {
	prefix := req.Prefix
	if req.ChainID != "" {
		cfg := subChainConfigById(req.ChainID)
		if cfg == nil {
			return []byte{}
		}
		switch req.Type {
		case "acc":
			prefix = cfg.Bech32PrefixAccAddr
		case "cons":
			prefix = cfg.Bech32PrefixConsAddr
		case "val":
			prefix = cfg.Bech32PrefixValAddr
		}
	}
	newaddr := convertAddress(req.Address, prefix)
	return []byte(newaddr)
}

// Cross-chain stubs (conversion implemented; execution not in TinyGo env yet)
func CrossChainTx(req xchain.MsgCrossChainCallRequest) []byte {
	// Prepare addresses for chain prefixes
	p := prepareCrossChainCallRequest(req)
	if p == nil {
		r := xchain.MsgCrossChainCallResponse{Error: "target chain configuration not found"}
		bz, _ := json.Marshal(&r)
		return bz
	}
	resp, err := xchain.ExecuteCrossChainTx(*p)
	if err != nil {
		r := xchain.MsgCrossChainCallResponse{Error: err.Error()}
		bz, _ := json.Marshal(&r)
		return bz
	}
	bz, _ := json.Marshal(&resp)
	return bz
}
func CrossChainQuery(req xchain.MsgCrossChainCallRequest) []byte {
	p := prepareCrossChainCallRequest(req)
	if p == nil {
		r := xchain.MsgCrossChainCallResponse{Error: "target chain configuration not found"}
		bz, _ := json.Marshal(&r)
		return bz
	}
	resp, err := xchain.ExecuteCrossChainQuery(*p)
	if err != nil {
		r := xchain.MsgCrossChainCallResponse{Error: err.Error()}
		bz, _ := json.Marshal(&r)
		return bz
	}
	bz, _ := json.Marshal(&resp)
	return bz
}
func CrossChainQueryNonDeterministic(req xchain.MsgCrossChainCallRequest) []byte {
	p := prepareCrossChainCallRequest(req)
	if p == nil {
		r := xchain.MsgCrossChainCallResponse{Error: "target chain configuration not found"}
		bz, _ := json.Marshal(&r)
		return bz
	}
	resp, err := xchain.ExecuteCrossChainQueryNonDeterministic(*p)
	if err != nil {
		r := xchain.MsgCrossChainCallResponse{Error: err.Error()}
		bz, _ := json.Marshal(&r)
		return bz
	}
	bz, _ := json.Marshal(&resp)
	return bz
}

// Helpers
func prepareCrossChainCallRequest(req xchain.MsgCrossChainCallRequest) *xchain.MsgCrossChainCallRequest {
	caller := wasmx.GetCaller()
	toCfg := subChainConfigById(req.ToChainId)
	if toCfg == nil {
		return nil
	}
	req.FromChainId = wasmx.GetChainId()
	// req.From is bech32 bytes for target chain prefix
	req.From = []byte(wasmx.AddrHumanizeMC([]byte(string(caller)), toCfg.Bech32PrefixAccAddr))
	// canonicalize and re-humanize to target prefix
	bz := wasmx.AddrCanonicalizeMC(string(req.To))
	req.To = []byte(wasmx.AddrHumanizeMC(bz, toCfg.Bech32PrefixAccAddr))
	req.TimeoutMs = CROSS_CHAIN_TIMEOUT_MS
	return &req
}

func convertAddress(sourceAddr wasmx.Bech32String, prefix string) string {
	bz := wasmx.AddrCanonicalizeMC(string(sourceAddr))
	return wasmx.AddrHumanizeMC(bz, prefix)
}

func subChainConfigById(chainId string) *consensus.ChainConfig {
	cd := GetChainData(chainId)
	if cd == nil {
		return nil
	}
	cfg := cd.Data.ChainConfig
	return &cfg
}

func tryRegisterUpperLevel(lastRegisteredLevel int32, _lastRegisteredChainId string, trynextlevel int32) {
	params := GetParams()
	levelchains := GetLevelChainIds(lastRegisteredLevel)
	count := len(levelchains)
	if count%int(params.MinValidatorsCount) > 0 {
		return
	}

	nextLevel := lastRegisteredLevel + 1
	upperlevels := GetLevelChainIds(nextLevel)
	expected := int(count / int(params.MinValidatorsCount))
	if len(upperlevels) >= expected {
		return
	}

	LoggerInfo("registering subchain", []string{"subchain_level", fmt.Sprintf("%d", nextLevel)})

	subchainIds := levelchains[count-int(params.MinValidatorsCount):]

	wasmxContractState := map[wasmx.Bech32String][]wasmx.ContractStorage{}
	// store CURRENT_LEVEL
	wasmxContractState[wasmx.Bech32String(wasmxwasmx.ADDR_MULTICHAIN_REGISTRY)] = append(
		wasmxContractState[wasmx.Bech32String(wasmxwasmx.ADDR_MULTICHAIN_REGISTRY)],
		wasmx.ContractStorage{Key: wasmx.HexString(hex.EncodeToString([]byte(CURRENT_LEVEL))), Value: []byte(fmt.Sprintf("%d", nextLevel))},
	)
	// store child ids for lastRegisteredLevel key
	bz, _ := json.Marshal(&subchainIds)
	wasmxContractState[wasmx.Bech32String(wasmxwasmx.ADDR_MULTICHAIN_REGISTRY)] = append(
		wasmxContractState[wasmx.Bech32String(wasmxwasmx.ADDR_MULTICHAIN_REGISTRY)],
		wasmx.ContractStorage{Key: wasmx.HexString(hex.EncodeToString([]byte(GetLevelChainIdsKey(lastRegisteredLevel)))), Value: []byte(string(bz))},
	)
	for _, id := range subchainIds {
		v := wasmx.SLoad(GetDataKey(id))
		wasmxContractState[wasmx.Bech32String(wasmxwasmx.ADDR_MULTICHAIN_REGISTRY)] = append(
			wasmxContractState[wasmx.Bech32String(wasmxwasmx.ADDR_MULTICHAIN_REGISTRY)],
			wasmx.ContractStorage{Key: wasmx.HexString(hex.EncodeToString([]byte(GetDataKey(id)))), Value: []byte(v)},
		)
	}

	newChainId := registerDefaultChainIdSubChainLevel(nextLevel)
	subchaindata := registerDefaultSubChainLevel(GetParams(), newChainId, nextLevel, wasmxContractState, params.LevelInitialBalance)

	// collect one validator from each composing subchain
	valInfos := []ValidatorInfo{}
	for _, sid := range subchainIds {
		valInfo := getChainValidatorInfoFromSubChain(sid, 0)
		if valInfo == nil {
			continue
		}
		val := valInfo.Validator
		// convert operator address to target chain prefix
		bz := wasmx.AddrCanonicalizeMC(string(val.OperatorAddress))
		newValAddr := wasmx.AddrHumanizeMC(bz, subchaindata.Data.ChainConfig.Bech32PrefixAccAddr)
		v := stakinglib.Validator{
			OperatorAddress:         wasmx.Bech32String(newValAddr),
			ConsensusPubkey:         val.ConsensusPubkey,
			Jailed:                  false,
			Status:                  stakinglib.BondedS,
			Tokens:                  params.LevelInitialBalance.Quo(sdkmath.NewInt(100)),
			DelegatorShares:         val.DelegatorShares,
			Description:             val.Description,
			UnbondingHeight:         val.UnbondingHeight,
			UnbondingTime:           val.UnbondingTime,
			Commission:              val.Commission,
			MinSelfDelegation:       val.MinSelfDelegation,
			UnbondingOnHoldRefCount: 0,
			UnbondingIds:            []uint64{},
		}
		// adjust peer address memo
		valInfo.P2PAddress = replacePeerOperatorAddress(valInfo.P2PAddress, wasmx.Bech32String(newValAddr))
		valInfos = append(valInfos, ValidatorInfo{Validator: v, OperatorPubkey: valInfo.OperatorPubkey, P2PAddress: valInfo.P2PAddress})
	}
	LoggerInfo("registering subchain with validators", []string{"subchain_level", fmt.Sprintf("%d", nextLevel), "subchain_id", newChainId.Full, "validator_count", fmt.Sprintf("%d", len(valInfos))})
	subchaindata = includeValidatorInfos(subchaindata, valInfos)
	SetChainData(subchaindata)

	LoggerInfo("initializing subchain", []string{"subchain_level", fmt.Sprintf("%d", nextLevel), "subchain_id", newChainId.Full})
	initSubChainInternalWithAuth(subchaindata, params.MinValidatorsCount, trynextlevel)
	LoggerInfo("initialized subchain", []string{"subchain_level", fmt.Sprintf("%d", nextLevel), "subchain_id", newChainId.Full})
}

func registerDefaultSubChainLevel(params Params, chainId consensus.ChainId, levelIndex int32, wasmxContractState map[wasmx.Bech32String][]wasmx.ContractStorage, initialBalance sdkmath.Int) SubChainData {
	denomUnit := fmt.Sprintf("lvl%d", levelIndex)
	req := RegisterDefaultSubChainRequest{DenomUnit: denomUnit, BaseDenomUnit: 18, ChainBaseName: chainId.BaseName, LevelIndex: uint32(levelIndex), InitialBalance: initialBalance, GenTxs: [][]byte{}}
	return registerDefaultSubChainInternal(params, chainId.Full, req, levelIndex, wasmxContractState)
}

func getChainBaseNameSubChainLevel(levelIndex int32) string {
	return fmt.Sprintf("level%d", levelIndex)
}

func registerDefaultChainIdSubChainLevel(levelIndex int32) consensus.ChainId {
	base := getChainBaseNameSubChainLevel(levelIndex)
	return registerDefaultChainId(base, levelIndex)
}

func registerDefaultChainId(chainBaseName string, levelIndex int32) consensus.ChainId {
	// start at 1 to leave space for level0 ids
	chainIndex := GetChainIdLast() + 1
	SetChainIdLast(chainIndex)
	return consensus.ChainId{Full: consensus.ChainIdToString(chainBaseName, uint32(levelIndex), int64(chainIndex), 1), BaseName: chainBaseName, Level: uint32(levelIndex), EvmID: chainIndex, ForkIndex: 1}
}

func buildDefaultSubChainGenesisInternal(params Params, chainId string, currentLevel int32, chainConfig consensus.ChainConfig, req RegisterDefaultSubChainRequest, wasmxContractState map[wasmx.Bech32String][]wasmx.ContractStorage, initialPorts consensus.NodePorts) consensus.InitSubChainDeterministicRequest {
	peers := []string{}
	defaultInitialHeight := int64(1)
	consensusParams := consensus.GetDefaultConsensusParams()

	// derive special addresses for this chain
	bootstrap := bech32FromHex(wasmx.ADDR_BOOTSTRAP_ACCOUNT, chainConfig.Bech32PrefixAccAddr)
	feeCollector := bech32FromHex(wasmx.ADDR_FEE_COLLECTOR, chainConfig.Bech32PrefixAccAddr)
	mint := bech32FromHex(wasmx.GOCORE_ADDR_MINT, chainConfig.Bech32PrefixAccAddr)

	genesisState := buildGenesisData(params, req.DenomUnit, req.BaseDenomUnit, bootstrap, feeCollector, mint, currentLevel, wasmxContractState, initialPorts, chainConfig.Bech32PrefixAccAddr)

	appStateBz, _ := json.Marshal(&genesisState)
	initChainReq := consensus.RequestInitChain{
		Time:            time.Now().UTC().Format(time.RFC3339Nano),
		ChainID:         chainId,
		ConsensusParams: consensusParams,
		Validators:      []consensus.ValidatorUpdate{},
		AppStateBytes:   appStateBz,
		InitialHeight:   defaultInitialHeight,
	}
	return consensus.InitSubChainDeterministicRequest{InitChainRequest: initChainReq, ChainConfig: chainConfig}
}

func registerDefaultSubChainInternal(params Params, chainId string, req RegisterDefaultSubChainRequest, levelIndex int32, wasmxContractState map[wasmx.Bech32String][]wasmx.ContractStorage) SubChainData {
	chainConfig := consensus.BuildChainConfig(req.DenomUnit, req.BaseDenomUnit, req.ChainBaseName)
	data := buildDefaultSubChainGenesisInternal(params, chainId, levelIndex, chainConfig, req, wasmxContractState, consensus.DefaultNodePorts(), chainConfig.Bech32PrefixAccAddr)
	return registerSubChainInternal(data, req.GenTxs, req.InitialBalance, levelIndex)
}

func registerSubChainInternal(data consensus.InitSubChainDeterministicRequest, genTxs [][]byte, initialBalance sdkmath.Int, levelIndex int32) SubChainData {
	chainId := data.InitChainRequest.ChainID
	AddChainId(chainId)
	chaindata := SubChainData{Data: data, GenTxs: genTxs, InitialBalance: initialBalance, Level: levelIndex, WasmxContractState: map[wasmx.Bech32String][]wasmx.ContractStorage{}}
	SetChainData(chaindata)
	for _, gt := range genTxs {
		registerSubChainValidatorInternal(chainId, gt)
	}

	ev := wasmx.Event{Type: EventTypeRegisterSubChain, Attributes: []wasmx.EventAttribute{{Key: AttributeKeyChainId, Value: chainId, Index: true}}}
	wasmx.EmitCosmosEvents([]wasmx.Event{ev})
	return chaindata
}

func registerSubChainValidatorInternal(chainId string, genTx []byte) {
	gtStr := string(decodeB64IfNeeded(genTx))
	var tx wasmx.SignedTransaction
	_ = json.Unmarshal([]byte(gtStr), &tx)
	msg := extractCreateValidatorMsg(tx)
	if msg == nil {
		Revert("invalid gentx: does not contain MsgCreateValidator")
		return
	}
	caller := wasmx.GetCaller()
	if !wasmx.AddrEquivalent(caller, wasmx.Bech32String(msg.ValidatorAddress)) {
		Revert(fmt.Sprintf("unauthorized: caller %s, validator %s", caller, msg.ValidatorAddress))
	}
	AddChainValidator(chainId, wasmx.Bech32String(msg.ValidatorAddress), genTx)

	chaindata := GetChainData(chainId)
	if chaindata == nil {
		Revert("no subchain found: " + chainId)
		return
	}
	chaindata.Data.Peers = append(chaindata.Data.Peers, tx.Body.Memo)
	chaindata.GenTxs = append(chaindata.GenTxs, genTx)
	SetChainData(*chaindata)
	LoggerInfo("registered new subchain validator", []string{"subchain_id", chainId, "address", msg.ValidatorAddress})

	ev := wasmx.Event{Type: EventTypeRegisterSubChainValidator, Attributes: []wasmx.EventAttribute{
		{Key: AttributeKeyChainId, Value: chainId, Index: true},
		{Key: AttributeKeyValidator, Value: string(msg.ValidatorAddress), Index: true},
	}}
	wasmx.EmitCosmosEvents([]wasmx.Event{ev})
}

func removeSubChain(_ string) {}

func initSubChainInternalWithAuth(chaindata SubChainData, minValidatorCount int32, trynextlevel int32) {
	chainId := chaindata.Data.InitChainRequest.ChainID
	addrs := GetChainValidatorAddresses(chainId)
	if len(addrs) < int(minValidatorCount) {
		Revert(fmt.Sprintf("subchain needs at least %d validators; has %d addresses registered", minValidatorCount, len(addrs)))
	}
	caller := wasmx.GetCaller()
	if !wasmx.AddrEquivalent(caller, addrs[0]) {
		Revert(fmt.Sprintf("unauthorized: caller %s, validator %s", caller, addrs[0]))
	}
	genTxs := GetChainValidators(chainId)
	initSubChainInternal(chaindata, genTxs, minValidatorCount, trynextlevel)
}

func initSubChainInternal(chaindata SubChainData, genTxs [][]byte, minValidatorCount int32, trynextlevel int32) {
	chainId := chaindata.Data.InitChainRequest.ChainID
	chaindata = initSubChainPrepareData(chaindata, genTxs, minValidatorCount)
	SetChainData(chaindata)

	// emit event
	data, _ := json.Marshal(&chaindata.Data)
	data64 := base64.StdEncoding.EncodeToString(data)
	ev := wasmx.Event{Type: EventTypeInitSubChain, Attributes: []wasmx.EventAttribute{
		{Key: AttributeKeyChainId, Value: chaindata.Data.InitChainRequest.ChainID, Index: true},
		{Key: AttributeKeyRequest, Value: data64, Index: false},
	}}
	wasmx.EmitCosmosEvents([]wasmx.Event{ev})

	// chain level bookkeeping
	AddLevelChainId(chaindata.Level, chainId)
	if trynextlevel > 0 {
		tryRegisterUpperLevel(chaindata.Level, chainId, trynextlevel-1)
	}
}

// initSubChainPrepareData updates chaindata with timestamp, included genTxs and wasmx state,
// validates minimum validator count, and marks it initialized. It does not persist or emit events.
func initSubChainPrepareData(chaindata SubChainData, genTxs [][]byte, minValidatorCount int32) SubChainData {
	// set timestamp
	chaindata.Data.InitChainRequest.Time = time.Now().UTC().Format(time.RFC3339Nano)

	// work on a local genesis map copy
	var genesisState consensus.GenesisState
	if err := json.Unmarshal(chaindata.Data.InitChainRequest.AppStateBytes, &genesisState); err != nil {
		Revert("invalid app_state_bytes: " + err.Error())
	}
	genesisState = includeGenTxs(genesisState, genTxs, chaindata.InitialBalance)
	genesisState = includeWasmxState(genesisState, chaindata.WasmxContractState)

	validatorCount := getValidatorCountFromGenesis(genesisState)
	if int32(validatorCount) < minValidatorCount {
		Revert(fmt.Sprintf("subchain needs at least %d validators; has %d registered in genesis", minValidatorCount, validatorCount))
	}
	newGenesisBz, err := json.Marshal(&genesisState)
	if err != nil {
		Revert("cannot marshal updated genesis: " + err.Error())
	}
	chaindata.Data.InitChainRequest.AppStateBytes = newGenesisBz
	chaindata.Initialized = true
	return chaindata
}

func includeWasmxState(genesisState consensus.GenesisState, wasmxContractState map[wasmx.Bech32String][]wasmx.ContractStorage) consensus.GenesisState {
	// update wasmx state
	bz, ok := genesisState[wasmx.MODULE_WASMX]
	if !ok {
		Revert("genesis state missing field: wasmx")
		return genesisState
	}
	var ws wasmxwasmx.GenesisState
	_ = json.Unmarshal(bz, &ws)
	// merge state
	for i := range ws.SystemContracts {
		c := &ws.SystemContracts[i]
		if nv, ok := wasmxContractState[c.Address]; ok {
			c.ContractState = mergeWasmxState(c.ContractState, nv)
		}
	}
	nbz, _ := json.Marshal(&ws)
	genesisState[wasmx.MODULE_WASMX] = nbz
	return genesisState
}

func mergeWasmxState(oldstate []wasmx.ContractStorage, newstate []wasmx.ContractStorage) []wasmx.ContractStorage {
	extant := map[string]string{}
	for _, kv := range oldstate {
		extant[string(kv.Key)] = string(kv.Value)
	}
	for _, kv := range newstate {
		if v, ok := extant[string(kv.Key)]; ok {
			if v != string(kv.Value) {
				Revert(fmt.Sprintf("genesis wasmx contract storage value mismatch: key %s", kv.Key))
			}
		} else {
			oldstate = append(oldstate, kv)
		}
	}
	return oldstate
}

func includeGenTxs(genesisState consensus.GenesisState, genTxs [][]byte, initialBalance sdkmath.Int) consensus.GenesisState {
	if len(genTxs) == 0 {
		return genesisState
	}

	// genutil
	genutil := consensus.GenutilGenesis{GenTxs: genTxs}
	bz, _ := json.Marshal(&genutil)
	genesisState[wasmx.MODULE_GENUTIL] = bz

	// cosmosmod balances and accounts
	bz2, ok := genesisState[wasmx.MODULE_COSMOSMOD]
	if !ok {
		Revert("genesis state missing field: cosmosmod")
	}
	var cm CosmosmodGenesisState
	_ = json.Unmarshal(bz2, &cm)

	baseDenom := cm.Bank.DenomInfo[0].Metadata.Base

	for _, gt := range genTxs {
		gtStr := string(decodeB64IfNeeded(gt))
		var tx wasmx.SignedTransaction
		if err := json.Unmarshal([]byte(gtStr), &tx); err != nil {
			continue
		}
		msg := extractCreateValidatorMsg(tx)
		if msg == nil {
			continue
		}
		// signer pubkey
		if len(tx.AuthInfo.SignerInfos) == 0 {
			Revert("genTx transaction has empty signer_infos")
		}
		signer := tx.AuthInfo.SignerInfos[0]
		cm = includeValidatorAccountInfo(cm, wasmx.Bech32String(msg.ValidatorAddress), signer.PublicKey, initialBalance)
	}
	nbz, _ := json.Marshal(&cm)
	genesisState[wasmx.MODULE_COSMOSMOD] = nbz
	return genesisState
}

func passCheckEIDActive(addr wasmx.Bech32String) bool {
	params := GetParams()
	if params.EnableEidCheck {
		return isEIDActive(addr)
	}
	return true
}

func isEIDActive(addr wasmx.Bech32String) bool {
	const signature = "9f8a13d7"
	bz := wasmx.Bech32StringToBytes(string(addr))
	padded := envutils.PaddLeftTo32(bz)
	calldata := "0x" + signature + hex.EncodeToString(padded)
	req := wasmx.CallRequest{To: wasmx.ROLE_EID_REGISTRY, Calldata: calldata, Value: nil, GasLimit: 100000000, IsQuery: true}
	resp := wasmx.CallEvm(req, MODULE_NAME)
	if resp.Success > 0 {
		return false
	}
	dbz, err := base64.StdEncoding.DecodeString(resp.Data)
	if err != nil {
		return false
	}
	for _, b := range dbz {
		if b != 0 {
			return true
		}
	}
	return false
}

// buildGenesisData assembles a minimal viable genesis state map
func buildGenesisData(params Params, denomUnit string, baseDenomUnit uint32, bootstrapAccountBech32 string, feeCollectorBech32 string, mintBech32 string, currentLevel int32, wasmxContractState map[wasmx.Bech32String][]wasmx.ContractStorage, initialPorts consensus.NodePorts, bech32PrefixAccAddr string) consensus.GenesisState {
	// Bank: set three base denoms
	gasBaseDenom := "a" + denomUnit
	stakingBaseDenom := "as" + denomUnit
	rewardsBaseDenom := "arw" + denomUnit
	bank := banklib.GenesisState{Params: banklib.Params{DefaultSendEnabled: true}, Balances: []banklib.Balance{}, Supply: []wasmx.Coin{}, DenomInfo: []banklib.DenomDeploymentInfo{
		{Metadata: banklib.Metadata{Base: gasBaseDenom}},
		{Metadata: banklib.Metadata{Base: stakingBaseDenom}},
		{Metadata: banklib.Metadata{Base: rewardsBaseDenom}},
	}}
	staking := stakinglib.GenesisState{Params: stakinglib.Params{BondDenom: stakingBaseDenom}}
	gov := govmod.GetDefaultGenesis(gasBaseDenom, stakingBaseDenom, rewardsBaseDenom)
	slashing := slashinglib.GenesisState{}
	distribution := distributionlib.GenesisState{BaseDenom: stakingBaseDenom, RewardsDenom: rewardsBaseDenom}
	auth := authlib.GenesisState{}

	cosmosmod := CosmosmodGenesisState{Staking: staking, Bank: bank, Gov: gov, Auth: auth, Slashing: slashing, Distribution: distribution}

	// wasmx module genesis from defaults helper
	wasmxGenesis := wasmxwasmx.GetDefaultGenesis(bootstrapAccountBech32, feeCollectorBech32, mintBech32, uint32(params.MinValidatorsCount), params.EnableEidCheck, int(currentLevel), string(mustJSON(initialPorts)), bech32PrefixAccAddr, stakingBaseDenom)
	// apply contract state overrides
	for i := range wasmxGenesis.SystemContracts {
		c := &wasmxGenesis.SystemContracts[i]
		if st, ok := wasmxContractState[c.Address]; ok {
			c.ContractState = st
		}
	}

	gen := consensus.GenesisState{}
	gen[wasmx.MODULE_COSMOSMOD] = mustJSON(cosmosmod)
	gen[wasmx.MODULE_AUTHZ] = wasmx.GetDefaultAuthzGenesis()
	gen[wasmx.MODULE_CIRCUIT] = wasmx.GetDefaultCircuitGenesis()
	gen[wasmx.MODULE_CRISIS] = wasmx.GetDefaultCrisisGenesis(stakingBaseDenom)
	gen[wasmx.MODULE_EVIDENCE] = wasmx.GetDefaultEvidenceGenesis()
	gen[wasmx.MODULE_GENUTIL] = mustJSON(consensus.GenutilGenesis{GenTxs: [][]byte{}})
	gen[wasmx.MODULE_GROUP] = wasmx.GetDefaultGroupGenesis()
	gen[wasmx.MODULE_MINT] = wasmx.GetDefaultMintGenesis(gasBaseDenom)
	gen[wasmx.MODULE_NETWORK] = wasmx.GetDefaultNetworkGenesis()
	gen[wasmx.MODULE_UPGRADE] = wasmx.GetDefaultUpgradeGenesis()
	gen[wasmx.MODULE_WASMX] = mustJSON(wasmxGenesis)
	gen[wasmx.MODULE_WEBSRV] = wasmx.GetDefaultWebsrvGenesis()
	// transfer module defaults can be added when available
	return gen
}

func includeValidatorAccountInfo(cosmosmodGenesis CosmosmodGenesisState, operatorAddress wasmx.Bech32String, operatorPubKey *wasmx.PublicKey, initialBalance sdkmath.Int) CosmosmodGenesisState {
	baseDenom := cosmosmodGenesis.Bank.DenomInfo[0].Metadata.Base
	// balance
	balance := banklib.Balance{Address: operatorAddress, Coins: []wasmx.Coin{{Denom: baseDenom, Amount: initialBalance}}}
	cosmosmodGenesis.Bank.Balances = append(cosmosmodGenesis.Bank.Balances, balance)
	// account
	var accPub *wasmx.PublicKey
	if operatorPubKey != nil {
		v := *operatorPubKey
		accPub = &v
	}
	account := authlib.BaseAccount{Address: operatorAddress, PubKey: accPub, AccountNumber: 0, Sequence: 0}
	encoded, _ := json.Marshal(&account)
	any := wasmx.AnyWrap{TypeURL: authlib.TypeUrl_BaseAccount, Value: base64.StdEncoding.EncodeToString(encoded)}
	cosmosmodGenesis.Auth.Accounts = append(cosmosmodGenesis.Auth.Accounts, any)
	return cosmosmodGenesis
}

func includeValidatorInfos(data SubChainData, validators []ValidatorInfo) SubChainData {
	if len(validators) == 0 {
		return data
	}
	var gs consensus.GenesisState
	_ = json.Unmarshal(data.Data.InitChainRequest.AppStateBytes, &gs)
	// cosmosmod must exist
	bz, ok := gs[wasmx.MODULE_COSMOSMOD]
	if !ok {
		Revert("genesis state missing field: cosmosmod")
	}
	var cosmosmod CosmosmodGenesisState
	_ = json.Unmarshal(bz, &cosmosmod)

	chainId := data.Data.InitChainRequest.ChainID
	peers := []string{}
	for _, vinfo := range validators {
		val := vinfo.Validator
		AddChainValidatorAddress(chainId, val.OperatorAddress)
		cosmosmod = includeValidatorAccountInfo(cosmosmod, val.OperatorAddress, vinfo.OperatorPubkey, data.InitialBalance)
		cosmosmod.Staking.Validators = append(cosmosmod.Staking.Validators, val)
		cosmosmod.Staking.Delegations = append(cosmosmod.Staking.Delegations, stakinglib.Delegation{DelegatorAddress: string(val.OperatorAddress), ValidatorAddress: val.OperatorAddress, Amount: val.Tokens})
		peers = append(peers, vinfo.P2PAddress)
	}
	gs[wasmx.MODULE_COSMOSMOD] = mustJSON(cosmosmod)
	data.Data.InitChainRequest.AppStateBytes = mustJSON(gs)
	data.Data.Peers = peers
	return data
}

func getValidatorCountFromGenesis(genesisState consensus.GenesisState) int {
	count := 0
	if bz, ok := genesisState[wasmx.MODULE_GENUTIL]; ok {
		var data consensus.GenutilGenesis
		_ = json.Unmarshal(bz, &data)
		count += len(data.GenTxs)
	}
	if bz, ok := genesisState[wasmx.MODULE_COSMOSMOD]; ok {
		var data CosmosmodGenesisState
		_ = json.Unmarshal(bz, &data)
		count += len(data.Staking.Validators)
	}
	return count
}

func getChainValidatorInfoFromSubChain(chainId string, index int) *ValidatorInfo {
	if v := getChainValidatorInfoFromGenTx(chainId, index); v != nil {
		return v
	}
	return getChainValidatorInfoFromGenesis(chainId, index)
}

func getChainValidatorInfoFromGenTx(chainId string, index int) *ValidatorInfo {
	genTxs := GetChainValidators(chainId)
	if len(genTxs) <= index {
		return nil
	}
	genTx := genTxs[index]
	gtStr := string(decodeB64IfNeeded(genTx))
	var tx wasmx.SignedTransaction
	if err := json.Unmarshal([]byte(gtStr), &tx); err != nil {
		return nil
	}
	msg := extractCreateValidatorMsg(tx)
	if msg == nil {
		Revert("invalid gentx: does not contain MsgCreateValidator")
		return nil
	}
	if len(tx.AuthInfo.SignerInfos) == 0 {
		Revert("genTx transaction has empty signer_infos")
	}
	signer := tx.AuthInfo.SignerInfos[0]
	v := stakinglib.GetValidatorFromMsgCreate(*msg)
	return &ValidatorInfo{Validator: v, OperatorPubkey: signer.PublicKey, P2PAddress: tx.Body.Memo}
}

func getChainValidatorInfoFromGenesis(chainId string, index int) *ValidatorInfo {
	chaindata := GetChainData(chainId)
	if chaindata == nil {
		return nil
	}
	var gs consensus.GenesisState
	_ = json.Unmarshal(chaindata.Data.InitChainRequest.AppStateBytes, &gs)
	bz, ok := gs[wasmx.MODULE_COSMOSMOD]
	if !ok {
		return nil
	}
	var data CosmosmodGenesisState
	_ = json.Unmarshal(bz, &data)
	if len(data.Staking.Validators) <= index {
		LoggerDebug("index oob validators", []string{"index", fmt.Sprintf("%d", index)})
		return nil
	}
	if len(chaindata.Data.Peers) <= index {
		LoggerDebug("index oob peers", []string{"index", fmt.Sprintf("%d", index)})
		return nil
	}
	if len(data.Auth.Accounts) <= index {
		LoggerDebug("index oob accounts", []string{"index", fmt.Sprintf("%d", index)})
		return nil
	}
	valid := data.Staking.Validators[index]
	peer := chaindata.Data.Peers[index]
	// decode anywrap value
	var ba authlib.BaseAccount
	abz, _ := base64.StdEncoding.DecodeString(data.Auth.Accounts[index].Value)
	_ = json.Unmarshal(abz, &ba)
	return &ValidatorInfo{Validator: valid, OperatorPubkey: ba.PubKey, P2PAddress: peer}
}

// Utilities
func extractCreateValidatorMsg(tx wasmx.SignedTransaction) *stakinglib.MsgCreateValidator {
	if len(tx.Body.Messages) == 0 {
		return nil
	}
	m := tx.Body.Messages[0]
	if m.TypeUrl != stakinglib.TypeUrl_MsgCreateValidator {
		return nil
	}
	var msg stakinglib.MsgCreateValidator
	if err := json.Unmarshal(m.Value, &msg); err != nil {
		return nil
	}
	return &msg
}

func bech32FromHex(hexstr string, prefix string) string {
	bz, _ := hex.DecodeString(strings.TrimPrefix(hexstr, "0x"))
	return wasmx.AddrHumanizeMC(bz, prefix)
}

func decodeB64IfNeeded(bz []byte) []byte {
	// If input is base64 string bytes (e.g., "eyJhIjo..."), try to decode; otherwise return as-is
	s := string(bz)
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		return bz
	}
	if dbz, err := base64.StdEncoding.DecodeString(s); err == nil {
		return dbz
	}
	return bz
}

func mustJSON(v any) []byte { bz, _ := json.Marshal(v); return bz }
