package utils

import (
	"fmt"
	"os"
	"path/filepath"

	mcfg "github.com/loredanacirstea/wasmx/config"
	menc "github.com/loredanacirstea/wasmx/encoding"
	"github.com/loredanacirstea/wasmx/multichain"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func CreateMockAppCreator(wasmVmMeta memc.IWasmVmMeta, appCreatorFactory multichain.NewAppCreator, index int) (*mcfg.MultiChainApp, func(chainId string, chainCfg *menc.ChainConfig) mcfg.MythosApp) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	tempNodeHome := filepath.Join(userHomeDir, fmt.Sprintf(".mythostmp_%d", index))
	return multichain.CreateNoLoggerAppCreator(wasmVmMeta, appCreatorFactory, tempNodeHome)
}
