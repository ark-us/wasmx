package taylor

import (
	_ "embed"
	"strings"
)

var (
	//go:embed sdk.tay
	SdkTay string

	//go:embed simple_storage.tay
	simpleStorageTay string

	//go:embed erc20.tay
	eRC20Tay string
)

var (
	SimpleStorageTay = strings.ReplaceAll(simpleStorageTay, `(import "sdk")`, SdkTay)
	ERC20Tay         = strings.ReplaceAll(eRC20Tay, `(import "sdk")`, SdkTay)
)
