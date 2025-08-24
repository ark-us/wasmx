package keeper_test

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/vm/precompiles"

	testdata "github.com/loredanacirstea/mythos-tests/testdata/classic"
	wasmxtest "github.com/loredanacirstea/mythos-tests/testdata/wasmx"
)

func (suite *KeeperTestSuite) TestEwasmCallToPriviledged() {
	SkipFixmeTests(suite.T(), "TestEwasmCallToPriviledged")
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	evmcode, err := hex.DecodeString(testdata.CallGeneral)
	s.Require().NoError(err)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	_, contractAddress1 := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "callgeneralwasm1", nil)

	storagebin := wasmxtest.WasmxSimpleStorage
	codeId := appA.StoreCode(sender, storagebin, nil)
	contractAddress2 := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage", nil)

	msg := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	// res := appA.ExecuteContract(sender, contractAddress2, types.WasmxExecutionMessage{Data: data}, nil, nil)

	// Execute nested calls
	// value := `0000000000000000000000000000000000000000000000000000000000000009`
	// data := `0000000000000000000000000000000000000000000000000000000000000002` + `000000000000000000000000` + hex.EncodeToString(contractAccount2.Bytes()) + `000000000000000000000000` + hex.EncodeToString(contractAccount3.Bytes()) + value

	data := `0000000000000000000000000000000000000000000000000000000000000001` + `000000000000000000000000` + hex.EncodeToString(contractAddress2.Bytes()) + hex.EncodeToString(msg)

	deps := []string{types.EvmAddressFromAcc(contractAddress1.Bytes()).Hex(), types.EvmAddressFromAcc(contractAddress2.Bytes()).Hex()}
	res := appA.ExecuteContract(sender, contractAddress1, types.WasmxExecutionMessage{Data: appA.Hex2bz(data)}, nil, deps)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000003", string(res.Data))

	queryres, err := appA.App.WasmxKeeper.QuerySmart(
		appA.Context(),
		contractAddress2,
		[]byte(`{"get":{"key":"hello"}}`),
	)
	suite.Require().NoError(err)
	suite.Require().Equal("", hex.EncodeToString(queryres))

	// 	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	// 	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAccount1, keybz)
	// 	suite.Require().Equal(value, hex.EncodeToString(queryres))

	// 	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAccount2, keybz)
	// 	suite.Require().Equal(value, hex.EncodeToString(queryres))

	// 	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAccount3, keybz)
	// 	suite.Require().Equal(value, hex.EncodeToString(queryres))

}

func (suite *KeeperTestSuite) TestUpgradeRolesStaking() {
	// test upgrading staking contract
	// test storage migration or instantiation with other values
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)
	valAccount := simulation.Account{
		PrivKey: s.Chain().SenderPrivKey,
		PubKey:  s.Chain().SenderPrivKey.PubKey(),
		Address: s.Chain().SenderAccount.GetAddress(),
	}

	appA := s.AppContext()
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixed, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(valAccount.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	valid, err := appA.App.StakingKeeper.GetAllValidators(appA.Context())
	s.Require().NoError(err)

	rolesAddr := appA.AccBech32Codec().BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_ROLES))
	wasmbin := precompiles.GetPrecompileByLabel(appA.AddressCodec(), types.STAKING_v001)
	codeId := appA.StoreCode(sender, wasmbin, nil)

	newAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "newstaking", nil)

	newlabel := types.STAKING_v001 + "2"
	title := "Register staking"
	description := "Register staking"
	authority := appA.MustAccAddressToString(authtypes.NewModuleAddress(types.ROLE_GOVERNANCE))
	newAddressStr := newAddress.String()

	msg := []byte(fmt.Sprintf(`{"SetContractForRoleGov":{"role":"%s","label":"%s","contract_address":"%s","action_type":0}}`, types.ROLE_STAKING, newlabel, newAddressStr))
	msgbz, err := json.Marshal(&types.WasmxExecutionMessage{Data: msg})
	s.Require().NoError(err)

	proposal := &types.MsgExecuteContract{
		Sender:   authority,
		Contract: rolesAddr.String(),
		Msg:      msgbz,
	}
	appA.PassGovProposal(valAccount, sender, []sdk.Msg{proposal}, "", title, description, false)

	resp := appA.App.WasmxKeeper.GetRoleLabelByContract(appA.Context(), newAddress)
	s.Require().Equal(newlabel, resp)

	role := appA.App.WasmxKeeper.GetRoleByLabel(appA.Context(), newlabel)
	s.Require().Equal(newAddressStr, role.Addresses[role.Primary])
	s.Require().Equal(types.ROLE_STAKING, role.Role)
	s.Require().Equal(newlabel, role.Labels[0])

	roleAddr, err := appA.App.WasmxKeeper.GetAddressOrRole(appA.Context(), types.ROLE_STAKING)
	s.Require().NoError(err)
	s.Require().Equal(newAddress.String(), roleAddr.String())

	// query staking contract
	validPost, err := appA.App.StakingKeeper.GetAllValidators(appA.Context())
	s.Require().NoError(err)
	s.Require().Equal(valid, validPost)
}

