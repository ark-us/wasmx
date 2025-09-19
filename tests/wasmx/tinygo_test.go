package keeper_test

import (
	_ "embed"
	"encoding/hex"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"

	py "github.com/loredanacirstea/mythos-tests/testdata/python"
	tinygo "github.com/loredanacirstea/mythos-tests/testdata/tinygo"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
)

func (suite *KeeperTestSuite) TestWasiTinygoDeterminism() {
	wasmbinD := tinygo.TinyGoDeterministic
	wasmbinNonD := tinygo.TinyGoNonDeterministic
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	codeId := appA.StoreCode(sender, wasmbinD, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "TinyGoDeterministic", nil)

	codeId2 := appA.StoreCode(sender, wasmbinNonD, nil)
	contractAddress2 := appA.InstantiateCode(sender, codeId2, types.WasmxExecutionMessage{Data: []byte{}}, "TinyGoNonDeterministic", nil)

	qresD := appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte{}}, nil, nil)
	bz, _ := hex.DecodeString(qresD)
	fmt.Println("deterministic:", string(bz))
	qres := appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte{}}, nil, nil)
	s.Require().Equal(qresD, qres)

	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte{}}, nil, nil)
	s.Require().Equal(qresD, qres)

	// nondeterministic contract without role should be deterministic
	qres = appA.WasmxQuery(sender, contractAddress2, types.WasmxExecutionMessage{Data: []byte{}}, nil, nil)
	s.Require().Equal(qresD, qres)

	qres = appA.WasmxQuery(sender, contractAddress2, types.WasmxExecutionMessage{Data: []byte{}}, nil, nil)
	s.Require().Equal(qresD, qres)

	// set a role to have access to protected deterministic APIs
	msg := []byte(fmt.Sprintf(`{"SetRole":{"role":{"role":"%s","storage_type":0,"primary":0,"multiple":false,"labels":["%s"],"addresses":["%s"]}}}`, "somerole", "somerole", contractAddress2.String()))
	_, err := appA.App.NetworkKeeper.ExecuteContractInternal(appA.Context(), &networktypes.MsgExecuteContract{
		Sender:   appA.App.WasmxKeeper.GetAuthority(),
		Contract: types.ROLE_ROLES,
		Msg:      []byte(msg),
	})
	s.Require().NoError(err)

	label := appA.App.WasmxKeeper.GetRoleLabelByContract(appA.Context(), contractAddress2)
	s.Require().Equal("somerole", label)

	role := appA.App.WasmxKeeper.GetRoleByLabel(appA.Context(), label)
	s.Require().NotNil(role)
	s.Require().Equal("somerole", role.Role)

	appA.Chain.CommitBlock()

	// now it is undeterministic
	isDifferent := false
	for i := 0; i < 10; i++ {
		qres = appA.WasmxQuery(sender, contractAddress2, types.WasmxExecutionMessage{Data: []byte{}}, nil, nil)
		bz, _ := hex.DecodeString(qres)
		fmt.Println("nondeterministic:", string(bz))
		if qresD != qres {
			isDifferent = true
			break
		}
	}
	s.Require().True(isDifferent, "should be nondeterministic")
}

func (suite *KeeperTestSuite) TestWasiTinygoSimpleStorage() {
	wasmbin := tinygo.TinyGoSimpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	source := []byte("somesource")
	codeId := appA.StoreCodeWithSource(sender, wasmbin, nil, source)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte("hello")}, "tinygoSimpleStorage", nil)

	key := []byte("storagekey")
	value := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("hello"), value)

	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte(`{"store":{"key":"storagekey","value":"goodbye"}}`)}, nil, nil)

	value = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte(`goodbye`), value)

	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte(`{"load":{"key":"storagekey"}}`)}, nil, nil)
	s.Require().Equal([]byte("goodbye"), resp)

	appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte(`{"store":{"key":"storagekey","value":"hello"}}`)}, nil, nil)

	value = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("goodbye"), value)

	codeInfo, err := appA.App.WasmxKeeper.GetCodeInfo(appA.Context(), codeId)
	suite.Require().NoError(err)
	suite.Require().Equal(string(source), string(codeInfo.Source))
}

func (suite *KeeperTestSuite) TestWasiTinygoCallSimpleStorage() {
	SkipFixmeTests(suite.T(), "TestWasiTinygoCallSimpleStorage")
	wasmbin := tinygo.TinyGoSimpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	depsPy := []string{types.INTERPRETER_PYTHON}

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	codeId := appA.StoreCode(sender, py.PySimpleStorage, depsPy)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`"123"`)}, "simpleContractPy", nil)

	codeId2 := appA.StoreCode(sender, wasmbin, nil)
	contractAddressWrap := appA.InstantiateCode(sender, codeId2, types.WasmxExecutionMessage{Data: []byte("hello")}, "tinygoSimpleStorage", nil)

	key := []byte("pystore")
	value := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("123"), value)

	data := []byte(fmt.Sprintf(`{"wrapStore":{"address":"%s","key":"storagekey","value":"goodbye"}}`, contractAddress.String()))
	appA.ExecuteContract(sender, contractAddressWrap, types.WasmxExecutionMessage{Data: data}, nil, nil)

	value = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte(`goodbye`), value)

	resp := appA.WasmxQueryRaw(sender, contractAddressWrap, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"wrapLoad":{"address":"%s","key":"storagekey"}}`, contractAddress.String()))}, nil, nil)
	s.Require().Equal([]byte("goodbye23"), resp)
}
