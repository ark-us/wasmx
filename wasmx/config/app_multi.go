package config

import (
	"context"
	"fmt"
	"sync"

	menc "github.com/loredanacirstea/wasmx/encoding"
)

type ContextKey string

const MultiChainAppKey ContextKey = "MultiChainApp"

type MultiChainApp struct {
	apps    map[string]interface{}
	mtxApps sync.Mutex

	ChainIds []string
	NewApp   func(chainId string, chainCfg *menc.ChainConfig) MythosApp
	APICtx   APICtxI
}

func (m *MultiChainApp) SetAPICtx(apictx APICtxI) {
	m.APICtx = apictx
}

func (m *MultiChainApp) SetAppCreator(appCreator func(chainId string, chainCfg *menc.ChainConfig) MythosApp) {
	m.NewApp = appCreator
}

// use GetAppsUnsafe only in testing, otherwise set lock
func (m *MultiChainApp) GetAppsUnsafe() map[string]interface{} {
	return m.apps
}

func (m *MultiChainApp) GetApp(chainId string) (interface{}, error) {
	m.mtxApps.Lock()
	defer m.mtxApps.Unlock()
	app, ok := m.apps[chainId]
	if !ok {
		return nil, fmt.Errorf("app not found for chainId: %s", chainId)
	}
	return app, nil
}

func (m *MultiChainApp) SetApp(chainId string, app interface{}) {
	m.mtxApps.Lock()
	defer m.mtxApps.Unlock()
	m.apps[chainId] = app
	m.ChainIds = append(m.ChainIds, chainId)
}

func NewMultiChainApp(apps map[string]interface{}, chainIds []string) *MultiChainApp {
	return &MultiChainApp{
		apps:     apps,
		ChainIds: chainIds,
	}
}

func WithMultiChainAppEmpty(ctx context.Context) (context.Context, *MultiChainApp) {
	mapp := &MultiChainApp{apps: map[string]interface{}{}, ChainIds: []string{}}
	return context.WithValue(ctx, MultiChainAppKey, mapp), mapp
}

func WithMultiChainApp(ctx context.Context, mapp *MultiChainApp) context.Context {
	return context.WithValue(ctx, MultiChainAppKey, mapp)
}

func GetMultiChainApp(ctx context.Context) (*MultiChainApp, error) {
	mappi := ctx.Value(MultiChainAppKey)
	mapp := (mappi).(*MultiChainApp)
	if mapp == nil {
		return nil, fmt.Errorf("multichainapp not set on context")
	}
	return mapp, nil
}
