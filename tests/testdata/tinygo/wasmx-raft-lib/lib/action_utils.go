package lib

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	blocks "github.com/loredanacirstea/wasmx-blocks/lib"
	consutils "github.com/loredanacirstea/wasmx-consensus-utils/lib"
	consensuswrap "github.com/loredanacirstea/wasmx-env-consensus/lib"
	typestnd "github.com/loredanacirstea/wasmx-env-consensus/lib"
	wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	stakinglib "github.com/loredanacirstea/wasmx-staking/lib"
)

// GetMajority: wrapper to storage's majority
func GetMajority(count int) int64 { return int64(count/2) + 1 }

// GetRandomInRange: deterministic-ish using block timestamp
func GetRandomInRange(min int64, max int64) (int64, error) {
	if max < min {
		return 0, errors.New("invalid range")
	}
	if min == max {
		return min, nil
	}
	ts := wasmx.GetTimestamp().UnixNano()
	span := max - min + 1
	if span <= 0 {
		return min, nil
	}
	return (ts%span + min), nil
}

// Wrappers with specific integer sizes
func GetRandomInRangeI64(min int64, max int64) (int64, error) { return GetRandomInRange(min, max) }
func GetRandomInRangeI32(min int32, max int32) (int32, error) {
	v, err := GetRandomInRange(int64(min), int64(max))
	return int32(v), err
}

// SignMessage signs string message with current state's ed25519 privkey (base64)
func SignMessage(msg string) (string, error) {
	st, err := GetCurrentState()
	if err != nil {
		return "", err
	}
	if len(st.ValidatorPrivkey) == 0 {
		return "", errors.New("empty validator privkey")
	}
	sig := wasmx.Ed25519Sign(st.ValidatorPrivkey, []byte(msg))
	return base64.StdEncoding.EncodeToString(sig), nil
}

// PrepareAppendEntryMessage builds the JSON payload to send over gRPC, base64-encoded
func PrepareAppendEntryMessage(nodeId int32, nextIndex int64, lastIndex int64, lastIndexToSend int64, node NodeInfo, data AppendEntry) (string, error) {
	datastrBz, err := json.Marshal(&data)
	if err != nil {
		return "", err
	}
	datastr := string(datastrBz)
	signature, err := SignMessage(datastr)
	if err != nil {
		return "", err
	}
	dataBase64 := base64.StdEncoding.EncodeToString([]byte(datastr))
	msgstr := fmt.Sprintf(`{"run":{"event":{"type":"receiveHeartbeat","params":[{"key":"entry","value":"%s"},{"key":"signature","value":"%s"}]}}}`, dataBase64, signature)
	LoggerDebug("diseminate append entry...", []string{"nodeId", Int32ToString(nodeId), "receiver", node.Address, "from", Int64ToString(nextIndex), "to", Int64ToString(lastIndexToSend), "last_index", Int64ToString(lastIndex)})
	return msgstr, nil
}

// SendGrpcJSONBase64 sends a base64-encoded JSON string via wasmxcore GrpcRequest
func SendGrpcJSONBase64(ip string, contract wasmx.Bech32String, jsonStr string) (*wasmxcore.GrpcResponse, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(jsonStr))
	return wasmxcore.GrpcRequest(ip, contract, encoded)
}

// InitChainSetup mirrors the AS init payload used by SetupNode
type InitChainSetup struct {
	ChainID          string                    `json:"chain_id"`
	Version          typestnd.Version          `json:"version"`
	AppHash          []byte                    `json:"app_hash"`
	LastResultsHash  []byte                    `json:"last_results_hash"`
	ValidatorAddress wasmx.HexString           `json:"validator_address"`
	ValidatorPrivkey []byte                    `json:"validator_privkey"`
	ValidatorPubkey  []byte                    `json:"validator_pubkey"`
	ConsensusParams  *typestnd.ConsensusParams `json:"consensus_params"`
	Peers            []string                  `json:"peers"`
	NodeIndex        int32                     `json:"node_index"`
}

