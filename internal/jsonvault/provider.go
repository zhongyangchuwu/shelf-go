package jsonvault

import (
	"strconv"

	"github.com/zhongyangchuwu/shelf-go/internal/age"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

type Provider struct{}

func (Provider) Open(options vault.Options) (vault.Repository, error) {
	return NewVault(options.Path, options)
}

func (Provider) ReadOrCreateIdentity(path string) (vault.Identity, error) {
	identity, err := age.ReadOrCreateIdentity(path)
	if err != nil {
		return vault.Identity{}, err
	}
	return vault.Identity{Value: identity.String(), Recipient: identity.Recipient()}, nil
}

func (Provider) CheckStatus(report *vault.Report, options vault.Options) {
	checkVaultRecipients(report, options.Recipients)
	checkVaultLoads(report, options)
}

func (Provider) CheckDoctor(report *vault.Report, options vault.Options) {
	CheckFile(report, options.Path)
	checkVaultLoads(report, options)
	CheckTracking(report, options.Path)
}

func checkVaultRecipients(report *vault.Report, recipients []string) {
	if len(recipients) == 0 {
		report.Fail("vault recipients", MissingRecipientsDetail())
		return
	}
	report.OK("vault recipients", strconv.Itoa(len(recipients))+" configured")
}

func checkVaultLoads(report *vault.Report, options vault.Options) {
	format, err := DetectFileFormat(options.Path)
	if err != nil {
		report.Fail("vault format", err.Error())
		return
	}
	switch format {
	case FileFormatMissing:
		report.Warn("vault format", FormatDetail(format, options.Path))
	case FileFormatEmpty:
		report.Warn("vault format", FormatDetail(format, options.Path))
	case FileFormatEncryptedVault:
		report.OK("vault format", "encrypted shelf-vault/v1")
	case FileFormatPlaintextStore:
		report.Fail("vault format", FormatDetail(format, options.Path))
		return
	case FileFormatUnsupportedVault:
		report.Fail("vault format", FormatDetail(format, options.Path))
		return
	default:
		report.Fail("vault format", FormatDetail(format, options.Path))
		return
	}
	vaultHandle, err := NewVault(options.Path, options)
	if err != nil {
		report.Fail("vault loads", LoadErrorDetail(err))
		return
	}
	if _, err := vaultHandle.Load(); err != nil {
		report.Fail("vault loads", LoadErrorDetail(err))
		return
	}
	report.OK("vault loads", options.Path)
}
