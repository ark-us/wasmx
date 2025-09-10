package lib

import (
	"encoding/base64"

	typestnd "github.com/loredanacirstea/wasmx-env-consensus/lib"
	crosschain "github.com/loredanacirstea/wasmx-env-crosschain/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// CurrentState mirrors the AS type; JSON tags match field names.
type CurrentState struct {
	ChainID             string                `json:"chain_id"`
	UniqueP2PID         string                `json:"unique_p2p_id"`
	Version             typestnd.Version      `json:"version"`
	AppHash             []byte                `json:"app_hash"`
	LastBlockID         typestnd.BlockID      `json:"last_block_id"`
	LastCommitHash      []byte                `json:"last_commit_hash"`
	LastResultsHash     []byte                `json:"last_results_hash"`
	LastRound           int64                 `json:"last_round"`
	LastBlockSigs       []typestnd.CommitSig  `json:"last_block_signatures"`
	LastTime            string                `json:"last_time"`
	ValidatorAddress    wasmx.HexString       `json:"validator_address"`
	ValidatorPrivkey    []byte                `json:"validator_privkey"`
	ValidatorPubkey     []byte                `json:"validator_pubkey"`
	NextHeight          int64                 `json:"nextHeight"`
	NextHash            []byte                `json:"nextHash"`
	LockedValue         int64                 `json:"lockedValue"`
	LockedRound         int64                 `json:"lockedRound"`
	ValidValue          int64                 `json:"validValue"`
	ValidRound          int64                 `json:"validRound"`
	ProposerQueue       []ValidatorQueueEntry `json:"proposerQueue"`
	ProposerQueueTermId int64                 `json:"proposerQueueTermId"`
	ProposerIndex       int32                 `json:"proposerIndex"`
	WeAreNotAlone       bool                  `json:"wearenotalone"`
}

type ValidatorQueueEntry struct {
	Address string `json:"address"`
	Index   int32  `json:"index"`
	Value   string `json:"value"` // BigInt as string placeholder
}

type GetProposerResponse struct {
	ProposerQueue []ValidatorQueueEntry `json:"proposerQueue"`
	ProposerIndex int32                 `json:"proposerIndex"`
}

// MempoolBatch mirrors the batching metadata.
type MempoolBatch struct {
	Txs           [][]byte `json:"txs"`
	CummulatedGas int64    `json:"cummulatedGas"`
	IsAtomicTx    bool     `json:"isAtomicTx"`
	IsLeader      bool     `json:"isLeader"`
	Full          bool     `json:"full"`
}

// MempoolTx mirrors AS with leader chain info.
type MempoolTx struct {
	Tx     []byte `json:"tx"`
	Gas    uint64 `json:"gas"`
	Leader string `json:"leader"`
}

// Mempool with improved tracking (tempsize, tobesent)
type Mempool struct {
	TempSize int32                `json:"tempsize"`
	Map      map[string]MempoolTx `json:"map"`
	Temp     map[string]bool      `json:"temp"`
	ToBeSent map[string]bool      `json:"tobesent"`
	Order    []string             `json:"order"`
}

func NewMempool(m map[string]MempoolTx, tempsize int32) Mempool {
	if m == nil {
		m = map[string]MempoolTx{}
	}
	return Mempool{TempSize: tempsize, Map: m, Temp: map[string]bool{}, ToBeSent: map[string]bool{}, Order: []string{}}
}

func (m *Mempool) HasSeen(txhash string) bool {
	if m == nil {
		return false
	}
	if m.Temp[txhash] {
		return true
	}
	_, ok := m.Map[txhash]
	return ok
}

func (m *Mempool) Seen(txhash string) {
	if m == nil {
		return
	}
	if m.Temp == nil {
		m.Temp = map[string]bool{}
	}
	m.Temp[txhash] = true
}

func (m *Mempool) Add(txhash string, tx []byte, gas uint64, leaderChainID string) {
	if m == nil {
		return
	}
	if m.Map == nil {
		m.Map = map[string]MempoolTx{}
	}
	if m.ToBeSent == nil {
		m.ToBeSent = map[string]bool{}
	}
	m.Map[txhash] = MempoolTx{Tx: tx, Gas: gas, Leader: leaderChainID}
	m.ToBeSent[txhash] = true
	m.Order = append(m.Order, txhash)
}

func (m *Mempool) Count() int {
	return len(m.Order)
}

func (m *Mempool) Remove(txhash string) {
	if m == nil {
		return
	}
	delete(m.Map, txhash)
	m.DropSeen()
	delete(m.ToBeSent, txhash)
	ndx := 0
	for i, h := range m.Order {
		if h == txhash {
			ndx = i
			break
		}
	}
	m.Order = m.Order[ndx+1:]
}

func (m *Mempool) MustBeSent(txhash string) bool {
	if m == nil {
		return false
	}
	if !m.ToBeSent[txhash] {
		return false
	}
	delete(m.ToBeSent, txhash)
	return true
}

func (m *Mempool) ClearMustBeSent() {
	if m != nil {
		m.ToBeSent = map[string]bool{}
	}
}

func (m *Mempool) DropSeen() {
	if m == nil {
		return
	}
	if m.Temp == nil {
		m.Temp = map[string]bool{}
	}
	// clear out temporary txhashes too if they get > TempSize
	extra := int32(len(m.Temp)) - m.TempSize
	if extra < 1 {
		return
	}
	// best effort: delete arbitrary keys until size fits
	for k := range m.Temp {
		delete(m.Temp, k)
		extra--
		if extra <= 0 {
			break
		}
	}
}

// base64Len: 4 chars represent 3 bytes
func base64Len(value []byte) int32 {
	return int32(base64.StdEncoding.EncodedLen(len(value)))
}

// Batch selects a set of txs for the proposal. Returns errors from crosschain check.
func (m *Mempool) Batch(maxGas int64, maxBytes int64, ourchain string) (MempoolBatch, error) {
	batch := MempoolBatch{Txs: [][]byte{}, CummulatedGas: 0, IsAtomicTx: false, IsLeader: false, Full: false}
	if m == nil {
		return batch, nil
	}
	var cummulatedBytes int64 = 0
	for _, txhash := range m.Order {
		tx, ok := m.Map[txhash]
		if !ok {
			continue
		}
		atomicInExec := false
		if tx.Leader != "" {
			if tx.Leader == ourchain {
				atomicInExec = true
				batch.IsLeader = true
			} else {
				ok, err := crosschain.IsAtomicTxInExecution(crosschain.MsgIsAtomicTxInExecutionRequest{SubChainId: tx.Leader, TxHash: []byte(txhash)})
				if err != nil {
					return batch, err
				}
				if ok {
					atomicInExec = true
				} else {
					LoggerDebug("atomic tx is not in execution on leader chain, skipping...", []string{"leader", tx.Leader, "txhash", txhash})
					continue
				}
			}
			batch.Txs = [][]byte{}
			batch.CummulatedGas = 0
			batch.IsAtomicTx = true
		}
		if maxGas > -1 && maxGas < (batch.CummulatedGas+int64(tx.Gas)) {
			break
		}
		bytelen := int64(base64Len(tx.Tx))
		if maxBytes < (cummulatedBytes + bytelen) {
			break
		}
		batch.Txs = append(batch.Txs, tx.Tx)
		batch.CummulatedGas += int64(tx.Gas)
		cummulatedBytes += bytelen
		if atomicInExec {
			return batch, nil
		}
	}
	return batch, nil
}

// IsBatchFull: quick check to see if we have enough tx in mempool
func (m *Mempool) IsBatchFull(maxGas int64, maxBytes int64) bool {
	if m == nil {
		return false
	}
	if maxGas == -1 {
		return false
	}
	var cummulatedGas int64 = 0
	var cummulatedBytes int64 = 0
	for _, txhash := range m.Order {
		tx, ok := m.Map[txhash]
		if !ok {
			continue
		}
		if maxGas < (cummulatedGas + int64(tx.Gas)) {
			return true
		}
		bytelen := int64(base64Len(tx.Tx))
		if maxBytes < (cummulatedBytes + bytelen) {
			return true
		}
		cummulatedGas += int64(tx.Gas)
		cummulatedBytes += bytelen
	}
	if (maxGas - cummulatedGas) < 50000 {
		return true
	}
	if (maxBytes - cummulatedBytes) < 1000 {
		return true
	}
	return false
}
