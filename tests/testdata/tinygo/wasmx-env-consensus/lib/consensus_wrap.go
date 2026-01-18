package consensus

import (
	"encoding/json"
	"fmt"

	utils "github.com/loredanacirstea/wasmx-env-utils"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

func CheckTx(req RequestCheckTx) (ResponseCheckTx, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return ResponseCheckTx{}, err
	}
	wasmx.LoggerDebugExtended(MODULE_NAME, "CheckTx", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(CheckTx_(utils.BytesToPackedPtr(bz)))
	wasmx.LoggerDebugExtended(MODULE_NAME, "CheckTx", []string{"response", string(out)})
	var resp ResponseCheckTx
	if err := json.Unmarshal(out, &resp); err != nil {
		return ResponseCheckTx{}, err
	}
	return resp, nil
}

func PrepareProposal(req RequestPrepareProposal) (ResponsePrepareProposal, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return ResponsePrepareProposal{}, err
	}
	wasmx.LoggerDebugExtended(MODULE_NAME, "PrepareProposal", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(PrepareProposal_(utils.BytesToPackedPtr(bz)))
	wasmx.LoggerDebugExtended(MODULE_NAME, "PrepareProposal", []string{"response", string(out)})
	var resp ResponsePrepareProposal
	if err := json.Unmarshal(out, &resp); err != nil {
		return ResponsePrepareProposal{}, err
	}
	return resp, nil
}

func ProcessProposal(req RequestProcessProposal) (ResponseProcessProposal, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return ResponseProcessProposal{}, err
	}
	wasmx.LoggerDebugExtended(MODULE_NAME, "ProcessProposal", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(ProcessProposal_(utils.BytesToPackedPtr(bz)))
	wasmx.LoggerDebugExtended(MODULE_NAME, "ProcessProposal", []string{"response", string(out)})
	var resp ResponseProcessProposal
	if err := json.Unmarshal(out, &resp); err != nil {
		return ResponseProcessProposal{}, err
	}
	return resp, nil
}

func OptimisticExecution(req RequestProcessProposal, resp ResponseProcessProposal) (ResponseOptimisticExecution, error) {
	reqbz, err := json.Marshal(&req)
	if err != nil {
		return ResponseOptimisticExecution{}, err
	}
	respbz, err := json.Marshal(&resp)
	if err != nil {
		return ResponseOptimisticExecution{}, err
	}
	wasmx.LoggerDebugExtended(MODULE_NAME, "OptimisticExecution", []string{"request", string(reqbz), "resp", string(respbz)})
	out := utils.PackedPtrToBytes(OptimisticExecution_(utils.BytesToPackedPtr(reqbz), utils.BytesToPackedPtr(respbz)))
	wasmx.LoggerDebugExtended(MODULE_NAME, "OptimisticExecution", []string{"response", string(out)})
	var r ResponseOptimisticExecution
	if err := json.Unmarshal(out, &r); err != nil {
		return ResponseOptimisticExecution{}, err
	}
	return r, nil
}

func FinalizeBlock(req WrapRequestFinalizeBlock) (ResponseFinalizeBlockWrap, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return ResponseFinalizeBlockWrap{}, err
	}
	wasmx.LoggerDebugExtended(MODULE_NAME, "FinalizeBlock", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(FinalizeBlock_(utils.BytesToPackedPtr(bz)))
	wasmx.LoggerDebugExtended(MODULE_NAME, "FinalizeBlock", []string{"response", string(out)})
	var wrap ResponseWrap
	if err := json.Unmarshal(out, &wrap); err != nil {
		return ResponseFinalizeBlockWrap{}, err
	}
	resp := ResponseFinalizeBlockWrap{Error: wrap.Error}
	if wrap.Error == "" && len(wrap.Data) > 0 {
		// data is base64-encoded JSON in host response
		var inner ResponseFinalizeBlock
		if err := json.Unmarshal(wrap.Data, &inner); err != nil {
			return ResponseFinalizeBlockWrap{}, fmt.Errorf("decode finalize block data: %w", err)
		}
		resp.Data = &inner
	}
	return resp, nil
}

func BeginBlock(req RequestFinalizeBlock) (ResponseBeginBlockWrap, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return ResponseBeginBlockWrap{}, err
	}
	wasmx.LoggerDebugExtended(MODULE_NAME, "BeginBlock", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(BeginBlock_(utils.BytesToPackedPtr(bz)))
	wasmx.LoggerDebugExtended(MODULE_NAME, "BeginBlock", []string{"response", string(out)})
	var wrap ResponseWrap
	if err := json.Unmarshal(out, &wrap); err != nil {
		return ResponseBeginBlockWrap{}, err
	}
	resp := ResponseBeginBlockWrap{Error: wrap.Error}
	if wrap.Error == "" && len(wrap.Data) > 0 {
		var inner ResponseBeginBlock
		if err := json.Unmarshal(wrap.Data, &inner); err != nil {
			return ResponseBeginBlockWrap{}, fmt.Errorf("decode begin block data: %w", err)
		}
		resp.Data = &inner
	}
	return resp, nil
}

func EndBlock(metadata string) (ResponseFinalizeBlockWrap, error) {
	wasmx.LoggerDebugExtended(MODULE_NAME, "EndBlock", []string{"metadata", metadata})
	out := utils.PackedPtrToBytes(EndBlock_(utils.StringToPackedPtr(metadata)))
	wasmx.LoggerDebugExtended(MODULE_NAME, "EndBlock", []string{"response", string(out)})
	var wrap ResponseWrap
	if err := json.Unmarshal(out, &wrap); err != nil {
		return ResponseFinalizeBlockWrap{}, err
	}
	resp := ResponseFinalizeBlockWrap{Error: wrap.Error}
	if wrap.Error == "" && len(wrap.Data) > 0 {
		var inner ResponseFinalizeBlock
		if err := json.Unmarshal(wrap.Data, &inner); err != nil {
			return ResponseFinalizeBlockWrap{}, err
		}
		resp.Data = &inner
	}
	return resp, nil
}

func Commit() (ResponseCommit, error) {
	wasmx.LoggerDebugExtended(MODULE_NAME, "Commit", nil)
	out := utils.PackedPtrToBytes(Commit_())
	wasmx.LoggerDebugExtended(MODULE_NAME, "Commit", []string{"response", string(out)})
	var resp ResponseCommit
	if err := json.Unmarshal(out, &resp); err != nil {
		return ResponseCommit{}, err
	}
	return resp, nil
}

func RollbackToVersion(height int64) error {
	wasmx.LoggerDebugExtended(MODULE_NAME, "RollbackToVersion", []string{"height", fmt.Sprintf("%d", height)})
	out := utils.PackedPtrToBytes(RollbackToVersion_(height))
	errStr := string(out)
	wasmx.LoggerDebugExtended(MODULE_NAME, "RollbackToVersion", []string{"err", errStr})
	if errStr != "" {
		return fmt.Errorf("%s", errStr)
	}
	return nil
}

func HeaderHash(header Header) ([]byte, error) {
	bz, err := json.Marshal(&header)
	if err != nil {
		return nil, err
	}
	out := utils.PackedPtrToBytes(HeaderHash_(utils.BytesToPackedPtr(bz)))
	return out, nil
}

func ValidatorsHash(validators []TendermintValidator) ([]byte, error) {
	payload := struct {
		Validators []TendermintValidator `json:"validators"`
	}{Validators: validators}
	bz, err := json.Marshal(&payload)
	if err != nil {
		return nil, err
	}
	out := utils.PackedPtrToBytes(ValidatorsHash_(utils.BytesToPackedPtr(bz)))
	return out, nil
}

func ConsensusParamsHash(params ConsensusParams) ([]byte, error) {
	bz, err := json.Marshal(&params)
	if err != nil {
		return nil, err
	}
	out := utils.PackedPtrToBytes(ConsensusParamsHash_(utils.BytesToPackedPtr(bz)))
	return out, nil
}

func BlockCommitVoteBytes(vote VoteTendermint) ([]byte, error) {
	bz, err := json.Marshal(&vote)
	if err != nil {
		return nil, err
	}
	out := utils.PackedPtrToBytes(BlockCommitVoteBytes_(utils.BytesToPackedPtr(bz)))
	return out, nil
}

// Snapshot helpers (TODO in AS: no host call)
func ApplySnapshotChunk(_ RequestApplySnapshotChunk) ResponseApplySnapshotChunk {
	return ResponseApplySnapshotChunk{}
}
func LoadSnapshotChunk(_ RequestLoadSnapshotChunk) ResponseLoadSnapshotChunk {
	return ResponseLoadSnapshotChunk{}
}
func OfferSnapshot(_ RequestOfferSnapshot) ResponseOfferSnapshot { return ResponseOfferSnapshot{} }
func ListSnapshots(_ RequestListSnapshots) ResponseListSnapshots { return ResponseListSnapshots{} }
