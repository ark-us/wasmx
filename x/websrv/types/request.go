package types

import (
	fmt "fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	aabi "github.com/ethereum/go-ethereum/accounts/abi"
)

var ModuleAddress = sdk.AccAddress([]byte(ModuleName))

var HttpRequestGetAbiStr = `[{"inputs":[{"components":[{"components":[{"internalType":"string","name":"Path","type":"string"},{"components":[{"internalType":"string","name":"Key","type":"string"},{"internalType":"string","name":"Value","type":"string"}],"internalType":"struct WebServer.RequestParam[]","name":"Params","type":"tuple[]"}],"internalType":"struct WebServer.RequestUrl","name":"Url","type":"tuple"}],"internalType":"struct WebServer.HttpRequestGet","name":"request","type":"tuple"}],"name":"get","outputs":[{"components":[{"internalType":"string","name":"Content","type":"string"},{"internalType":"string","name":"ContentType","type":"string"}],"internalType":"struct WebServer.HttpRequestGetResponse","name":"","type":"tuple"}],"stateMutability":"view","type":"function"},{"inputs":[{"components":[{"components":[{"internalType":"string","name":"Path","type":"string"},{"components":[{"internalType":"string","name":"Key","type":"string"},{"internalType":"string","name":"Value","type":"string"}],"internalType":"struct WebServer.RequestParam[]","name":"Params","type":"tuple[]"}],"internalType":"struct WebServer.RequestUrl","name":"Url","type":"tuple"}],"internalType":"struct WebServer.HttpRequestGet","name":"request","type":"tuple"}],"name":"getStr","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[{"components":[{"components":[{"internalType":"string","name":"Path","type":"string"},{"components":[{"internalType":"string","name":"Key","type":"string"},{"internalType":"string","name":"Value","type":"string"}],"internalType":"struct WebServer.RequestParam[]","name":"Params","type":"tuple[]"}],"internalType":"struct WebServer.RequestUrl","name":"Url","type":"tuple"}],"internalType":"struct WebServer.HttpRequestGet","name":"request","type":"tuple"}],"name":"getStr2","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"}]`

var HttpRequestGetAbi aabi.ABI

func init() {
	abi, err := aabi.JSON(strings.NewReader(HttpRequestGetAbiStr))
	if err != nil {
		panic(err)
	}
	HttpRequestGetAbi = abi
}

func RequestGetEncodeAbi(request HttpRequestGet) ([]byte, error) {
	return HttpRequestGetAbi.Pack(
		"get",
		request,
	)
}

type HttpRequestGetResponseContract struct {
	Content     string `json:"Content"`
	ContentType string `json:"ContentType"`
}

type HttpRequestGetResponseContractAbi struct {
	Message HttpRequestGetResponseContract `json:"message"`
}

func ResponseGetDecodeAbi(data []byte) (*HttpRequestGetResponse, error) {
	unpacked, err := HttpRequestGetAbi.Methods["get"].Outputs.Unpack(data)
	if err != nil {
		return nil, err
	}

	var tuple HttpRequestGetResponseContractAbi
	err = HttpRequestGetAbi.Methods["get"].Outputs.Copy(&tuple, unpacked)
	if err != nil {
		return nil, err
	}
	r := tuple.Message
	fmt.Println("----response", r)

	response := HttpRequestGetResponse{
		Content:     []byte(r.Content),
		ContentType: r.ContentType,
	}
	return &response, nil
}

// func ResponseGetDecodeAbi(data []byte) (string, error) {
// 	result, err := HttpRequestGetAbi.Methods["getStr"].Outputs.Unpack(data)
// 	if err != nil {
// 		return "", err
// 	}
// 	content := result[0].(string)
// 	return content, nil
// }