func (suite *KeeperTestSuite) TestUpgradeCacheRolesContract() {
	// test upgrading roles contract
	// test storage migration or instantiation with other values
	// test upgrade cache

	// upgrading roles contract: must be initialized only with existent roles
	// because we do not emit events or hooks
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)
	valAccount := simulation.Account{
		PrivKey: s.Chain().SenderPrivKey,
		PubKey:  s.Chain().SenderPrivKey.PubKey(),
		Address: s.Chain().SenderAccount.GetAddress(),
	}

	appA := s.AppContext()
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixed, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(valAccount.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	rolesAddr := appA.AccBech32Codec().BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_ROLES))
	wasmbin := precompiles.GetPrecompileByLabel(appA.AddressCodec(), types.ROLES_v001)
	codeId := appA.StoreCode(sender, wasmbin, nil)

	// setup roles in genesis
	rolesbz := appA.QueryContract(sender, rolesAddr, []byte(`{"GetRoles":{}}`), nil, nil)

	var roles types.RolesGenesis
	err := json.Unmarshal(rolesbz, &roles)
	s.Require().NoError(err)

	newRoles := &types.RolesGenesis{}
	newrolesinitbz, err := json.Marshal(&newRoles)
	s.Require().NoError(err)

	newAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: newrolesinitbz}, "newroles", nil)

	newrolesbz := appA.QueryContract(sender, newAddress, []byte(`{"GetRoles":{}}`), nil, nil)

	var roles2 types.RolesGenesis
	err = json.Unmarshal(newrolesbz, &roles2)
	s.Require().NoError(err)
	s.Require().Equal(1, len(roles2.Roles))

	newlabel := types.ROLES_v001 + "2"

	// Register contract role proposal
	title := "Register roles"
	description := "Register roles"
	authority := appA.MustAccAddressToString(authtypes.NewModuleAddress(types.ROLE_GOVERNANCE))
	newAddressStr := newAddress.String()

	msg := []byte(fmt.Sprintf(`{"SetContractForRoleGov":{"role":"%s","label":"%s","contract_address":"%s","action_type":0}}`, types.ROLE_ROLES, newlabel, newAddressStr))
	msgbz, err := json.Marshal(&types.WasmxExecutionMessage{Data: msg})
	s.Require().NoError(err)

	proposal := &types.MsgExecuteContract{
		Sender:   authority,
		Contract: rolesAddr.String(),
		Msg:      msgbz,
	}
	appA.PassGovProposal(valAccount, sender, []sdk.Msg{proposal}, "Register roles contract", title, description, false)

	cached, err := appA.App.WasmxKeeper.GetSystemBootstrapData(appA.Context())
	s.Require().NoError(err)
	s.Require().NotNil(cached)
	s.Require().Equal(newAddress.String(), cached.RoleAddress)

	newlabel = "roles_rolesv0.0.1" // label set by contract
	resp := appA.App.WasmxKeeper.GetRoleLabelByContract(appA.Context(), newAddress)
	s.Require().Equal(newlabel, resp)

	role := appA.App.WasmxKeeper.GetRoleByLabel(appA.Context(), newlabel)
	s.Require().Equal(newAddressStr, role.Addresses[0])
	s.Require().Equal(types.ROLE_ROLES, role.Role)
	s.Require().Equal(newlabel, role.Labels[0])

	// make a generic transaction that uses roles contract
	codeId2 := appA.StoreCode(sender, wasmxtest.WasmxSimpleStorage, nil)
	appA.InstantiateCode(sender, codeId2, types.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage", nil)
}

