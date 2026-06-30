package app

import (
	"strconv"

	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/jsonvault"
)

type Report = jsonvault.Report
type Level = jsonvault.Level

const (
	LevelOK   = jsonvault.LevelOK
	LevelWarn = jsonvault.LevelWarn
	LevelFail = jsonvault.LevelFail
)

func ResolveStatus(configPathFlag, vaultPathFlag string) (Report, error) {
	runtime, err := ResolveRuntime(configPathFlag, vaultPathFlag)
	if err != nil {
		return nil, err
	}
	return Status(runtime), nil
}

func ResolveDoctor(configPathFlag, vaultPathFlag string) (Runtime, Report, error) {
	runtime, err := ResolveRuntime(configPathFlag, vaultPathFlag)
	if err != nil {
		return Runtime{}, nil, err
	}
	return runtime, Doctor(runtime), nil
}

func Status(runtime config.Runtime) jsonvault.Report {
	var report jsonvault.Report
	report.OK("config", runtime.ConfigPath)
	report.OK("vault path", runtime.VaultPath)
	checkVaultRecipients(&report, runtime.Recipients)
	checkVaultLoads(&report, runtime.VaultPath, runtime.Recipients, runtime.IdentityPaths)
	return report
}

func Doctor(runtime config.Runtime) jsonvault.Report {
	var report jsonvault.Report
	jsonvault.CheckFile(&report, runtime.VaultPath)
	checkVaultLoads(&report, runtime.VaultPath, runtime.Recipients, runtime.IdentityPaths)
	jsonvault.CheckTracking(&report, runtime.VaultPath)
	return report
}

func checkVaultRecipients(report *jsonvault.Report, recipients []string) {
	if len(recipients) == 0 {
		report.Fail("vault recipients", jsonvault.MissingRecipientsDetail())
		return
	}
	report.OK("vault recipients", strconv.Itoa(len(recipients))+" configured")
}

func checkVaultLoads(report *jsonvault.Report, vaultPath string, recipients, identityPaths []string) {
	format, err := jsonvault.DetectFileFormat(vaultPath)
	if err != nil {
		report.Fail("vault format", err.Error())
		return
	}
	switch format {
	case jsonvault.FileFormatMissing:
		report.Warn("vault format", jsonvault.FormatDetail(format, vaultPath))
	case jsonvault.FileFormatEmpty:
		report.Warn("vault format", jsonvault.FormatDetail(format, vaultPath))
	case jsonvault.FileFormatEncryptedVault:
		report.OK("vault format", "encrypted shelf-vault/v1")
	case jsonvault.FileFormatPlaintextStore:
		report.Fail("vault format", jsonvault.FormatDetail(format, vaultPath))
		return
	case jsonvault.FileFormatUnsupportedVault:
		report.Fail("vault format", jsonvault.FormatDetail(format, vaultPath))
		return
	default:
		report.Fail("vault format", jsonvault.FormatDetail(format, vaultPath))
		return
	}
	vaultHandle, err := jsonvault.NewVault(vaultPath, jsonvault.VaultOptions{Recipients: recipients, IdentityPaths: identityPaths})
	if err != nil {
		report.Fail("vault loads", jsonvault.LoadErrorDetail(err))
		return
	}
	if _, err := vaultHandle.Load(); err != nil {
		report.Fail("vault loads", jsonvault.LoadErrorDetail(err))
		return
	}
	report.OK("vault loads", vaultPath)
}
