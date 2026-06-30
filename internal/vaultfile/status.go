package vaultfile

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

func HasFailures(report Report) bool {
	for _, check := range report {
		if check.Level == LevelFail {
			return true
		}
	}
	return false
}

func CheckFile(report *Report, vaultPath string) {
	info, err := os.Stat(vaultPath)
	if os.IsNotExist(err) {
		report.Warn("vault file exists", vaultPath+" will be created on first write")
		return
	}
	if err != nil {
		report.Fail("vault file exists", err.Error())
		return
	}
	if info.IsDir() {
		report.Fail("vault file is regular file", vaultPath+" is a directory")
		return
	}
	report.OK("vault file exists", vaultPath)
	detail, secure := fileModeDetail(info)
	if secure {
		report.OK("vault file mode", detail)
	} else {
		report.Warn("vault file mode", detail)
	}
}

func fileModeDetail(info os.FileInfo) (string, bool) {
	if runtime.GOOS == "windows" {
		return "not checked on windows", true
	}
	mode := info.Mode().Perm()
	if mode&0o077 == 0 {
		return mode.String(), true
	}
	return mode.String() + " is broader than 0600", false
}

func MissingRecipientsDetail() string {
	return "no age recipients configured; run shelf vault init --force --recipient AGE_RECIPIENT --identity PATH before creating or updating secrets"
}

func FormatDetail(format FileFormat, path string) string {
	switch format {
	case FileFormatMissing:
		return path + " is missing; run shelf vault init or write a secret after configuring recipients"
	case FileFormatEmpty:
		return path + " is empty; run shelf vault init or write a secret after configuring recipients"
	case FileFormatPlaintextStore:
		return "plaintext JSON store; run shelf vault migrate --from " + path + " --to <vault.age>, update config, then move/delete/archive the plaintext source"
	case FileFormatUnsupportedVault:
		return "unsupported shelf vault format; upgrade Shelf if this vault came from a newer version, or restore a compatible encrypted backup"
	default:
		return "unsupported file content; choose a valid vault path or restore a compatible encrypted backup"
	}
}

func LoadErrorDetail(err error) string {
	message := err.Error()
	switch {
	case strings.Contains(message, "no age identity paths"):
		return message + "; add identity_paths in config or run shelf vault init --identity PATH"
	case strings.Contains(message, "read age identity"):
		return message + "; fix identity_paths or identity file permissions"
	case strings.Contains(message, "parse age identity") || strings.Contains(message, "no age identities loaded"):
		return message + "; fix the identity file contents or run shelf vault init --identity PATH"
	case strings.Contains(message, "no configured age identity matched"):
		return message + "; configure the age identity that matches this vault recipient"
	case strings.Contains(message, "could not decrypt vault"):
		return message + "; verify identity_paths match the vault recipient or restore a known-good encrypted backup"
	case strings.Contains(message, "invalid decrypted store"):
		return message + "; restore a known-good encrypted backup"
	default:
		return message
	}
}

func CheckTracking(report *Report, vaultPath string) {
	format, formatErr := DetectFileFormat(vaultPath)
	abs, err := filepath.Abs(vaultPath)
	if err != nil {
		report.Warn("git tracking", err.Error())
		return
	}
	rootBytes, err := exec.Command("git", "-C", filepath.Dir(abs), "rev-parse", "--show-toplevel").Output()
	if err != nil {
		report.OK("git tracking", "vault file is not inside a Git worktree")
		return
	}
	root := strings.TrimSpace(string(rootBytes))
	rel, err := filepath.Rel(root, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		report.OK("git tracking", "vault file is outside Git worktree")
		return
	}
	tracked := exec.Command("git", "-C", root, "ls-files", "--error-unmatch", "--", rel).Run() == nil
	if formatErr != nil {
		report.Warn("git tracking", formatErr.Error())
		return
	}
	if tracked && format == FileFormatPlaintextStore {
		report.Fail("git tracking", "tracked plaintext secret store is unsafe: "+rel)
		return
	}
	if tracked && format == FileFormatEncryptedVault {
		report.OK("git tracking", "tracked vault is encrypted: "+rel)
		return
	}
	if tracked {
		report.Warn("git tracking", "tracked vault path is not confirmed encrypted: "+rel)
		return
	}
	report.OK("git tracking", "vault file is not tracked by ordinary git")
}
