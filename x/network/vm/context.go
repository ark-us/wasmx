package vm

import (
	"bufio"
	"fmt"
	"math/big"

	log "cosmossdk.io/log"

	network "github.com/libp2p/go-libp2p/core/network"

	wasmxtypes "mythos/v1/x/wasmx/types"
	wasmxvm "mythos/v1/x/wasmx/vm"
	vmtypes "mythos/v1/x/wasmx/vm/types"
)

// main stream
func (c *Context) handleStream(stream network.Stream) {
	// Create a buffer stream for non-blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readDataStd(c.Context.Ctx.Logger(), rw, "mainstream", c.handleMessage)
	// go writeData(rw, "mainstream")
}

// peer stream
func (c *Context) listenPeerStream(stream network.Stream, peeraddrstr string) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	// go writeDataStd(rw, "peerstream: "+peeraddrstr)
	go readDataStd(c.Context.Ctx.Logger(), rw, "peerstream: "+peeraddrstr, c.handleMessage)
	c.Context.Ctx.Logger().Debug("Connected to:", peeraddrstr)
}

func (c *Context) handleMessage(msg []byte) {
	// TODO address should be in the request
	addr, err := c.Context.CosmosHandler.GetAddressOrRole(c.Context.Ctx, wasmxtypes.ROLE_CONSENSUS)
	if err != nil {
		c.Context.Ctx.Logger().Debug(fmt.Sprintf("p2p message execution failed: %s", "cannot find consensus role address"))
	}
	contractContext := wasmxvm.GetContractContext(c.Context, addr)

	req := vmtypes.CallRequest{
		To:       c.Context.Env.Contract.Address,
		From:     c.Context.Env.Contract.Address, // TODO from?
		Value:    big.NewInt(0),
		GasLimit: big.NewInt(100000000), // TODO
		Calldata: msg,
		Bytecode: contractContext.ContractInfo.Bytecode,
		CodeHash: contractContext.ContractInfo.CodeHash,
		IsQuery:  false,
	}
	success, data := wasmxvm.WasmxCall(c.Context, req)
	if success > 0 {
		c.Context.Ctx.Logger().Debug(fmt.Sprintf("p2p message execution failed: %s", string(data)))
	}
}

func readDataStd(logger log.Logger, rw *bufio.ReadWriter, frompeer string, handleMessage func(msg []byte)) {
	logger.Debug("reading stream data from peer", "peer", frompeer)
	for {
		msgbz, err := rw.ReadBytes('\n')
		if err != nil {
			logger.Debug("Error reading from buffer", "peer", frompeer)
			panic(err)
		}

		if len(msgbz) == 0 {
			return
		}
		if string(msgbz) != "\n" {
			handleMessage(msgbz)
		}
	}
}
