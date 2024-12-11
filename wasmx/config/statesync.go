package config

// Note protocol must be different than any other protocol ID and per chain ID
var StateSyncProtocolId = "statesync"

func GetStateSyncProtocolId(chainId string) string {
	return StateSyncProtocolId + "_" + chainId
}
