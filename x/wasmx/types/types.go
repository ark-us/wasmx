package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	defaultMemoryCacheSize    uint32 = 100 // in MiB
	defaultSmartQueryGasLimit uint64 = 3_000_000
	defaultContractDebugMode         = false
)

// WasmConfig is the extra config required for wasm
type WasmConfig struct {
	// SimulationGasLimit is the max gas to be used in a tx simulation call.
	// When not set the consensus max block gas is used instead
	SimulationGasLimit *uint64
	// SimulationGasLimit is the max gas to be used in a smart query contract call
	SmartQueryGasLimit uint64
	// MemoryCacheSize in MiB not bytes
	MemoryCacheSize uint32
	// ContractDebugMode log what contract print
	ContractDebugMode bool
}

// DefaultWasmConfig returns the default settings for WasmConfig
func DefaultWasmConfig() WasmConfig {
	return WasmConfig{
		SmartQueryGasLimit: defaultSmartQueryGasLimit,
		MemoryCacheSize:    defaultMemoryCacheSize,
		ContractDebugMode:  defaultContractDebugMode,
	}
}

func (c CodeInfo) ValidateBasic() error {
	if len(c.CodeHash) == 0 {
		return sdkerrors.Wrap(ErrEmpty, "code hash")
	}
	if _, err := sdk.AccAddressFromBech32(c.Creator); err != nil {
		return sdkerrors.Wrap(err, "creator")
	}
	return nil
}

// NewCodeInfo fills a new CodeInfo struct
func NewCodeInfo(codeHash []byte, creator sdk.AccAddress, deps []string, metadata CodeMetadata) CodeInfo {
	return CodeInfo{
		CodeHash: codeHash,
		Creator:  creator.String(),
		Deps:     deps,
		Metadata: metadata,
	}
}

// NewEnv initializes the environment for a contract instance
func NewEnv(ctx sdk.Context, denom string, contractAddr sdk.AccAddress, codeHash []byte, bytecode []byte, info MessageInfo) Env {
	// safety checks before casting below
	if ctx.BlockHeight() < 0 {
		panic("Block height must never be negative")
	}
	nano := ctx.BlockTime().UnixNano()
	if nano < 1 {
		panic("Block (unix) time must never be empty or negative ")
	}

	blockGasLimit := uint64(10_000_000)
	if ctx.BlockGasMeter() != nil {
		blockGasLimit = ctx.BlockGasMeter().Limit()
	}

	chainId, err := ParseChainID(ctx.ChainID())
	if err != nil {
		panic("Chain ID cannot be parsed for wasmx")
	}

	env := Env{
		Chain: ChainInfo{
			ChainIdFull: ctx.ChainID(),
			ChainId:     chainId,
			Denom:       denom,
		},
		Block: BlockInfo{
			Height:    uint64(ctx.BlockHeight()),
			Timestamp: uint64(nano),
			GasLimit:  blockGasLimit,
			Hash:      PaddLeftTo32(ctx.HeaderHash().Bytes()), // TODO fixme
			Proposer:  sdk.AccAddress(ctx.BlockHeader().ProposerAddress),
		},
		Transaction: &TransactionInfo{
			GasPrice: big.NewInt(1), // TODO
		},
		Contract: EnvContractInfo{
			Address:  contractAddr,
			CodeHash: codeHash,
			Bytecode: bytecode,
		},
		CurrentCall: info,
	}
	return env
}

// NewInfo initializes the MessageInfo for a contract instance
func NewInfo(origin sdk.AccAddress, creator sdk.AccAddress, deposit sdk.Coins) MessageInfo {
	funds := big.NewInt(0)
	if len(deposit) > 0 {
		funds = deposit[0].Amount.BigInt()
	}
	return MessageInfo{
		Sender: creator,
		Funds:  funds,
		Origin: origin,
		// TODO!!
		GasLimit: big.NewInt(20000000),
	}
}

// NewContractInfo creates a new instance of a given WASM contract info
func NewContractInfo(CodeId uint64, creator sdk.AccAddress, provenance sdk.AccAddress, initMsg []byte, label string) ContractInfo {
	info := ContractInfo{
		CodeId:      CodeId,
		Creator:     creator.String(),
		Label:       label,
		InitMessage: initMsg,
		// Created: createdAt,
		// TODO tx hash
	}
	if provenance != nil {
		info.Provenance = provenance.String()
	}
	return info
}

// validatable is an optional interface that can be implemented by an ContractInfoExtension to enable validation
type validatable interface {
	ValidateBasic() error
}

// ValidateBasic does syntax checks on the data. If an extension is set and has the `ValidateBasic() error` method, then
// the method is called as well. It is recommend to implement `ValidateBasic` so that the data is verified in the setter
// but also in the genesis import process.
func (c *ContractInfo) ValidateBasic() error {
	if c.CodeId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "code id")
	}
	if _, err := sdk.AccAddressFromBech32(c.Creator); err != nil {
		return sdkerrors.Wrap(err, "creator")
	}
	if err := ValidateLabel(c.Label); err != nil {
		return sdkerrors.Wrap(err, "label")
	}
	if c.Provenance != "" {
		if _, err := sdk.AccAddressFromBech32(c.Provenance); err != nil {
			return sdkerrors.Wrap(err, "creator")
		}
	}
	return nil
}