// initChain initializes the chain current state and sets consensus params for next height
func InitChain(req InitChainSetup) error {
	LoggerDebug("start chain init", nil)
	empty := typestnd.BlockID{Hash: wasmx.HexString(""), Parts: typestnd.PartSetHeader{Total: 0, Hash: wasmx.HexString("")}}
	st := CurrentState{
		ChainID:          req.ChainID,
		Version:          req.Version,
		AppHash:          req.AppHash,
		LastBlockID:      empty,
		LastCommitHash:   []byte(""),
		LastResultsHash:  req.LastResultsHash,
		ValidatorAddress: req.ValidatorAddress,
		ValidatorPrivkey: req.ValidatorPrivkey,
		ValidatorPubkey:  req.ValidatorPubkey,
	}
	stbz, _ := json.Marshal(&st)
	LoggerDebug("set current state", []string{"state", string(stbz)})
	if err := SetCurrentState(st); err != nil {
		return err
	}
	// store consensus params for next block
	if err := setConsensusParams(LOG_START+1, req.ConsensusParams); err != nil {
		return err
	}
	LoggerDebug("current state set", nil)
	return nil
}

// Helpers to extract a param value by key
func GetParam(params []string, key string) (string, bool) {
	return "", false
}

// Parse string to int helpers
func ParseI64(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) }
func ParseI32(s string) (int32, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	return int32(v), err
}

// Node helpers
func getNodeByAddress(addr string, nodes []NodeInfo) (*NodeInfo, int) {
	for i := range nodes {
		if nodes[i].Address == addr {
			return &nodes[i], i
		}
	}
	return nil, -1
}

func getNodeIdByAddress(addr string, nodes []NodeInfo) int32 {
	_, idx := getNodeByAddress(addr, nodes)
	return int32(idx)
}

// parseNodeAddress parses `<address>@/ip4/<host>/tcp/<port>/p2p/<id>` into NodeInfo
type NodeInfoResponse struct {
	NodeInfo NodeInfo
	Error    string
}

func parseNodeAddress(peeraddr string) NodeInfoResponse {
	resp := NodeInfoResponse{}
	parts1 := strings.Split(peeraddr, "@")
	if len(parts1) != 2 {
		resp.Error = "invalid node format: missing @"
		return resp
	}
	addr := parts1[0]
	parts2 := strings.Split(parts1[1], "/")
	if len(parts2) < 7 {
		resp.Error = "invalid node format: missing ip4/tcp/p2p segments"
		return resp
	}
	host := parts2[2]
	port := parts2[4]
	p2pid := parts2[6]
	resp.NodeInfo = NodeInfo{Address: addr, Node: NetworkNode{ID: p2pid, Host: host, Port: port, IP: ""}, OutOfSync: false}
	return resp
}

// Hook calls
func callHookContract(hookName string, data string) error {
	return callHookContractInternal(wasmx.ROLE_HOOKS, hookName, data)
}

func callHookNonCContract(hookName string, data string) error {
	return callHookContractInternal(wasmx.ROLE_HOOKS_NONC, hookName, data)
}

func callHookContractInternal(contractRole string, hookName string, data string) error {
	dataBase64 := base64.StdEncoding.EncodeToString([]byte(data))
	payload := struct {
		RunHook struct {
			Hook string `json:"hook"`
			Data string `json:"data"`
		} `json:"RunHook"`
	}{}
	payload.RunHook.Hook = hookName
	payload.RunHook.Data = dataBase64
	bz, err := json.Marshal(&payload)
	if err != nil {
		return err
	}
	addr := wasmx.GetAddressByRole(contractRole)
	ok, out := wasmx.CallSimple(addr, bz, false, MODULE_NAME)
	if !ok {
		// do not fail, but log as AS does
		LoggerError("hooks failed", []string{"error", string(out)})
	}
	return nil
}