func (suite *KeeperTestSuite) TestUpgradeCacheContractsRegistry() {
	// test upgrade contracts registry
	// test data migration
	// test upgrade cache
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)
	valAccount := simulation.Account{
		PrivKey: s.Chain().SenderPrivKey,
		PubKey:  s.Chain().SenderPrivKey.PubKey(),
		Address: s.Chain().SenderAccount.GetAddress(),
	}

	appA := s.AppContext()
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixed, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(valAccount.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	wasmbin := precompiles.GetPrecompileByLabel(appA.AddressCodec(), types.STORAGE_CONTRACTS_v001)
	codeId := appA.StoreCode(sender, wasmbin, nil)
	newAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{"code_infos":[],"contract_infos":[]}`)}, "newregistry", nil)

	newlabel := types.STORAGE_CONTRACTS_v001 + "2"

	// Register contract registry
	title := "Register registry"
	description := "Register registry"
	authority := appA.MustAccAddressToString(authtypes.NewModuleAddress(types.ROLE_GOVERNANCE))
	newAddressStr := newAddress.String()

	contractInfo, codeInfo, _, err := appA.App.WasmxKeeper.ContractInstance(appA.Context(), newAddress)
	s.Require().NoError(err)
	s.Require().NotNil(codeInfo)
	s.Require().NotNil(contractInfo)

	rolesAddr := appA.AccBech32Codec().BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_ROLES))

	msg := []byte(fmt.Sprintf(`{"SetContractForRoleGov":{"role":"%s","label":"%s","contract_address":"%s","action_type":0}}`, types.ROLE_STORAGE_CONTRACTS, newlabel, newAddressStr))
	msgbz, err := json.Marshal(&types.WasmxExecutionMessage{Data: msg})
	s.Require().NoError(err)

	proposal := &types.MsgExecuteContract{
		Sender:   authority,
		Contract: rolesAddr.String(),
		Msg:      msgbz,
	}
	appA.PassGovProposal(valAccount, sender, []sdk.Msg{proposal}, "", title, description, false)

	resp := appA.App.WasmxKeeper.GetRoleLabelByContract(appA.Context(), newAddress)
	s.Require().Equal(newlabel, resp)

	role := appA.App.WasmxKeeper.GetRoleByLabel(appA.Context(), newlabel)
	s.Require().Equal(newAddressStr, role.Addresses[0])
	s.Require().Equal(newlabel, role.Labels[0])
	s.Require().Equal(types.ROLE_STORAGE_CONTRACTS, role.Role)

	// check cached code & contract info
	cached, err := appA.App.WasmxKeeper.GetSystemBootstrapData(appA.Context())
	s.Require().NoError(err)
	s.Require().NotNil(cached)
	s.Require().Equal(newAddress.String(), cached.CodeRegistryAddress)
	s.Require().NotNil(cached.CodeRegistryCodeInfo)
	s.Require().NotNil(cached.CodeRegistryContractInfo)

	s.Require().True(bytes.Equal(codeInfo.CodeHash, cached.CodeRegistryCodeInfo.CodeHash))
	s.Require().True(bytes.Equal([]byte(codeInfo.CodeHash), cached.CodeRegistryCodeInfo.CodeHash))
	s.Require().Equal([]string(codeInfo.Deps), cached.CodeRegistryCodeInfo.Deps)
	s.Require().True(bytes.Equal([]byte(codeInfo.InterpretedBytecodeDeployment), cached.CodeRegistryCodeInfo.InterpretedBytecodeDeployment))
	s.Require().True(bytes.Equal([]byte(codeInfo.InterpretedBytecodeRuntime), cached.CodeRegistryCodeInfo.InterpretedBytecodeRuntime))
	s.Require().Equal(codeInfo.MeteringOff, cached.CodeRegistryCodeInfo.MeteringOff)
	s.Require().Equal(codeInfo.Pinned, cached.CodeRegistryCodeInfo.Pinned)
	s.Require().True(bytes.Equal([]byte(codeInfo.RuntimeHash), cached.CodeRegistryCodeInfo.RuntimeHash))
	s.Require().Equal(codeInfo.Creator, cached.CodeRegistryCodeInfo.Creator)

	s.Require().Equal(contractInfo.CodeId, cached.CodeRegistryContractInfo.CodeId)
	s.Require().Equal(contractInfo.Creator, cached.CodeRegistryContractInfo.Creator)
	s.Require().True(bytes.Equal([]byte(contractInfo.InitMessage), []byte(cached.CodeRegistryContractInfo.InitMessage)))
	s.Require().Equal(contractInfo.Label, cached.CodeRegistryContractInfo.Label)
	s.Require().Equal(contractInfo.Provenance, cached.CodeRegistryContractInfo.Provenance)
	s.Require().Equal(contractInfo.StorageType, cached.CodeRegistryContractInfo.StorageType)
}
