package secret

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

type EditRequest struct {
	Path   string
	Editor string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type Editable struct {
	GroupPath   string          `json:"group_path"`
	Key         string          `json:"key"`
	Value       json.RawMessage `json:"value"`
	Env         string          `json:"env,omitempty"`
	Description string          `json:"description,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
}

func NewEditable(path string, secret vault.Secret) (Editable, error) {
	id, err := vault.ParseSecretID(path)
	if err != nil {
		return Editable{}, err
	}
	return Editable{GroupPath: id.GroupPath, Key: id.Key, Value: secret.Value, Env: secret.Env, Description: secret.Description, Tags: secret.Tags}, nil
}

func (e Editable) Secret() (vault.SecretID, vault.Secret) {
	return vault.SecretID{GroupPath: e.GroupPath, Key: e.Key}, vault.Secret{Value: e.Value, Env: e.Env, Description: e.Description, Tags: e.Tags}
}

func Edit(st *vault.Store, req EditRequest) error {
	secret, ok := st.Get(req.Path)
	if !ok {
		return fmt.Errorf("secret not found: %s", req.Path)
	}
	editable, err := NewEditable(req.Path, secret)
	if err != nil {
		return err
	}
	bytes, err := json.MarshalIndent(editable, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := createPrivateTempFile("edit.json")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer removePrivateTempFile(tmpName)
	if _, err := tmp.Write(append(bytes, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	editorCmd := exec.Command("sh", "-c", "$SHELF_EDITOR \"$SHELF_EDIT_FILE\"")
	editorCmd.Env = append(os.Environ(), "SHELF_EDITOR="+req.Editor, "SHELF_EDIT_FILE="+tmpName)
	editorCmd.Stdin = req.Stdin
	editorCmd.Stdout = req.Stdout
	editorCmd.Stderr = req.Stderr
	if err := editorCmd.Run(); err != nil {
		return err
	}
	edited, err := os.ReadFile(tmpName)
	if err != nil {
		return err
	}
	var updated Editable
	if err := json.Unmarshal(edited, &updated); err != nil {
		return err
	}
	id, updatedSecret := updated.Secret()
	return st.Update(req.Path, id, updatedSecret)
}
