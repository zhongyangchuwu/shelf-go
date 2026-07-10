package app

import (
	"errors"
	"fmt"
	"os"

	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultjson"
)

type Level string

const (
	LevelOK   Level = "ok"
	LevelWarn Level = "warn"
	LevelFail Level = "fail"
)

type Check struct {
	Level  Level
	Name   string
	Detail string
}

type Report []Check

func (r *Report) OK(name, detail string) {
	*r = append(*r, Check{Level: LevelOK, Name: name, Detail: detail})
}

func (r *Report) Warn(name, detail string) {
	*r = append(*r, Check{Level: LevelWarn, Name: name, Detail: detail})
}

func (r *Report) Fail(name, detail string) {
	*r = append(*r, Check{Level: LevelFail, Name: name, Detail: detail})
}

func (a *App) ResolveDoctor(configPathFlag, vaultPathFlag string) (Runtime, Report, error) {
	runtime, err := ResolveRuntime(configPathFlag, vaultPathFlag)
	if err != nil {
		return Runtime{}, nil, err
	}
	return runtime, a.Doctor(runtime), nil
}

func (a *App) Doctor(runtime config.Runtime) Report {
	var report Report
	report.OK("vault path", runtime.VaultPath)
	info, err := os.Stat(runtime.VaultPath)
	if errors.Is(err, os.ErrNotExist) {
		report.Fail("vault exists", "run shelf setup or create a secret")
		return report
	}
	if err != nil {
		report.Fail("vault stat", err.Error())
		return report
	}
	if info.IsDir() {
		report.Fail("vault file", "path is a directory")
		return report
	}
	mode := info.Mode().Perm()
	if mode&0o077 != 0 {
		report.Warn("vault permissions", fmt.Sprintf("%03o", mode))
	} else {
		report.OK("vault permissions", fmt.Sprintf("%03o", mode))
	}
	if _, err := vaultjson.NewRepository(runtime.VaultPath).Load(); err != nil {
		report.Fail("vault load", err.Error())
	} else {
		report.OK("vault load", "valid item vault")
	}
	return report
}
