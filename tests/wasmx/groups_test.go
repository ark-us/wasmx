package keeper_test

import (
	_ "embed"
	"encoding/base64"
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

	// Store and instantiate wasmx-groups contract
	groupsWasm := precompiles.GetPrecompileByLabel(appA.AddressCodec(), "groups_0.0.1")
	s.Require().NotNil(groupsWasm, "groups contract binary not found")

	codeId := appA.StoreCode(sender, groupsWasm, nil)

	// Initialize groups contract with identity contract reference
	initMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"identity_contract":"%s","groups":[]}`, identityAddr.String())),
	}
	groupsContract := appA.InstantiateCode(sender, codeId, initMsg, "wasmx_groups", nil)

	return &GroupsTestSetup{
		AppA:           appA,
		Sender:         sender,
		ValAccount:     valAccount,
		GroupsContract: groupsContract,
		IdentityAddr:   identityAddr,
		GovContract:    govAddr,
		UserID:         registerResp.UserID,
	}
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
		Config struct {
			IdentityContract string `json:"identity_contract"`
		} `json:"config"`
	}
	err := json.Unmarshal(configBz, &configResp)
	s.Require().NoError(err)
	s.Require().Equal(setup.IdentityAddr.String(), configResp.Config.IdentityContract)
}

func (suite *KeeperTestSuite) TestGroupsCreateGroup() {
	setup := suite.SetupGroupsTest()
	appA := setup.AppA

	// Create a group
	createGroupMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"create_group":{"name":"Test Group","description":"A test group","admins":["%s"],"protocol":{"governance_contract":"%s"}}}`,
			setup.UserID, setup.GovContract.String())),
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
		} `json:"group"`
	}
	err = json.Unmarshal(groupBz, &groupResp)
	s.Require().NoError(err)
	s.Require().Equal("Test Group", groupResp.Group.Name)
	s.Require().Equal("A test group", groupResp.Group.Description)
	s.Require().Equal(uint64(0), groupResp.Group.MemberCount)
}

func (suite *KeeperTestSuite) TestGroupsAddMemberByAdmin() {
	setup := suite.SetupGroupsTest()
	appA := setup.AppA

	// Create a new user to add as member
	newUser := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(newUser.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	newUserID := suite.RegisterUser(setup, newUser)

	// Create a group with sender as admin
	createGroupMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"create_group":{"name":"Admin Test Group","description":"Testing admin add member","admins":["%s"],"protocol":{"governance_contract":"%s"}}}`,
			setup.UserID, setup.GovContract.String())),
	}

	res := appA.ExecuteContract(setup.Sender, setup.GroupsContract, createGroupMsg, nil, nil)
	s.Require().True(res.IsOK(), "create_group failed: %s", res.Log)

	var createResp struct {
		GroupID string `json:"group_id"`
	}
	err := appA.DecodeExecuteResponse(res, &createResp)
	s.Require().NoError(err)

	// Add member by admin
	addMemberMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"add_member":{"group_id":"%s","user_id":"%s"}}`, createResp.GroupID, newUserID)),
	}

	res = appA.ExecuteContract(setup.Sender, setup.GroupsContract, addMemberMsg, nil, nil)
	s.Require().True(res.IsOK(), "add_member failed: %s", res.Log)

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

	// Create a group
	createGroupMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"create_group":{"name":"Remove Test Group","description":"Testing remove member","admins":["%s"],"protocol":{"governance_contract":"%s"}}}`,
			setup.UserID, setup.GovContract.String())),
	}

	res := appA.ExecuteContract(setup.Sender, setup.GroupsContract, createGroupMsg, nil, nil)
	s.Require().True(res.IsOK())

	var createResp struct {
		GroupID string `json:"group_id"`
	}
	_ = appA.DecodeExecuteResponse(res, &createResp)

	// Add member
	addMemberMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"add_member":{"group_id":"%s","user_id":"%s"}}`, createResp.GroupID, newUserID)),
	}
	res = appA.ExecuteContract(setup.Sender, setup.GroupsContract, addMemberMsg, nil, nil)
	s.Require().True(res.IsOK())

	// Remove member
	removeMemberMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"remove_member":{"group_id":"%s","user_id":"%s"}}`, createResp.GroupID, newUserID)),
	}
	res = appA.ExecuteContract(setup.Sender, setup.GroupsContract, removeMemberMsg, nil, nil)
	s.Require().True(res.IsOK(), "remove_member failed: %s", res.Log)

	// Verify member was removed
	isMemberQuery := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"query_is_member":{"group_id":"%s","user_id":"%s"}}`, createResp.GroupID, newUserID)),
	}
	memberBz := appA.WasmxQueryRaw(setup.Sender, setup.GroupsContract, isMemberQuery, nil, nil)

	var isMemberResp struct {
		IsMember bool `json:"is_member"`
	}
	_ = json.Unmarshal(memberBz, &isMemberResp)
	s.Require().False(isMemberResp.IsMember)
}

