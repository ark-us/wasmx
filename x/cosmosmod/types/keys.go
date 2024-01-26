package types

const (
	// ModuleName defines the module name
	ModuleName = "cosmosmod"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

const (
	paramsPrefix = iota + 1
	otherPrefix
)

var (
	KeyParamsPrefix = []byte{paramsPrefix}
	KeyOtherPrefix  = []byte{otherPrefix}
)
