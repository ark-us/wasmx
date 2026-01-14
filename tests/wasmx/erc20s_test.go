package keeper_test

import (
	"encoding/json"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/vm/precompiles"
)

func (suite *KeeperTestSuite) TestERC20I_MintAndTransfer() {
	sender := suite.GetRandomAccount()
	recipient := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(recipient.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	senderAddr := appA.BytesToAccAddressPrefixed(sender.Address).String()
	recipientAddr := appA.BytesToAccAddressPrefixed(recipient.Address).String()

	senderUserID := ut.RegisterIdentityUser(&appA, sender, senderAddr, sender.PubKey.Bytes())
	recipientUserID := ut.RegisterIdentityUser(&appA, recipient, recipientAddr, recipient.PubKey.Bytes())

	wasmbin := precompiles.GetPrecompileByLabel(appA.AddressCodec(), types.ERC20I_v001)
	codeId := appA.StoreCode(sender, wasmbin, nil)

	instMsg := map[string]interface{}{
		"admins":     []string{senderUserID},
		"minters":    []string{senderUserID},
		"name":       "Identity Token",
		"symbol":     "IDT",
		"decimals":   6,
		"base_denom": "uidt",
	}
	instMsgBytes, err := json.Marshal(instMsg)
	suite.Require().NoError(err)

	contractAddr := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: instMsgBytes}, "erc20i", nil)

	mintMsg := []byte(fmt.Sprintf(`{"mint":{"to":"%s","value":"1000"}}`, recipientUserID))
	appA.ExecuteContract(sender, contractAddr, types.WasmxExecutionMessage{Data: mintMsg}, nil, nil)

	balanceResp := appA.WasmxQueryRaw(sender, contractAddr, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, recipientUserID))}, nil, nil)
	var balanceOut struct {
		Balance struct {
			Denom  string      `json:"denom"`
			Amount sdkmath.Int `json:"amount"`
		} `json:"balance"`
	}
	err = json.Unmarshal(balanceResp, &balanceOut)
	suite.Require().NoError(err)
	suite.Require().Equal("IDT", balanceOut.Balance.Denom)
	suite.Require().Equal(sdkmath.NewInt(1000), balanceOut.Balance.Amount)

	transferMsg := []byte(fmt.Sprintf(`{"transfer":{"to":"%s","value":"250"}}`, senderUserID))
	appA.ExecuteContract(recipient, contractAddr, types.WasmxExecutionMessage{Data: transferMsg}, nil, nil)

	balanceResp = appA.WasmxQueryRaw(sender, contractAddr, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, recipientUserID))}, nil, nil)
	err = json.Unmarshal(balanceResp, &balanceOut)
	suite.Require().NoError(err)
	suite.Require().Equal(sdkmath.NewInt(750), balanceOut.Balance.Amount)

	balanceResp = appA.WasmxQueryRaw(sender, contractAddr, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, senderUserID))}, nil, nil)
	err = json.Unmarshal(balanceResp, &balanceOut)
	suite.Require().NoError(err)
	suite.Require().Equal(sdkmath.NewInt(250), balanceOut.Balance.Amount)
}

// =============================================================================
// ERC20GOV TEST HELPERS
// =============================================================================

// ERC20GovTestSetup contains setup data for ERC20Gov tests
type ERC20GovTestSetup struct {
	AppA         *ut.AppContext
	Sender       simulation.Account
	ValAccount   simulation.Account
	ERC20Gov     mcodec.AccAddressPrefixed
	GovContract  mcodec.AccAddressPrefixed
	SenderUserID string
}

