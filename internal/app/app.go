package app

import (
	"github.com/zhongyangchuwu/shelf-go/internal/jsonvault"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

type App struct {
	vaults vault.Provider
}

func New(vaults vault.Provider) *App {
	return &App{vaults: vaults}
}
func NewDefault() *App {
	return New(jsonvault.Provider{})
}

func (a *App) vaultOptions(runtime Runtime) vault.Options {
	return vault.Options{Path: runtime.VaultPath, Recipients: runtime.Recipients, IdentityPaths: runtime.IdentityPaths}
}