// Signature verification helpers
func verifyMessage(nodeIndex int32, signatureStr string, msg string) (bool, error) {
	nodes, err := GetNodeIPs()
	if err != nil {
		return false, err
	}
	if int(nodeIndex) < 0 || int(nodeIndex) >= len(nodes) {
		return false, nil
	}
	return VerifyMessageByAddr(nodes[nodeIndex].Address, signatureStr, msg)
}

func VerifyMessageByAddr(addr string, signatureStr string, msg string) (bool, error) {
	pubKey, err := getConsensusKeyByAddr(addr)
	if err != nil {
		return false, err
	}
	if pubKey == nil {
		return false, nil
	}
	sig, err := base64.StdEncoding.DecodeString(signatureStr)
	if err != nil {
		return false, err
	}
	return wasmx.Ed25519Verify(pubKey.GetKey().Key, sig, []byte(msg)), nil
}

func verifyMessageBytesByAddr(addr string, signatureStr string, msg []byte) (bool, error) {
	pubKey, err := getConsensusKeyByAddr(addr)
	if err != nil {
		return false, err
	}
	if pubKey == nil {
		return false, nil
	}
	sig, err := base64.StdEncoding.DecodeString(signatureStr)
	if err != nil {
		return false, err
	}
	return wasmx.Ed25519Verify(pubKey.GetKey().Key, sig, msg), nil
}

func getConsensusKeyByAddr(addr string) (*wasmx.PublicKey, error) {
	vals, err := GetAllValidators()
	if err != nil {
		return nil, err
	}
	for i := range vals {
		if string(vals[i].OperatorAddress) == addr {
			return vals[i].ConsensusPubkey, nil
		}
	}
	return nil, nil
}

// Storage contract calls (by role "storage")
func callStorage(calldata string, isQuery bool) (wasmx.CallResponse, error) {
	addr := wasmx.GetAddressByRole("storage")
	ok, data := wasmx.CallSimple(addr, []byte(calldata), isQuery, MODULE_NAME)
	resp := wasmx.CallResponse{Success: 0, Data: string(data)}
	if !ok {
		resp.Success = 1
	}
	return resp, nil
}

func setFinalizedBlock(blockData []byte, hash string, txhashes [][]byte, indexedTopics []blocks.IndexedTopic) error {
	payload := blocks.CallDataSetBlock{Value: blockData, Hash: hash, Txhashes: make([]string, len(txhashes)), IndexedTopics: indexedTopics}
	for i := range txhashes {
		payload.Txhashes[i] = base64.StdEncoding.EncodeToString(txhashes[i])
	}
	calld := blocks.CallData{SetBlock: &payload}
	bz, err := json.Marshal(&calld)
	if err != nil {
		return err
	}
	resp, err := callStorage(string(bz), false)
	if err != nil {
		return err
	}
	if resp.Success > 0 {
		return errors.New("could not set finalized block: " + resp.Data)
	}
	return nil
}

func getLastBlockIndex() (int64, error) {
	calld := blocks.CallData{GetLastBlockIndex: &blocks.CallDataGetLastBlockIndex{}}
	bz, err := json.Marshal(&calld)
	if err != nil {
		return 0, err
	}
	resp, err := callStorage(string(bz), true)
	if err != nil {
		return 0, err
	}
	if resp.Success > 0 {
		return 0, errors.New("could not get last block index")
	}
	var res blocks.LastBlockIndexResult
	if err := json.Unmarshal([]byte(resp.Data), &res); err != nil {
		return 0, err
	}
	return res.Index, nil
}

func getFinalBlock(index int64) (string, error) {
	calld := blocks.CallData{GetBlockByIndex: &blocks.CallDataGetBlockByIndex{Index: index}}
	bz, err := json.Marshal(&calld)
	if err != nil {
		return "", err
	}
	resp, err := callStorage(string(bz), true)
	if err != nil {
		return "", err
	}
	if resp.Success > 0 {
		return "", fmt.Errorf("could not get finalized block: %d", index)
	}
	return resp.Data, nil
}

