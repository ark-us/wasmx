package types

import (
	fmt "fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	aabi "github.com/ethereum/go-ethereum/accounts/abi"
)

var ModuleAddress = sdk.AccAddress([]byte(ModuleName))

var HttpRequestGetAbiStr = `[{"inputs":[{"components":[{"components":[{"internalType":"enum HeaderOpt","name":"HeaderType","type":"uint8"},{"internalType":"string","name":"Value","type":"string"}],"internalType":"struct HeaderItem[]","name":"Header","type":"tuple[]"},{"components":[{"internalType":"string","name":"Key","type":"string"},{"internalType":"string","name":"Value","type":"string"}],"internalType":"struct RequestQueryParam[]","name":"QueryParams","type":"tuple[]"}],"internalType":"struct HttpRequest","name":"request","type":"tuple"}],"name":"get","outputs":[{"components":[{"components":[{"internalType":"enum HeaderOpt","name":"HeaderType","type":"uint8"},{"internalType":"string","name":"Value","type":"string"}],"internalType":"struct HeaderItem[]","name":"Header","type":"tuple[]"},{"internalType":"string","name":"Content","type":"string"}],"internalType":"struct HttpResponse","name":"","type":"tuple"}],"stateMutability":"view","type":"function"},{"inputs":[{"components":[{"components":[{"internalType":"enum HeaderOpt","name":"HeaderType","type":"uint8"},{"internalType":"string","name":"Value","type":"string"}],"internalType":"struct HeaderItem[]","name":"Header","type":"tuple[]"},{"components":[{"internalType":"string","name":"Key","type":"string"},{"internalType":"string","name":"Value","type":"string"}],"internalType":"struct RequestQueryParam[]","name":"QueryParams","type":"tuple[]"}],"internalType":"struct HttpRequest","name":"request","type":"tuple"}],"name":"post","outputs":[{"components":[{"components":[{"internalType":"enum HeaderOpt","name":"HeaderType","type":"uint8"},{"internalType":"string","name":"Value","type":"string"}],"internalType":"struct HeaderItem[]","name":"Header","type":"tuple[]"},{"internalType":"string","name":"Content","type":"string"}],"internalType":"struct HttpResponse","name":"","type":"tuple"}],"stateMutability":"payable","type":"function"}]`

var HttpRequestGetAbi aabi.ABI

func init() {
	abi, err := aabi.JSON(strings.NewReader(HttpRequestGetAbiStr))
	if err != nil {
		panic(err)
	}
	HttpRequestGetAbi = abi
}

func RequestGetEncodeAbi(request HttpRequest) ([]byte, error) {
	return HttpRequestGetAbi.Pack(
		"get",
		request,
	)
}


type HttpResponseAbi struct {
	Message HttpResponse `json:"message"`
}

func ResponseGetDecodeAbi(data []byte) (*HttpResponse, error) {
	unpacked, err := HttpRequestGetAbi.Methods["get"].Outputs.Unpack(data)
	if err != nil {
		return nil, err
	}

	var tuple HttpResponseAbi
	err = HttpRequestGetAbi.Methods["get"].Outputs.Copy(&tuple, unpacked)
	if err != nil {
		return nil, err
	}
	r := tuple.Message
	fmt.Println("----response", r)

	return &tuple.Message, nil
}

// func ResponseGetDecodeAbi(data []byte) (string, error) {
// 	result, err := HttpRequestGetAbi.Methods["getStr"].Outputs.Unpack(data)
// 	if err != nil {
// 		return "", err
// 	}
// 	content := result[0].(string)
// 	return content, nil
// }

var CGIsol = `// SPDX-License-Identifier: MIT

pragma solidity >=0.7.0 <0.9.0;

struct RequestQueryParam {
    string Key;
    string Value;
}

struct HttpRequest {
    HeaderItem[] Header;
    RequestQueryParam[] QueryParams;
}

struct HttpResponse {
    HeaderItem[] Header;
    string Content;
}

enum HeaderOpt {
    ContentType,          // Content-Type            // "text/html; charset=UTF-8" Indicates the media type of the resource.
    ContentEncoding,      // Content-Encoding        // Used to specify the compression algorithm.
    ContentLanguage,      // Content-Language        // "en" Describes the human language(s) intended for the audience, so that it allows a user to differentiate according to the users' own preferred language.
    Location,             // Indicates the URL to redirect a page to.
    Status,               // "200"
    WWWAuthenticate,      // WWW-Authenticate        // Defines the authentication method that should be used to access a resource.
    Authorization,        //Authorization           // Contains the credentials to authenticate a user-agent with a server.
    ContentLength,        // "0" Content-Length          // The size of the resource, in decimal number of bytes.
    ContentLocation,       // "/" Content-Location
    GatewayInterface,     // Gateway-Interface "CGI/1.1"
    Connection,           //   Controls whether the network connection stays open after the current transaction finishes.
    KeepAlive,            // Keep-Alive              // Controls how long a persistent connection should stay open.
    Cookie,               // Cookie                  // Contains stored HTTP cookies previously sent by the server with the Set-Cookie header.
    SetCookie,            // Set-Cookie              // Send cookies from the server to the user-agent.
    AccessControlAllowOrigin,  // Access-Control-Allow-Origin   // Indicates whether the response can be shared.
    Server,               // "Mythos"
    AuthType,
    Accept,               //
    RequestMethod,        // "GET"
    HttpHost,             // "example.com"
    PathInfo,             // "/foo/bar"
    QueryString,          // "var1=value1&var2=with%20percent%20encoding"
    RemoteAddr,
    ServerPort,           // "80"
    AcceptPushPolicy,     // Accept-Push-Policy      // A client can express the desired push policy for a request by sending an Accept-Push-Policy header field in the request.
    AcceptSignature      // Accept-Signature        // A client can send the Accept-Signature header field to indicate intention to take advantage of any available signatures and to indicate what kinds of signatures it supports.
}

struct HeaderItem {
    HeaderOpt HeaderType;
    string Value;
}

abstract contract CGI {
    function get(HttpRequest memory request) virtual public view returns (HttpResponse memory) {}
    function post(HttpRequest memory request) virtual public payable returns (HttpResponse memory) {}
}`
