package server

import (
	"net/http"

	"github.com/tendermint/tendermint/libs/log"

	websrvmodulekeeper "wasmx/x/websrv/keeper"
)

// TODO context from cli
func StartWebsrv(
	websrv websrvmodulekeeper.Keeper,
	logger log.Logger,
	addr string,
) {
	httpSrvDone := make(chan struct{}, 1)
	go func() {
		if err := websrv.Init(addr); err != nil {
			if err == http.ErrServerClosed {
				close(httpSrvDone)
				return
			}
			logger.Error("Error serving websrv", "err", err)
		}
	}()
}
