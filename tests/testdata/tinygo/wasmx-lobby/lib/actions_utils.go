package lib

import (
	"fmt"
	"sort"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

func SignMessage(validatorPrivkey string, msgstr string) (string, error) {
	return wasmx.Ed25519Sign(validatorPrivkey, msgstr)
}

func GetTopicLobby() (string, error) {
	level, err := GetCurrentLevel()
	if err != nil {
		return "", err
	}
	return GetTopicLevel(level, ROOM_LOBBY), nil
}

func GetTopicNewChain(newchainId string) string {
	return GetTopic(newchainId, ROOM_NEW_CHAIN)
}

func GetProtocolId() string {
	return PROTOCOL_ID
}

func GetTopicLevel(level int32, topic string) string {
	return fmt.Sprintf("%d_%s", level, topic)
}

func GetTopic(chainId string, topic string) string {
	return topic + "_" + chainId
}

func WrapValidators(validators []PotentialValidator, signatures []string) []PotentialValidatorWithSignature {
	v := make([]PotentialValidatorWithSignature, len(validators))
	for i := 0; i < len(validators); i++ {
		v[i] = PotentialValidatorWithSignature{
			Validator: validators[i],
			Signature: signatures[i],
		}
	}
	return v
}

func UnwrapValidators(tempdata MsgNewChainResponse, allvalid []PotentialValidatorWithSignature) MsgNewChainResponse {
	tempdata.Msg.Validators = make([]PotentialValidator, len(allvalid))
	tempdata.Signatures = make([]string, len(allvalid))
	for i := 0; i < len(allvalid); i++ {
		tempdata.Msg.Validators[i] = allvalid[i].Validator
		tempdata.Signatures[i] = allvalid[i].Signature
	}
	return tempdata
}

func SortValidators(validators []PotentialValidatorWithSignature) []PotentialValidatorWithSignature {
	sort.Slice(validators, func(i, j int) bool {
		return validators[i].Validator.AddressBytes < validators[j].Validator.AddressBytes
	})
	return validators
}

func SortValidatorsSimple(validators []PotentialValidator) []PotentialValidator {
	sort.Slice(validators, func(i, j int) bool {
		return validators[i].AddressBytes < validators[j].AddressBytes
	})
	return validators
}

func MergeValidators(validators []PotentialValidatorWithSignature, validators2 []PotentialValidatorWithSignature) []PotentialValidatorWithSignature {
	m := make(map[string]bool)
	for i := 0; i < len(validators); i++ {
		m[validators[i].Validator.AddressBytes] = true
	}
	for i := 0; i < len(validators2); i++ {
		if _, exists := m[validators2[i].Validator.AddressBytes]; exists {
			continue
		}
		validators = append(validators, validators2[i])
		m[validators2[i].Validator.AddressBytes] = true
	}
	return validators
}
