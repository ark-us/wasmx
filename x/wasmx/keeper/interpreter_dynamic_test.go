package keeper_test

import (
	_ "embed"
	"encoding/hex"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	testdata "mythos/v1/x/wasmx/keeper/testdata/classic"
	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm/precompiles"
)

func (suite *KeeperTestSuite) TestDynamicInterpreter() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1_000_000_000_000_000_000)
	valAccount := simulation.Account{
		PrivKey: s.chainA.SenderPrivKey,
		PubKey:  s.chainA.SenderPrivKey.PubKey(),
		Address: s.chainA.SenderAccount.GetAddress(),
	}

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), valAccount.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	wasmbin := precompiles.GetPrecompileByLabel(types.INTERPRETER_EVM_SHANGHAI)
	codeId := appA.StoreCode(sender, wasmbin, nil)
	interpreterAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "newinterpreter", nil)

	newlabel := types.INTERPRETER_EVM_SHANGHAI + "2"

	// Register contract role proposal
	proposal := types.NewRegisterRoleProposal("Register interpreter", "Register interpreter", "interpreter", newlabel, interpreterAddress.String())

	appA.PassGovProposal(valAccount, sender, proposal)

	resp := appA.App.WasmxKeeper.GetRoleLabelByContract(appA.Context(), interpreterAddress)
	s.Require().Equal(newlabel, resp)

	role := appA.App.WasmxKeeper.GetRoleByLabel(appA.Context(), newlabel)
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
	_, contractAddress := appA.Deploy(sender, evmcode, []string{newlabel}, types.WasmxExecutionMessage{Data: initvaluebz}, nil, "simpleStorage")

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(initvalue, hex.EncodeToString(queryres))

	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(setHex + "0000000000000000000000000000000000000000000000000000000000000006")}, nil, nil)

	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal("0000000000000000000000000000000000000000000000000000000000000006", hex.EncodeToString(queryres))

}

func (suite *KeeperTestSuite) TestWasmxDebug() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	evmcode, err := hex.DecodeString(testdata.SimpleStorage)
	s.Require().NoError(err)
	initvalue := "0000000000000000000000000000000000000000000000000000000000000009"
	initvaluebz, err := hex.DecodeString(initvalue)
	s.Require().NoError(err)
	_, caddr := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: initvaluebz}, nil, "simpleStorage")
	qres, memsnap, err := appA.WasmxQueryDebug(sender, caddr, types.WasmxExecutionMessage{Data: appA.Hex2bz("6d4ce63c")}, nil, nil)
	s.Require().NoError(err)
	moduleMem := parseMem(memsnap)
	s.Require().Equal(int32(117), moduleMem.Pc, "wrong pc")
	s.Require().Equal(byte(0x5b), moduleMem.PcOpcode, "wrong opcode")
	s.Require().Equal(initvalue, qres)
}

type EWasmMemory struct {
	Stack             []byte
	WordCount         int64
	InterpreterMemory []byte
	ContractMemory    []byte
	Pc                int32
	PcOpcode          byte
	BytecodeOffset    int64
	PcOffset          int64
}

func parseMem(mem []byte) EWasmMemory {
	start := 0
	stackSize := 1024 * 32 // 1024 slots - 32768 bytes
	fmt.Println("stackSize", len(mem), stackSize)
	// stack := make([]byte, stackSize)
	// copy(stack, mem[start: stackSize])
	stack := mem[start:stackSize]
	fmt.Println("stack", len(stack))

	start = stackSize
	fmt.Println(hex.EncodeToString(mem[start : start+32]))
	// how many 32 byte words are stored in memory
	wordCount := big.NewInt(0).SetBytes(mem[start : start+4]).Int64()
	fmt.Println("wordCount", wordCount)
	if wordCount == 0 {
		wordCount = 2000
	}

	// mem cost, scratch space, keccak, bytecode, actual used memory
	// start += 4 + 4 + 32 + 1024 + 28800
	start = 62632
	interpreterMemory := mem[start:(int64(start) + wordCount*32)]

	// interpreterMemory = mem[start:(start + 60000)]
	fmt.Println("--interpreterMemory", hex.EncodeToString(interpreterMemory))

	// interpreter-specific
	memOffset := 0x140
	pcOffset := 0x160
	bytecodeOffset := 0x100

	memPtrBz := mloadEwasmMem(interpreterMemory, memOffset)
	pcPtrBz := mloadEwasmMem(interpreterMemory, pcOffset)
	bytecodePtrBz := mloadEwasmMem(interpreterMemory, bytecodeOffset)

	fmt.Println("--memPtrBz-", memPtrBz)
	fmt.Println("--pcPtrBz-", pcPtrBz)
	fmt.Println("--bytecodePtrBz-", bytecodePtrBz)
	memPtr := big.NewInt(0).SetBytes(memPtrBz).Int64()
	pcPtr := big.NewInt(0).SetBytes(pcPtrBz).Int64()
	bytecodePtr := big.NewInt(0).SetBytes(bytecodePtrBz).Int64()

	truepc := pcPtr - bytecodePtr

	fmt.Println("--memPtr-", memPtr)
	fmt.Println("--pcPtr-", pcPtr)
	fmt.Println("--bytecodePtr-", bytecodePtr)
	fmt.Println("--truepc-", truepc)

	pc := mloadEwasmMem(interpreterMemory, int(pcPtr))
	fmt.Println("--pcPtr-", pc, hex.EncodeToString(pc))

	return EWasmMemory{
		Stack:             stack,
		WordCount:         wordCount,
		InterpreterMemory: interpreterMemory,
		ContractMemory:    interpreterMemory[memPtr:],
		Pc:                int32(truepc),
		PcOpcode:          pc[0],
		BytecodeOffset:    bytecodePtr,
		PcOffset:          pcPtr,
	}
}

func mloadEwasmMem(mem []byte, offset int) []byte {
	return mem[offset : offset+32]
}
