package testdata

import (
	_ "embed"
	"strings"

	aabi "github.com/ethereum/go-ethereum/accounts/abi"
)

var (
	//go:embed erc20.json
	Erc20bz []byte

	//go:embed cw20.json
	Cw20bz []byte

	//go:embed forwardEvm.json
	ForwardEvmbz []byte
)

var Erc20Str string
var Cw20Str string
var ForwardEvmStr string
var Erc20Abi aabi.ABI
var ForwardEvmAbi aabi.ABI

func init() {
	var err error
	Erc20Str = string(Erc20bz)
	Cw20Str = string(Cw20bz)
	ForwardEvmStr = string(ForwardEvmbz)

	Erc20Abi, err = aabi.JSON(strings.NewReader(Erc20Str))
	if err != nil {
		panic("Erc20Abi ABI decoding failure")
	}

	ForwardEvmAbi, err = aabi.JSON(strings.NewReader(ForwardEvmStr))
	if err != nil {
		panic("ForwardEvmStr ABI decoding failure")
	}
}
