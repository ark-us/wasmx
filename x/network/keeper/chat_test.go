package keeper_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simulation "github.com/cosmos/cosmos-sdk/types/simulation"

	abci "github.com/cometbft/cometbft/abci/types"

	// ibctesting "mythos/v1/testutil/ibc"

	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

type NetworkNode struct {
	Id   string `json:"id"`
	Host string `json:"host"`
	Port string `json:"port"`
	Ip   string `json:"ip"`
}

type NodeInfo struct {
	Address sdk.AccAddress `json:"address"`
	Node    NetworkNode    `json:"node"`
}

type ChatMessage struct {
	RoomId    string         `json:"roomId"`
	Message   string         `json:"message"`
	Timestamp string         `json:"timestamp"`
	Sender    sdk.AccAddress `json:"sender"`
}

type ChatRoom struct {
	RoomId string     `json:"roomId"`
	Peers  []NodeInfo `json:"peers"`
}

type ChatHeader struct {
	Height     int64     `json:"height"`
	Time       time.Time `json:"time"`
	ParentHash []byte    `json:"parent_hash"`
	DataHash   []byte    `json:"data_hash"`
}

type ChatBlock struct {
	Header ChatHeader `json:"header"`
	Data   []byte     `json:"data"`
}

func (suite *KeeperTestSuite) TestChat() {
	sender := suite.GetRandomAccount()
	sender2 := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(10_000_000_000)
	appA := s.AppContext()
	denom := appA.Chain.Config.BaseDenom
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(denom, initBalance))
	appA.Faucet.Fund(appA.Context(), sender2.Address, sdk.NewCoin(denom, initBalance))
	chatAddress, found := appA.App.WasmxKeeper.GetContractAddressByRole(appA.Context(), wasmxtypes.ROLE_CHAT)
	s.Require().True(found)

	msg := []byte(`{"JoinRoom":{"roomId":"room101"}}`)
	_, err := suite.broadcastMessage(msg, sender, chatAddress)
	suite.Require().NoError(err)

	// we just need sender2 to send a transaction, in order to have its public key registered
	contractAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_IDENTITY)
	internalmsg := wasmxtypes.WasmxExecutionMessage{Data: appA.Hex2bz("aa0000000000000000000000000000000000000000000000000000000077")}
	appA.ExecuteContract(sender2, contractAddress, internalmsg, nil, nil)

	msg = []byte(`{"GetRooms":{}}`)
	qresp, err := suite.App().NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   sender.Address.String(),
		Contract: chatAddress.String(),
		Msg:      msg,
	})
	suite.Require().NoError(err)
	qrespbz := appA.QueryDecode(qresp.Data)
	var rooms []ChatRoom
	err = json.Unmarshal(qrespbz, &rooms)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(rooms))
	suite.Require().Equal("room101", rooms[0].RoomId)
	suite.Require().Equal(1, len(rooms[0].Peers))
	suite.Require().Equal(sender.Address, rooms[0].Peers[0].Address)

	suite.sendMessage(sender, "room101", "hello1", chatAddress)
	suite.sendMessage(sender, "room101", "hello2", chatAddress)
	suite.sendMessage(sender, "room101", "hello3", chatAddress)
	suite.sendMessage(sender, "room101", "hello4", chatAddress)
	suite.sendMessage(sender, "room101", "hello5", chatAddress)
	suite.sendMessage(sender, "room101", "hello6", chatAddress)
	suite.sendMessage(sender, "room101", "hello7", chatAddress)
	suite.sendMessage(sender, "room101", "hello8", chatAddress)
	suite.sendMessage(sender, "room101", "hello9", chatAddress)
	suite.sendMessage(sender, "room101", "hello10", chatAddress)
	suite.sendMessage(sender, "room101", "hello11", chatAddress)

	block := suite.getBlock(sender, "room101", 11, chatAddress)
	suite.Require().Equal(int64(11), block.Header.Height)
	decodedTx, err := appA.App.TxDecode(block.Data)
	suite.Require().NoError(err)
	msgExec, ok := decodedTx.GetMsgs()[0].(*wasmxtypes.MsgExecuteContract)
	suite.Require().True(ok)
	var msgexecmsg wasmxtypes.WasmxExecutionMessage
	err = json.Unmarshal(msgExec.Msg, &msgexecmsg)
	suite.Require().NoError(err)
	suite.Require().Contains(string(msgexecmsg.Data), "hello11")
	suite.sendMessageFromOther(sender2, "room101", "hello12", chatAddress)

	msg = []byte(`{"GetBlock":{"roomId":"room101","index":12}}`)
	qrespbz = suite.queryChat(sender, chatAddress, msg)
	err = json.Unmarshal(qrespbz, &block)
	suite.Require().NoError(err)

	msg = []byte(`{"GetBlocks":{"roomId":"room101"}}`)
	qrespbz = suite.queryChat(sender, chatAddress, msg)
	blocksbz := qrespbz
	var blocks []ChatBlock
	err = json.Unmarshal(qrespbz, &blocks)
	suite.Require().NoError(err)
	suite.Require().Equal(12, len(blocks))

	msg = []byte(`{"GetMessages":{"roomId":"room101"}}`)
	qrespbz = suite.queryChat(sender, chatAddress, msg)
	var messages []ChatMessage
	err = json.Unmarshal(qrespbz, &messages)
	suite.Require().NoError(err)
	suite.Require().Equal(12, len(messages))
	suite.Require().Equal("hello1", messages[0].Message)
	suite.Require().Equal("hello12", messages[11].Message)

	// Test chat verifier
	appA = s.AppContext()
	verifierAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_CHAT_VERIFIER)
	msg = []byte(fmt.Sprintf(`{"VerifyConversation":{"blocks":%s}}`, string(blocksbz)))
	qresp, err = suite.App().NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   sender.Address.String(),
		Contract: verifierAddress.String(),
		Msg:      msg,
	})
	suite.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	suite.Require().Equal(uint8(1), qrespbz[0])
}