func setConsensusParams(height int64, value *typestnd.ConsensusParams) error {
	var params []byte
	if value != nil {
		bz, err := json.Marshal(value)
		if err != nil {
			return err
		}
		params = bz
	}
	calld := blocks.CallData{SetConsensusParams: &blocks.CallDataSetConsensusParams{Height: height, Params: params}}
	bz, err := json.Marshal(&calld)
	if err != nil {
		return err
	}
	resp, err := callStorage(string(bz), false)
	if err != nil {
		return err
	}
	if resp.Success > 0 {
		return errors.New("could not set consensus params: " + resp.Data)
	}
	return nil
}

func getConsensusParams(height int64) (*typestnd.ConsensusParams, error) {
	calld := blocks.CallData{GetConsensusParams: &blocks.CallDataGetConsensusParams{Height: height}}
	bz, err := json.Marshal(&calld)
	if err != nil {
		return nil, err
	}
	resp, err := callStorage(string(bz), true)
	if err != nil {
		return nil, err
	}
	if resp.Success > 0 {
		return nil, errors.New("could not get consensus params: " + resp.Data)
	}
	if resp.Data == "" {
		return nil, nil
	}
	var params typestnd.ConsensusParams
	if err := json.Unmarshal([]byte(resp.Data), &params); err != nil {
		return nil, err
	}
	return &params, nil
}

// UpdateConsensusParams merges updates into existing params and stores them for height+1
func updateConsensusParams(height int64, updates *typestnd.ConsensusParams) error {
	if updates == nil {
		return setConsensusParams(height+1, nil)
	}
	params, err := getConsensusParams(height)
	if err != nil {
		return err
	}
	if params == nil {
		params = &typestnd.ConsensusParams{}
	}
	// selective updates mirroring AS logic
	if updates.ABCI.VoteExtensionsEnableHeight != 0 {
		params.ABCI.VoteExtensionsEnableHeight = updates.ABCI.VoteExtensionsEnableHeight
	}
	if updates.Block.MaxBytes != 0 {
		params.Block.MaxBytes = updates.Block.MaxBytes
	}
	if updates.Block.MaxGas != 0 {
		params.Block.MaxGas = updates.Block.MaxGas
	}
	if updates.Evidence.MaxAgeDuration != 0 {
		params.Evidence.MaxAgeDuration = updates.Evidence.MaxAgeDuration
	}
	if updates.Evidence.MaxAgeNumBlocks != 0 {
		params.Evidence.MaxAgeNumBlocks = updates.Evidence.MaxAgeNumBlocks
	}
	if updates.Evidence.MaxBytes != 0 {
		params.Evidence.MaxBytes = updates.Evidence.MaxBytes
	}
	if len(updates.Validator.PubKeyTypes) > 0 {
		params.Validator.PubKeyTypes = updates.Validator.PubKeyTypes
	}
	if updates.Version.App != 0 {
		params.Version.App = updates.Version.App
	}
	return setConsensusParams(height+1, params)
}

// Expose select consensus-utils helpers
func extractIndexedTopics(resp typestnd.ResponseFinalizeBlock, txhashes [][]byte) []blocks.IndexedTopic {
	topics := consutils.ExtractIndexedTopics(resp, txhashes)
	out := make([]blocks.IndexedTopic, len(topics))
	for i := range topics {
		out[i] = blocks.IndexedTopic{Topic: topics[i].Topic, Values: topics[i].Values}
	}
	return out
}

// verifyBlockProposalMeta verifies header hash equals request hash when header is available in metainfo
func verifyBlockProposalMeta(wrap typestnd.RequestProcessProposalWithMetaInfo) error {
	if wrap.Metainfo == nil {
		return nil
	}
	hbz, ok := wrap.Metainfo["header"]
	if !ok || len(hbz) == 0 {
		return nil
	}
	var header typestnd.Header
	if err := json.Unmarshal(hbz, &header); err != nil {
		return err
	}
	hhash, err := consensuswrap.HeaderHash(header)
	if err != nil {
		return err
	}
	if len(hhash) == 0 || len(wrap.Request.Hash) == 0 {
		return nil
	}
	if base64.StdEncoding.EncodeToString(hhash) != base64.StdEncoding.EncodeToString(wrap.Request.Hash) {
		return errors.New("header hash mismatch with proposal hash")
	}
	return nil
}

