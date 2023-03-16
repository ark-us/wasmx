package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	aabi "github.com/ethereum/go-ethereum/accounts/abi"
)

var ModuleAddress = sdk.AccAddress([]byte(ModuleName))

var HttpRequestGetAbiStr = `[{"inputs":[{"components":[{"components":[{"internalType":"enum HeaderOption","name":"HeaderType","type":"uint8"},{"internalType":"string","name":"Value","type":"string"}],"internalType":"struct HeaderItem[]","name":"Header","type":"tuple[]"},{"components":[{"internalType":"string","name":"Key","type":"string"},{"internalType":"string","name":"Value","type":"string"}],"internalType":"struct RequestQueryParam[]","name":"QueryParams","type":"tuple[]"}],"internalType":"struct HttpRequest","name":"request","type":"tuple"}],"name":"get","outputs":[{"components":[{"components":[{"internalType":"enum HeaderOption","name":"HeaderType","type":"uint8"},{"internalType":"string","name":"Value","type":"string"}],"internalType":"struct HeaderItem[]","name":"Header","type":"tuple[]"},{"internalType":"string","name":"Content","type":"string"}],"internalType":"struct HttpResponse","name":"","type":"tuple"}],"stateMutability":"view","type":"function"},{"inputs":[{"components":[{"internalType":"enum HeaderOption","name":"HeaderType","type":"uint8"},{"internalType":"string","name":"Value","type":"string"}],"internalType":"struct HeaderItem[]","name":"header","type":"tuple[]"},{"internalType":"enum HeaderOption","name":"headerType","type":"uint8"}],"name":"getHeaderValue","outputs":[{"internalType":"string","name":"value","type":"string"}],"stateMutability":"pure","type":"function"},{"inputs":[{"components":[{"components":[{"internalType":"enum HeaderOption","name":"HeaderType","type":"uint8"},{"internalType":"string","name":"Value","type":"string"}],"internalType":"struct HeaderItem[]","name":"Header","type":"tuple[]"},{"components":[{"internalType":"string","name":"Key","type":"string"},{"internalType":"string","name":"Value","type":"string"}],"internalType":"struct RequestQueryParam[]","name":"QueryParams","type":"tuple[]"}],"internalType":"struct HttpRequest","name":"request","type":"tuple"}],"name":"post","outputs":[{"components":[{"components":[{"internalType":"enum HeaderOption","name":"HeaderType","type":"uint8"},{"internalType":"string","name":"Value","type":"string"}],"internalType":"struct HeaderItem[]","name":"Header","type":"tuple[]"},{"internalType":"string","name":"Content","type":"string"}],"internalType":"struct HttpResponse","name":"","type":"tuple"}],"stateMutability":"payable","type":"function"}]`

var HttpRequestGetAbi aabi.ABI

func init() {
	abi, err := aabi.JSON(strings.NewReader(HttpRequestGetAbiStr))
	if err != nil {
		panic(err)
	}
	HttpRequestGetAbi = abi
}

// type HeaderOption uint8

const (
	Undefined                   uint8 = iota // 0
	Proto                                    // 1 // "HTTP/1.1"
	ProtoMajor                               // 2 // 1
	ProtoMinor                               // 3 // 1
	Request_Method                           // 4 // request_method Request-Method; "GET"
	Http_Host                                // 5 // http_host Http-Host; "example.com"
	Path_Info                                // 6 // path_info Path-Info; "/foo/bar"
	Query_String                             // 7 // query_string Query-String; "var1=value1&var2=with%20percent%20encoding"
	Location                                 // 8 // location Location indicates the URL to redirect a page to
	Content_Type                             // 9 // content_type Content-Type indicates the media type of the resource;  "text/html; charset=UTF-8"
	Content_Encoding                         // 10 // content_encoding Content-Encoding is used to specify the compression algorithm
	Content_Language                         // 11 // content_language Content-Language describes the human language(s) intended for the audience; "en"
	Content_Length                           // 12 // content_length Content-Length indicates the size of the resource, in decimal number of bytes; "0"
	Content_Location                         // 13 // content_location Content-Location "/"
	Status                                   // 14 // status Status indicates the status code response; "200"
	StatusCode                               // 15 // status Status-Code indicates the status code response; "200"
	WWW_Authenticate                         // 16 // www_authenticate WWW-Authenticate defines the authentication method that should be used to access a resource
	Authorization                            // 17 // authorization Authorization contains the credentials to authenticate a user-agent with a server
	Auth_Type                                // 18 // auth_type Auth-Type
	Accept                                   // 19 // accept Accept
	Connection                               // 20 // connection Connection controls whether the network connection stays open after the current transaction finishes
	Keep_Alive                               // 21 // keep_alive Keep-Alive controls how long a persistent connection should stay open
	Cookie                                   // 22 // cookie Cookie contains stored HTTP cookies previously sent by the server with the Set-Cookie header
	Set_Cookie                               // 23 // set_cookie Set-Cookie send cookies from the server to the user-agent
	Access_Control_Allow_Origin              // 24 // access_control_allow_origin Access-Control-Allow-Origin indicates whether the response can be shared
	Server                                   // 25 // server Server
	Remote_Addr                              // 26 // remote_addr Remote-Addr
	Server_Port                              // 27 // server_port ServerPort; "80"
	Accept_Push_Policy                       // 28 // status Accept-Push-Policy
	Accept_Signature                         // 29 // accept_signature Accept-Signature indicates the intention to take advantage of
	// any available signatures and to indicate what kinds of signatures it supports
)

