package vm

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"strings"

	sdkerr "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	aabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	vmtypes "mythos/v1/x/wasmx/vm/types"
)

// type UnpackedArgs = map[string]interface{}

var BigIntElem = &big.Int{}
var BigIntType = reflect.TypeOf(BigIntElem)
var AddressElem = common.Address{}
var AddressType = reflect.TypeOf(AddressElem)

type UnpackedArgs struct {
	Args map[string]interface{}
	// GetAlias func(addr sdk.AccAddress) sdk.AccAddress
}

func (b UnpackedArgs) MarshalJSON() ([]byte, error) {
	keys := make([]string, 0, len(b.Args))
	for key := range b.Args {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	pairs := []string{}
	for _, key := range keys {
		val := fmt.Sprintf(`"%s":`, key)
		value := b.Args[key]
		v, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		vtype := reflect.TypeOf(value)
		switch vtype {
		case BigIntType:
			val += `"`
			val += string(v)
			val += `"`
		case AddressType:
			addr := value.(common.Address)
			// TODO
			// account := b.GetAlias(sdk.AccAddress(addr.Bytes()))
			account := sdk.AccAddress(addr.Bytes())
			val += `"` + account.String() + `"`
		default:
			val += string(v)
		}
		pairs = append(pairs, val)
	}
	marshalledV := fmt.Sprintf(`{%s}`, strings.Join(pairs, ","))
	return []byte(marshalledV), nil
}

// func (b big.Int) MarshalJSON() ([]byte, error) {
// }

type BigInt struct {
	*big.Int
}

func (b *BigInt) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, b.String())), nil
}

func (b *BigInt) UnmarshalJSON(data []byte) error {
	if b.Int == nil {
		b.Int = new(big.Int)
	}

	// We receive the data enclosed with quotes, so we need to unquote it first
	unquotedData, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	_, ok := b.SetString(unquotedData, 10)
	if !ok {
		return fmt.Errorf("bigInt UnmarshalJSON: invalid number %s", unquotedData)
	}
	return nil
}

type ProxyInterfacesEvmToJson struct {
	TargetContract common.Address `json:"targetContract"`
	MethodName     string         `json:"methodName"`
	Input          []byte         `json:"input"`
}

func ProxyInterfaces(context *Context, input []byte) ([]byte, error) {
	sig := input[0:4]
	calld := input[4:]

	method, err := ProxyInterfacesAbi.MethodById(sig)
	if err != nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "method not found")
	}
	switch method.Name {
	case "EvmToJson":
		return EvmToJson(method, context, calld)
	case "JsonToEvm":
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "not implemented")
	case "EvmToJsonCall":
		return EvmToJsonCall(method, context, calld)
	case "JsonToEvmCall":
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "not implemented")
	}
	return nil, sdkerr.Wrapf(sdkerr.Error{}, "invalid method")
}

func EvmToJsonCall(method *aabi.Method, context *Context, calld []byte) ([]byte, error) {
	var data ProxyInterfacesEvmToJson
	unpacked, err := method.Inputs.Unpack(calld)
	if err != nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "cannot unpack")
	}
	err = method.Inputs.Copy(&data, unpacked)
	if err != nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "cannot unpack")
	}

	// get contract abi
	contractAddress := sdk.AccAddress(data.TargetContract.Bytes())
	abi, err := getContractAbi(context, contractAddress)
	if err != nil {
		return nil, err
	}
	fabi, found := abi.Methods[data.MethodName]
	if !found {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "method not found")
	}

	input, err := evmToJsonInner(fabi, contractAddress, data.MethodName, data.Input)
	if err != nil {
		return nil, err
	}

	// do call
	handler := context.GetCosmosHandler()
	if handler == nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "invalid cosmos handler")
	}
	req := vmtypes.CallRequest{
		To:       contractAddress,
		From:     context.Env.CurrentCall.Sender,
		Value:    context.Env.CurrentCall.Funds,
		Calldata: input,
		GasLimit: context.Env.CurrentCall.GasLimit,
		IsQuery:  false,
	}

	var success int32
	var returnData []byte
	// Send funds
	if req.Value.BitLen() > 0 {
		err = handler.SendCoin(req.To, req.Value)
	}
	if err != nil {
		success = int32(2)
	} else {
		contractContext := GetContractContext(context, contractAddress)
		if contractContext == nil {
			// ! we return success here in case the contract does not exist
			success = int32(0)
		} else {
			req.Bytecode = contractContext.ContractInfo.Bytecode
			req.CodeHash = contractContext.ContractInfo.CodeHash
			success, returnData = WasmxCall(context, req)
		}
	}

	response, err := method.Outputs.Pack(success == 0, returnData)
	if err != nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "EvmToJsonCall return data pack failed")
	}
	return response, nil
}

