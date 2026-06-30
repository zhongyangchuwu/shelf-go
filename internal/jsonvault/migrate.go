package jsonvault

import (
	"bytes"
	"fmt"
	"os"
)

func (v *Vault) MigratePlaintext(sourcePath string, force bool) error {
	before, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read plaintext source: %w", err)
	}
	st, err := Load(sourcePath)
	if err != nil {
		return fmt.Errorf("load plaintext source: %w", err)
	}
	format, err := DetectFileFormat(v.Path())
	if err != nil {
		return fmt.Errorf("inspect target vault: %w", err)
	}
	if format != FileFormatMissing && format != FileFormatEmpty && !force {
		return fmt.Errorf("target vault already exists; pass --force to replace %s", v.Path())
	}
	if format == FileFormatPlaintextStore {
		return fmt.Errorf("target vault is plaintext JSON; choose a different --to path or move it before migration")
	}
	if err := v.Save(st); err != nil {
		return fmt.Errorf("write encrypted vault: %w", err)
	}
	verified, err := v.Load()
	if err != nil {
		return fmt.Errorf("verify encrypted vault: %w", err)
	}
	if len(verified.Data.Secrets) != len(st.Data.Secrets) {
		return fmt.Errorf("verify encrypted vault: secret count mismatch")
	}
	after, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("re-read plaintext source: %w", err)
	}
	if !bytes.Equal(before, after) {
		return fmt.Errorf("plaintext source changed during migration")
	}
	return nil
}
