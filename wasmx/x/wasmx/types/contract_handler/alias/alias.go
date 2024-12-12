package alias

import (
	"strings"

	aabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/loredanacirstea/wasmx/v1/x/wasmx/types"
	cch "github.com/loredanacirstea/wasmx/v1/x/wasmx/types/contract_handler"
)

type RegisterRequest struct {
	EthAddress common.Address `json:"ethAddress"`
	CoinType   uint32         `json:"coinType"`
}

type GetCosmosAddressRequest struct {
	EthAddress common.Address `json:"ethAddress"`
	CoinType   uint32         `json:"coinType"`
}

type GetCosmosAddressResponse = struct {
	CosmAddress common.Address `json:"cosmAddress"`
	Found       bool           `json:"found"`
}

var AliasAbi aabi.ABI
var err error

func init() {
	AliasAbi, err = aabi.JSON(strings.NewReader(AliasAbiJson))
	if err != nil {
		panic("Alias ABI decoding failure")
	}
}

type AliasHandler struct{}

func NewAliasHandler() AliasHandler {
	return AliasHandler{}
}

func (a AliasHandler) Encode(req cch.ContractHandlerMessage) (*types.WasmxExecutionMessage, error) {
	bz, err := AliasAbi.Pack(req.Method, req.Msg)
	if err != nil {
		return nil, err
	}
	return &types.WasmxExecutionMessage{Data: bz}, nil
}

func (a AliasHandler) Decode(method string, data []byte) (any, error) {
	unpacked, err := AliasAbi.Unpack(method, data)
	if err != nil {
		return nil, err
	}
	return unpacked, nil
}

// func ParseCosmosAddressResponse(data any) GetCosmosAddressResponse {
// 	qres := *(*GetCosmosAddressResponse)(unsafe.Pointer(&data.([]interface{})[0]))
// 	return qres
// }
