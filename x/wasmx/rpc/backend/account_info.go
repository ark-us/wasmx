package backend

import (
	"encoding/hex"

	"github.com/pkg/errors"

	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	app "mythos/v1/app"
	rpctypes "mythos/v1/x/wasmx/rpc/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

// GetCode returns the contract code at the given address and block number.
func (b *Backend) GetCode(_address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (hexutil.Bytes, error) {
	blockNum, err := b.BlockNumberFromTendermint(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	address := wasmxtypes.AccAddressFromEvm(_address)
	req := &wasmxtypes.QueryContractInfoRequest{
		Address: address.String(),
	}

	res, err := b.queryClient.ContractInfo(rpctypes.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	reqCode := &wasmxtypes.QueryCodeInfoRequest{
		CodeId: res.CodeId,
	}

	resCode, err := b.queryClient.CodeInfo(rpctypes.ContextWithHeight(blockNum.Int64()), reqCode)
	if err != nil || resCode == nil {
		return nil, err
	}

	if len(resCode.CodeInfo.InterpretedBytecodeRuntime) > 0 {
		return resCode.CodeInfo.InterpretedBytecodeRuntime, nil
	}

	return nil, nil
}

// GetStorageAt returns the contract storage at the given address, block number, and key.
func (b *Backend) GetStorageAt(_address common.Address, key string, blockNrOrHash rpctypes.BlockNumberOrHash) (hexutil.Bytes, error) {
	blockNum, err := b.BlockNumberFromTendermint(blockNrOrHash)
	if err != nil {
		return nil, err
	}
	// TODO check if 0x is used, same for value
	keybz, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}

	address := wasmxtypes.AccAddressFromEvm(_address)
	req := &wasmxtypes.QueryRawContractStateRequest{
		Address:   address.String(),
		QueryData: keybz,
	}

	res, err := b.queryClient.RawContractState(rpctypes.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	value := common.HexToHash(hex.EncodeToString(res.Data))
	return value.Bytes(), nil
}

// GetBalance returns the provided account's balance up to the provided block number.
func (b *Backend) GetBalance(_address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (*hexutil.Big, error) {
	blockNum, err := b.BlockNumberFromTendermint(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	address := wasmxtypes.AccAddressFromEvm(_address)
	req := &banktypes.QueryBalanceRequest{
		Address: address.String(),
		// TODO
		Denom: app.BaseDenom,
	}

	_, err = b.TendermintBlockByNumber(blockNum)
	if err != nil {
		return nil, err
	}

	res, err := b.queryClient.Bank.Balance(rpctypes.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	val := res.Balance.Amount

	// balance can only be negative in case of pruned node
	if val.IsNegative() {
		return nil, errors.New("couldn't fetch balance. Node state is pruned")
	}

	return (*hexutil.Big)(val.BigInt()), nil
}

// GetTransactionCount returns the number of transactions at the given address up to the given block number.
func (b *Backend) GetTransactionCount(address common.Address, blockNum rpctypes.BlockNumber) (*hexutil.Uint64, error) {
	n := hexutil.Uint64(0)
	bn, err := b.BlockNumber()
	if err != nil {
		return &n, err
	}
	height := blockNum.Int64()

	currentHeight := int64(bn) //#nosec G701 -- checked for int overflow already
	if height > currentHeight {
		return &n, sdkerr.Wrapf(
			sdkerrors.ErrInvalidHeight,
			"cannot query with height in the future (current: %d, queried: %d); please provide a valid height",
			currentHeight, height,
		)
	}
	// Get nonce (sequence) from account
	from := sdk.AccAddress(address.Bytes())
	accRet := b.clientCtx.AccountRetriever

	err = accRet.EnsureExists(b.clientCtx, from)
	if err != nil {
		// account doesn't exist yet, return 0
		return &n, nil
	}

	includePending := blockNum == rpctypes.EthPendingBlockNumber
	nonce, err := b.getAccountNonce(address, includePending, blockNum.Int64(), b.logger)
	if err != nil {
		return nil, err
	}

	n = hexutil.Uint64(nonce)
	return &n, nil
}
