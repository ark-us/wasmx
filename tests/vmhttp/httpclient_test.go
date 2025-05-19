package keeper_test

import (
	_ "embed"
	"encoding/json"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/mythos-tests/vmhttp/testdata"
	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	"github.com/loredanacirstea/wasmx/x/vmhttpclient"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

type CalldataTestHttpClient struct {
	HttpRequest *vmhttpclient.HttpRequestWrap `json:"HttpRequest"`
}

func (suite *KeeperTestSuite) TestHttpClient() {
	wasmbin := testdata.WasmxTestHttp
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "httpclient", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "httpclient", contractAddress, sender)

	msg := &CalldataTestHttpClient{
		HttpRequest: &vmhttpclient.HttpRequestWrap{
			Request: vmhttpclient.HttpRequest{
				Method: "GET",
				Url:    "http://httpbin.org/get?key=1&value=hello",
				Data:   []byte{},
			},
			ResponseHandler: vmhttpclient.ResponseHandler{
				MaxSize: int64(1 << 20), // 1MB
			},
		}}
	data, err := json.Marshal(msg)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp := &vmhttpclient.HttpResponseWrap{}
	suite.parseQueryResponse(qres, qresp)

	suite.Require().Equal("", qresp.Error)
	suite.Require().Equal("200 OK", qresp.Data.Status)
	suite.Require().Equal(200, qresp.Data.StatusCode)
	suite.Require().Greater(qresp.Data.ContentLength, int64(0))
	suite.Require().Equal([]string{"application/json"}, qresp.Data.Header.Values("Content-Type"))
	// suite.Require().Contains(string(qresp.Data.Data), `"args":{"key":"1","value":"hello"}`)

	msg = &CalldataTestHttpClient{
		HttpRequest: &vmhttpclient.HttpRequestWrap{
			Request: vmhttpclient.HttpRequest{
				Method: "POST",
				Url:    "http://httpbin.org/post",
				Data:   []byte(`{"a":1}`),
			},
			ResponseHandler: vmhttpclient.ResponseHandler{
				MaxSize: int64(1 << 20), // 1MB
			},
		}}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp = &vmhttpclient.HttpResponseWrap{}
	suite.parseQueryResponse(qres, qresp)

	suite.Require().Equal("", qresp.Error)
	suite.Require().Equal("200 OK", qresp.Data.Status)
	suite.Require().Equal(200, qresp.Data.StatusCode)
	suite.Require().Greater(qresp.Data.ContentLength, int64(0))
	suite.Require().Equal([]string{"application/json"}, qresp.Data.Header.Values("Content-Type"))
	// suite.Require().Contains(string(qresp.Data.Data), `"data": "{\"a\":1}"`)
}
