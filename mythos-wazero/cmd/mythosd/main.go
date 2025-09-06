package main

import (
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	runtime "github.com/loredanacirstea/wasmx-wazero"

	"github.com/loredanacirstea/wasmx/app"
	cmd "github.com/loredanacirstea/wasmx/cmdutils"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd(&runtime.WazeroVmMeta{}, app.DefaultNodeHome, nil)
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
