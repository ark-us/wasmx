package consensus

import (
	"encoding/base64"

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
// DefaultFinalizeResponseEventsParse filters successful tx results and extracts consensus-related events.
// Matches AssemblyScript signature: (txResults, endBlockEvents)
func DefaultFinalizeResponseEventsParse(txResults []ExecTxResult, endBlockEvents []wasmx.Event) FinalizedCoreEventsInfo {
	roleConsensus := false
	consensusContract := ""
	consensusLabel := ""
	createdValidators := []CreatedValidator{}
	initChainRequests := [][]byte{}

	// Scan successful tx results for create validator / init subchain
	for x := 0; x < len(txResults); x++ {
		if txResults[x].Code != uint32(CodeTypeOk) {
			continue
		}
		evs := txResults[x].Events
		for i := 0; i < len(evs); i++ {
			ev := evs[i]
			switch ev.Type {
			case eventTypeCreateValidator:
				for j := range ev.Attributes {
					if ev.Attributes[j].Key == "validator" {
						createdValidators = append(createdValidators, CreatedValidator{OperatorAddress: wasmx.Bech32String(ev.Attributes[j].Value), TxIndex: int32(x)})
					}
				}
			case eventTypeInitSubchain:
				for j := range ev.Attributes {
					if ev.Attributes[j].Key == eventAttrInitSubchainReq {
						// keep as []byte for callers; values are base64-strings in events
						v, err := base64.StdEncoding.DecodeString(ev.Attributes[j].Value)
						if err != nil {
							wasmx.LoggerError("consensus", "init subchain decode error", []string{"error", err.Error()})
						} else {
							initChainRequests = append(initChainRequests, v)
						}
					}
				}
			}
		}
	}

	// Scan end-block events for consensus role registration
	for i := 0; i < len(endBlockEvents); i++ {
		ev := endBlockEvents[i]
		if ev.Type == wasmx.EventTypeRegisterRole {
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
				wasmx.LoggerInfo(MODULE_NAME, "found new consensus contract", []string{"address", consensusContract, "label", consensusLabel})
				break
			} else {
				consensusContract = ""
				consensusLabel = ""
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

// AggregateEvents concatenates end-block events with successful tx events
// Mirrors the AssemblyScript consensus/assembly/utils.ts aggregateEvents
func AggregateEvents(txResults []ExecTxResult, events []wasmx.Event) []wasmx.Event {
	res := make([]wasmx.Event, 0, len(events))
	res = append(res, events...)
	for i := 0; i < len(txResults); i++ {
		res = append(res, txResults[i].Events...)
	}
	return res
}
