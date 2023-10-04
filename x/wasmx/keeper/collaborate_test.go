package keeper_test

import (
	_ "embed"
	"fmt"
	"strconv"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/cometbft/cometbft/abci/types"

	"mythos/v1/x/wasmx/keeper/testutil"
	"mythos/v1/x/wasmx/types"
)

var (
	//go:embed testdata/python/forward.py
	forwardPy []byte

	//go:embed testdata/js/forward.js
	forwardJs []byte

	//go:embed testdata/tinygo/forward.wasm
	forwardGo []byte
)

// Python -> JavaScript -> Tinygo wasm -> AssemblyScript -> EVM -> CosmWasm
func (suite *KeeperTestSuite) TestVMCollaboration() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	depsPy := []string{types.INTERPRETER_PYTHON}
	depsJs := []string{types.INTERPRETER_JS}

	codeIdPy := appA.StoreCode(sender, forwardPy, depsPy)
	contractAddressPy := appA.InstantiateCode(sender, codeIdPy, types.WasmxExecutionMessage{Data: []byte(``)}, "forwardPy", nil)

	codeIdJs := appA.StoreCode(sender, forwardJs, depsJs)
	contractAddressJs := appA.InstantiateCode(sender, codeIdJs, types.WasmxExecutionMessage{Data: []byte(``)}, "forwardJs", nil)

	codeIdGo := appA.StoreCode(sender, forwardGo, nil)
	contractAddressGo := appA.InstantiateCode(sender, codeIdGo, types.WasmxExecutionMessage{Data: []byte{}}, "forwardGo", nil)

	// TODO
	// evmcode, err := hex.DecodeString(testdata.ForwardContract)
	// s.Require().NoError(err)
	// _, contractAddressEvm := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "ForwardContract", &types.CodeMetadata{Abi: interfacesTestdata.ForwardEvmStr})
	// forwardHex := "d5c4c85c" // "forward(string,address[])",
	// forwardGetHex := "7f2f2ad8" // "forward_get(address[])"

	// go
	data := []byte(`{"forward":["multi-vm transaction: ",[]}`)
	resp := appA.ExecuteContract(sender, contractAddressGo, types.WasmxExecutionMessage{Data: data}, nil, nil)
	expected := "multi-vm transaction: tinygo"
	s.Require().Contains(string(resp.GetData()), expected)
	evs := appA.GetWasmxEvents(resp.GetEvents())
	s.Require().Equal(1, len(evs))
	for _, attr := range evs[0].GetAttributes() {
		if attr.Key == types.AttributeKeyDependency {
			s.Require().Equal(types.WASI_SNAPSHOT_PREVIEW1, attr.Value)
		}
		if attr.Key == types.AttributeKeyData {
			s.Require().Equal(expected, string(appA.Hex2bz(attr.Value)))
		}
	}

	// js -> py
	data = []byte(fmt.Sprintf(`{"forward":["multi-vm transaction: ",["%s"]]}`, contractAddressPy.String()))
	resp = appA.ExecuteContract(sender, contractAddressJs, types.WasmxExecutionMessage{Data: data}, nil, nil)
	expected = "multi-vm transaction: javascript -> python"
	s.Require().Contains(string(resp.GetData()), expected)
	checkLogs(appA, resp.GetEvents(), []string{types.INTERPRETER_JS, types.INTERPRETER_PYTHON}, []string{"multi-vm transaction: javascript", expected})

	// py -> js
	data = []byte(fmt.Sprintf(`{"forward":["multi-vm transaction: ",["%s"]]}`, contractAddressJs.String()))
	resp = appA.ExecuteContract(sender, contractAddressPy, types.WasmxExecutionMessage{Data: data}, nil, nil)
	expected = "multi-vm transaction: python -> javascript"
	s.Require().Contains(string(resp.GetData()), expected)
	checkLogs(appA, resp.GetEvents(), []string{types.INTERPRETER_PYTHON, types.INTERPRETER_JS}, []string{"multi-vm transaction: python", expected})

	// py -> js -> go
	data = []byte(fmt.Sprintf(`{"forward":["multi-vm transaction: ",["%s","%s"]]}`, contractAddressJs.String(), contractAddressGo.String()))
	resp = appA.ExecuteContract(sender, contractAddressPy, types.WasmxExecutionMessage{Data: data}, nil, nil)
	expected = "multi-vm transaction: python -> javascript -> tinygo"
	s.Require().Contains(string(resp.GetData()), expected)
	checkLogs(appA, resp.GetEvents(), []string{types.INTERPRETER_PYTHON, types.INTERPRETER_JS, types.WASI_SNAPSHOT_PREVIEW1}, []string{"multi-vm transaction: python", "multi-vm transaction: python -> javascript", expected})

	// TODO
	// // js -> evm
	// // contractAddressEvm
	// fmt.Println("====js -> evm=====")
	// data = []byte(fmt.Sprintf(`{"forward":["multi-vm transaction: ",["%s"]]}`, contractAddressEvm.String()))
	// resp = appA.ExecuteContract(sender, contractAddressJs, types.WasmxExecutionMessage{Data: data}, nil, nil)

	// get go
	data = []byte(`{"forward_get":[[]]}`)
	qres := appA.WasmxQueryRaw(sender, contractAddressGo, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("tinygo", string(qres))

	// get py -> js
	data = []byte(fmt.Sprintf(`{"forward_get":[["%s"]]}`, contractAddressJs.String()))
	qres = appA.WasmxQueryRaw(sender, contractAddressPy, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("python -> javascript", string(qres))

	// get js -> py
	data = []byte(fmt.Sprintf(`{"forward_get":[["%s"]]}`, contractAddressPy.String()))
	qres = appA.WasmxQueryRaw(sender, contractAddressJs, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("javascript -> python", string(qres))

	// get py -> js -> go
	data = []byte(fmt.Sprintf(`{"forward_get":[["%s","%s"]]}`, contractAddressJs.String(), contractAddressGo.String()))
	qres = appA.WasmxQueryRaw(sender, contractAddressPy, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("python -> javascript -> tinygo", string(qres))
}

func checkLogs(appA testutil.AppContext, events []abci.Event, deps []string, data []string) {
	var err error
	evs := appA.GetWasmxEvents(events)
	s.Require().Equal(len(deps), len(evs))
	logindex := int64(0)
	for _, attr := range evs[0].GetAttributes() {
		if attr.Key == types.AttributeKeyIndex {
			logindex, err = strconv.ParseInt(attr.Value, 10, 64)
			s.Require().NoError(err)
		}
		expectedDep := deps[logindex]
		expectedData := data[logindex]
		if attr.Key == types.AttributeKeyDependency {
			s.Require().Equal(expectedDep, attr.Value)
		}
		if attr.Key == types.AttributeKeyData {
			s.Require().Equal(expectedData, string(appA.Hex2bz(attr.Value)))
		}
	}
}
