package app

import (
	"strconv"

	"github.com/zhongyangchuwu/shelf-go/internal/adapters/shelfvault"
	"github.com/zhongyangchuwu/shelf-go/internal/config"
)

type Report = shelfvault.Report
type Level = shelfvault.Level

const (
	LevelOK   = shelfvault.LevelOK
	LevelWarn = shelfvault.LevelWarn
	LevelFail = shelfvault.LevelFail
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

func Status(runtime config.Runtime) shelfvault.Report {
	var report shelfvault.Report
	report.OK("config", runtime.ConfigPath)
	report.OK("vault path", runtime.VaultPath)
	checkVaultRecipients(&report, runtime.Recipients)
	checkVaultLoads(&report, runtime.VaultPath, runtime.Recipients, runtime.IdentityPaths)
	return report
}

func Doctor(runtime config.Runtime) shelfvault.Report {
	var report shelfvault.Report
	shelfvault.CheckFile(&report, runtime.VaultPath)
	checkVaultLoads(&report, runtime.VaultPath, runtime.Recipients, runtime.IdentityPaths)
	shelfvault.CheckTracking(&report, runtime.VaultPath)
	return report
}

func checkVaultRecipients(report *shelfvault.Report, recipients []string) {
	if len(recipients) == 0 {
		report.Fail("vault recipients", shelfvault.MissingRecipientsDetail())
		return
	}
	report.OK("vault recipients", strconv.Itoa(len(recipients))+" configured")
}

func checkVaultLoads(report *shelfvault.Report, vaultPath string, recipients, identityPaths []string) {
	format, err := shelfvault.DetectFileFormat(vaultPath)
	if err != nil {
		report.Fail("vault format", err.Error())
		return
	}
	switch format {
	case shelfvault.FileFormatMissing:
		report.Warn("vault format", shelfvault.FormatDetail(format, vaultPath))
	case shelfvault.FileFormatEmpty:
		report.Warn("vault format", shelfvault.FormatDetail(format, vaultPath))
	case shelfvault.FileFormatEncryptedVault:
		report.OK("vault format", "encrypted shelf-vault/v1")
	case shelfvault.FileFormatPlaintextStore:
		report.Fail("vault format", shelfvault.FormatDetail(format, vaultPath))
		return
	case shelfvault.FileFormatUnsupportedVault:
		report.Fail("vault format", shelfvault.FormatDetail(format, vaultPath))
		return
	default:
		report.Fail("vault format", shelfvault.FormatDetail(format, vaultPath))
		return
	}
	vaultHandle, err := shelfvault.NewVault(vaultPath, shelfvault.VaultOptions{Recipients: recipients, IdentityPaths: identityPaths})
	if err != nil {
		report.Fail("vault loads", shelfvault.LoadErrorDetail(err))
		return
	}
	if _, err := vaultHandle.Load(); err != nil {
		report.Fail("vault loads", shelfvault.LoadErrorDetail(err))
		return
	}
	report.OK("vault loads", vaultPath)
}
