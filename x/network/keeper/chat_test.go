package keeper_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simulation "github.com/cosmos/cosmos-sdk/types/simulation"

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

func (suite *KeeperTestSuite) TestChat() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(10_000_000_000)
	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))

	msg := []byte(`{"JoinRoom":{"roomId":"room101"}}`)
	_, err := suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   sender.Address.String(),
		Contract: wasmxtypes.ROLE_CHAT,
		Msg:      msg,
	})
	suite.Require().NoError(err)

	msg = []byte(`{"GetRooms":{}}`)
	qresp, err := suite.App().NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   sender.Address.String(),
		Contract: wasmxtypes.ROLE_CHAT,
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

	suite.sendMessage(sender, "room101", "hello1")
	suite.sendMessage(sender, "room101", "hello2")
	suite.sendMessage(sender, "room101", "hello3")
	suite.sendMessage(sender, "room101", "hello4")
	suite.sendMessage(sender, "room101", "hello5")
	suite.sendMessage(sender, "room101", "hello6")
	suite.sendMessage(sender, "room101", "hello7")
	suite.sendMessage(sender, "room101", "hello8")
	suite.sendMessage(sender, "room101", "hello9")
	suite.sendMessage(sender, "room101", "hello10")
	suite.sendMessage(sender, "room101", "hello11")
	suite.sendMessageFromOther(sender, "room101", "hello12")

	msg = []byte(`{"GetMessages":{"roomId":"room101"}}`)
	qresp, err = suite.App().NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   sender.Address.String(),
		Contract: wasmxtypes.ROLE_CHAT,
		Msg:      msg,
	})
	suite.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	var messages []ChatMessage
	err = json.Unmarshal(qrespbz, &messages)
	suite.Require().NoError(err)
	suite.Require().Equal(12, len(messages))
	suite.Require().Equal("hello1", messages[0].Message)
	suite.Require().Equal("hello12", messages[11].Message)
}

func (suite *KeeperTestSuite) sendMessage(sender simulation.Account, room string, msgstr string) {
	appA := s.AppContext()
	msg := []byte(fmt.Sprintf(`{"SendMessage":{"roomId":"%s","message":"%s"}}`, room, msgstr))
	_, err := suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   sender.Address.String(),
		Contract: wasmxtypes.ROLE_CHAT,
		Msg:      msg,
	})
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) sendMessageFromOther(sender simulation.Account, room string, msgstr string) {
	appA := s.AppContext()
	msg64 := base64.StdEncoding.EncodeToString([]byte(msgstr))
	msg := []byte(fmt.Sprintf(`{"roomId":"%s","message":"%s","timestamp":"2024-03-19T12:20:26.924Z","sender":{"id":"12D3KooWRgN1dUKvDM9zu8gB7DQhLYVqEEpxBTRK3YYtr8Bj92xD","host":"","port":"","ip":"/ip4/127.0.0.1/tcp/5001"}}`, room, msg64))
	_, err := suite.App().NetworkKeeper.ExecuteEntryPoint(appA.Context(), wasmxtypes.ENTRY_POINT_P2P_MSG, &types.MsgExecuteContract{
		Sender:   sender.Address.String(),
		Contract: wasmxtypes.ROLE_CHAT,
		Msg:      msg,
	})
	suite.Require().NoError(err)
}
