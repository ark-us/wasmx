package consensus

import (
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

type CreatedValidator struct {
	OperatorAddress wasmx.Bech32String `json:"operator_address"`
	TxIndex         int32              `json:"txindex"`
}

type FinalizedCoreEventsInfo struct {
	ConsensusContract string             `json:"consensusContract"`
	ConsensusLabel    string             `json:"consensusLabel"`
	CreatedValidators []CreatedValidator `json:"createdValidators"`
	InitChainRequests [][]byte           `json:"initChainRequests"`
}

const (
	eventTypeCreateValidator = "create_validator"
	eventTypeInitSubchain    = "init_subchain"
	eventAttrInitSubchainReq = "init_subchain_request"
)

// DefaultFinalizeResponseEventsParse filters successful tx results and extracts consensus-related events.
func DefaultFinalizeResponseEventsParse(txResults []ExecTxResult) FinalizedCoreEventsInfo {
	roleConsensus := false
	consensusContract := ""
	consensusLabel := ""
	createdValidators := []CreatedValidator{}
	initChainRequests := [][]byte{}

	for x := 0; x < len(txResults); x++ {
		if txResults[x].Code != uint32(CodeTypeOk) {
			continue
		}
		evs := txResults[x].Events
		for i := 0; i < len(evs); i++ {
			ev := evs[i]
			switch ev.Type {
			case wasmx.EventTypeRegisterRole:
				for j := range ev.Attributes {
					a := ev.Attributes[j]
					switch a.Key {
					case wasmx.AttributeKeyRole:
						roleConsensus = (a.Value == "consensus")
					case wasmx.AttributeKeyContractAddress:
						consensusContract = a.Value
					case wasmx.AttributeKeyRoleLabel:
						consensusLabel = a.Value
					}
				}
				if roleConsensus {
					wasmx.LoggerInfo("wasmx_consensus", "found new consensus contract", []string{"address", consensusContract, "label", consensusLabel})
				} else {
					consensusContract = ""
					consensusLabel = ""
				}
			case eventTypeCreateValidator:
				for j := range ev.Attributes {
					if ev.Attributes[j].Key == "validator" {
						createdValidators = append(createdValidators, CreatedValidator{OperatorAddress: wasmx.Bech32String(ev.Attributes[j].Value), TxIndex: int32(x)})
					}
				}
			case eventTypeInitSubchain:
				for j := range ev.Attributes {
					if ev.Attributes[j].Key == eventAttrInitSubchainReq {
						// values are base64-strings in events; keep as []byte for callers
						initChainRequests = append(initChainRequests, []byte(ev.Attributes[j].Value))
					}
				}
			}
		}
	}

	return FinalizedCoreEventsInfo{
		ConsensusContract: consensusContract,
		ConsensusLabel:    consensusLabel,
		CreatedValidators: createdValidators,
		InitChainRequests: initChainRequests,
	}
}
