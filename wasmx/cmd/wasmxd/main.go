package main

import (
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"wasmx/v1/app"
	memc "wasmx/v1/x/wasmx/vm/memory/common"

	cmd "wasmx/v1/cmdutils"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd(memc.WasmRuntimeMockVmMeta{})
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