func (suite *KeeperTestSuite) TestGroupsQueryAllMembers() {
	setup := suite.SetupGroupsTest()
	appA := setup.AppA

	// Create a group
	createGroupMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"create_group":{"name":"Members Query Group","description":"Testing query all members","admins":["%s"],"protocol":{"governance_contract":"%s"}}}`,
			setup.UserID, setup.GovContract.String())),
	}

	res := appA.ExecuteContract(setup.Sender, setup.GroupsContract, createGroupMsg, nil, nil)
	s.Require().True(res.IsOK())

	var createResp struct {
		GroupID string `json:"group_id"`
	}
	_ = appA.DecodeExecuteResponse(res, &createResp)

	// Add multiple members
	for i := 0; i < 3; i++ {
		newUser := suite.GetRandomAccount()
		initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
		appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(newUser.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
		userID := suite.RegisterUser(setup, newUser)

		addMemberMsg := types.WasmxExecutionMessage{
			Data: []byte(fmt.Sprintf(`{"add_member":{"group_id":"%s","user_id":"%s"}}`, createResp.GroupID, userID)),
		}
		res = appA.ExecuteContract(setup.Sender, setup.GroupsContract, addMemberMsg, nil, nil)
		s.Require().True(res.IsOK())
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
	s.Require().Equal(3, len(membersResp.Members))
}

func (suite *KeeperTestSuite) TestGroupsQueryVotingProtocol() {
	setup := suite.SetupGroupsTest()
	appA := setup.AppA

	// Create a group with specific protocol
	createGroupMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"create_group":{"name":"Protocol Query Group","description":"Testing voting protocol query","admins":["%s"],"protocol":{"governance_contract":"%s","protocol_id":"custom_protocol"}}}`,
			setup.UserID, setup.GovContract.String())),
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
		Data: []byte(fmt.Sprintf(`{"create_group":{"name":"Admin Query Group","description":"Testing is_admin query","admins":["%s"],"protocol":{"governance_contract":"%s"}}}`,
			setup.UserID, setup.GovContract.String())),
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

	// Create a group with NO admins (governance-only control)
	createGroupMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"create_group":{"name":"Gov Controlled Group","description":"Testing governance-controlled membership","admins":[],"protocol":{"governance_contract":"%s"}}}`,
			setup.GovContract.String())),
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
	s.Require().True(isMemberResp.IsMember, "member should have been added via governance")
}

func (suite *KeeperTestSuite) TestGroupsGetUserGroups() {
	setup := suite.SetupGroupsTest()
	appA := setup.AppA

	// Create a new user
	newUser := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(newUser.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	newUserID := suite.RegisterUser(setup, newUser)

	// Create multiple groups and add user to them
	groupIDs := []string{}
	for i := 0; i < 3; i++ {
		createGroupMsg := types.WasmxExecutionMessage{
			Data: []byte(fmt.Sprintf(`{"create_group":{"name":"User Group %d","description":"Test group %d","admins":["%s"],"protocol":{"governance_contract":"%s"}}}`,
				i, i, setup.UserID, setup.GovContract.String())),
		}

		res := appA.ExecuteContract(setup.Sender, setup.GroupsContract, createGroupMsg, nil, nil)
		s.Require().True(res.IsOK())

		var createResp struct {
			GroupID string `json:"group_id"`
		}
		_ = appA.DecodeExecuteResponse(res, &createResp)
		groupIDs = append(groupIDs, createResp.GroupID)

		// Add user to group
		addMemberMsg := types.WasmxExecutionMessage{
			Data: []byte(fmt.Sprintf(`{"add_member":{"group_id":"%s","user_id":"%s"}}`, createResp.GroupID, newUserID)),
		}
		res = appA.ExecuteContract(setup.Sender, setup.GroupsContract, addMemberMsg, nil, nil)
		s.Require().True(res.IsOK())
	}

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
