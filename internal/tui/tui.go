// Package tui provides Shelf's optional interactive terminal workspace.
package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"github.com/zhongyangchuwu/shelf-go/internal/clipboard"
)

type Service interface {
	SecretCatalog(configPathFlag, vaultPathFlag string) ([]app.SecretCatalogItem, error)
	SecretInfo(configPathFlag, vaultPathFlag, selector string) (app.SecretInfo, error)
	CreateSecret(configPathFlag, vaultPathFlag string, req app.CreateSecretRequest) (string, error)
	UpdateSecret(configPathFlag, vaultPathFlag string, req app.UpdateSecretRequest) error
	GetSecretField(configPathFlag, vaultPathFlag string, req app.GetSecretFieldRequest) (string, error)
	RemoveSecret(configPathFlag, vaultPathFlag, selector string) error
}

type Options struct {
	ConfigPath string
	VaultPath  string
}

func Run(service Service, clipboardWriter clipboard.Writer, options Options) error {
	if service == nil {
		return fmt.Errorf("TUI service is required")
	}
	if clipboardWriter == nil {
		return fmt.Errorf("TUI clipboard writer is required")
	}
	_, err := tea.NewProgram(NewModel(service, clipboardWriter, options)).Run()
	return err
}
