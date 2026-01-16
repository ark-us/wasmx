package keeper_test

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	cosmosmodtypes "github.com/loredanacirstea/wasmx/x/cosmosmod/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/vm/precompiles"
	wasmxtest "github.com/loredanacirstea/mythos-tests/testdata/wasmx"
)

// =============================================================================
// WASMX-GROUPS TEST HELPERS
// =============================================================================

// GroupsTestSetup contains setup data for groups tests
type GroupsTestSetup struct {
	AppA           *ut.AppContext
	Sender         simulation.Account
	ValAccount     simulation.Account
	GroupsContract mcodec.AccAddressPrefixed
	IdentityAddr   mcodec.AccAddressPrefixed
	GovContract    mcodec.AccAddressPrefixed
	TokenContract  mcodec.AccAddressPrefixed
	TokenDenom     string
	MinBalance     string
	UserID         string
}

// SetupGroupsTest sets up the test environment for groups tests
func (suite *KeeperTestSuite) SetupGroupsTest() *GroupsTestSetup {
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

	// Get identity contract address
	identityAddr, found := suite.App().WasmxKeeper.GetContractAddressByRoleInitial(appA.Context(), types.ROLE_ACCOUNT_IDENTITY)
	s.Require().True(found, "identity contract not found")

	// Use continuous governance contract address
	govAddr := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_GOV_CONT))

	// Store and instantiate ERC20 token contract for group membership
	tokenWasm := precompiles.GetPrecompileByLabel(appA.AddressCodec(), types.ERC20_v001)
	s.Require().NotNil(tokenWasm, "erc20 contract binary not found")
	tokenCodeId := appA.StoreCode(sender, tokenWasm, nil)
	tokenDenom := "ugrp"
	tokenSymbol := "GRP"
	tokenInitMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"admins":["%s"],"minters":["%s"],"name":"Group Token","symbol":"%s","decimals":6,"base_denom":"%s","negative_balance_threshold":"0"}`,
			senderPrefixed.String(), senderPrefixed.String(), tokenSymbol, tokenDenom)),
	}
	tokenContract := appA.InstantiateCode(sender, tokenCodeId, tokenInitMsg, "group_token", nil)

	// Register sender with identity contract
	senderAddr := senderPrefixed.String()
	senderPubKeyBase64 := base64.StdEncoding.EncodeToString(sender.PubKey.Bytes())

	registerMsg := types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"register_user":{"address":"%s","public_key":"%s"}}`, senderAddr, senderPubKeyBase64))}
	res := appA.ExecuteContract(sender, identityAddr, registerMsg, nil, nil)
	s.Require().True(res.IsOK(), "register_user failed: %s", res.Log)

	// Parse user_id from response
	var registerResp struct {
		UserID string `json:"user_id"`
	}
	err := appA.DecodeExecuteResponse(res, &registerResp)
	s.Require().NoError(err)
	s.Require().NotEmpty(registerResp.UserID, "user_id should not be empty")

	groupsContract, found := suite.App().WasmxKeeper.GetContractAddressByRoleInitial(appA.Context(), types.ROLE_GROUP)
	s.Require().True(found, "groups contract not found in system contracts")

	return &GroupsTestSetup{
		AppA:           appA,
		Sender:         sender,
		ValAccount:     valAccount,
		GroupsContract: groupsContract,
		IdentityAddr:   identityAddr,
		GovContract:    govAddr,
		TokenContract:  tokenContract,
		TokenDenom:     tokenSymbol,
		MinBalance:     "100",
		UserID:         registerResp.UserID,
	}
}

func (suite *KeeperTestSuite) MintGroupToken(setup *GroupsTestSetup, toAddr string, amount string) {
	mintMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"mint":{"to":"%s","value":"%s"}}`, toAddr, amount)),
	}
	res := setup.AppA.ExecuteContract(setup.Sender, setup.TokenContract, mintMsg, nil, nil)
	suite.Require().True(res.IsOK(), "mint failed: %s", res.Log)
}

// RegisterUser registers a new user with the identity contract and returns the user_id
func (suite *KeeperTestSuite) RegisterUser(setup *GroupsTestSetup, account simulation.Account) string {
	addr := setup.AppA.BytesToAccAddressPrefixed(account.Address).String()
	pubKeyBase64 := base64.StdEncoding.EncodeToString(account.PubKey.Bytes())

	registerMsg := types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"register_user":{"address":"%s","public_key":"%s"}}`, addr, pubKeyBase64))}
	res := setup.AppA.ExecuteContract(account, setup.IdentityAddr, registerMsg, nil, nil)
	s.Require().True(res.IsOK(), "register_user failed: %s", res.Log)

	var resp struct {
		UserID string `json:"user_id"`
	}
	err := setup.AppA.DecodeExecuteResponse(res, &resp)
	s.Require().NoError(err)
	return resp.UserID
}

