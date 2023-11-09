package keeper_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"

	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

var tstoreprefix = []byte{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 40}
var bzkey = []byte{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 40, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}

func (suite *KeeperTestSuite) TestSetValidators() {
	ctx := context.Background()
	client, conn := grpcClient(suite.T(), ctx)
	defer conn.Close()
	resp, err := client.SetValidators(ctx, &types.MsgSetValidators{})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", resp)

	fmt.Println("-----storage before-execution---")
	app := suite.GetApp(suite.chainA)
	bz, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")
	tstorer := app.CommitMultiStore().GetKVStore(app.GetMKey(wasmxtypes.MemStoreKey))
	fmt.Println("-----GET-----0000000000000000000000000000000000000000000000000000000000000001", tstorer.Get(append(tstoreprefix, bz...)))
	bz, _ = hex.DecodeString("b10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6")
	fmt.Println("------GET----b10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6", tstorer.Get(append(tstoreprefix, bz...)))

	// Test for output here.
	fmt.Println("=====GetValidators====")
	resp2, err := client.GetValidators(ctx, &types.MsgGetValidators{})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", resp2)
}

func (suite *KeeperTestSuite) TestSetValidators2() {
	ctx := context.Background()
	client, conn := grpcClient(suite.T(), ctx)
	resp, err := client.SetValidators(ctx, &types.MsgSetValidators{})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", resp)
	conn.Close()

	ctx = context.Background()
	client, conn = grpcClient(suite.T(), ctx)
	fmt.Println("=====GetValidators====")
	resp2, err := client.GetValidators(ctx, &types.MsgGetValidators{})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", resp2)
	conn.Close()
}
