package keeper

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"testing"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	mcodec "wasmx/v1/codec"
)

type BankKeeper interface {
	SendCoinsPrefixed(ctx context.Context, fromAddr mcodec.AccAddressPrefixed, toAddr mcodec.AccAddressPrefixed, amt sdk.Coins) error
	SendCoinsFromModuleToAccountPrefixed(ctx context.Context, senderModule string, recipientAddr mcodec.AccAddressPrefixed, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
}

type TestFaucet struct {
	t                testing.TB
	bankKeeper       BankKeeper
	sender           mcodec.AccAddressPrefixed
	balance          sdk.Coins
	minterModuleName string
}

func NewTestFaucet(t testing.TB, accBech32Codec mcodec.AccBech32Codec, ctx sdk.Context, bankKeeper BankKeeper, minterModuleName string, initialAmount ...sdk.Coin) *TestFaucet {
	require.NotEmpty(t, initialAmount)
	r := &TestFaucet{t: t, bankKeeper: bankKeeper, minterModuleName: minterModuleName}
	_, _, addr := keyPubAddr()
	r.sender = accBech32Codec.BytesToAccAddressPrefixed(addr)
	r.Mint(ctx, r.sender, initialAmount...)
	r.balance = initialAmount
	return r
}

func (f *TestFaucet) Mint(parentCtx sdk.Context, addr mcodec.AccAddressPrefixed, amounts ...sdk.Coin) {
	require.NotEmpty(f.t, amounts)
	ctx := parentCtx.WithEventManager(sdk.NewEventManager()) // discard all faucet related events
	err := f.bankKeeper.MintCoins(ctx, f.minterModuleName, amounts)
	require.NoError(f.t, err)
	err = f.bankKeeper.SendCoinsFromModuleToAccountPrefixed(ctx, f.minterModuleName, addr, amounts)
	require.NoError(f.t, err)
	f.balance = f.balance.Add(amounts...)
}

func (f *TestFaucet) Fund(parentCtx sdk.Context, receiver mcodec.AccAddressPrefixed, amounts ...sdk.Coin) {
	require.NotEmpty(f.t, amounts)
	// ensure faucet is always filled
	if !f.balance.IsAllGTE(amounts) {
		f.Mint(parentCtx, f.sender, amounts...)
	}
	ctx := parentCtx.WithEventManager(sdk.NewEventManager()) // discard all faucet related events
	err := f.bankKeeper.SendCoinsPrefixed(ctx, f.sender, receiver, amounts)
	require.NoError(f.t, err)
	f.balance = f.balance.Sub(amounts...)
}

func (f *TestFaucet) NewFundedRandomAccount(accBech32Codec mcodec.AccBech32Codec, ctx sdk.Context, amounts ...sdk.Coin) sdk.AccAddress {
	_, _, addr := keyPubAddr()
	addrPrefixed := accBech32Codec.BytesToAccAddressPrefixed(addr)
	f.Fund(ctx, addrPrefixed, amounts...)
	return addr
}

type BurnerExampleInitMsg struct {
	Payout sdk.AccAddress `json:"payout"`
}

func (m BurnerExampleInitMsg) GetBytes(t testing.TB) []byte {
	initMsgBz, err := json.Marshal(m)
	require.NoError(t, err)
	return initMsgBz
}

var keyCounter uint64

// we need to make this deterministic (same every test run), as encoded address size and thus gas cost,
// depends on the actual bytes (due to ugly CanonicalAddress encoding)
func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.AccAddress) {
	keyCounter++
	seed := make([]byte, 8)
	binary.BigEndian.PutUint64(seed, keyCounter)

	key := ed25519.GenPrivKeyFromSecret(seed)
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}