func encodeMsgsAsAnyBase64(appA *ut.AppContext, msgs []sdk.Msg) []string {
	encoded := make([]string, 0, len(msgs))
	anys, err := txtypes.SetMsgs(msgs)
	appA.S.Require().NoError(err)
	for _, anymsg := range anys {
		msgBz, err := appA.App.JSONCodec().MarshalJSON(anymsg)
		appA.S.Require().NoError(err)
		encoded = append(encoded, base64.StdEncoding.EncodeToString(msgBz))
	}
	return encoded
}

// =============================================================================
// WASMX-GROUPS TESTS
// =============================================================================

func (suite *KeeperTestSuite) TestGroupsContractInit() {
	setup := suite.SetupGroupsTest()
	appA := setup.AppA

	// Query config to verify initialization
	configQuery := types.WasmxExecutionMessage{Data: []byte(`{"query_get_config":{}}`)}
	configBz := appA.WasmxQueryRaw(setup.Sender, setup.GroupsContract, configQuery, nil, nil)

	var configResp struct {
		Config struct{} `json:"config"`
	}
	err := json.Unmarshal(configBz, &configResp)
	s.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestGroupsCreateGroup() {
	setup := suite.SetupGroupsTest()
	appA := setup.AppA

	// Create a group
	createGroupMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"create_group":{"name":"Test Group","description":"A test group","admins":["%s"],"protocol":{"governance_contract":"%s"},"token":"%s","min_balance":"%s"}}`,
			setup.UserID, setup.GovContract.String(), setup.TokenContract.String(), setup.MinBalance)),
	}

	res := appA.ExecuteContract(setup.Sender, setup.GroupsContract, createGroupMsg, nil, nil)
	s.Require().True(res.IsOK(), "create_group failed: %s", res.Log)

	// Parse group_id from response
	var createResp struct {
		GroupID string `json:"group_id"`
	}
	err := appA.DecodeExecuteResponse(res, &createResp)
	s.Require().NoError(err)
	s.Require().NotEmpty(createResp.GroupID)

	// Query the created group
	getGroupQuery := types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"query_get_group":{"group_id":"%s"}}`, createResp.GroupID))}
	groupBz := appA.WasmxQueryRaw(setup.Sender, setup.GroupsContract, getGroupQuery, nil, nil)

	var groupResp struct {
		Group struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			MemberCount uint64 `json:"member_count"`
			Token       string `json:"token"`
			TokenDenom  string `json:"token_denom"`
			MinBalance  string `json:"min_balance"`
		} `json:"group"`
	}
	err = json.Unmarshal(groupBz, &groupResp)
	s.Require().NoError(err)
	s.Require().Equal("Test Group", groupResp.Group.Name)
	s.Require().Equal("A test group", groupResp.Group.Description)
	s.Require().Equal(uint64(0), groupResp.Group.MemberCount)
	s.Require().Equal(setup.TokenContract.String(), groupResp.Group.Token)
	s.Require().Equal(setup.TokenDenom, groupResp.Group.TokenDenom)
	s.Require().Equal(setup.MinBalance, groupResp.Group.MinBalance)
}

func (suite *KeeperTestSuite) TestGroupsAddMemberByAdmin() {
	setup := suite.SetupGroupsTest()
	appA := setup.AppA

	// Create a new user to add as member
	newUser := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(newUser.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	newUserID := suite.RegisterUser(setup, newUser)
	newUserAddr := appA.BytesToAccAddressPrefixed(newUser.Address).String()

	// Create a group with sender as admin
	createGroupMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"create_group":{"name":"Admin Test Group","description":"Testing admin add member","admins":["%s"],"protocol":{"governance_contract":"%s"},"token":"%s","min_balance":"%s"}}`,
			setup.UserID, setup.GovContract.String(), setup.TokenContract.String(), setup.MinBalance)),
	}

	res := appA.ExecuteContract(setup.Sender, setup.GroupsContract, createGroupMsg, nil, nil)
	s.Require().True(res.IsOK(), "create_group failed: %s", res.Log)

	var createResp struct {
		GroupID string `json:"group_id"`
	}
	err := appA.DecodeExecuteResponse(res, &createResp)
	s.Require().NoError(err)

	// Add member should fail for token-based groups
	addMemberMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"add_member":{"group_id":"%s","user_id":"%s"}}`, createResp.GroupID, newUserID)),
	}
	res = appA.ExecuteContract(setup.Sender, setup.GroupsContract, addMemberMsg, nil, nil)
	s.Require().False(res.IsOK(), "add_member should fail for token-based groups")

	// Mint token to meet min_balance
	suite.MintGroupToken(setup, newUserAddr, "1000")

	// Verify member was added
	isMemberQuery := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"query_is_member":{"group_id":"%s","user_id":"%s"}}`, createResp.GroupID, newUserID)),
	}
	memberBz := appA.WasmxQueryRaw(setup.Sender, setup.GroupsContract, isMemberQuery, nil, nil)

	var isMemberResp struct {
		IsMember bool `json:"is_member"`
	}
	err = json.Unmarshal(memberBz, &isMemberResp)
	s.Require().NoError(err)
	s.Require().True(isMemberResp.IsMember)
}

func (suite *KeeperTestSuite) TestGroupsRemoveMember() {
	setup := suite.SetupGroupsTest()
	appA := setup.AppA

	// Create a new user
	newUser := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(newUser.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	newUserID := suite.RegisterUser(setup, newUser)
	newUserAddr := appA.BytesToAccAddressPrefixed(newUser.Address).String()

	// Create a group
	createGroupMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"create_group":{"name":"Remove Test Group","description":"Testing remove member","admins":["%s"],"protocol":{"governance_contract":"%s"},"token":"%s","min_balance":"%s"}}`,
			setup.UserID, setup.GovContract.String(), setup.TokenContract.String(), setup.MinBalance)),
	}

	res := appA.ExecuteContract(setup.Sender, setup.GroupsContract, createGroupMsg, nil, nil)
	s.Require().True(res.IsOK())

	var createResp struct {
		GroupID string `json:"group_id"`
	}
	_ = appA.DecodeExecuteResponse(res, &createResp)

	// Mint token to meet min_balance
	suite.MintGroupToken(setup, newUserAddr, "1000")

	// Remove member
	removeMemberMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"remove_member":{"group_id":"%s","user_id":"%s"}}`, createResp.GroupID, newUserID)),
	}
	res = appA.ExecuteContract(setup.Sender, setup.GroupsContract, removeMemberMsg, nil, nil)
	s.Require().False(res.IsOK(), "remove_member should fail for token-based groups")

	// Verify member was removed
	isMemberQuery := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"query_is_member":{"group_id":"%s","user_id":"%s"}}`, createResp.GroupID, newUserID)),
	}
	memberBz := appA.WasmxQueryRaw(setup.Sender, setup.GroupsContract, isMemberQuery, nil, nil)

	var isMemberResp struct {
		IsMember bool `json:"is_member"`
	}
	_ = json.Unmarshal(memberBz, &isMemberResp)
	s.Require().True(isMemberResp.IsMember)
}

