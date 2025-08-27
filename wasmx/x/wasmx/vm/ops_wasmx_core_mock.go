package vm

import (
	"fmt"

	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func MockWithPanic(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 0)
	return returns, fmt.Errorf("wasmx core not allowed")
}
