package keeper_test

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestWasmxIdentityAllowances() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	senderAddr := appA.BytesToAccAddressPrefixed(sender.Address).String()
	senderPubKeyBytes := sender.PubKey.Bytes()
	senderPubKeyBase64 := base64.StdEncoding.EncodeToString(senderPubKeyBytes)

	// Get identity contract address by role
	identityAddr, found := suite.App().WasmxKeeper.GetContractAddressByRoleInitial(appA.Context(), wasmxtypes.ROLE_ACCOUNT_IDENTITY)
	s.Require().True(found, "identity contract not found")

	// Step 1: Register sender with identity contract
	registerMsg := wasmxtypes.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"register_user":{"address":"%s","public_key":"%s"}}`, senderAddr, senderPubKeyBase64))}
	res := appA.ExecuteContract(sender, identityAddr, registerMsg, nil, nil)
	s.Require().True(res.IsOK(), "register_user failed: %s", res.Log)

	// Parse user_id from response
	var registerResp struct {
		UserID string `json:"user_id"`
	}
	err := appA.DecodeExecuteResponse(res, &registerResp)
	s.Require().NoError(err)
	s.Require().NotEmpty(registerResp.UserID, "user_id should not be empty")
	userID := registerResp.UserID

	// Step 2: Create a new account (ephemeral address)
	newAccount := suite.GetRandomAccount()
	newAddr := appA.BytesToAccAddressPrefixed(newAccount.Address).String()
	newPubKeyBytes := newAccount.PubKey.Bytes()
	newPubKeyBase64 := base64.StdEncoding.EncodeToString(newPubKeyBytes)

	// Step 3: Add the new address with DefaultAllowances
	allowanceAmount := "1000000000000000000" // 1 token with 18 decimals
	baseDenom := appA.Chain.Config.BaseDenom

	addAddressMsg := wasmxtypes.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"add_address":{"user_id":"%s","address":"%s","public_key":"%s","default_allowances":[{"denom":"%s","amount":"%s"}]}}`,
		userID, newAddr, newPubKeyBase64, baseDenom, allowanceAmount))}

	res = appA.ExecuteContract(sender, identityAddr, addAddressMsg, nil, nil)
	s.Require().True(res.IsOK(), "add_address failed: %s", res.Log)

	// Step 4: Query ERC20 allowance to verify it was set
	// Query bank for ERC20 address
	bankAddr, found := suite.App().WasmxKeeper.GetContractAddressByRoleInitial(appA.Context(), wasmxtypes.ROLE_BANK)
	s.Require().True(found, "bank contract not found")

	bankQueryMsg := wasmxtypes.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"GetAddressByDenom":{"denom":"%s"}}`, baseDenom))}
	bankRespBz := appA.WasmxQueryRaw(sender, bankAddr, bankQueryMsg, nil, nil)

	var erc20AddrResp struct {
		Address string `json:"address"`
	}
	err = json.Unmarshal(bankRespBz, &erc20AddrResp)
	s.Require().NoError(err)
	erc20AddrStr := erc20AddrResp.Address
	s.Require().NotEmpty(erc20AddrStr, "ERC20 address should not be empty")

	erc20Addr, err := appA.AddressCodec().StringToBytes(erc20AddrStr)
	s.Require().NoError(err)
	erc20AddrPrefixed := appA.BytesToAccAddressPrefixed(erc20Addr)

	// Query ERC20 allowance
	allowanceQueryMsg := wasmxtypes.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"allowance":{"owner":"%s","spender":"%s"}}`, senderAddr, newAddr))}
	allowanceRespBz := appA.WasmxQueryRaw(sender, erc20AddrPrefixed, allowanceQueryMsg, nil, nil)

	var allowanceResult struct {
		Remaining sdkmath.Int `json:"remaining"`
	}
	err = json.Unmarshal(allowanceRespBz, &allowanceResult)
	s.Require().NoError(err)

	expectedAllowance, _ := sdkmath.NewIntFromString(allowanceAmount)
	s.Require().True(allowanceResult.Remaining.Equal(expectedAllowance),
		"allowance mismatch: expected %s, got %s", expectedAllowance.String(), allowanceResult.Remaining.String())
}
