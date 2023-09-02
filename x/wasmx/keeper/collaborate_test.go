package keeper_test

import (
	_ "embed"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
	initBalance := sdk.NewInt(1_000_000_000_000_000_000)

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

	// evmcode, err := hex.DecodeString(testdata.ForwardContract)
	// s.Require().NoError(err)
	// _, contractAddressEvm := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "ForwardContract", &types.CodeMetadata{Abi: interfacesTestdata.ForwardEvmStr})
	// forwardHex := "d5c4c85c" // "forward(string,address[])",
	// forwardGetHex := "7f2f2ad8" // "forward_get(address[])"

	// go
	data := []byte(fmt.Sprintf(`{"forward":["multi-vm transaction: ",[]}`))
	resp := appA.ExecuteContract(sender, contractAddressGo, types.WasmxExecutionMessage{Data: data}, nil, nil)
	fmt.Println("-----GetLog--go-", resp.GetLog())
	fmt.Println("-----data---", resp.GetData(), string(resp.GetData()))

	// js -> py
	data = []byte(fmt.Sprintf(`{"forward":["multi-vm transaction: ",["%s"]]}`, contractAddressPy.String()))
	resp = appA.ExecuteContract(sender, contractAddressJs, types.WasmxExecutionMessage{Data: data}, nil, nil)
	fmt.Println("-----GetLog-js -> py--", resp.GetLog())
	fmt.Println("-----data---", resp.GetData(), string(resp.GetData()))

	// py -> js
	data = []byte(fmt.Sprintf(`{"forward":["multi-vm transaction: ",["%s"]]}`, contractAddressJs.String()))
	resp = appA.ExecuteContract(sender, contractAddressPy, types.WasmxExecutionMessage{Data: data}, nil, nil)
	fmt.Println("-----GetLog-py -> js--", resp.GetLog())
	fmt.Println("-----data---", resp.GetData(), string(resp.GetData()))

	// py -> js -> go
	fmt.Println("====py -> js -> go=====")
	data = []byte(fmt.Sprintf(`{"forward":["multi-vm transaction: ",["%s","%s"]]}`, contractAddressJs.String(), contractAddressGo.String()))
	resp = appA.ExecuteContract(sender, contractAddressPy, types.WasmxExecutionMessage{Data: data}, nil, nil)
	fmt.Println("-----GetLog--py -> js -> go-", resp.GetLog())
	fmt.Println("-----data---", resp.GetData(), string(resp.GetData()))

	// // js -> evm
	// // contractAddressEvm
	// fmt.Println("====js -> evm=====")
	// data = []byte(fmt.Sprintf(`{"forward":["multi-vm transaction: ",["%s"]]}`, contractAddressEvm.String()))
	// resp = appA.ExecuteContract(sender, contractAddressJs, types.WasmxExecutionMessage{Data: data}, nil, nil)
	// fmt.Println("-----GetLog--js -> evm-", resp.GetLog())
	// fmt.Println("-----data---", resp.GetData(), string(resp.GetData()))

	// get go
	data = []byte(`{"forward_get":[[]]}`)
	qres := appA.WasmxQueryRaw(sender, contractAddressGo, types.WasmxExecutionMessage{Data: data}, nil, nil)
	fmt.Println("-----qres---", string(qres))

	// get py -> js
	data = []byte(fmt.Sprintf(`{"forward_get":[["%s"]]}`, contractAddressJs.String()))
	qres = appA.WasmxQueryRaw(sender, contractAddressPy, types.WasmxExecutionMessage{Data: data}, nil, nil)
	fmt.Println(string(qres))
	// s.Require().Equal("", string(qres))

	// get js -> py
	data = []byte(fmt.Sprintf(`{"forward_get":[["%s"]]}`, contractAddressPy.String()))
	qres = appA.WasmxQueryRaw(sender, contractAddressJs, types.WasmxExecutionMessage{Data: data}, nil, nil)
	fmt.Println(string(qres))
	// s.Require().Equal("", string(qres))

	// get py -> js -> go
	data = []byte(fmt.Sprintf(`{"forward_get":[["%s","%s"]]}`, contractAddressJs.String(), contractAddressGo.String()))
	qres = appA.WasmxQueryRaw(sender, contractAddressPy, types.WasmxExecutionMessage{Data: data}, nil, nil)
	fmt.Println(string(qres))
	s.Require().Equal("python -> javascript -> tinygo", string(qres))
}
