package keeper_test

import (
	"context"
	"fmt"
	"log"

	"mythos/v1/x/network/types"
)

func (suite *KeeperTestSuite) TestSetValidators() {
	ctx := context.Background()
	client, conn := grpcClient(suite.T(), ctx)
	defer conn.Close()
	resp, err := client.SetValidators(ctx, &types.MsgSetValidators{})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", resp)
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
