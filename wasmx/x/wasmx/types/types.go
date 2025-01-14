package types

import (
	"fmt"
	"math/big"

	mcodec "github.com/loredanacirstea/wasmx/codec"

	address "cosmossdk.io/core/address"
	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	defaultMemoryCacheSize    uint32 = 100 // in MiB
	defaultSmartQueryGasLimit uint64 = 1_000_000_000
	defaultContractDebugMode         = false
	DefaultBlockGasLimit      uint64 = 100_000_000
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
		return sdkerr.Wrap(ErrEmpty, "code hash")
	}
	return nil
}

func (c CodeInfo) ValidateWithAddress(addressCodec address.Codec) error {
	if err := c.ValidateBasic(); err != nil {
		return err
	}
	if _, err := addressCodec.StringToBytes(c.Creator); err != nil {
		return sdkerr.Wrap(err, "creator")
	}
	return nil
}

// NewCodeInfo fills a new CodeInfo struct
func NewCodeInfo(codeHash []byte, creatorBech32 string, deps []string, metadata CodeMetadata, pinned bool, meteringOff bool) CodeInfo {
	return CodeInfo{
		CodeHash:    codeHash,
		Creator:     creatorBech32,
		Deps:        deps,
		Metadata:    metadata,
		Pinned:      pinned,
		MeteringOff: meteringOff,
	}
}

// NewEnv initializes the environment for a contract instance
func NewEnv(accBech32Codec mcodec.AccBech32Codec, ctx sdk.Context, denom string, contractAddr mcodec.AccAddressPrefixed, codeHash []byte, bytecode []byte, systemDeps []string, info MessageInfo) (Env, error) {
	// safety checks before casting below
	if ctx.BlockHeight() < 0 {
		return Env{}, fmt.Errorf("block height must never be negative: %s", ctx.ChainID())
	}
	nano := ctx.BlockTime().UnixNano()
	if nano < 1 {
		return Env{}, fmt.Errorf("block (unix) time must never be empty or negative: chainId %s, height: %d", ctx.ChainID(), ctx.BlockHeight())
	}

	blockGasLimit := DefaultBlockGasLimit
	if ctx.BlockGasMeter() != nil {
		blockGasLimit = ctx.BlockGasMeter().Limit()
	}

	chainId, err := ParseEvmChainID(ctx.ChainID())
	if err != nil {
		return Env{}, fmt.Errorf("chain ID cannot be parsed for wasmx: %s: %s", ctx.ChainID(), err.Error())
	}
	// TODO add Transaction.hash
	// txhash := tmhash.Sum(ctx.TxBytes())

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
			Hash:      PaddLeftTo32(ctx.HeaderHash()),                                              // TODO fixme
			Proposer:  accBech32Codec.BytesToAccAddressPrefixed(ctx.BlockHeader().ProposerAddress), // TODO fixme
		},
		Transaction: &TransactionInfo{
			GasPrice: big.NewInt(1), // TODO
		},
		Contract: EnvContractInfo{
			Address:    contractAddr,
			CodeHash:   codeHash,
			Bytecode:   bytecode,
			SystemDeps: systemDeps,
		},
		CurrentCall: info,
	}
	return env, nil
}

// NewInfo initializes the MessageInfo for a contract instance
func NewInfo(ctx sdk.Context, origin mcodec.AccAddressPrefixed, creator mcodec.AccAddressPrefixed, deposit sdk.Coins) MessageInfo {
	funds := big.NewInt(0)
	if len(deposit) > 0 {
		funds = deposit[0].Amount.BigInt()
	}
	gasMeter := ctx.GasMeter()
	return MessageInfo{
		Sender:   creator,
		Funds:    funds,
		Origin:   origin,
		GasLimit: big.NewInt(int64(gasMeter.Limit())),
	}
}

// NewContractInfo creates a new instance of a given WASM contract info
func NewContractInfo(CodeId uint64, creatorBech32 string, provenanceBech32 string, initMsg []byte, label string) ContractInfo {
	info := ContractInfo{
		CodeId:      CodeId,
		Creator:     creatorBech32,
		Label:       label,
		InitMessage: initMsg,
		StorageType: ContractStorageType_CoreConsensus,
		// Created: createdAt,
		// TODO tx hash
	}
	if provenanceBech32 != "" {
		info.Provenance = provenanceBech32
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
		return sdkerr.Wrap(ErrEmpty, "code id")
	}
	if err := ValidateLabel(c.Label); err != nil {
		return sdkerr.Wrap(err, "label")
	}
	return nil
}

func (c *ContractInfo) ValidateWithAddress(addressCodec address.Codec) error {
	if err := c.ValidateBasic(); err != nil {
		return err
	}
	if _, err := addressCodec.StringToBytes(c.Creator); err != nil {
		return sdkerr.Wrap(err, "creator")
	}
	if c.Provenance != "" {
		if _, err := addressCodec.StringToBytes(c.Provenance); err != nil {
			return sdkerr.Wrap(err, "provenance")
		}
	}
	return nil
}
