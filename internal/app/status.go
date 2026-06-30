package app

import (
	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

type Report = vault.Report
type Level = vault.Level

const (
	LevelOK   = vault.LevelOK
	LevelWarn = vault.LevelWarn
	LevelFail = vault.LevelFail
)

func (a *App) ResolveStatus(configPathFlag, vaultPathFlag string) (Report, error) {
	runtime, err := ResolveRuntime(configPathFlag, vaultPathFlag)
	if err != nil {
		return nil, err
	}
	return a.Status(runtime), nil
}

func (a *App) ResolveDoctor(configPathFlag, vaultPathFlag string) (Runtime, Report, error) {
	runtime, err := ResolveRuntime(configPathFlag, vaultPathFlag)
	if err != nil {
		return Runtime{}, nil, err
	}
	return runtime, a.Doctor(runtime), nil
}

func (a *App) Status(runtime config.Runtime) vault.Report {
	var report vault.Report
	report.OK("config", runtime.ConfigPath)
	report.OK("vault path", runtime.VaultPath)
	a.vaults.CheckStatus(&report, a.vaultOptions(runtime))
	return report
}

func (a *App) Doctor(runtime config.Runtime) vault.Report {
	var report vault.Report
	a.vaults.CheckDoctor(&report, a.vaultOptions(runtime))
	return report
}
