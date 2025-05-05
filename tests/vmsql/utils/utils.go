package utils

import (
	_ "embed"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/stretchr/testify/require"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	wt "github.com/loredanacirstea/wasmx/testutil/wasmx"
	"github.com/loredanacirstea/wasmx/x/vmsql"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/vm/precompiles"
)

// KeeperTestSuiteInterface defines the common functionality needed by test suites
type KeeperTestSuiteInterface interface {
	// Commit commits the current block
	Commit()
	// GetRandomAccount returns a random account for testing
	GetRandomAccount() simulation.Account
	// Require returns the require object for assertions
	Require() *require.Assertions

	Chain() wt.TestChain
}

func DeployDType(suite KeeperTestSuiteInterface, appA wt.AppContext, sender simulation.Account) (mcodec.AccAddressPrefixed, error) {
	wasmbin := precompiles.GetPrecompileByLabel(appA.AddressCodec(), types.DTYPE_v001)
	codeId := appA.StoreCode(sender, wasmbin, nil)
	cmdi := &vmsql.InstantiateDType{Dir: "", Driver: "sqlite3"}
	data, err := json.Marshal(cmdi)
	if err != nil {
		return mcodec.AccAddressPrefixed{}, err
	}
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: data}, "dtype", nil)

	// set a role to have access to protected APIs
	RegisterRole(suite, appA, "dtype", contractAddress, sender)

	cmd := &vmsql.CalldataDType{Initialize: cmdi}
	data, err = json.Marshal(cmd)
	if err != nil {
		return mcodec.AccAddressPrefixed{}, err
	}
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	cmd = &vmsql.CalldataDType{InitializeTokens: &vmsql.InitializeTokens{}}
	data, err = json.Marshal(cmd)
	if err != nil {
		return mcodec.AccAddressPrefixed{}, err
	}
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 50000000, nil)

	cmd = &vmsql.CalldataDType{InitializeIdentity: &vmsql.InitializeIdentity{}}
	data, err = json.Marshal(cmd)
	if err != nil {
		return mcodec.AccAddressPrefixed{}, err
	}
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 50000000, nil)
	return contractAddress, nil
}

func RegisterRole(suite KeeperTestSuiteInterface, appA wt.AppContext, rolename string, contractAddress mcodec.AccAddressPrefixed, sender simulation.Account) {
	title := "Register " + rolename
	description := "Register " + rolename
	rolesAddr := appA.AccBech32Codec().BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_ROLES))

	valAccount := simulation.Account{
		PrivKey: suite.Chain().SenderPrivKey,
		PubKey:  suite.Chain().SenderPrivKey.PubKey(),
		Address: suite.Chain().SenderAccount.GetAddress(),
	}
	initBalance := sdkmath.NewInt(wt.DEFAULT_BALANCE).MulRaw(500000)
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(valAccount.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	// register new role
	msg := []byte(fmt.Sprintf(`{"SetRole":{"role":{"role":"%s","storage_type":0,"primary":0,"multiple":false,"labels":["%s"],"addresses":["%s"]}}}`, rolename, rolename, contractAddress))
	msgbz, err := json.Marshal(&types.WasmxExecutionMessage{Data: msg})
	suite.Require().NoError(err)
	exec := &types.MsgExecuteContract{
		Sender:   appA.App.WasmxKeeper.GetAuthority(),
		Contract: rolesAddr.String(),
		Msg:      msgbz,
	}
	appA.PassGovProposal(valAccount, sender, []sdk.Msg{exec}, "", title, description, false)

	resp := appA.App.WasmxKeeper.GetRoleLabelByContract(appA.Context(), contractAddress)
	suite.Require().Equal(rolename, resp)

	role := appA.App.WasmxKeeper.GetRoleByLabel(appA.Context(), rolename)
	suite.Require().Equal(contractAddress.String(), role.Addresses[0])
	suite.Require().Equal(rolename, role.Labels[0])
	suite.Require().Equal(rolename, role.Role)
}