type RequestQueryParam struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

type HeaderItem struct {
	HeaderType uint8  `json:"HeaderType"`
	Value      string `json:"Value"`
}

type HttpRequest struct {
	Header      []HeaderItem        `json:"Header"`
	QueryParams []RequestQueryParam `json:"QueryParams"`
}

type HttpResponse struct {
	Header  []HeaderItem `json:"Header"`
	Content string       `json:"Content"`
}

type HttpResponseAbi struct {
	Message HttpResponse `json:"message"`
}

func RequestGetEncodeAbi(request HttpRequest) ([]byte, error) {
	return HttpRequestGetAbi.Pack(
		"get",
		request,
	)
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

enum HeaderOption {
    Proto,                // "HTTP/1.1"
    ProtoMajor,           // 1
    ProtoMinor,           // 1
    RequestMethod,        // "GET"
    HttpHost,             // "example.com"
    PathInfo,             // "/foo/bar"
    QueryString,          // "var1=value1&var2=with%20percent%20encoding"
    Location,             // Indicates the URL to redirect a page to.
    ContentType,          // Content-Type            // "text/html; charset=UTF-8" Indicates the media type of the resource.
    ContentEncoding,      // Content-Encoding        // Used to specify the compression algorithm.
    ContentLanguage,      // Content-Language        // "en" Describes the human language(s) intended for the audience, so that it allows a user to differentiate according to the users' own preferred language.
    ContentLength,        // "0" Content-Length          // The size of the resource, in decimal number of bytes.
    ContentLocation,       // "/" Content-Location
    Status,               // "200 OK"
    StatusCode,           //  200
    WWWAuthenticate,      // WWW-Authenticate        // Defines the authentication method that should be used to access a resource.
    Authorization,        //Authorization           // Contains the credentials to authenticate a user-agent with a server.
    AuthType,
    Accept,               //
    Connection,           //   Controls whether the network connection stays open after the current transaction finishes; "close"
    KeepAlive,            // Keep-Alive              // Controls how long a persistent connection should stay open.
    Cookie,               // Cookie                  // Contains stored HTTP cookies previously sent by the server with the Set-Cookie header.
    SetCookie,            // Set-Cookie              // Send cookies from the server to the user-agent.
    AccessControlAllowOrigin,  // Access-Control-Allow-Origin   // Indicates whether the response can be shared.
    Server,               // "Mythos"
    RemoteAddr,
    ServerPort,           // "80"
    AcceptPushPolicy,     // Accept-Push-Policy      // A client can express the desired push policy for a request by sending an Accept-Push-Policy header field in the request.
    AcceptSignature      // Accept-Signature        // A client can send the Accept-Signature header field to indicate intention to take advantage of any available signatures and to indicate what kinds of signatures it supports.
}

struct HeaderItem {
    HeaderOption HeaderType;
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

abstract contract CGI {
    function get(HttpRequest memory request) virtual public view returns (HttpResponse memory) {}
    function post(HttpRequest memory request) virtual public payable returns (HttpResponse memory) {}
}`
