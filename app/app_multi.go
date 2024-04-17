package app

import (
	"context"
	"fmt"
)

type ContextKey string

const MultiChainAppKey ContextKey = "MultiChainApp"

type MultiChainApp struct {
	Apps map[string]*App
}

func (m *MultiChainApp) GetApps() map[string]*App {
	return m.Apps
}

func (m *MultiChainApp) GetApp(chainId string) (*App, error) {
	app, ok := m.Apps[chainId]
	if !ok {
		return nil, fmt.Errorf("app not found for chainId: %s", chainId)
	}
	return app, nil
}

func (m *MultiChainApp) SetApp(chainId string, app *App) {
	m.Apps[chainId] = app
}

func NewMultiChainApp(apps map[string]*App) *MultiChainApp {
	return &MultiChainApp{
		Apps: apps,
	}
}

func WithMultiChainAppEmpty(ctx context.Context) (context.Context, *MultiChainApp) {
	mapp := &MultiChainApp{Apps: map[string]*App{}}
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
