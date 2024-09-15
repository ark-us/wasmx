package taylor

import (
	_ "embed"
	"strings"
)

var (
	//go:embed json.tay
	JsonTay string

	//go:embed sdk.tay
	SdkTay string

	//go:embed simple_storage.tay
	simpleStorageTay string

	//go:embed erc20.tay
	eRC20Tay string

	//go:embed json_test.tay
	jsonTestTay string
)

func applyImports(content string) string {
	content = strings.ReplaceAll(content, `(import "sdk")`, SdkTay)
	content = strings.ReplaceAll(content, `(import "json")`, JsonTay)
	return content
}

var (
	SimpleStorageTay = applyImports(simpleStorageTay)
	ERC20Tay         = applyImports(eRC20Tay)
	JsonTestTay      = applyImports(jsonTestTay)
)
