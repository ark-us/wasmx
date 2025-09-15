package lib

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	blocks "github.com/loredanacirstea/wasmx-blocks/lib"
	consutils "github.com/loredanacirstea/wasmx-consensus-utils/lib"
	consensuswrap "github.com/loredanacirstea/wasmx-env-consensus/lib"
	typestnd "github.com/loredanacirstea/wasmx-env-consensus/lib"
	wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
	p2p "github.com/loredanacirstea/wasmx-env-p2p/lib"
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
func PrepareAppendEntryMessage(nodeId int32, nextIndex int64, lastIndex int64, lastIndexToSend int64, node p2p.NodeInfo, data AppendEntry) (string, error) {
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
	LoggerDebug("diseminate append entry...", []string{"nodeId", Int32ToString(nodeId), "receiver", string(node.Address), "from", Int64ToString(nextIndex), "to", Int64ToString(lastIndexToSend), "last_index", Int64ToString(lastIndex)})
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
	empty := typestnd.BlockID{Hash: wasmx.HexString(hex.EncodeToString(req.AppHash)), Parts: typestnd.PartSetHeader{Total: 0, Hash: wasmx.HexString("")}}
	st := CurrentState{
		ChainID:          req.ChainID,
		Version:          req.Version,
		AppHash:          req.AppHash,
		LastBlockID:      empty,
		LastCommitHash:   []byte{},
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
func getNodeByAddress(addr wasmx.Bech32String, nodes []p2p.NodeInfo) (*p2p.NodeInfo, int) {
	for i := range nodes {
		if nodes[i].Address == addr {
			return &nodes[i], i
		}
	}
	return nil, -1
}

func getNodeIdByAddress(addr wasmx.Bech32String, nodes []p2p.NodeInfo) int32 {
	_, idx := getNodeByAddress(addr, nodes)
	return int32(idx)
}

// parseNodeAddress parses `<address>@/ip4/<host>/tcp/<port>/p2p/<id>` into NodeInfo
func parseNodeAddress(peeraddr string) p2p.NodeInfoResponse {
	resp := p2p.NodeInfoResponse{}
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
	resp.NodeInfo = &p2p.NodeInfo{Address: wasmx.Bech32String(addr), Node: p2p.NetworkNode{ID: p2pid, Host: host, Port: port, IP: ""}, OutOfSync: false}
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

func VerifyMessageByAddr(addr wasmx.Bech32String, signatureStr string, msg string) (bool, error) {
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

func verifyMessageBytesByAddr(addr wasmx.Bech32String, signatureStr string, msg []byte) (bool, error) {
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

func getConsensusKeyByAddr(addr wasmx.Bech32String) (*wasmx.PublicKey, error) {
	vals, err := GetAllValidators()
	if err != nil {
		return nil, err
	}
	for i := range vals {
		if vals[i].OperatorAddress == addr {
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

func setFinalizedBlock(blockData string, hash string, txhashes [][]byte, indexedTopics []blocks.IndexedTopic) error {
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

func GetLastBlockIndex() (int64, error) {
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
	// AS: Base64 encode JSON string (lines 184-187)
	params := []byte{}
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

func GetConsensusParams(height int64) (*typestnd.ConsensusParams, error) {
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
		// AS: Return default params when empty (lines 202-209)
		return &typestnd.ConsensusParams{
			Block:     typestnd.BlockParams{MaxBytes: 0, MaxGas: 0},
			Evidence:  typestnd.EvidenceParams{MaxAgeDuration: 0, MaxAgeNumBlocks: 0, MaxBytes: 0},
			Validator: typestnd.ValidatorParams{PubKeyTypes: []string{}},
			Version:   typestnd.VersionParams{App: 0},
			ABCI:      typestnd.ABCIParams{VoteExtensionsEnableHeight: 0},
		}, nil
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
	params, err := GetConsensusParams(height)
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

// https://github.com/cometbft/cometbft/blob/f4a803f14a2f5bc5c17d75fcd1131b9249bba133/state/validation.go
func verifyBlockProposal(data blocks.BlockEntry, processReq typestnd.RequestProcessProposal) error {
	// TODO? verify:
	// processReq.next_validators_hash
	// processReq.proposed_last_commit

	if len(data.Header) == 0 {
		return errors.New("header is empty")
	}

	var header typestnd.Header
	if err := json.Unmarshal(data.Header, &header); err != nil {
		return fmt.Errorf("failed to parse header: %v", err)
	}

	hash, err := consutils.GetHeaderHash(header)
	if err != nil {
		return fmt.Errorf("failed to compute header hash: %v", err)
	}

	reqHashHex := strings.ToUpper(hex.EncodeToString(processReq.Hash))
	headerHashHex := strings.ToUpper(hex.EncodeToString(hash))
	if headerHashHex != reqHashHex {
		return fmt.Errorf("header hash mismatch: expected %s, got %s", reqHashHex, headerHashHex)
	}

	currentState, err := GetCurrentState()
	if err != nil {
		return fmt.Errorf("failed to get current state: %v", err)
	}

	appHashHex := strings.ToUpper(hex.EncodeToString(currentState.AppHash))
	if string(header.AppHash) != appHashHex {
		return fmt.Errorf("header app_hash mismatch: expected %s, got %s", appHashHex, header.AppHash)
	}

	if header.Version.Block != typestnd.BlockProtocol {
		return fmt.Errorf("header version.block mismatch: expected %d, got %d", typestnd.BlockProtocol, header.Version.Block)
	}

	if header.Version.App != currentState.Version.Consensus.App {
		return fmt.Errorf("header version.app mismatch: expected %d, got %d", currentState.Version.Consensus.App, header.Version.App)
	}

	if header.ChainID != currentState.ChainID {
		return fmt.Errorf("header chain_id mismatch: expected %s, got %s", currentState.ChainID, header.ChainID)
	}

	// if (header.height != currentState.nextHeight) return `header height mismatch: expected ${currentState.nextHeight}, got ${header.height}`

	lastResultsHashHex := strings.ToUpper(hex.EncodeToString(currentState.LastResultsHash))
	if string(header.LastResultsHash) != lastResultsHashHex {
		return fmt.Errorf("header last_results_hash mismatch: expected %s, got %s", lastResultsHashHex, header.LastResultsHash)
	}

	if header.LastBlockID.Hash != currentState.LastBlockID.Hash {
		return fmt.Errorf("header last_block_id.hash mismatch: expected %s, got %s", currentState.LastBlockID.Hash, header.LastBlockID.Hash)
	}

	if header.LastBlockID.Parts.Hash != currentState.LastBlockID.Parts.Hash {
		return fmt.Errorf("header last_block_id.parts.hash mismatch: expected %s, got %s", currentState.LastBlockID.Parts.Hash, header.LastBlockID.Parts.Hash)
	}

	if header.LastBlockID.Parts.Total != currentState.LastBlockID.Parts.Total {
		return fmt.Errorf("header last_block_id.parts.total mismatch: expected %d, got %d", currentState.LastBlockID.Parts.Total, header.LastBlockID.Parts.Total)
	}

	txsHash := consutils.GetTxsHash(processReq.Txs)
	dataHashHex := strings.ToUpper(hex.EncodeToString(txsHash))
	if string(header.DataHash) != dataHashHex {
		return fmt.Errorf("header data_hash mismatch: expected %s, got %s", dataHashHex, header.DataHash)
	}

	cparams, err := GetConsensusParams(0)
	if err != nil {
		return fmt.Errorf("failed to get consensus params: %v", err)
	}
	if cparams == nil {
		return errors.New("consensus params is nil")
	}

	consensusHash, err := consutils.GetConsensusParamsHash(*cparams)
	if err != nil {
		return fmt.Errorf("failed to compute consensus params hash: %v", err)
	}
	consensusHashHex := strings.ToUpper(hex.EncodeToString(consensusHash))
	if string(header.ConsensusHash) != consensusHashHex {
		return fmt.Errorf("header consensus_hash mismatch: expected %s, got %s", consensusHashHex, header.ConsensusHash)
	}

	// TODO see other time constraints that fit our protocol
	// if (Date.fromString(header.time).getTime() <= Date.fromString(currentState.last_time).getTime()) {
	//     return `header time mismatch: expected higher than ${currentState.last_time}, got ${header.time}`
	// }
	// TODO set an upper time bound

	// TODO
	// header.last_commit_hash
	// header.next_validators_hash
	// header.validators_hash
	// header.evidence_hash
	// TODO validate commit format
	return nil
}

func hexEqual(hexStr wasmx.HexString, bz []byte) bool {
	if len(bz) == 0 {
		return hexStr == ""
	}
	enc := hex.EncodeToString(bz)
	return strings.EqualFold(string(hexStr), enc)
}

// DecodeTx decodes Cosmos tx from bytes to SignedTransaction using host
func DecodeTx(tx []byte) (wasmx.SignedTransaction, error) {
	res := wasmx.DecodeCosmosTxFromBytes(tx)
	return res, nil
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

// updateValidators calls staking module to update validators. Matches AS implementation lines 212-219.
func updateValidators(updates []typestnd.ValidatorUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	payload := map[string]any{
		"UpdateValidators": map[string]any{
			"updates": updates,
		},
	}

	bz, err := json.Marshal(&payload)
	if err != nil {
		return err
	}

	resp, err := callStaking(string(bz), true)
	if err != nil {
		return err
	}

	if resp.Success > 0 {
		return fmt.Errorf("could not update validators")
	}

	return nil
}

func getCurrentValidator() typestnd.ValidatorInfo {
	st, _ := GetCurrentState()
	return typestnd.ValidatorInfo{Address: wasmx.HexString(st.ValidatorAddress), PubKey: st.ValidatorPubkey, VotingPower: 0, ProposerPriority: 0}
}

func checkValidatorsUpdate(validators []typestnd.ValidatorInfo, validatorInfo typestnd.ValidatorInfo, nodeId int32) error {
	// Add comprehensive bounds checking for nodeId
	if nodeId < 0 || int(nodeId) >= len(validators) {
		return fmt.Errorf("validator index out of range: nodeId=%d, validators length=%d", nodeId, len(validators))
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

// getValidatorByHexAddr queries the staking module for a validator by hex address
func getValidatorByHexAddr(addr wasmx.HexString) (stakinglib.Validator, error) {
	payload := map[string]any{"ValidatorByHexAddr": map[string]any{"validator_addr": string(addr)}}
	bz, err := json.Marshal(&payload)
	if err != nil {
		return stakinglib.Validator{}, err
	}
	resp, err := callStaking(string(bz), true)
	if err != nil {
		return stakinglib.Validator{}, err
	}
	if resp.Success > 0 {
		return stakinglib.Validator{}, errors.New(resp.Data)
	}
	if resp.Data == "" {
		return stakinglib.Validator{}, fmt.Errorf("validator not found: %s", addr)
	}
	LoggerDebug("ValidatorByHexAddr", []string{"addr", string(addr), "data", resp.Data})
	var result struct {
		Validator stakinglib.Validator `json:"validator"`
	}
	if err := json.Unmarshal([]byte(resp.Data), &result); err != nil {
		return stakinglib.Validator{}, err
	}
	return result.Validator, nil
}

// getCommitSigsFromLeaderSignature creates a single commit signature for the Leader in RAFT consensus
// This is the RAFT equivalent of Tendermint's getCommitSigsFromPrecommitArray
func getCommitSigsFromLeaderSignature(blockCommit typestnd.BlockCommit, validatorInfos []typestnd.TendermintValidator) ([]typestnd.CommitSig, error) {
	if len(validatorInfos) == 0 {
		return []typestnd.CommitSig{}, nil
	}

	// Get current state to access Leader's private key
	st, err := GetCurrentState()
	if err != nil {
		return nil, fmt.Errorf("failed to get current state: %v", err)
	}

	// Get current Leader node ID
	leaderId, err := GetCurrentNodeId()
	if err != nil {
		return nil, fmt.Errorf("failed to get current node ID: %v", err)
	}

	// Find the Leader's validator info
	var leaderValidatorInfo *typestnd.TendermintValidator
	leaderValidatorIndex := int(leaderId)
	if leaderValidatorIndex >= 0 && leaderValidatorIndex < len(validatorInfos) {
		leaderValidatorInfo = &validatorInfos[leaderValidatorIndex]
	} else {
		// Try to find by address if index doesn't work
		for i := range validatorInfos {
			if validatorInfos[i].HexAddress == st.ValidatorAddress {
				leaderValidatorInfo = &validatorInfos[i]
				break
			}
		}
	}

	if leaderValidatorInfo == nil {
		return nil, fmt.Errorf("leader validator not found in validator set")
	}
	hash, err := hex.DecodeString(string(blockCommit.BlockID.Hash))
	if err != nil {
		return nil, fmt.Errorf("cannot decode blockID hash %s", blockCommit.BlockID.Hash)
	}
	validAddr, err := hex.DecodeString(string(st.ValidatorAddress))
	if err != nil {
		return nil, fmt.Errorf("cannot decode validator address %s", blockCommit.BlockID.Hash)
	}

	// Create vote data for signing (similar to Tendermint precommit)
	vote := typestnd.VoteTendermint{
		Type:             typestnd.SIGNED_MSG_TYPE_PRECOMMIT,
		Height:           blockCommit.Height,
		Round:            blockCommit.Round,
		BlockID:          GetBlockIDProto(hash),
		Timestamp:        time.Now().UTC().Format(time.RFC3339Nano),
		ValidatorAddress: validAddr,
		ValidatorIndex:   int32(leaderValidatorIndex),
	}

	// Get the canonical vote bytes for signing
	voteBytes, err := consensuswrap.BlockCommitVoteBytes(vote)
	if err != nil {
		return nil, fmt.Errorf("failed to get vote bytes: %v", err)
	}

	// Sign the vote bytes with Leader's private key
	signature := wasmx.Ed25519Sign(st.ValidatorPrivkey, voteBytes)

	// Get validator's consensus public key hex
	validator, err := getValidatorByHexAddr(wasmx.HexString(st.ValidatorAddress))
	if err != nil {
		return nil, fmt.Errorf("failed to get validator: %v", err)
	}

	if validator.ConsensusPubkey == nil {
		return nil, fmt.Errorf("validator missing consensus public key")
	}

	consKey := wasmx.Ed25519PubToHex(validator.ConsensusPubkey.GetKey().Key)

	// Create commit signature for the Leader
	commitSig := typestnd.CommitSig{
		BlockIDFlag:      typestnd.BlockIDFlagCommit,
		ValidatorAddress: wasmx.HexString(hex.EncodeToString(consKey)),
		Timestamp:        time.Now().UTC().Format(time.RFC3339Nano),
		Signature:        signature,
	}

	// Return array with single signature (RAFT has only Leader)
	return []typestnd.CommitSig{commitSig}, nil
}

// appendLogInternalVerified creates a BlockEntry and LogEntryAggregate, then appends it to the log
func appendLogInternalVerified(processReq typestnd.RequestProcessProposal, header typestnd.Header, blockCommit typestnd.BlockCommit, optimisticExecution bool, meta map[string][]byte, validatorSet typestnd.TendermintValidators) error {
	// AS: Create RequestProcessProposalWithMetaInfo with meta as-is (lines 1476-1477)
	if meta == nil {
		meta = make(map[string][]byte)
	}

	// Create RequestProcessProposalWithMetaInfo - matches AS line 1477
	wrap := typestnd.RequestProcessProposalWithMetaInfo{
		Request:             processReq,
		OptimisticExecution: optimisticExecution,
		Metainfo:            meta,
	}

	blockData, err := json.Marshal(&wrap)
	if err != nil {
		return fmt.Errorf("failed to marshal block data: %v", err)
	}

	blockHeader, err := json.Marshal(&header)
	if err != nil {
		return fmt.Errorf("failed to marshal block header: %v", err)
	}

	commit, err := json.Marshal(&blockCommit)
	if err != nil {
		return fmt.Errorf("failed to marshal block commit: %v", err)
	}

	termId, err := GetTermId()
	if err != nil {
		return fmt.Errorf("failed to get term ID: %v", err)
	}

	leaderId, err := GetCurrentNodeId()
	if err != nil {
		return fmt.Errorf("failed to get current node ID: %v", err)
	}

	validator, err := getValidatorByHexAddr(wasmx.HexString(processReq.ProposerAddress))
	if err != nil {
		return fmt.Errorf("failed to get validator: %v", err)
	}

	contractAddress := wasmx.GetAddressBz()

	validatorSetBytes, err := json.Marshal(&validatorSet)
	if err != nil {
		return fmt.Errorf("failed to marshal validator set: %v", err)
	}

	// Create BlockEntry
	blockEntry := blocks.BlockEntry{
		Index:           processReq.Height,
		ReaderContract:  contractAddress,
		WriterContract:  contractAddress,
		Data:            blockData,
		Header:          blockHeader,
		ProposerAddress: validator.OperatorAddress,
		LastCommit:      commit,
		Evidence:        []byte(`{"evidence":[]}`),
		Result:          []byte{},
		ValidatorInfo:   validatorSetBytes,
	}

	// Create LogEntryAggregate
	entry := LogEntryAggregate{
		Index:    processReq.Height,
		TermID:   termId,
		LeaderID: leaderId,
		Data:     blockEntry,
	}

	return AppendLogEntry(entry)
}

// getSelfNodeInfo returns the NodeInfo for the current node
func getSelfNodeInfo() (p2p.NodeInfo, error) {
	nodeIps, err := GetNodeIPs()
	if err != nil {
		return p2p.NodeInfo{}, fmt.Errorf("failed to get node IPs: %v", err)
	}

	ourId, err := GetCurrentNodeId()
	if err != nil {
		return p2p.NodeInfo{}, fmt.Errorf("failed to get current node ID: %v", err)
	}

	if len(nodeIps) <= int(ourId) {
		return p2p.NodeInfo{}, fmt.Errorf("index out of range: nodes count %d, our node id is %d", len(nodeIps), ourId)
	}

	return nodeIps[ourId], nil
}

// prepareAppendEntry prepares AppendEntry data structure for sending to followers
func prepareAppendEntry(nodeIps []p2p.NodeInfo, nextIndex int64, lastIndex int64) (AppendEntry, error) {
	entries := make([]LogEntryAggregate, 0)
	for i := nextIndex; i <= lastIndex; i++ {
		entry, err := GetLogEntryAggregate(i)
		if err != nil {
			return AppendEntry{}, fmt.Errorf("failed to get log entry at index %d: %v", i, err)
		}
		if entry != nil {
			entries = append(entries, *entry)
		}
	}

	previousEntry, err := GetLogEntryObj(nextIndex - 1)
	if err != nil {
		return AppendEntry{}, fmt.Errorf("failed to get previous log entry at index %d: %v", nextIndex-1, err)
	}

	lastCommitIndex, err := GetLastBlockIndex()
	if err != nil {
		return AppendEntry{}, fmt.Errorf("failed to get commit index: %v", err)
	}

	termId, err := GetTermId()
	if err != nil {
		return AppendEntry{}, fmt.Errorf("failed to get term ID: %v", err)
	}

	leaderId, err := GetCurrentNodeId()
	if err != nil {
		return AppendEntry{}, fmt.Errorf("failed to get current node ID: %v", err)
	}

	data := AppendEntry{
		TermID:       termId,
		LeaderID:     leaderId,
		PrevLogIndex: nextIndex - 1,
		PrevLogTerm:  previousEntry.TermID,
		Entries:      entries,
		LeaderCommit: lastCommitIndex,
		NodeIPs:      nodeIps,
	}

	return data, nil
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

func GetBlockID(hash []byte) typestnd.BlockID {
	hashhex := wasmx.HexString(hex.EncodeToString(hash))
	return typestnd.BlockID{
		Hash: hashhex,
		Parts: typestnd.PartSetHeader{
			Total: 1,
			Hash:  hashhex,
		},
	}
}

func GetBlockIDProto(hash []byte) typestnd.BlockIDProto {
	return typestnd.BlockIDProto{
		Hash: hash,
		PartSetHeader: typestnd.PartSetHeaderProto{
			Total: 1,
			Hash:  hash,
		},
	}
}

// getTendermintVote converts ValidatorProposalVote to VoteTendermint
func getTendermintVote(data ValidatorProposalVote) (typestnd.VoteTendermint, error) {
	hash := data.Hash
	// cometbft expects hash: []byte => nil is []byte{}
	hashBytes := []byte{}
	if string(hash) != "nil" && string(hash) != "" {
		var err error
		hashBytes, err = base64.StdEncoding.DecodeString(string(hash))
		if err != nil {
			return typestnd.VoteTendermint{}, fmt.Errorf("failed to decode hash: %v", err)
		}
	}

	// Convert validator address to bytes
	validatorAddr, err := hex.DecodeString(string(data.ValidatorAddress))
	if err != nil {
		return typestnd.VoteTendermint{}, fmt.Errorf("failed to decode validator address: %v", err)
	}

	return typestnd.VoteTendermint{
		Type:             data.Type,
		Height:           data.Index,
		Round:            data.TermId,
		BlockID:          GetBlockIDProto(hashBytes),
		Timestamp:        data.Timestamp.Format(time.RFC3339Nano),
		ValidatorAddress: validatorAddr,
		ValidatorIndex:   data.ValidatorIndex,
	}, nil
}

// buildPrecommitMessage creates a ValidatorProposalVote for precommit
func buildPrecommitMessage() (ValidatorProposalVote, error) {
	// Get current node ID
	ourId, err := GetCurrentNodeId()
	if err != nil {
		return ValidatorProposalVote{}, err
	}

	// Get current state
	state, err := GetCurrentState()
	if err != nil {
		return ValidatorProposalVote{}, err
	}

	// Get term ID
	termId, err := GetTermId()
	if err != nil {
		return ValidatorProposalVote{}, err
	}

	// Get validator address from state
	validatorAddr := state.ValidatorAddress

	// Create the precommit vote
	vote := ValidatorProposalVote{
		Type:             typestnd.SIGNED_MSG_TYPE_PRECOMMIT,
		TermId:           int64(termId),
		ValidatorAddress: wasmx.Bech32String(validatorAddr),
		ValidatorIndex:   ourId,
		Index:            state.NextHeight,
		Hash:             state.NextHash,
		Timestamp:        time.Now().UTC(),
		ChainId:          state.ChainID,
	}

	return vote, nil
}

// preparePrecommitMessage creates a message and signature for precommit
func preparePrecommitMessage(data ValidatorProposalVote) ([]byte, error) {
	// Convert to JSON
	// dataBytes, err := json.Marshal(data)
	// if err != nil {
	// 	return "", "", fmt.Errorf("failed to marshal ValidatorProposalVote: %v", err)
	// }
	// dataStr := string(dataBytes)

	// Get the tendermint vote for signing
	commit, err := getTendermintVote(data)
	if err != nil {
		return nil, fmt.Errorf("failed to get tendermint vote: %v", err)
	}

	// Get vote bytes for signing
	voteBytes, err := consensuswrap.BlockCommitVoteBytes(commit)
	if err != nil {
		return nil, fmt.Errorf("failed to get vote bytes: %v", err)
	}

	// Get current state to access private key
	state, err := GetCurrentState()
	if err != nil {
		return nil, fmt.Errorf("failed to get current state: %v", err)
	}

	// Sign the vote bytes
	signature := wasmx.Ed25519Sign(state.ValidatorPrivkey, voteBytes)
	// signatureB64 := base64.StdEncoding.EncodeToString(signature)

	// Create the message
	// dataBase64 := base64.StdEncoding.EncodeToString([]byte(dataStr))
	// msgstr := fmt.Sprintf(`{"run":{"event":{"type":"receivePrecommit","params":[{"key": "entry","value":"%s"},{"key": "signature","value":"%s"}]}}}`, dataBase64, signatureB64)

	// return msgstr, signatureB64, nil
	return signature, nil
}

func BuildValidatorCommitVote() (*ValidatorCommitVote, error) {
	// get the current proposal & vote on the block hash
	data, err := buildPrecommitMessage()
	if err != nil {
		return nil, err
	}
	signature, err := preparePrecommitMessage(data)
	if err != nil {
		return nil, err
	}
	LoggerDebug("sending precommit", []string{"index", Int64ToString(data.Index), "hash", hex.EncodeToString(data.Hash), "term_id", Int64ToString(data.TermId)})

	return &ValidatorCommitVote{
		Vote:        data,
		BlockIdFlag: typestnd.BlockIDFlagCommit,
		Signature:   signature,
	}, nil
}

// getLastBlockCommit creates a BlockCommit from the current state
// This mirrors the Tendermint implementation and provides last block commit information
func getLastBlockCommit(state CurrentState) typestnd.BlockCommit {
	return typestnd.BlockCommit{
		Height:     state.NextHeight - 1,
		Round:      state.LastRound,
		BlockID:    state.LastBlockID,
		Signatures: state.LastBlockSigs,
	}
}

func getCommitSigsFromPrecommitArray(st CurrentState, height int64, blockhash []byte, termId int64) ([]typestnd.CommitSig, error) {
	// Record round and construct commit signatures for this finalized block.
	// RAFT has a single leader signing; followers are Absent.
	// These signatures will be included as LastCommit in the next block proposal.
	// Compute signatures array sized to current active validator set order.
	validators, err := GetAllValidators()
	if err != nil {
		return nil, err
	}
	activeInfos, err := consutils.GetActiveValidatorInfo(validators)
	if err != nil {
		return nil, err
	}
	validatorInfos := consutils.SortTendermintValidators(activeInfos)
	// default all as Absent
	sigs := make([]typestnd.CommitSig, len(validatorInfos))
	leaderIdx := -1
	var t time.Time
	for i := range validatorInfos {
		addrHex := validatorInfos[i].HexAddress
		sigs[i] = typestnd.CommitSig{
			BlockIDFlag:      typestnd.BlockIDFlagAbsent,
			ValidatorAddress: addrHex,
			Timestamp:        t.Format(time.RFC3339),
			Signature:        []byte{},
		}
		if strings.EqualFold(string(addrHex), string(st.ValidatorAddress)) {
			leaderIdx = i
		}
	}

	// Build leader precommit if we are in the active set and have a privkey
	if leaderIdx >= 0 && len(st.ValidatorPrivkey) > 0 {
		// Build BlockIDProto from the finalized block hash bytes (must be 32 bytes)
		bidp := GetBlockIDProto(blockhash)
		timestamp := time.Now().UTC().Format(time.RFC3339Nano)
		vote := typestnd.VoteTendermint{
			Type:             2, // SIGNED_MSG_TYPE_PRECOMMIT
			Height:           height,
			Round:            termId,
			BlockID:          bidp,
			Timestamp:        timestamp,
			ValidatorAddress: []byte{},
			ValidatorIndex:   int32(leaderIdx),
		}
		if voteBytes, err4 := consensuswrap.BlockCommitVoteBytes(vote); err4 == nil {
			leaderSig := wasmx.Ed25519Sign(st.ValidatorPrivkey, voteBytes)
			// Set leader commit signature
			sigs[leaderIdx] = typestnd.CommitSig{BlockIDFlag: typestnd.BlockIDFlagCommit, ValidatorAddress: st.ValidatorAddress, Timestamp: timestamp, Signature: leaderSig}
		}
	}
	return sigs, nil
}

// storageBootstrapAfterStateSync updates storage contract after state sync
func storageBootstrapAfterStateSync(height int64, lastHeightChanged int64, consensusParams typestnd.ConsensusParams) error {
	// Marshal consensus params
	paramsBytes, err := json.Marshal(consensusParams)
	if err != nil {
		return fmt.Errorf("failed to marshal consensus params: %v", err)
	}
	paramsBase64 := base64.StdEncoding.EncodeToString(paramsBytes)

	// Create calldata for bootstrapAfterStateSync
	calldata := fmt.Sprintf(`{"bootstrapAfterStateSync":{"last_block_height":%d,"last_height_changed":%d,"params":"%s"}}`,
		height, lastHeightChanged, paramsBase64)

	// Call storage contract
	resp, err := callStorage(calldata, false)
	if err != nil {
		return fmt.Errorf("could not bootstrap storage err: %s", err.Error())
	}
	if resp.Success > 0 {
		return fmt.Errorf("could not bootstrap storage: %s", resp.Data)
	}

	LoggerDebug("storage bootstrap after state sync completed", []string{
		"height", Int64ToString(height),
		"last_height_changed", Int64ToString(lastHeightChanged),
	})

	return nil
}

func WeAreNotAlone(state CurrentState) bool {
	nodes, err := GetNodeIPs()
	if err != nil {
		return false
	}
	return WeAreNotAloneInternal(nodes, state)
}

func WeAreNotAloneInternal(nodes []p2p.NodeInfo, state CurrentState) bool {
	if len(nodes) > 1 {
		return true
	}
	return state.WeAreNotAlone
}