func (suite *KeeperTestSuite) getBlock(sender simulation.Account, room string, index int64, contractAddress sdk.AccAddress) ChatBlock {
	msg := []byte(fmt.Sprintf(`{"GetBlock":{"roomId":"%s","index":%d}}`, room, index))
	qrespbz := suite.queryChat(sender, contractAddress, msg)
	var value ChatBlock
	err := json.Unmarshal(qrespbz, &value)
	suite.Require().NoError(err)
	return value
}

func (suite *KeeperTestSuite) queryChat(sender simulation.Account, contractAddress sdk.AccAddress, msg []byte) []byte {
	appA := s.AppContext()
	qresp, err := suite.App().NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   sender.Address.String(),
		Contract: contractAddress.String(),
		Msg:      msg,
	})
	suite.Require().NoError(err)
	return appA.QueryDecode(qresp.Data)
}

func (suite *KeeperTestSuite) sendMessage(sender simulation.Account, room string, msgstr string, contractAddress sdk.AccAddress) {
	msg := []byte(fmt.Sprintf(`{"SendMessage":{"roomId":"%s","message":"%s"}}`, room, msgstr))

	_, err := suite.broadcastMessage(msg, sender, contractAddress)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) sendMessageFromOther(sender simulation.Account, room string, msgstr string, contractAddress sdk.AccAddress) {
	appA := s.AppContext()

	msg := []byte(fmt.Sprintf(`{"SendMessage":{"roomId":"%s","message":"%s"}}`, room, msgstr))
	msgwrap := &wasmxtypes.WasmxExecutionMessage{Data: msg}
	msgbz, err := json.Marshal(msgwrap)
	suite.Require().NoError(err)
	msgexec := &wasmxtypes.MsgExecuteContract{
		Sender:   sender.Address.String(),
		Contract: contractAddress.String(),
		Msg:      msgbz,
	}
	txdata := appA.PrepareCosmosTx(sender, []sdk.Msg{msgexec}, nil, nil)

	block := ChatBlock{
		Header: ChatHeader{Height: 3, Time: time.Now(), ParentHash: []byte{1, 2}, DataHash: []byte{1, 2}},
		Data:   txdata,
	}
	blockbz, err := json.Marshal(&block)

	msgreceived64 := base64.StdEncoding.EncodeToString(blockbz)
	msgreceived := []byte(fmt.Sprintf(`{"roomId":"%s","message":"%s","timestamp":"2024-03-19T12:20:26.924Z","sender":{"id":"12D3KooWRgN1dUKvDM9zu8gB7DQhLYVqEEpxBTRK3YYtr8Bj92xD","host":"","port":"","ip":"/ip4/127.0.0.1/tcp/5001"}}`, room, msgreceived64))
	_, err = suite.App().NetworkKeeper.ExecuteEntryPoint(appA.Context(), wasmxtypes.ENTRY_POINT_P2P_MSG, &types.MsgExecuteContract{
		Sender:   sender.Address.String(),
		Contract: contractAddress.String(),
		Msg:      msgreceived,
	})
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) broadcastMessage(msg []byte, sender simulation.Account, contractAddress sdk.AccAddress) (*abci.ExecTxResult, error) {
	appA := s.AppContext()
	msgwrap := &wasmxtypes.WasmxExecutionMessage{Data: msg}
	msgbz, err := json.Marshal(msgwrap)
	suite.Require().NoError(err)
	msgexec := &wasmxtypes.MsgExecuteContract{
		Sender:   sender.Address.String(),
		Contract: contractAddress.String(),
		Msg:      msgbz,
	}
	return appA.BroadcastTxAsync(sender, msgexec)
}