// verifyBlockProposal checks header.DataHash, LastCommitHash (if commit present), LastResultsHash, and AppHash with finalize response
// Name aligned with AssemblyScript (verifyBlockProposal)
func verifyBlockProposal(wrap typestnd.RequestProcessProposalWithMetaInfo, fin *typestnd.ResponseFinalizeBlock) error {
	if wrap.Metainfo == nil {
		return nil
	}
	hbz, ok := wrap.Metainfo["header"]
	if !ok || len(hbz) == 0 {
		return nil
	}
	var header typestnd.Header
	if err := json.Unmarshal(hbz, &header); err != nil {
		return err
	}
	// Data hash: txs merkle
	txsHash := consutils.GetTxsHash(wrap.Request.Txs)
	if !hexEqual(string(header.DataHash), txsHash) {
		return errors.New("data hash mismatch with txs merkle")
	}
	// Commit hash if provided in metainfo
	if cbz, ok := wrap.Metainfo["commit"]; ok && len(cbz) > 0 {
		var commit typestnd.BlockCommit
		if err := json.Unmarshal(cbz, &commit); err == nil {
			ch := consutils.GetCommitHash(commit)
			if !hexEqual(string(header.LastCommitHash), ch) {
				return errors.New("last commit hash mismatch")
			}
		}
	}
	// Validators hash using active set from staking (best effort)
	if vlist, err := GetAllValidators(); err == nil && len(vlist) > 0 {
		if tvals, err2 := consutils.GetActiveValidatorInfo(vlist); err2 == nil && len(tvals) > 0 {
			if vh, err3 := consensuswrap.ValidatorsHash(tvals); err3 == nil {
				if !hexEqual(string(header.ValidatorsHash), vh) {
					return errors.New("validators hash mismatch")
				}
			}
		}
	}
	// Results hash from finalize response
	if fin != nil {
		rh := consutils.GetResultsHash(fin.TxResults)
		if !hexEqual(string(header.LastResultsHash), rh) {
			return errors.New("results hash mismatch")
		}
		// App hash
		if len(fin.AppHash) > 0 && !hexEqual(string(header.AppHash), fin.AppHash) {
			return errors.New("app hash mismatch")
		}
	}
	return nil
}

func hexEqual(hexStr string, bz []byte) bool {
	if len(bz) == 0 {
		return hexStr == ""
	}
	enc := hex.EncodeToString(bz)
	return strings.EqualFold(hexStr, enc)
}

// Tendermint helpers
func getBlockID(hash []byte) typestnd.BlockID {
	hexhash := strings.ToUpper(hex.EncodeToString(hash))
	return typestnd.BlockID{Hash: wasmx.HexString(hexhash), Parts: typestnd.PartSetHeader{Total: 1, Hash: wasmx.HexString(hexhash)}}
}

// decodeTx decodes Cosmos tx from bytes to SignedTransaction using host
func decodeTx(tx []byte) (wasmx.SignedTransaction, error) {
	res := wasmx.DecodeCosmosTxFromBytes(tx)
	return res, nil
}

// buildBlockEntry composes a BlockEntry JSON matching AS semantics from finalize data and wrapped proposal
func buildBlockEntry(height int64, wrap typestnd.RequestProcessProposalWithMetaInfo, finResp *typestnd.ResponseFinalizeBlock, proposer string, validators *typestnd.TendermintValidators) ([]byte, error) {
	// header and commit may be present in metainfo
	var header []byte
	var commit []byte
	if wrap.Metainfo != nil {
		if v, ok := wrap.Metainfo["header"]; ok {
			header = v
		}
		if v, ok := wrap.Metainfo["commit"]; ok {
			commit = v
		}
	}
	// validator set
	var valbz []byte
	if validators != nil {
		vb, err := json.Marshal(validators)
		if err == nil {
			valbz = vb
		}
	}
	// data is the wrap JSON
	data, err := json.Marshal(&wrap)
	if err != nil {
		return nil, err
	}
	entry := blocks.BlockEntry{
		Index:           height,
		ReaderContract:  wasmx.GetAddress(),
		WriterContract:  wasmx.GetAddress(),
		Data:            data,
		Header:          header,
		ProposerAddress: wasmx.Bech32String(proposer),
		LastCommit:      commit,
		Evidence:        []byte(`{"evidence":[]}`),
		Result:          string(data),
		ValidatorInfo:   valbz,
	}
	return json.Marshal(&entry)
}