func EvmToJson(method *aabi.Method, context *Context, calld []byte) ([]byte, error) {
	var data ProxyInterfacesEvmToJson
	unpacked, err := method.Inputs.Unpack(calld)
	if err != nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "cannot unpack")
	}
	err = method.Inputs.Copy(&data, unpacked)
	if err != nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "cannot unpack")
	}

	// get contract abi
	contractAddress := sdk.AccAddress(data.TargetContract.Bytes())
	abi, err := getContractAbi(context, contractAddress)
	if err != nil {
		return nil, err
	}
	fabi, found := abi.Methods[data.MethodName]
	if !found {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "method not found")
	}

	return evmToJsonInner(fabi, contractAddress, data.MethodName, data.Input)
}

func evmToJsonInner(fabi aabi.Method, contractAddress sdk.AccAddress, methodName string, input []byte) ([]byte, error) {
	v := map[string]interface{}{}
	err := fabi.Inputs.UnpackIntoMap(v, input)
	if err != nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "cannot unpack input")
	}

	vWrap := UnpackedArgs{Args: v}
	vstr, err := json.Marshal(vWrap)
	if err != nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "cannot marshal")
	}
	jsonCalldata := []byte(fmt.Sprintf(`{"%s":%s}`, methodName, string(vstr)))
	return jsonCalldata, nil
}

func getContractAbi(context *Context, contractAddress sdk.AccAddress) (*aabi.ABI, error) {
	handler := context.GetCosmosHandler()
	if handler == nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "invalid cosmos handler")
	}
	codeInfo := handler.GetCodeInfo(contractAddress)
	// TODO check codeInfo.GetMetadata()
	abiStr := codeInfo.GetMetadata().Abi
	if abiStr == "" {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "empty abi")
	}

	abi, err := aabi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "invalid abi")
	}
	return &abi, nil
}

var ProxyInterfacesAbi aabi.ABI

func init() {
	var err error
	ProxyInterfacesAbi, err = aabi.JSON(strings.NewReader(ProxyInterfacesAbiJsonStr))
	if err != nil {
		panic("ProxyInterfacesAbi decoding failure")
	}
}

var ProxyInterfacesAbiJsonStr = `[{"inputs":[{"internalType":"address","name":"targetContract","type":"address"},{"internalType":"string","name":"jsonCalldata","type":"string"}],"name":"JsonToEvmCall","outputs":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"data","type":"bytes"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"targetContract","type":"address"},{"internalType":"string","name":"methodName","type":"string"},{"internalType":"bytes","name":"input","type":"bytes"}],"name":"EvmToJsonCall","outputs":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"data","type":"bytes"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"targetContract","type":"address"},{"internalType":"string","name":"methodName","type":"string"},{"internalType":"bytes","name":"input","type":"bytes"}],"name":"EvmToJson","outputs":[{"internalType":"string","name":"jsonCalldata","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"targetContract","type":"address"},{"internalType":"string","name":"jsonCalldata","type":"string"}],"name":"JsonToEvm","outputs":[{"internalType":"bytes","name":"evmCalldata","type":"bytes"}],"stateMutability":"view","type":"function"}]`

var ProxyInterfacesInterface = `
interface ProxyInterfaces {
    function EvmToJson(
        address targetContract,
        string memory methodName,
        bytes memory input
    ) external view returns (string memory jsonCalldata);
    function EvmToJsonCall(
        address targetContract,
        string memory methodName,
        bytes memory input
    ) external payable returns (bool success, bytes memory data);
    function JsonToEvm(
        address targetContract,
        string memory jsonCalldata
    ) external view returns (bytes memory evmCalldata);
    function JsonToEvmCall(
        address targetContract,
        string memory jsonCalldata
    ) external payable returns (bool success, bytes memory data);
}
`
