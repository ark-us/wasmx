package vm

import (
	_ "embed"

	"encoding/json"

	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	multiaddr "github.com/multiformats/go-multiaddr"

	"github.com/second-state/WasmEdge-go/wasmedge"

	host "github.com/libp2p/go-libp2p/core/host"

	ed25519 "github.com/cometbft/cometbft/crypto/ed25519"

	vmtypes "mythos/v1/x/wasmx/vm"
	asmem "mythos/v1/x/wasmx/vm/memory/assemblyscript"
)

func StartNodeWithIdentity(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--StartNodeWithIdentity--")
	ctx := _context.(*Context)
	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	fmt.Println("--StartNodeWithIdentity--", string(requestbz))
	var req StartNodeWithIdentityRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	node, err := startNodeWithIdentityInternal(req.PrivateKey, req.Port, req.ProtocolId)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.node = &node

	ctx.Context.GoRoutineGroup.Go(func() error {
		intervalEnded := make(chan bool, 1)
		defer close(intervalEnded)
		go func(ctx_ *Context) {
			fmt.Println("goroutine node started")
			defer fmt.Println("goroutine node finished")

			err := startNodeListeners(*ctx_.node, req.ProtocolId)
			if err != nil {
				intervalEnded <- true
			}
		}(ctx)

		select {
		case <-ctx.Context.GoContextParent.Done():
			return nil
		case <-intervalEnded:
			return nil
		}
		return nil
	})

	response := StartNodeWithIdentityResponse{Error: "", Data: make([]byte, 0)}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.Context.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func ConnectPeer(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--ConnectPeer--")
	ctx := _context.(*Context)
	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	fmt.Println("--ConnectPeer--", string(requestbz))
	var req ConnectPeerRequest
	err = json.Unmarshal(requestbz, &req)

	stream, err := connectPeerInternal(*ctx.node, req.ProtocolId, req.Peer)
	ctx.streams[req.Peer] = stream

	ctx.Context.GoRoutineGroup.Go(func() error {
		intervalEnded := make(chan bool, 1)
		defer close(intervalEnded)
		go func(ctx_ *Context) {
			fmt.Println(fmt.Sprintf("goroutine peer connect started: %s", req.Peer))
			defer fmt.Println(fmt.Sprintf("goroutine peer connect finished: %s", req.Peer))

			fmt.Println("--connectPeerInternal ctx_--", ctx_)
			fmt.Println("--connectPeerInternal ctx_.node--", ctx_.node)
			stream, found := ctx_.streams[req.Peer]
			fmt.Println("--connectPeerInternal found--", found)
			err := listenPeerStream(stream, req.Peer)
			fmt.Println("connect peer err", err)
			if err != nil {
				intervalEnded <- true
			}

		}(ctx)

		select {
		case <-ctx.Context.GoContextParent.Done():
			return nil
		case <-intervalEnded:
			return nil
		}
		return nil
	})

	response := ConnectPeerResponse{}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.Context.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func SendMessage(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--SendMessage--")
	ctx := _context.(*Context)
	ptr, err := asmem.AllocateWriteMem(ctx.Context.MustGetVmFromContext(), callframe, make([]byte, 32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func SendMessageToPeers(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--SendMessageToPeers--")
	ctx := _context.(*Context)
	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	fmt.Println("--SendMessageToPeers--", string(requestbz))

	var req SendMessageToPeersRequest
	err = json.Unmarshal(requestbz, &req)
	fmt.Println("--SendMessageToPeers err--", err)
	fmt.Println("--SendMessageToPeers req--", req)
	if err != nil {
		fmt.Println("--SendMessageToPeers is err!--", err)
		return nil, wasmedge.Result_Fail
	}
	fmt.Println("--SendMessageToPeers 11--")

	for _, peer := range req.Peers {
		_, found := ctx.streams[peer]
		if !found {
			stream, err := connectPeerInternal(*ctx.node, req.ProtocolId, peer)
			if err != nil {
				fmt.Println("--connectPeerInternal is err!--", err)
				return nil, wasmedge.Result_Fail
			}
			ctx.streams[peer] = stream
		}
	}

	ctx.Context.GoRoutineGroup.Go(func() error {
		intervalEnded := make(chan bool, 1)
		defer close(intervalEnded)
		go func(ctx_ *Context) {
			fmt.Println(fmt.Sprintf("goroutine peers send message started: %s", req.Peers))
			defer fmt.Println(fmt.Sprintf("goroutine peers send message finished: %s", req.Peers))

			for _, peer := range req.Peers {
				stream, found := ctx.streams[peer]
				if !found {
					fmt.Println("stream not found: ", peer)
					intervalEnded <- true
				}
				err := sendMessageToPeersInternal(*ctx_.node, stream, peer, req.Msg)
				fmt.Println("send message err", err)
				if err != nil {
					intervalEnded <- true
				}
			}

			// intervalEnded <- true
		}(ctx)

		select {
		case <-ctx.Context.GoContextParent.Done():
			return nil
		case <-intervalEnded:
			return nil
		}
		return nil
	})

	// for _, peer := range req.Peers {
	// 	sendMessageToPeersInternal(*ctx.node, req.ProtocolId, peer, req.Msg)
	// }

	fmt.Println("--SendMessageToPeers 22--")

	response := SendMessageToPeersResponse{}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.Context.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	fmt.Println("--SendMessageToPeers END--")
	return returns, wasmedge.Result_Success
}

func CloseNode(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--CloseNode--")
	// ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func BuildWasmxP2P1(context *vmtypes.Context) *wasmedge.Module {
	ctx := &Context{Context: *context}
	env := wasmedge.NewModule(HOST_WASMX_ENV_P2P)
	functype__i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	env.AddFunction("StartNodeWithIdentity", wasmedge.NewFunction(functype_i32_i32, StartNodeWithIdentity, ctx, 0))
	env.AddFunction("CloseNode", wasmedge.NewFunction(functype__i32, CloseNode, ctx, 0))
	env.AddFunction("ConnectPeer", wasmedge.NewFunction(functype_i32_i32, ConnectPeer, ctx, 0))
	env.AddFunction("SendMessage", wasmedge.NewFunction(functype_i32_i32, SendMessage, ctx, 0))
	env.AddFunction("SendMessageToPeers", wasmedge.NewFunction(functype_i32_i32, SendMessageToPeers, ctx, 0))
	return env
}

func startNodeWithIdentityInternal(_pk []byte, port string, protocolID string) (host.Host, error) {
	pk := ed25519.PrivKey(_pk)
	pkcrypto, err := crypto.UnmarshalEd25519PrivateKey(pk)
	if err != nil {
		return nil, err
	}
	identity := libp2p.Identity(pkcrypto)

	// start a libp2p node that listens on a random local TCP port,
	// but without running the built-in ping protocol
	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/"+port),
		libp2p.Ping(false),
		identity,
	)

	if err != nil {
		return nil, err
	}
	return node, nil
}

func startNodeListeners(node host.Host, protocolID string) error {
	// configure our own ping protocol
	// pingService := &ping.PingService{Host: node}
	// node.SetStreamHandler(ping.ID, pingService.PingHandler)

	// print the node's PeerInfo in multiaddr format
	peerInfo := peerstore.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}
	fmt.Println("peer ID:", peerInfo.ID)
	addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		return err
	}
	fmt.Println("libp2p node address:", addrs[0])
	fmt.Println("libp2p node address:", addrs)
	node.SetStreamHandler(protocol.ID(protocolID), handleStream)
	return nil
}

func connectPeerInternal(node host.Host, protocolID string, peeraddrstr string) (network.Stream, error) {
	fmt.Println("--connectPeerInternal--", node, protocolID, peeraddrstr)
	ctx := context.Background()
	peeraddr, err := multiaddr.NewMultiaddr(peeraddrstr)
	if err != nil {
		return nil, err
	}
	fmt.Println("--connectPeerInternal peeraddr--", peeraddr)
	peer, err := peerstore.AddrInfoFromP2pAddr(peeraddr)
	if err != nil {
		return nil, err
	}
	fmt.Println("--connectPeerInternal peer--", peer)
	if err := node.Connect(context.Background(), *peer); err != nil {
		return nil, err
	}

	// open a stream, this stream will be handled by handleStream other end
	stream, err := node.NewStream(ctx, peer.ID, protocol.ID(protocolID))
	if err != nil {
		fmt.Println("Stream open failed", err)
		return nil, err
	}
	return stream, nil
}

func listenPeerStream(stream network.Stream, peeraddrstr string) error {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	// go writeDataStd(rw, "peerstream: "+peeraddrstr)
	go readDataStd(rw, "peerstream: "+peeraddrstr)
	fmt.Println("Connected to:", peeraddrstr)
	return nil
}

func sendMessageToPeersInternal(node host.Host, stream network.Stream, peeraddrstr string, msg []byte) error {
	fmt.Println("--sendMessageToPeersInternal--", string(msg))
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	fmt.Println("--sendMessageToPeersInternal - Connected to:", peeraddrstr)

	err := writeData(rw, msg)
	return err
}

func handleStream(stream network.Stream) {
	fmt.Println("Got a new stream!")

	// Create a buffer stream for non-blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	// fmt.Println("Received message", str, err)

	go readDataStd(rw, "mainstream")
	// go writeData(rw, "mainstream")

	// 'stream' will stay open until you close it (or the other side closes it).
}

func writeData(rw *bufio.ReadWriter, msg []byte) error {
	fmt.Println("writing stream data: ", string(msg))
	// _, err := rw.Write(msg)
	_, err := rw.WriteString(string(msg) + "\n")
	if err != nil {
		fmt.Println("Error writing to buffer")
		return err
	}
	err = rw.Flush()
	if err != nil {
		fmt.Println("Error flushing buffer")
		return err
	}
	return nil
}

func readDataStd(rw *bufio.ReadWriter, frompeer string) {
	fmt.Println("reading stream data from peer: ", frompeer)
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer")
			panic(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			// fmt.Printf("\x1b[32m%s\x1b[0m> ", str+" - ")
			fmt.Printf(str)
		}

	}
}

func writeDataStd(rw *bufio.ReadWriter, msg string) {
	fmt.Println("writing stream data: ", msg)
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}