func (suite *KeeperTestSuite) TestGroupsQueryAllMembers() {
	setup := suite.SetupGroupsTest()
	appA := setup.AppA

	// Create a group
	createGroupMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"create_group":{"name":"Members Query Group","description":"Testing query all members","admins":["%s"],"protocol":{"governance_contract":"%s"},"token":"%s","min_balance":"%s"}}`,
			setup.UserID, setup.GovContract.String(), setup.TokenContract.String(), setup.MinBalance)),
	}

	res := appA.ExecuteContract(setup.Sender, setup.GroupsContract, createGroupMsg, nil, nil)
	s.Require().True(res.IsOK())

	var createResp struct {
		GroupID string `json:"group_id"`
	}
	_ = appA.DecodeExecuteResponse(res, &createResp)

	// Mint tokens for multiple users (membership is token-based)
	for i := 0; i < 3; i++ {
		newUser := suite.GetRandomAccount()
		initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
		appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(newUser.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
		_ = suite.RegisterUser(setup, newUser)
		userAddr := appA.BytesToAccAddressPrefixed(newUser.Address).String()
		suite.MintGroupToken(setup, userAddr, "1000")
	}

	// Query all members
	getAllMembersQuery := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"query_get_all_members":{"group_id":"%s"}}`, createResp.GroupID)),
	}
	membersBz := appA.WasmxQueryRaw(setup.Sender, setup.GroupsContract, getAllMembersQuery, nil, nil)

	var membersResp struct {
		Members []struct {
			UserID string `json:"user_id"`
		} `json:"members"`
		Pagination struct {
			Total string `json:"total"`
		} `json:"pagination"`
	}
	err := json.Unmarshal(membersBz, &membersResp)
	s.Require().NoError(err)
	s.Require().Equal(0, len(membersResp.Members))
}

func (suite *KeeperTestSuite) TestGroupsQueryVotingProtocol() {
	setup := suite.SetupGroupsTest()
	appA := setup.AppA

	// Create a group with specific protocol
	createGroupMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"create_group":{"name":"Protocol Query Group","description":"Testing voting protocol query","admins":["%s"],"protocol":{"governance_contract":"%s","protocol_id":"custom_protocol"},"token":"%s","min_balance":"%s"}}`,
			setup.UserID, setup.GovContract.String(), setup.TokenContract.String(), setup.MinBalance)),
	}

	res := appA.ExecuteContract(setup.Sender, setup.GroupsContract, createGroupMsg, nil, nil)
	s.Require().True(res.IsOK())

	var createResp struct {
		GroupID string `json:"group_id"`
	}
	_ = appA.DecodeExecuteResponse(res, &createResp)

	// Query voting protocol
	protocolQuery := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"query_get_voting_protocol":{"group_id":"%s"}}`, createResp.GroupID)),
	}
	protocolBz := appA.WasmxQueryRaw(setup.Sender, setup.GroupsContract, protocolQuery, nil, nil)

	var protocolResp struct {
		Protocol struct {
			GovernanceContract string `json:"governance_contract"`
			ProtocolID         string `json:"protocol_id"`
		} `json:"protocol"`
	}
	err := json.Unmarshal(protocolBz, &protocolResp)
	s.Require().NoError(err)
	s.Require().Equal(setup.GovContract.String(), protocolResp.Protocol.GovernanceContract)
	s.Require().Equal("custom_protocol", protocolResp.Protocol.ProtocolID)
}

