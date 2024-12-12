package main

import (
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/loredanacirstea/wasmx/app"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"

	cmd "github.com/loredanacirstea/wasmx/cmdutils"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd(memc.WasmRuntimeMockVmMeta{})
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
