package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	mcodec "wasmx/v1/codec"
)

// AddressGenerator abstract address generator to be used for a single contract address
type AddressGenerator func(ctx sdk.Context, codeID uint64, checksum []byte) mcodec.AccAddressPrefixed

// UInt64LengthPrefix prepend big endian encoded byte length
func UInt64LengthPrefix(bz []byte) []byte {
	return append(sdk.Uint64ToBigEndian(uint64(len(bz))), bz...)
}

// EwasmClassicAddressGenerator generates a contract address using codeID and instanceID sequence and increments sequence
func (k *Keeper) EwasmClassicAddressGenerator(creator mcodec.AccAddressPrefixed) AddressGenerator {
	cdcacc := mcodec.MustUnwrapAccBech32Codec(k.AddressCodec())
	return func(ctx sdk.Context, _ uint64, _ []byte) mcodec.AccAddressPrefixed {
		existingAcct, err := k.GetAccountKeeper().GetAccountPrefixed(ctx, creator)
		if err != nil || existingAcct == nil {
			// create an empty account (so we don't have issues later)
			existingAcct, err = k.GetAccountKeeper().NewAccountWithAddressPrefixed(ctx, creator)
			if err != nil {
				panic(fmt.Sprintf("cannot create new account: %s", err.Error()))
			}
			if existingAcct == nil {
				panic("created nil account")
			}
		}
		nonce := existingAcct.GetSequence()
		existingAcct.SetSequence(nonce + 1)
		err = k.GetAccountKeeper().SetAccountPrefixed(ctx, existingAcct)
		if err != nil {
			panic(fmt.Sprintf("cannot set new account: %s", err.Error()))
		}
		return cdcacc.BytesToAccAddressPrefixed(EwasmBuildContractAddressClassic(creator.Bytes(), nonce))
	}
}

// EwasmBuildContractAddressClassic builds an sdk account address for a contract.
func EwasmBuildContractAddressClassic(creator sdk.AccAddress, nonce uint64) sdk.AccAddress {
	creatorAddress := common.BytesToAddress(creator.Bytes())
	contractAddr := crypto.CreateAddress(creatorAddress, nonce)
	return sdk.AccAddress(contractAddr.Bytes())
}

// EwasmBuildContractAddressPredictable builds an sdk account address for a contract.
func EwasmBuildContractAddressPredictable(creator sdk.AccAddress, salt []byte, checksum []byte) sdk.AccAddress {
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

// EwasmPredictableAddressGenerator generates a predictable contract address
func (k *Keeper) EwasmPredictableAddressGenerator(creator mcodec.AccAddressPrefixed, salt []byte, _ []byte, _ bool) AddressGenerator {
	cdcacc := mcodec.MustUnwrapAccBech32Codec(k.AddressCodec())

	return func(ctx sdk.Context, _ uint64, checksum []byte) mcodec.AccAddressPrefixed {
		return cdcacc.BytesToAccAddressPrefixed(EwasmBuildContractAddressPredictable(creator.Bytes(), salt, checksum))
	}
}