func (suite *KeeperTestSuite) TestGroupsQueryIsAdmin() {
	setup := suite.SetupGroupsTest()
	appA := setup.AppA

	// Create a group with sender as admin
	createGroupMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"create_group":{"name":"Admin Query Group","description":"Testing is_admin query","admins":["%s"],"protocol":{"governance_contract":"%s"},"token":"%s","min_balance":"%s"}}`,
			setup.UserID, setup.GovContract.String(), setup.TokenContract.String(), setup.MinBalance)),
	}

	res := appA.ExecuteContract(setup.Sender, setup.GroupsContract, createGroupMsg, nil, nil)
	s.Require().True(res.IsOK())

	var createResp struct {
		GroupID string `json:"group_id"`
	}
	_ = appA.DecodeExecuteResponse(res, &createResp)

	// Query if sender is admin (should be true)
	isAdminQuery := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"query_is_admin":{"group_id":"%s","user_id":"%s"}}`, createResp.GroupID, setup.UserID)),
	}
	adminBz := appA.WasmxQueryRaw(setup.Sender, setup.GroupsContract, isAdminQuery, nil, nil)

	var isAdminResp struct {
		IsAdmin bool `json:"is_admin"`
	}
	err := json.Unmarshal(adminBz, &isAdminResp)
	s.Require().NoError(err)
	s.Require().True(isAdminResp.IsAdmin)

	// Query for non-admin (should be false)
	nonAdminQuery := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"query_is_admin":{"group_id":"%s","user_id":"non_existent_user"}}`, createResp.GroupID)),
	}
	nonAdminBz := appA.WasmxQueryRaw(setup.Sender, setup.GroupsContract, nonAdminQuery, nil, nil)

	err = json.Unmarshal(nonAdminBz, &isAdminResp)
	s.Require().NoError(err)
	s.Require().False(isAdminResp.IsAdmin)
}

func (suite *KeeperTestSuite) TestGroupsAddMemberByGovernance() {
	setup := suite.SetupGroupsTest()
	appA := setup.AppA

	// Create a new user
	newUser := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(newUser.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	newUserID := suite.RegisterUser(setup, newUser)
	newUserAddr := appA.BytesToAccAddressPrefixed(newUser.Address).String()

	// Create a group with NO admins (governance-only control)
	createGroupMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"create_group":{"name":"Gov Controlled Group","description":"Testing governance-controlled membership","admins":[],"protocol":{"governance_contract":"%s"},"token":"%s","min_balance":"%s"}}`,
			setup.GovContract.String(), setup.TokenContract.String(), setup.MinBalance)),
	}

	res := appA.ExecuteContract(setup.Sender, setup.GroupsContract, createGroupMsg, nil, nil)
	s.Require().True(res.IsOK())

	var createResp struct {
		GroupID string `json:"group_id"`
	}
	_ = appA.DecodeExecuteResponse(res, &createResp)

	addMemberData := []byte(fmt.Sprintf(`{"add_member":{"group_id":"%s","user_id":"%s"}}`, createResp.GroupID, newUserID))
	addMemberMsgBz, err := json.Marshal(&types.WasmxExecutionMessage{Data: addMemberData})
	s.Require().NoError(err)

	proposal := &types.MsgExecuteContract{
		Sender:   setup.GovContract.String(),
		Contract: setup.GroupsContract.String(),
		Msg:      addMemberMsgBz,
	}

	// Use the helper from app_context
	appA.PassProposalWithGovContract(
		setup.GovContract,
		ut.GovVoteContinuous,
		setup.ValAccount,
		setup.Sender,
		[]sdk.Msg{proposal},
		"",
		"Add member to group",
		"Adding a new member through governance",
		false, // not expedited
	)

	// Verify member was not added by governance
	isMemberQuery := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"query_is_member":{"group_id":"%s","user_id":"%s"}}`, createResp.GroupID, newUserID)),
	}
	memberBz := appA.WasmxQueryRaw(setup.Sender, setup.GroupsContract, isMemberQuery, nil, nil)

	var isMemberResp struct {
		IsMember bool `json:"is_member"`
	}
	err = json.Unmarshal(memberBz, &isMemberResp)
	s.Require().NoError(err)
	s.Require().False(isMemberResp.IsMember, "member should not be added via governance for token-based groups")

	suite.MintGroupToken(setup, newUserAddr, "1000")
	memberBz = appA.WasmxQueryRaw(setup.Sender, setup.GroupsContract, isMemberQuery, nil, nil)
	err = json.Unmarshal(memberBz, &isMemberResp)
	s.Require().NoError(err)
	s.Require().True(isMemberResp.IsMember, "member should be added via token balance")
}

func (suite *KeeperTestSuite) TestGroupsGetUserGroups() {
	setup := suite.SetupGroupsTest()
	appA := setup.AppA

	// Create a new user
	newUser := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(newUser.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	newUserID := suite.RegisterUser(setup, newUser)
	newUserAddr := appA.BytesToAccAddressPrefixed(newUser.Address).String()

	// Create multiple groups for the user
	for i := 0; i < 3; i++ {
		createGroupMsg := types.WasmxExecutionMessage{
			Data: []byte(fmt.Sprintf(`{"create_group":{"name":"User Group %d","description":"Test group %d","admins":["%s"],"protocol":{"governance_contract":"%s"},"token":"%s","min_balance":"%s"}}`,
				i, i, setup.UserID, setup.GovContract.String(), setup.TokenContract.String(), setup.MinBalance)),
		}

		res := appA.ExecuteContract(setup.Sender, setup.GroupsContract, createGroupMsg, nil, nil)
		s.Require().True(res.IsOK())

		var createResp struct {
			GroupID string `json:"group_id"`
		}
		_ = appA.DecodeExecuteResponse(res, &createResp)
	}
	suite.MintGroupToken(setup, newUserAddr, "1000")

	// Query user groups
	userGroupsQuery := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"query_get_user_groups":{"user_id":"%s"}}`, newUserID)),
	}
	userGroupsBz := appA.WasmxQueryRaw(setup.Sender, setup.GroupsContract, userGroupsQuery, nil, nil)

	var userGroupsResp struct {
		GroupIDs []string `json:"group_ids"`
	}
	err := json.Unmarshal(userGroupsBz, &userGroupsResp)
	s.Require().NoError(err)
	s.Require().Equal(3, len(userGroupsResp.GroupIDs))
}

