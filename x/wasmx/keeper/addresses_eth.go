package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// AddressGenerator abstract address generator to be used for a single contract address
type AddressGenerator func(ctx sdk.Context, codeID uint64, checksum []byte) sdk.AccAddress

// UInt64LengthPrefix prepend big endian encoded byte length
func UInt64LengthPrefix(bz []byte) []byte {
	return append(sdk.Uint64ToBigEndian(uint64(len(bz))), bz...)
}

// EwasmClassicAddressGenerator generates a contract address using codeID and instanceID sequence
func (k Keeper) EwasmClassicAddressGenerator(creator sdk.AccAddress) AddressGenerator {
	return func(ctx sdk.Context, _ uint64, _ []byte) sdk.AccAddress {
		existingAcct := k.accountKeeper.GetAccount(ctx, creator)
		if existingAcct == nil {
			// create an empty account (so we don't have issues later)
			existingAcct = k.accountKeeper.NewAccountWithAddress(ctx, creator)
			k.accountKeeper.SetAccount(ctx, existingAcct)
		}
		nonce := existingAcct.GetSequence()
		return EwasmBuildContractAddressClassic(creator, nonce)
	}
}

// EwasmBuildContractAddressClassic builds an sdk account address for a contract.
func EwasmBuildContractAddressClassic(creator sdk.AccAddress, nonce uint64) sdk.AccAddress {
	creatorAddress := common.BytesToAddress(creator.Bytes())
	contractAddr := crypto.CreateAddress(creatorAddress, nonce)
	return sdk.AccAddress(contractAddr.Bytes())
}

// EwasmPredictableAddressGenerator generates a predictable contract address
func (k Keeper) EwasmPredictableAddressGenerator(creator sdk.AccAddress, salt []byte, _ []byte, _ bool) AddressGenerator {
	return func(ctx sdk.Context, _ uint64, checksum []byte) sdk.AccAddress {
		if len(checksum) != 32 {
			panic("invalid checksum")
		}
		if err := sdk.VerifyAddressFormat(creator); err != nil {
			panic(fmt.Sprintf("creator: %s", err))
		}

		if len(salt) != 32 {
			panic(fmt.Sprintf("salt is not 32 bytes"))
		}

		creatorAddress := common.BytesToAddress(creator.Bytes())

		var salt32 [32]byte
		copy(salt32[:], salt)

		contractAddr := crypto.CreateAddress2(creatorAddress, salt32, checksum)
		return sdk.AccAddress(contractAddr.Bytes())
	}
}
