package app

import (
	"strconv"

	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultfile"
)

type Report = vaultfile.Report
type Level = vaultfile.Level

const (
	LevelOK   = vaultfile.LevelOK
	LevelWarn = vaultfile.LevelWarn
	LevelFail = vaultfile.LevelFail
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

func Status(runtime config.Runtime) vaultfile.Report {
	var report vaultfile.Report
	report.OK("config", runtime.ConfigPath)
	report.OK("vault path", runtime.VaultPath)
	checkVaultRecipients(&report, runtime.Recipients)
	checkVaultLoads(&report, runtime.VaultPath, runtime.Recipients, runtime.IdentityPaths)
	return report
}

func Doctor(runtime config.Runtime) vaultfile.Report {
	var report vaultfile.Report
	vaultfile.CheckFile(&report, runtime.VaultPath)
	checkVaultLoads(&report, runtime.VaultPath, runtime.Recipients, runtime.IdentityPaths)
	vaultfile.CheckTracking(&report, runtime.VaultPath)
	return report
}

func checkVaultRecipients(report *vaultfile.Report, recipients []string) {
	if len(recipients) == 0 {
		report.Fail("vault recipients", vaultfile.MissingRecipientsDetail())
		return
	}
	report.OK("vault recipients", strconv.Itoa(len(recipients))+" configured")
}

func checkVaultLoads(report *vaultfile.Report, vaultPath string, recipients, identityPaths []string) {
	format, err := vaultfile.DetectFileFormat(vaultPath)
	if err != nil {
		report.Fail("vault format", err.Error())
		return
	}
	switch format {
	case vaultfile.FileFormatMissing:
		report.Warn("vault format", vaultfile.FormatDetail(format, vaultPath))
	case vaultfile.FileFormatEmpty:
		report.Warn("vault format", vaultfile.FormatDetail(format, vaultPath))
	case vaultfile.FileFormatEncryptedVault:
		report.OK("vault format", "encrypted shelf-vault/v1")
	case vaultfile.FileFormatPlaintextStore:
		report.Fail("vault format", vaultfile.FormatDetail(format, vaultPath))
		return
	case vaultfile.FileFormatUnsupportedVault:
		report.Fail("vault format", vaultfile.FormatDetail(format, vaultPath))
		return
	default:
		report.Fail("vault format", vaultfile.FormatDetail(format, vaultPath))
		return
	}
	vaultHandle, err := vaultfile.NewVault(vaultPath, vaultfile.VaultOptions{Recipients: recipients, IdentityPaths: identityPaths})
	if err != nil {
		report.Fail("vault loads", vaultfile.LoadErrorDetail(err))
		return
	}
	if _, err := vaultHandle.Load(); err != nil {
		report.Fail("vault loads", vaultfile.LoadErrorDetail(err))
		return
	}
	report.OK("vault loads", vaultPath)
}
