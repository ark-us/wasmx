package consensus

// #include <stdlib.h>
import "C"

// Packed pointer convention identical to wasmx-env (int64: high 32 bits ptr, low 32 bits len)

//go:wasm-module consensus
//export memory_ptrlen_i64_1
func memory_ptrlen_i64_1() {}

//go:wasm-module consensus
//export wasmx_env_i64_2
func wasmx_env_i64_2() {}

// Consensus host API imports

//go:wasmimport consensus CheckTx
func CheckTx_(reqPtr int64) int64

//go:wasmimport consensus PrepareProposal
func PrepareProposal_(reqPtr int64) int64

//go:wasmimport consensus OptimisticExecution
func OptimisticExecution_(reqPtr int64, respPtr int64) int64

//go:wasmimport consensus ProcessProposal
func ProcessProposal_(reqPtr int64) int64

//go:wasmimport consensus FinalizeBlock
func FinalizeBlock_(reqPtr int64) int64

//go:wasmimport consensus BeginBlock
func BeginBlock_(reqPtr int64) int64

//go:wasmimport consensus EndBlock
func EndBlock_(dataPtr int64) int64

//go:wasmimport consensus Commit
func Commit_() int64

//go:wasmimport consensus RollbackToVersion
func RollbackToVersion_(height int64) int64

//go:wasmimport consensus HeaderHash
func HeaderHash_(dataPtr int64) int64

//go:wasmimport consensus ValidatorsHash
func ValidatorsHash_(dataPtr int64) int64

//go:wasmimport consensus ConsensusParamsHash
func ConsensusParamsHash_(dataPtr int64) int64

//go:wasmimport consensus BlockCommitVoteBytes
func BlockCommitVoteBytes_(dataPtr int64) int64

// Snapshot related (declared but not used in AS wrapper)
//
//go:wasmimport consensus ApplySnapshotChunk
func ApplySnapshotChunk_(dataPtr int64) int64

//go:wasmimport consensus LoadSnapshotChunk
func LoadSnapshotChunk_(dataPtr int64) int64

//go:wasmimport consensus OfferSnapshot
func OfferSnapshot_(dataPtr int64) int64

//go:wasmimport consensus ListSnapshots
func ListSnapshots_(dataPtr int64) int64
