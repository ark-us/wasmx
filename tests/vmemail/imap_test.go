package keeper_test

import (
	_ "embed"
	"encoding/json"

	"github.com/emersion/go-imap/v2"
	_ "github.com/mattn/go-sqlite3"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/mythos-tests/vmemail/testdata"
	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	vmimap "github.com/loredanacirstea/wasmx-vmimap"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

type Calldata struct {
	ConnectWithPassword *vmimap.ImapConnectionSimpleRequest `json:"ConnectWithPassword"`
	ConnectOAuth2       *vmimap.ImapConnectionOauth2Request `json:"ConnectOAuth2"`
	Close               *vmimap.ImapCloseRequest            `json:"Close"`
	Fetch               *vmimap.ImapFetchRequest            `json:"Fetch"`
	Listen              *vmimap.ImapListenRequest           `json:"Listen"`
	CreateFolder        *vmimap.ImapCreateFolderRequest     `json:"CreateFolder"`
}

func (suite *KeeperTestSuite) TestImap() {
	wasmbin := testdata.WasmxTestImap
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "imaptest", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "someemailrole", contractAddress, sender)

	msg := &Calldata{
		ConnectWithPassword: &vmimap.ImapConnectionSimpleRequest{
			Id:            "conn1",
			ImapServerUrl: "mail.mail.provable.dev:993",
			Username:      "test@mail.provable.dev",
			Password:      "uwsawW3A6**yB^kp",
		}}
	data, err := json.Marshal(msg)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resc := &vmimap.ImapConnectionResponse{}
	err = appA.DecodeExecuteResponse(res, resc)
	suite.Require().NoError(err)
	suite.Require().Equal("", resc.Error)

	msg = &Calldata{
		Fetch: &vmimap.ImapFetchRequest{
			Id:          "conn1",
			Folder:      "INBOX",
			SeqSet:      imap.SeqSetNum(1),
			UidSet:      make(imap.UIDSet, 0),
			FetchFilter: nil,
		}}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp := &vmimap.ImapFetchResponse{}
	err = json.Unmarshal(qres, qresp)
	suite.Require().NoError(err)
	suite.Require().Equal(qresp.Error, "")
	suite.Require().Equal(1, len(qresp.Data))

	msg = &Calldata{
		CreateFolder: &vmimap.ImapCreateFolderRequest{
			Id:   "conn1",
			Path: "INBOX2/mysubfolder",
		}}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rescf := &vmimap.ImapCreateFolderResponse{}
	err = appA.DecodeExecuteResponse(res, rescf)
	suite.Require().NoError(err)
	suite.Require().Equal("", rescf.Error)

	msg = &Calldata{
		Listen: &vmimap.ImapListenRequest{
			Id:     "conn1",
			Folder: "INBOX",
		}}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resl := &vmimap.ImapListenResponse{}
	err = appA.DecodeExecuteResponse(res, resl)
	suite.Require().NoError(err)
	suite.Require().Equal("", resl.Error)

	msg = &Calldata{
		Close: &vmimap.ImapCloseRequest{
			Id: "conn1",
		}}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rescl := &vmimap.ImapCloseResponse{}
	err = appA.DecodeExecuteResponse(res, rescl)
	suite.Require().NoError(err)
	suite.Require().Equal("", rescl.Error)
}
