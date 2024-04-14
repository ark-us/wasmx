package keeper_test

import (
	_ "embed"
	"encoding/hex"
	"math/big"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	testdata "mythos/v1/x/wasmx/keeper/testdata/classic"
	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm/precompiles"
)

func (suite *KeeperTestSuite) TestDynamicInterpreter() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)
	valAccount := simulation.Account{
		PrivKey: s.Chain().SenderPrivKey,
		PubKey:  s.Chain().SenderPrivKey.PubKey(),
		Address: s.Chain().SenderAccount.GetAddress(),
	}

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), valAccount.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	wasmbin := precompiles.GetPrecompileByLabel(types.INTERPRETER_EVM_SHANGHAI)
	codeId := appA.StoreCode(sender, wasmbin, nil)
	interpreterAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "newinterpreter", nil)

	newlabel := types.INTERPRETER_EVM_SHANGHAI + "2"

	// Register contract role proposal
	title := "Register interpreter"
	description := "Register interpreter"
	authority := authtypes.NewModuleAddress(types.ROLE_GOVERNANCE).String()
	proposal := &types.MsgRegisterRole{Authority: authority, Title: title, Description: description, Role: "interpreter", Label: newlabel, ContractAddress: interpreterAddress.String()}
	appA.PassGovProposal(valAccount, sender, []sdk.Msg{proposal}, "", title, description, false)

	resp := appA.App.WasmxKeeper.GetRoleLabelByContract(appA.Context(), appA.Context().ChainID(), interpreterAddress)
	s.Require().Equal(newlabel, resp)

	role := appA.App.WasmxKeeper.GetRoleByLabel(appA.Context(), appA.Context().ChainID(), newlabel)
	s.Require().Equal(interpreterAddress.String(), role.ContractAddress)
	s.Require().Equal(newlabel, role.Label)
	s.Require().Equal("interpreter", role.Role)

	// use this interpreter to execute contract
	setHex := `60fe47b1`
	evmcode, err := hex.DecodeString(testdata.SimpleStorage)
	s.Require().NoError(err)

	initvalue := "0000000000000000000000000000000000000000000000000000000000000009"
	initvaluebz, err := hex.DecodeString(initvalue)
	s.Require().NoError(err)
	_, contractAddress := appA.Deploy(sender, evmcode, []string{newlabel}, types.WasmxExecutionMessage{Data: initvaluebz}, nil, "simpleStorage", nil)

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(initvalue, hex.EncodeToString(queryres))

	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(setHex + "0000000000000000000000000000000000000000000000000000000000000006")}, nil, nil)

	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal("0000000000000000000000000000000000000000000000000000000000000006", hex.EncodeToString(queryres))

}

func (suite *KeeperTestSuite) TestWasmxDebug() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	evmcode, err := hex.DecodeString(testdata.SimpleStorage)
	s.Require().NoError(err)
	initvalue := "0000000000000000000000000000000000000000000000000000000000000009"
	initvaluebz, err := hex.DecodeString(initvalue)
	s.Require().NoError(err)
	_, caddr := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: initvaluebz}, nil, "simpleStorage", nil)
	qres, memsnap, err := appA.WasmxQueryDebug(sender, caddr, types.WasmxExecutionMessage{Data: appA.Hex2bz("6d4ce63c")}, nil, nil)
	s.Require().NoError(err)
	moduleMem := parseMem(memsnap)
	s.Require().Equal(int32(117), moduleMem.Pc, "wrong pc")
	s.Require().Equal(byte(0x5b), moduleMem.PcOpcode, "wrong opcode")
	s.Require().Equal(initvalue, qres)
	// fmt.Println(moduleMem.Stack.ToString())
	// fmt.Println("--pc", hex.EncodeToString(moduleMem.InterpreterMemory[moduleMem.PcOffset:moduleMem.PcOffset+64]))
}

func (suite *KeeperTestSuite) TestWasmxDebugPush16() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	evmcode, err := hex.DecodeString("6019600d60003960196000f3fe6fc84a6e6ec1e7f30f5c812eeba420f76960005260206000f3")
	s.Require().NoError(err)
	_, caddr := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "push16", nil)
	qres, memsnap, err := appA.WasmxQueryDebug(sender, caddr, types.WasmxExecutionMessage{Data: []byte{}}, nil, nil)
	s.Require().NoError(err)
	moduleMem := parseMem(memsnap)

	s.Require().Equal("00000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769", qres)
	s.Require().Equal(int32(25), moduleMem.Pc, "wrong pc")
	s.Require().Equal(byte(0x00), moduleMem.PcOpcode, "wrong opcode")
}

type EWasmStack [][32]byte
type EWasmMemory struct {
	Stack             EWasmStack
	WordCount         int64
	InterpreterMemory []byte
	ContractMemory    []byte
	Pc                int32
	PcOpcode          byte
	BytecodeOffset    int64
	PcOffset          int64
}

func (s EWasmStack) ToString() string {
	strs := make([]string, len(s))
	for i, item := range s {
		strs[i] = hex.EncodeToString(item[:])
	}
	return strings.Join(strs, "\n")
}

func parseMem(mem []byte) EWasmMemory {
	if len(mem) == 0 {
		return EWasmMemory{}
	}
	start := 0
	stackSize := 2048 * 32 // 2048 slots - 65536 bytes
	stack := mem[start:stackSize]
	start = stackSize
	// how many 32 byte words are stored in memory
	wordCount := big.NewInt(0).SetBytes(mem[start : start+4]).Int64()
	if wordCount == 0 {
		wordCount = 2000
	}

	// mem cost, scratch space, keccak, bytecode, actual used memory
	// start += 4 + 4 + 32 + 1024 + 40960
	start = 107560
	interpreterMemory := mem[start:(int64(start) + wordCount*32)]

	// interpreter-specific
	memOffset := 0x140
	pcOffset := 0x160
	bytecodeOffset := 0x100

	memPtrBz := mloadEwasmMem(interpreterMemory, memOffset)
	pcPtrBz := mloadEwasmMem(interpreterMemory, pcOffset)
	bytecodePtrBz := mloadEwasmMem(interpreterMemory, bytecodeOffset)
	memPtr := big.NewInt(0).SetBytes(memPtrBz).Int64()
	pcPtr := big.NewInt(0).SetBytes(pcPtrBz).Int64()
	bytecodePtr := big.NewInt(0).SetBytes(bytecodePtrBz).Int64()
	truepc := pcPtr - bytecodePtr
	pc := mloadEwasmMem(interpreterMemory, int(pcPtr))

	return EWasmMemory{
		Stack:             parseStack(stack),
		WordCount:         wordCount,
		InterpreterMemory: interpreterMemory,
		ContractMemory:    interpreterMemory[memPtr:],
		Pc:                int32(truepc),
		PcOpcode:          pc[0],
		BytecodeOffset:    bytecodePtr,
		PcOffset:          pcPtr,
	}
}

func parseStack(data []byte) [][32]byte {
	chunkSize := 32
	numChunks := len(data) / chunkSize
	chunks := make([][32]byte, numChunks)

	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := (i + 1) * chunkSize
		chunk := make([]byte, 32)
		for j, x := range data[start:end] {
			chunk[31-j] = x
		}
		copy(chunks[i][:], chunk)
	}
	return chunks
}

func mloadEwasmMem(mem []byte, offset int) []byte {
	return mem[offset : offset+32]
}