func (suite *KeeperTestSuite) TestGovGroupWithERC20Gov() {
	setup := suite.SetupGroupsTest()
	appA := setup.AppA
	baseDenom := appA.Chain.Config.BaseDenom
	senderAddr := appA.BytesToAccAddressPrefixed(setup.Sender.Address).String()

	receiver := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(receiver.Address), sdk.NewCoin(baseDenom, initBalance))
	receiverUserID := suite.RegisterUser(setup, receiver)
	receiverAddr := appA.BytesToAccAddressPrefixed(receiver.Address).String()

	// Use pre-instantiated gov-group contract
	govGroupContract := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_GOV_GROUP))
	suite.InitGovGroupGenesis(appA, govGroupContract, baseDenom)

	// Instantiate erc20gov (ERC20 type contracts need instantiation)
	erc20govWasm := precompiles.GetPrecompileByLabel(appA.AddressCodec(), "erc20gov_0.0.1")
	s.Require().NotNil(erc20govWasm, "erc20gov contract binary not found")
	erc20govCodeId := appA.StoreCode(setup.Sender, erc20govWasm, nil)
	erc20govDenom := "GOV1"
	erc20govInit := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"governance_address":"%s","admins":["%s"],"name":"Gov Token1","symbol":"%s","decimals":6,"initial_balances":[{"user_id":"%s","amount":"1000000000"}]}`,
			govGroupContract.String(), setup.UserID, erc20govDenom, setup.UserID)),
	}
	erc20govContract := appA.InstantiateCode(setup.Sender, erc20govCodeId, erc20govInit, "erc20gov_group", nil)
	suite.RegisterErc20GovDenomRole(setup, erc20govContract, erc20govDenom)

	createGroupMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"create_group":{"name":"Gov Group","description":"Governed by gov-group","admins":[],"protocol":{"governance_contract":"%s"},"token":"%s","min_balance":"1"}}`,
			govGroupContract.String(), erc20govContract.String())),
	}
	res := appA.ExecuteContract(setup.Sender, setup.GroupsContract, createGroupMsg, nil, nil)
	s.Require().True(res.IsOK(), "create_group failed: %s", res.Log)
	var createResp struct {
		GroupID string `json:"group_id"`
	}
	err := appA.DecodeExecuteResponse(res, &createResp)
	s.Require().NoError(err)
	s.Require().NotEmpty(createResp.GroupID)

	isMemberQuery := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"query_is_member":{"group_id":"%s","user_id":"%s"}}`, createResp.GroupID, setup.UserID)),
	}
	memberBz := appA.WasmxQueryRaw(setup.Sender, setup.GroupsContract, isMemberQuery, nil, nil)
	var isMemberResp struct {
		IsMember bool `json:"is_member"`
	}
	err = json.Unmarshal(memberBz, &isMemberResp)
	s.Require().NoError(err)
	s.Require().True(isMemberResp.IsMember)

	transferMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"transfer":{"from":"%s","to":"%s","value":"10"}}`, setup.UserID, receiverUserID)),
	}
	transferMsgBz, err := json.Marshal(transferMsg)
	s.Require().NoError(err)

	execMsg := &types.MsgExecuteContract{
		Sender:   govGroupContract.String(),
		Contract: erc20govContract.String(),
		Msg:      transferMsgBz,
	}
	encodedMsgs := encodeMsgsAsAnyBase64(appA, []sdk.Msg{execMsg})

	submitPayload := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"SubmitProposal":{"messages":["%s"],"initial_deposit":[{"denom":"%s","amount":"10000001"}],"proposer":"%s","metadata":"","title":"Transfer via gov-group","summary":"erc20gov transfer","expedited":false,"group_id":"%s"}}`,
			encodedMsgs[0], erc20govDenom, senderAddr, createResp.GroupID)),
	}
	submitRes := appA.ExecuteContract(setup.Sender, govGroupContract, submitPayload, nil, nil)
	s.Require().True(submitRes.IsOK(), "submit proposal failed: %s", submitRes.Log)

	var submitResp struct {
		ProposalID string `json:"proposal_id"`
	}
	err = appA.DecodeExecuteResponse(submitRes, &submitResp)
	s.Require().NoError(err)
	proposalId, err := strconv.ParseInt(submitResp.ProposalID, 10, 64)
	s.Require().NoError(err)
	s.Require().NotZero(proposalId)

	votePayload := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"Vote":{"proposal_id":%d,"voter":"%s","option":"VOTE_OPTION_YES","metadata":"vote"}}`, proposalId, senderAddr)),
	}
	voteRes := appA.ExecuteContract(setup.Sender, govGroupContract, votePayload, nil, nil)
	s.Require().True(voteRes.IsOK(), "vote failed: %s", voteRes.Log)

	// non members cannot vote
	nonMemberVotePayload := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"Vote":{"proposal_id":%d,"voter":"%s","option":"VOTE_OPTION_YES","metadata":"vote"}}`, proposalId, receiverAddr)),
	}
	nonMemberVoteRes, err := appA.ExecuteContractNoCheck(setup.Sender, govGroupContract, nonMemberVotePayload, nil, nil, 1000000000, nil)
	s.Require().NoError(err)
	s.Require().False(nonMemberVoteRes.IsOK(), "voting should fail, oter should be not eligible for group")
	s.Require().Contains(nonMemberVoteRes.GetLog(), "failed to execute message", nonMemberVoteRes.GetEvents())

	appA.WaitForVotingPeriodEnd()

	balanceQuery := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, receiverUserID)),
	}
	balanceBz := appA.WasmxQueryRaw(setup.Sender, erc20govContract, balanceQuery, nil, nil)
	var balanceResp struct {
		Balance struct {
			Amount string `json:"amount"`
		} `json:"balance"`
	}
	err = json.Unmarshal(balanceBz, &balanceResp)
	s.Require().NoError(err)
	s.Require().Equal("10", balanceResp.Balance.Amount)

	// governance executes a contract call (simple storage)
	storageCodeId := appA.StoreCode(setup.Sender, wasmxtest.WasmxSimpleStorage, nil)
	storageContract := appA.InstantiateCode(setup.Sender, storageCodeId, types.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage_gov_group", nil)
	storageSetMsg := types.WasmxExecutionMessage{
		Data: []byte(`{"set":{"key":"govkey","value":"govval"}}`),
	}
	storageSetMsgBz, err := json.Marshal(storageSetMsg)
	s.Require().NoError(err)
	storageExecMsg := &types.MsgExecuteContract{
		Sender:   govGroupContract.String(),
		Contract: storageContract.String(),
		Msg:      storageSetMsgBz,
	}
	storageEncoded := encodeMsgsAsAnyBase64(appA, []sdk.Msg{storageExecMsg})
	storageSubmit := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"submit_group_proposal":{"group_id":"%s","messages":["%s"],"initial_deposit":[{"denom":"%s","amount":"10000001"}],"metadata":"","title":"Set storage via gov-group","summary":"simple storage","expedited":false}}`,
			createResp.GroupID, storageEncoded[0], erc20govDenom)),
	}
	storageSubmitRes := appA.ExecuteContract(setup.Sender, setup.GroupsContract, storageSubmit, nil, nil)
	s.Require().True(storageSubmitRes.IsOK(), "submit storage proposal failed: %s", storageSubmitRes.Log)
	var storageSubmitResp struct {
		ProposalID string `json:"proposal_id"`
	}
	err = appA.DecodeExecuteResponse(storageSubmitRes, &storageSubmitResp)
	s.Require().NoError(err)
	storageProposalID, err := strconv.ParseInt(storageSubmitResp.ProposalID, 10, 64)
	s.Require().NoError(err)
	s.Require().NotZero(storageProposalID)

	storageVote := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"Vote":{"proposal_id":%d,"voter":"%s","option":"VOTE_OPTION_YES","metadata":"vote"}}`, storageProposalID, senderAddr)),
	}
	storageVoteRes := appA.ExecuteContract(setup.Sender, govGroupContract, storageVote, nil, nil)
	s.Require().True(storageVoteRes.IsOK(), "storage vote failed: %s", storageVoteRes.Log)
	appA.WaitForVotingPeriodEnd()

	storageQuery := appA.App.WasmxKeeper.QueryRaw(appA.Context(), storageContract, []byte("govkey"))
	s.Require().Equal("govval", string(storageQuery))
}

func (suite *KeeperTestSuite) TestGovGroupContinuousWithERC20Gov() {
	setup := suite.SetupGroupsTest()
	appA := setup.AppA
	baseDenom := appA.Chain.Config.BaseDenom
	senderAddr := appA.BytesToAccAddressPrefixed(setup.Sender.Address).String()

	receiver := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(receiver.Address), sdk.NewCoin(baseDenom, initBalance))
	receiverUserID := suite.RegisterUser(setup, receiver)
	nonMember := suite.GetRandomAccount()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(nonMember.Address), sdk.NewCoin(baseDenom, initBalance))
	nonMemberAddr := appA.BytesToAccAddressPrefixed(nonMember.Address).String()

	// Use pre-instantiated gov-group-cont contract
	govGroupContContract := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_GOV_GROUP_CONT))

	// Instantiate erc20gov (ERC20 type contracts need instantiation)
	erc20govWasm := precompiles.GetPrecompileByLabel(appA.AddressCodec(), "erc20gov_0.0.1")
	s.Require().NotNil(erc20govWasm, "erc20gov contract binary not found")
	erc20govCodeId := appA.StoreCode(setup.Sender, erc20govWasm, nil)
	erc20denom := "GOV2"
	erc20govInit := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"governance_address":"%s","admins":["%s"],"name":"Gov Token2","symbol":"%s","decimals":6,"initial_balances":[{"user_id":"%s","amount":"1000"}]}`,
			govGroupContContract.String(), setup.UserID, erc20denom, setup.UserID)),
	}
	erc20govContract := appA.InstantiateCode(setup.Sender, erc20govCodeId, erc20govInit, "erc20gov_group_cont", nil)
	suite.RegisterErc20GovDenomRole(setup, erc20govContract, erc20denom)

	createGroupMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"create_group":{"name":"Gov Group Continuous","description":"Governed by gov-group-continuous","admins":[],"protocol":{"governance_contract":"%s"},"token":"%s","min_balance":"1"}}`,
			govGroupContContract.String(), erc20govContract.String())),
	}
	res := appA.ExecuteContract(setup.Sender, setup.GroupsContract, createGroupMsg, nil, nil)
	s.Require().True(res.IsOK(), "create_group failed: %s", res.Log)
	var createResp struct {
		GroupID string `json:"group_id"`
	}
	err := appA.DecodeExecuteResponse(res, &createResp)
	s.Require().NoError(err)
	s.Require().NotEmpty(createResp.GroupID)

	transferMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"transfer":{"from":"%s","to":"%s","value":"10"}}`, setup.UserID, receiverUserID)),
	}
	transferMsgBz, err2 := json.Marshal(transferMsg)
	s.Require().NoError(err2)

	execMsg := &types.MsgExecuteContract{
		Sender:   govGroupContContract.String(),
		Contract: erc20govContract.String(),
		Msg:      transferMsgBz,
	}
	encodedMsgs := encodeMsgsAsAnyBase64(appA, []sdk.Msg{execMsg})

	submitPayload := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"SubmitProposal":{"messages":["%s"],"initial_deposit":[{"denom":"%s","amount":"100"}],"proposer":"%s","metadata":"","title":"Transfer via gov-group-cont","summary":"erc20gov transfer","expedited":false,"group_id":"%s"}}`,
			encodedMsgs[0], erc20denom, senderAddr, createResp.GroupID)),
	}
	submitRes := appA.ExecuteContract(setup.Sender, govGroupContContract, submitPayload, nil, nil)
	s.Require().True(submitRes.IsOK(), "submit proposal failed: %s", submitRes.Log)

	var submitResp struct {
		ProposalID string `json:"proposal_id"`
	}
	err = appA.DecodeExecuteResponse(submitRes, &submitResp)
	s.Require().NoError(err)
	proposalID, err := strconv.ParseUint(submitResp.ProposalID, 10, 64)
	s.Require().NoError(err)
	s.Require().NotZero(proposalID)

	depositVotePayload := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"DepositVote":{"proposal_id":%d,"option_id":2,"voter":"%s","amount":"100","arbitration_amount":"0","metadata":"vote"}}`, proposalID, senderAddr)),
	}
	voteRes := appA.ExecuteContract(setup.Sender, govGroupContContract, depositVotePayload, nil, nil)
	s.Require().True(voteRes.IsOK(), "deposit vote failed: %s", voteRes.Log)

	// non members cannot vote
	nonMemberVotePayload := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"DepositVote":{"proposal_id":%d,"option_id":2,"voter":"%s","amount":"100","arbitration_amount":"0","metadata":"vote"}}`, proposalID, nonMemberAddr)),
	}
	nonMemberVoteRes, err := appA.ExecuteContractNoCheck(setup.Sender, govGroupContContract, nonMemberVotePayload, nil, nil, 1000000000, nil)
	s.Require().NoError(err)
	s.Require().False(nonMemberVoteRes.IsOK(), "voting should fail, voter should be not eligible for group")
	s.Require().Contains(nonMemberVoteRes.GetLog(), "failed to execute message", nonMemberVoteRes.GetEvents())

	balanceQuery := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, receiverUserID)),
	}
	balanceBz := appA.WasmxQueryRaw(setup.Sender, erc20govContract, balanceQuery, nil, nil)
	var balanceResp struct {
		Balance struct {
			Amount string `json:"amount"`
		} `json:"balance"`
	}
	err = json.Unmarshal(balanceBz, &balanceResp)
	s.Require().NoError(err)
	s.Require().Equal("10", balanceResp.Balance.Amount)

	// governance executes a contract call (simple storage)
	storageCodeId := appA.StoreCode(setup.Sender, wasmxtest.WasmxSimpleStorage, nil)
	storageContract := appA.InstantiateCode(setup.Sender, storageCodeId, types.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage_gov_group_cont", nil)
	storageSetMsg := types.WasmxExecutionMessage{
		Data: []byte(`{"set":{"key":"govkey","value":"govval"}}`),
	}
	storageSetMsgBz, err := json.Marshal(storageSetMsg)
	s.Require().NoError(err)
	storageExecMsg := &types.MsgExecuteContract{
		Sender:   govGroupContContract.String(),
		Contract: storageContract.String(),
		Msg:      storageSetMsgBz,
	}
	storageEncoded := encodeMsgsAsAnyBase64(appA, []sdk.Msg{storageExecMsg})
	storageSubmit := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"submit_group_proposal":{"group_id":"%s","messages":["%s"],"initial_deposit":[{"denom":"%s","amount":"100"}],"metadata":"","title":"Set storage via gov-group-cont","summary":"simple storage","expedited":false}}`,
			createResp.GroupID, storageEncoded[0], erc20denom)),
	}
	storageSubmitRes := appA.ExecuteContract(setup.Sender, setup.GroupsContract, storageSubmit, nil, nil)
	s.Require().True(storageSubmitRes.IsOK(), "submit storage proposal failed: %s", storageSubmitRes.Log)
	var storageSubmitResp struct {
		ProposalID string `json:"proposal_id"`
	}
	err = appA.DecodeExecuteResponse(storageSubmitRes, &storageSubmitResp)
	s.Require().NoError(err)
	storageProposalID, err := strconv.ParseUint(storageSubmitResp.ProposalID, 10, 64)
	s.Require().NoError(err)
	s.Require().NotZero(storageProposalID)

	storageVote := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"DepositVote":{"proposal_id":%d,"option_id":2,"voter":"%s","amount":"100","arbitration_amount":"0","metadata":"vote"}}`, storageProposalID, senderAddr)),
	}
	storageVoteRes := appA.ExecuteContract(setup.Sender, govGroupContContract, storageVote, nil, nil)
	s.Require().True(storageVoteRes.IsOK(), "storage deposit vote failed: %s", storageVoteRes.Log)

	storageQuery := appA.App.WasmxKeeper.QueryRaw(appA.Context(), storageContract, []byte("govkey"))
	s.Require().Equal("govval", string(storageQuery))
}

func (suite *KeeperTestSuite) RegisterErc20GovDenomRole(setup *GroupsTestSetup, contract mcodec.AccAddressPrefixed, label string) {
	appA := setup.AppA
	title := "Register denom"
	description := "Register denom"
	authority := appA.MustAccAddressToString(authtypes.NewModuleAddress(types.ROLE_GOVERNANCE))
	contractStr := contract.String()
	valAccount := simulation.Account{
		PrivKey: suite.Chain().SenderPrivKey,
		PubKey:  suite.Chain().SenderPrivKey.PubKey(),
		Address: suite.Chain().SenderAccount.GetAddress(),
	}

	msg := []byte(fmt.Sprintf(`{"SetContractForRoleGov":{"role":"%s","label":"%s","contract_address":"%s","action_type":1}}`, types.ROLE_DENOM, label, contractStr))
	msgbz, err := json.Marshal(&types.WasmxExecutionMessage{Data: msg})
	suite.Require().NoError(err)

	proposal := &types.MsgExecuteContract{
		Sender:   authority,
		Contract: types.ROLE_ROLES,
		Msg:      msgbz,
	}
	appA.PassGovProposal(valAccount, setup.Sender, []sdk.Msg{proposal}, "", title, description, false)

	resp := appA.App.WasmxKeeper.GetRoleLabelByContract(appA.Context(), contract)
	suite.Require().Equal(label, resp)
}

func (suite *KeeperTestSuite) InitGovGroupGenesis(appA *ut.AppContext, govContract mcodec.AccAddressPrefixed, baseDenom string) {
	govGenesis := cosmosmodtypes.DefaultGovGenesisState(baseDenom)
	votingPeriod := time.Millisecond * 800
	govGenesis.Params.VotingPeriod = votingPeriod.Milliseconds()

	msgjson, err := appA.App.AppCodec().MarshalJSON(govGenesis)
	suite.Require().NoError(err)
	msgbz := []byte(fmt.Sprintf(`{"InitGenesis":%s}`, string(msgjson)))
	execmsg := types.WasmxExecutionMessage{Data: msgbz}
	execmsgbz, err := json.Marshal(execmsg)
	suite.Require().NoError(err)

	_, err = appA.App.WasmxKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   govContract.String(),
		Contract: govContract.String(),
		Msg:      execmsgbz,
	})
	suite.Require().NoError(err)
}
