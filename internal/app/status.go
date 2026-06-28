package app

import (
	"strconv"

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

func Status(runtime config.Runtime) vault.Report {
	var report vault.Report
	report.OK("config", runtime.ConfigPath)
	report.OK("vault path", runtime.VaultPath)
	checkVaultRecipients(&report, runtime.Recipients)
	checkVaultLoads(&report, runtime.VaultPath, runtime.Recipients, runtime.IdentityPaths)
	return report
}

func Doctor(runtime config.Runtime) vault.Report {
	var report vault.Report
	vault.CheckFile(&report, runtime.VaultPath)
	checkVaultLoads(&report, runtime.VaultPath, runtime.Recipients, runtime.IdentityPaths)
	vault.CheckTracking(&report, runtime.VaultPath)
	return report
}

func checkVaultRecipients(report *vault.Report, recipients []string) {
	if len(recipients) == 0 {
		report.Fail("vault recipients", vault.MissingRecipientsDetail())
		return
	}
	report.OK("vault recipients", strconv.Itoa(len(recipients))+" configured")
}

func checkVaultLoads(report *vault.Report, vaultPath string, recipients, identityPaths []string) {
	format, err := vault.DetectFileFormat(vaultPath)
	if err != nil {
		report.Fail("vault format", err.Error())
		return
	}
	switch format {
	case vault.FileFormatMissing:
		report.Warn("vault format", vault.FormatDetail(format, vaultPath))
	case vault.FileFormatEmpty:
		report.Warn("vault format", vault.FormatDetail(format, vaultPath))
	case vault.FileFormatEncryptedVault:
		report.OK("vault format", "encrypted shelf-vault/v1")
	case vault.FileFormatPlaintextStore:
		report.Fail("vault format", vault.FormatDetail(format, vaultPath))
		return
	case vault.FileFormatUnsupportedVault:
		report.Fail("vault format", vault.FormatDetail(format, vaultPath))
		return
	default:
		report.Fail("vault format", vault.FormatDetail(format, vaultPath))
		return
	}
	vaultHandle, err := vault.NewVault(vaultPath, vault.VaultOptions{Recipients: recipients, IdentityPaths: identityPaths})
	if err != nil {
		report.Fail("vault loads", vault.LoadErrorDetail(err))
		return
	}
	if _, err := vaultHandle.Load(); err != nil {
		report.Fail("vault loads", vault.LoadErrorDetail(err))
		return
	}
	report.OK("vault loads", vaultPath)
}