// doOptimisticExecution runs optimistic execution on the current proposal
func doOptimisticExecution(processReq typestnd.RequestProcessProposal, processResp typestnd.ResponseProcessProposal) (typestnd.ResponseOptimisticExecution, error) {
	return consensuswrap.OptimisticExecution(processReq, processResp)
}

// callContract helper (role/module aware wrapper)
func callContract(to wasmx.Bech32String, calldata string, isQuery bool, moduleName string) (wasmx.CallResponse, error) {
	ok, data := wasmx.CallSimple(to, []byte(calldata), isQuery, moduleName)
	resp := wasmx.CallResponse{Success: 0, Data: string(data)}
	if !ok {
		resp.Success = 1
	}
	return resp, nil
}

// callStaking wrapper expecting module role
func callStaking(calldata string, isQuery bool) (wasmx.CallResponse, error) {
	addr := wasmx.GetAddressByRole(wasmx.ROLE_STAKING)
	return callContract(addr, calldata, isQuery, MODULE_NAME)
}

// GetAllValidators queries staking module for validators. The exact query envelope may differ; best-effort parsing.
func GetAllValidators() ([]stakinglib.Validator, error) {
	payload := map[string]any{"GetAllValidators": map[string]any{}}
	bz, err := json.Marshal(&payload)
	if err != nil {
		return nil, err
	}
	resp, err := callStaking(string(bz), true)
	if err != nil {
		return nil, err
	}
	if resp.Success > 0 || resp.Data == "" {
		return []stakinglib.Validator{}, nil
	}
	var out struct {
		Validators []stakinglib.Validator `json:"validators"`
	}
	if err := json.Unmarshal([]byte(resp.Data), &out); err != nil {
		return []stakinglib.Validator{}, nil
	}
	return out.Validators, nil
}

func getCurrentValidator() typestnd.ValidatorInfo {
	st, _ := GetCurrentState()
	return typestnd.ValidatorInfo{Address: wasmx.HexString(st.ValidatorAddress), PubKey: st.ValidatorPubkey, VotingPower: 0, ProposerPriority: 0}
}

func checkValidatorsUpdate(validators []typestnd.ValidatorInfo, validatorInfo typestnd.ValidatorInfo, nodeId int32) error {
	if int(nodeId) >= len(validators) {
		return errors.New("validator index out of range")
	}
	if validators[nodeId].Address != validatorInfo.Address {
		return errors.New("register node response has wrong validator address")
	}
	// compare pub_key (base64 comparison by bytes)
	if base64.StdEncoding.EncodeToString(validators[nodeId].PubKey) != base64.StdEncoding.EncodeToString(validatorInfo.PubKey) {
		return errors.New("register node response has wrong validator pub_key")
	}
	return nil
}

// initializeIndexArrays sets NextIndex to last+1 and MatchIndex to LOG_START for len
func InitializeIndexArrays(lenNodes int) error {
	last, err := GetLastLogIndex()
	if err != nil {
		return err
	}
	nextIndex := make([]int64, lenNodes)
	matchIndex := make([]int64, lenNodes)
	for i := 0; i < lenNodes; i++ {
		nextIndex[i] = last + 1
		matchIndex[i] = LOG_START
	}
	if err := SetNextIndexArray(nextIndex); err != nil {
		return err
	}
	return SetMatchIndexArray(matchIndex)
}