// SetupERC20GovTest sets up the test environment for ERC20Gov tests
func (suite *KeeperTestSuite) SetupERC20GovTest() *ERC20GovTestSetup {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	valAccount := simulation.Account{
		PrivKey: s.Chain().SenderPrivKey,
		PubKey:  s.Chain().SenderPrivKey.PubKey(),
		Address: s.Chain().SenderAccount.GetAddress(),
	}

	appAVal := s.AppContext()
	appA := &appAVal

	// Fund accounts
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixed, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(valAccount.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	// Use continuous governance contract address
	govAddr := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_GOV_CONT))

	// Register sender with identity contract
	senderAddr := senderPrefixed.String()
	senderUserID := ut.RegisterIdentityUser(appA, sender, senderAddr, sender.PubKey.Bytes())

	// Store and instantiate wasmx-erc20gov contract
	erc20govWasm := precompiles.GetPrecompileByLabel(appA.AddressCodec(), "erc20gov_0.0.1")
	s.Require().NotNil(erc20govWasm, "erc20gov contract binary not found")

	codeId := appA.StoreCode(sender, erc20govWasm, nil)

	// Initialize erc20gov contract with governance address and sender as admin
	initMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"governance_address":"%s","admins":["%s"],"name":"Gov Token","symbol":"GOV","decimals":6}`,
			govAddr.String(), senderUserID)),
	}
	erc20govContract := appA.InstantiateCode(sender, codeId, initMsg, "wasmx_erc20gov", nil)

	return &ERC20GovTestSetup{
		AppA:         appA,
		Sender:       sender,
		ValAccount:   valAccount,
		ERC20Gov:     erc20govContract,
		GovContract:  govAddr,
		SenderUserID: senderUserID,
	}
}

// =============================================================================
// ERC20GOV TESTS
// =============================================================================

func (suite *KeeperTestSuite) TestERC20Gov() {
	setup := suite.SetupERC20GovTest()
	appA := setup.AppA

	// Create users
	recipient := suite.GetRandomAccount()
	nonAdmin := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(recipient.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(nonAdmin.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	recipientAddr := appA.BytesToAccAddressPrefixed(recipient.Address).String()
	nonAdminAddr := appA.BytesToAccAddressPrefixed(nonAdmin.Address).String()
	recipientUserID := ut.RegisterIdentityUser(appA, recipient, recipientAddr, recipient.PubKey.Bytes())
	nonAdminUserID := ut.RegisterIdentityUser(appA, nonAdmin, nonAdminAddr, nonAdmin.PubKey.Bytes())

	// Verify initial member count is 0
	memberCountResp := appA.WasmxQueryRaw(setup.Sender, setup.ERC20Gov, types.WasmxExecutionMessage{Data: []byte(`{"memberCount":{}}`)}, nil, nil)
	var memberCountOut struct {
		Count uint64 `json:"count"`
	}
	err := json.Unmarshal(memberCountResp, &memberCountOut)
	s.Require().NoError(err)
	s.Require().Equal(uint64(0), memberCountOut.Count, "initial member count should be 0")

	// =========================================================================
	// Test 1: Direct mint should FAIL (caller is not governance)
	// =========================================================================
	directMintMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"mint":{"to":"%s","value":"1000"}}`, recipientUserID)),
	}
	res, _ := appA.ExecuteContractNoCheck(setup.Sender, setup.ERC20Gov, directMintMsg, nil, nil, ut.DEFAULT_GAS_LIMIT, nil)
	s.Require().False(res.IsOK(), "direct mint should fail without governance")
	s.Require().Contains(res.Log, "revert")

	// =========================================================================
	// Test 2: Mint via governance should SUCCEED
	// =========================================================================
	mintData := []byte(fmt.Sprintf(`{"mint":{"to":"%s","value":"1000"}}`, recipientUserID))
	mintMsgBz, _ := json.Marshal(&types.WasmxExecutionMessage{Data: mintData})

	appA.PassProposalWithGovContract(
		setup.GovContract,
		ut.GovVoteContinuous,
		setup.ValAccount,
		setup.Sender,
		[]sdk.Msg{&types.MsgExecuteContract{
			Sender:   setup.GovContract.String(),
			Contract: setup.ERC20Gov.String(),
			Msg:      mintMsgBz,
		}},
		"",
		"Mint tokens",
		"Minting 1000 GOV tokens",
		false,
	)

	// Verify balance and member count after mint
	var balanceOut struct {
		Balance struct {
			Denom  string      `json:"denom"`
			Amount sdkmath.Int `json:"amount"`
		} `json:"balance"`
	}
	balanceResp := appA.WasmxQueryRaw(setup.Sender, setup.ERC20Gov, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, recipientUserID))}, nil, nil)
	_ = json.Unmarshal(balanceResp, &balanceOut)
	s.Require().Equal("GOV", balanceOut.Balance.Denom)
	s.Require().Equal(sdkmath.NewInt(1000), balanceOut.Balance.Amount)

	memberCountResp = appA.WasmxQueryRaw(setup.Sender, setup.ERC20Gov, types.WasmxExecutionMessage{Data: []byte(`{"memberCount":{}}`)}, nil, nil)
	_ = json.Unmarshal(memberCountResp, &memberCountOut)
	s.Require().Equal(uint64(1), memberCountOut.Count, "member count should be 1 after mint")

	// =========================================================================
	// Test 3: Direct transfer should FAIL (caller is not governance)
	// =========================================================================
	directTransferMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"transfer":{"from":"%s","to":"%s","value":"100"}}`, recipientUserID, setup.SenderUserID)),
	}
	res, _ = appA.ExecuteContractNoCheck(setup.Sender, setup.ERC20Gov, directTransferMsg, nil, nil, ut.DEFAULT_GAS_LIMIT, nil)
	s.Require().False(res.IsOK(), "direct transfer should fail without governance")
	s.Require().Contains(res.Log, "revert")

	// =========================================================================
	// Test 4: Transfer via governance should SUCCEED
	// =========================================================================
	transferData := []byte(fmt.Sprintf(`{"transfer":{"from":"%s","to":"%s","value":"250"}}`, recipientUserID, setup.SenderUserID))
	transferMsgBz, _ := json.Marshal(&types.WasmxExecutionMessage{Data: transferData})

	appA.PassProposalWithGovContract(
		setup.GovContract,
		ut.GovVoteContinuous,
		setup.ValAccount,
		setup.Sender,
		[]sdk.Msg{&types.MsgExecuteContract{
			Sender:   setup.GovContract.String(),
			Contract: setup.ERC20Gov.String(),
			Msg:      transferMsgBz,
		}},
		"",
		"Transfer tokens",
		"Transferring 250 GOV tokens",
		false,
	)

	// Verify balances after transfer
	balanceResp = appA.WasmxQueryRaw(setup.Sender, setup.ERC20Gov, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, recipientUserID))}, nil, nil)
	_ = json.Unmarshal(balanceResp, &balanceOut)
	s.Require().Equal(sdkmath.NewInt(750), balanceOut.Balance.Amount)

	balanceResp = appA.WasmxQueryRaw(setup.Sender, setup.ERC20Gov, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, setup.SenderUserID))}, nil, nil)
	_ = json.Unmarshal(balanceResp, &balanceOut)
	s.Require().Equal(sdkmath.NewInt(250), balanceOut.Balance.Amount)

	// Verify member count is now 2
	memberCountResp = appA.WasmxQueryRaw(setup.Sender, setup.ERC20Gov, types.WasmxExecutionMessage{Data: []byte(`{"memberCount":{}}`)}, nil, nil)
	_ = json.Unmarshal(memberCountResp, &memberCountOut)
	s.Require().Equal(uint64(2), memberCountOut.Count, "member count should be 2 after transfer")

	// =========================================================================
	// Test 5: TransferFrom by admin should SUCCEED
	// =========================================================================
	transferFromMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"transferFrom":{"from":"%s","to":"%s","value":"100"}}`, recipientUserID, setup.SenderUserID)),
	}
	res = appA.ExecuteContract(setup.Sender, setup.ERC20Gov, transferFromMsg, nil, nil)
	s.Require().True(res.IsOK(), "transferFrom by admin should succeed: %s", res.Log)

	balanceResp = appA.WasmxQueryRaw(setup.Sender, setup.ERC20Gov, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, recipientUserID))}, nil, nil)
	_ = json.Unmarshal(balanceResp, &balanceOut)
	s.Require().Equal(sdkmath.NewInt(650), balanceOut.Balance.Amount)

	// =========================================================================
	// Test 6: TransferFrom by non-admin should FAIL
	// =========================================================================
	transferFromMsg = types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"transferFrom":{"from":"%s","to":"%s","value":"100"}}`, recipientUserID, nonAdminUserID)),
	}
	res, _ = appA.ExecuteContractNoCheck(nonAdmin, setup.ERC20Gov, transferFromMsg, nil, nil, ut.DEFAULT_GAS_LIMIT, nil)
	s.Require().False(res.IsOK(), "transferFrom by non-admin should fail")
	s.Require().Contains(res.Log, "revert")

	// =========================================================================
	// Test 7: Direct burn should FAIL (caller is not governance)
	// =========================================================================
	directBurnMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"burn":{"from":"%s","value":"100"}}`, recipientUserID)),
	}
	res, _ = appA.ExecuteContractNoCheck(setup.Sender, setup.ERC20Gov, directBurnMsg, nil, nil, ut.DEFAULT_GAS_LIMIT, nil)
	s.Require().False(res.IsOK(), "direct burn should fail without governance")
	s.Require().Contains(res.Log, "revert")

	// =========================================================================
	// Test 8: Burn via governance should SUCCEED
	// =========================================================================
	burnData := []byte(fmt.Sprintf(`{"burn":{"from":"%s","value":"150"}}`, recipientUserID))
	burnMsgBz, _ := json.Marshal(&types.WasmxExecutionMessage{Data: burnData})

	appA.PassProposalWithGovContract(
		setup.GovContract,
		ut.GovVoteContinuous,
		setup.ValAccount,
		setup.Sender,
		[]sdk.Msg{&types.MsgExecuteContract{
			Sender:   setup.GovContract.String(),
			Contract: setup.ERC20Gov.String(),
			Msg:      burnMsgBz,
		}},
		"",
		"Burn tokens",
		"Burning 150 GOV tokens",
		false,
	)

	// Verify balance after burn
	balanceResp = appA.WasmxQueryRaw(setup.Sender, setup.ERC20Gov, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, recipientUserID))}, nil, nil)
	_ = json.Unmarshal(balanceResp, &balanceOut)
	s.Require().Equal(sdkmath.NewInt(500), balanceOut.Balance.Amount)

	// Verify total supply
	supplyResp := appA.WasmxQueryRaw(setup.Sender, setup.ERC20Gov, types.WasmxExecutionMessage{Data: []byte(`{"totalSupply":{}}`)}, nil, nil)
	var supplyOut struct {
		Supply struct {
			Amount sdkmath.Int `json:"amount"`
		} `json:"supply"`
	}
	_ = json.Unmarshal(supplyResp, &supplyOut)
	s.Require().Equal(sdkmath.NewInt(850), supplyOut.Supply.Amount) // 1000 - 150 burned
}
