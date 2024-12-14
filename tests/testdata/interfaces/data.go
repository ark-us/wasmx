package testdata

import (
	_ "embed"
	"strings"

	aabi "github.com/ethereum/go-ethereum/accounts/abi"
)

var (
	//go:embed erc20.json
	erc20bz []byte

	//go:embed cw20.json
	cw20bz []byte

	//go:embed forwardEvm.json
	forwardEvmbz []byte
)

var Erc20Str string
var Cw20Str string
var ForwardEvmStr string
var Erc20Abi aabi.ABI
var ForwardEvmAbi aabi.ABI

func init() {
	var err error
	Erc20Str = string(erc20bz)
	Cw20Str = string(cw20bz)
	ForwardEvmStr = string(forwardEvmbz)

	Erc20Abi, err = aabi.JSON(strings.NewReader(Erc20Str))
	if err != nil {
		panic("Erc20Abi ABI decoding failure")
	}

	ForwardEvmAbi, err = aabi.JSON(strings.NewReader(ForwardEvmStr))
	if err != nil {
		panic("ForwardEvmStr ABI decoding failure")
	}
}
