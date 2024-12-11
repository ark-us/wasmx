package types

import (
	"strconv"

	sdkerr "cosmossdk.io/errors"
)

// Codes for wasm contract errors
var (
	DefaultCodespace = ModuleName

	// Note: code 1 is reserved for ErrInternal in the core cosmos sdk

	// ErrCreateFailed error for wasm code that has already been uploaded or failed
	ErrCreateFailed = sdkerr.Register(DefaultCodespace, 2, "create wasm contract failed")

	// ErrAccountExists error for a contract account that already exists
	ErrAccountExists = sdkerr.Register(DefaultCodespace, 3, "contract account already exists")

	// ErrInstantiateFailed error for rust instantiate contract failure
	ErrInstantiateFailed = sdkerr.Register(DefaultCodespace, 4, "instantiate wasm contract failed")

	// ErrExecuteFailed error for rust execution contract failure
	ErrExecuteFailed = sdkerr.Register(DefaultCodespace, 5, "execute wasm contract failed")

	// ErrGasLimit error for out of gas
	ErrGasLimit = sdkerr.Register(DefaultCodespace, 6, "insufficient gas")

	// ErrInvalidGenesis error for invalid genesis file syntax
	ErrInvalidGenesis = sdkerr.Register(DefaultCodespace, 7, "invalid genesis")

	// ErrNotFound error for an entry not found in the store
	ErrNotFound = sdkerr.Register(DefaultCodespace, 8, "not found")

	// ErrQueryFailed error for rust smart query contract failure
	ErrQueryFailed = sdkerr.Register(DefaultCodespace, 9, "query wasm contract failed")

	// ErrInvalidMsg error when we cannot process the error returned from the contract
	ErrInvalidMsg = sdkerr.Register(DefaultCodespace, 10, "invalid CosmosMsg from the contract")

	// ErrMigrationFailed error for rust execution contract failure
	ErrMigrationFailed = sdkerr.Register(DefaultCodespace, 11, "migrate wasm contract failed")

	// ErrEmpty error for empty content
	ErrEmpty = sdkerr.Register(DefaultCodespace, 12, "empty")

	// ErrLimit error for content that exceeds a limit
	ErrLimit = sdkerr.Register(DefaultCodespace, 13, "exceeds limit")

	// ErrInvalid error for content that is invalid in this context
	ErrInvalid = sdkerr.Register(DefaultCodespace, 14, "invalid")

	// ErrDuplicate error for content that exists
	ErrDuplicate = sdkerr.Register(DefaultCodespace, 15, "duplicate")

	// ErrMaxIBCChannels error for maximum number of ibc channels reached
	ErrMaxIBCChannels = sdkerr.Register(DefaultCodespace, 16, "max transfer channels")

	// ErrUnsupportedForContract error when a capability is used that is not supported for/ by this contract
	ErrUnsupportedForContract = sdkerr.Register(DefaultCodespace, 17, "unsupported for this contract")

	// ErrPinContractFailed error for pinning contract failures
	ErrPinContractFailed = sdkerr.Register(DefaultCodespace, 18, "pinning contract failed")

	// ErrUnpinContractFailed error for unpinning contract failures
	ErrUnpinContractFailed = sdkerr.Register(DefaultCodespace, 19, "unpinning contract failed")

	// ErrUnknownMsg error by a message handler to show that it is not responsible for this message type
	ErrUnknownMsg = sdkerr.Register(DefaultCodespace, 20, "unknown message from the contract")

	// ErrInvalidEvent error if an attribute/event from the contract is invalid
	ErrInvalidEvent = sdkerr.Register(DefaultCodespace, 21, "invalid event")

	//  error if an address does not belong to a contract (just for registration)
	_ = sdkerr.Register(DefaultCodespace, 22, "no such contract")

	// code 23 -26 were used for json parser

	// ErrExceedMaxQueryStackSize error if max query stack size is exceeded
	ErrExceedMaxQueryStackSize = sdkerr.Register(DefaultCodespace, 27, "max query stack size exceeded")

	ErrInvalidChainID = sdkerr.Register(DefaultCodespace, 28, "invalid chain ID")

	ErrInvalidRoute = sdkerr.Register(DefaultCodespace, 29, "invalid route")

	ErrUnauthorizedAddress = sdkerr.Register(DefaultCodespace, 30, "unauthorized address")

	ErrInvalidCoreContractCall = sdkerr.Register(DefaultCodespace, 31, "invalid core contract call")

	// ErrInvalidCode error if an attribute/event from the contract is invalid
	_ = sdkerr.Register(DefaultCodespace, 45, "invalid code id")
)

type ErrNoSuchContract struct {
	Addr string
}

func (m *ErrNoSuchContract) Error() string {
	return "no such contract: " + m.Addr
}

func (m *ErrNoSuchContract) ABCICode() uint32 {
	return 22
}

func (m *ErrNoSuchContract) Codespace() string {
	return DefaultCodespace
}

type ErrInvalidCode struct {
	CodeID uint64
}

func (m *ErrInvalidCode) Error() string {
	return "invalid code id: " + strconv.FormatUint(m.CodeID, 10)
}

func (m *ErrInvalidCode) ABCICode() uint32 {
	return 45
}

func (m *ErrInvalidCode) Codespace() string {
	return DefaultCodespace
}