func CreateDb(
	suite KeeperTestSuiteInterface,
	appA wt.AppContext,
	sender simulation.Account,
	contractAddress mcodec.AccAddressPrefixed,
	connFile string,
	connDriver string,
	connName string,
	dbname string,
) vmsql.TableIdentifier {
	// insert new database with connection
	cmd := &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{Identifier: vmsql.TableIdentifier{
		DbConnectionName: vmsql.DTypeConnection,
		DbName:           vmsql.DbName,
		TableName:        vmsql.DTypeDbConnName,
	},
		Data: []byte(fmt.Sprintf(`{"connection":"%s","driver":"%s","name":"%s"}`, connFile, connDriver, connName)),
	}}
	data, err := json.Marshal(cmd)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss := &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	newconnId := resss.Responses[0].LastInsertId

	// new db definition
	cmd = &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{Identifier: vmsql.TableIdentifier{
		DbConnectionName: vmsql.DTypeConnection,
		DbName:           vmsql.DbName,
		TableName:        vmsql.DTypeDbName,
	},
		Data: []byte(fmt.Sprintf(`{"name":"%s","connection_id":%d}`, dbname, newconnId)),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	newDbId := resss.Responses[0].LastInsertId

	return vmsql.TableIdentifier{
		DbConnectionId:   newconnId,
		DbId:             newDbId,
		DbConnectionName: connName,
		DbName:           dbname,
	}
}

func CreateTable(
	suite KeeperTestSuiteInterface,
	appA wt.AppContext,
	sender simulation.Account,
	contractAddress mcodec.AccAddressPrefixed,
	dbId int64,
	tablename string,
) int64 {

	// insert new table definition
	cmd := &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{Identifier: vmsql.TableIdentifier{
		DbConnectionName: vmsql.DTypeConnection,
		DbName:           vmsql.DbName,
		TableName:        vmsql.DTypeTableName,
	},
		Data: []byte(fmt.Sprintf(`{"name":"%s","db_id":%d}`, tablename, dbId)),
	}}
	data, err := json.Marshal(cmd)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss := &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	tableId := resss.Responses[0].LastInsertId

	return tableId
}

func CreateFields(
	suite KeeperTestSuiteInterface,
	appA wt.AppContext,
	sender simulation.Account,
	contractAddress mcodec.AccAddressPrefixed,
	fields []string,
) {
	// insert table fields
	for _, fielddef := range fields {
		cmd := &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{Identifier: vmsql.TableIdentifier{
			DbConnectionName: vmsql.DTypeConnection,
			DbName:           vmsql.DbName,
			TableName:        vmsql.DTypeFieldName,
		},
			Data: []byte(fielddef),
		}}
		data, err := json.Marshal(cmd)
		suite.Require().NoError(err)
		res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
		resss := &vmsql.SqlExecuteBatchResponse{}
		err = appA.DecodeExecuteResponse(res, resss)
		suite.Require().NoError(err)
		suite.Require().Equal("", resss.Error)
		suite.Require().Equal(1, len(resss.Responses))
	}
}

func InstantiateTable(
	suite KeeperTestSuiteInterface,
	appA wt.AppContext,
	sender simulation.Account,
	contractAddress mcodec.AccAddressPrefixed,
	newconnId int64,
	tableId int64,
) {
	// start database connection
	cmd := &vmsql.CalldataDType{Connect: &vmsql.ConnectRequest{Id: newconnId}}
	data, err := json.Marshal(cmd)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssc := &vmsql.SqlConnectionResponse{}
	err = appA.DecodeExecuteResponse(res, resssc)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssc.Error)

	// create table
	cmd = &vmsql.CalldataDType{CreateTable: &vmsql.CreateTableDTypeRequest{
		TableId: tableId,
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss := &vmsql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssc.Error)
}
