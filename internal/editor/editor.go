package editor

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// SetupTmpFileWithEditor creates a temp file in your configured TempDir and
// finds out if the `EDITOR` environment variable is set properly.
// It then sets up the file in that editor and returns a scanner to process the
// entered text.
// The caller is responsible to call the cleanup function after they are done processing.
func SetupTmpFileWithEditor(prefill string) (*bufio.Scanner, func(), error) {
	editor, exists := os.LookupEnv("EDITOR")
	if !exists {
		return nil, func() {}, errors.New("expecting `EDITOR` environment variable to be set")
	}

	tmpFile, err := os.CreateTemp(os.TempDir(), "tcc-oncall-create-*")
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to create temp file for editing: %w", err)
	}

	cleanup := func() {
		os.Remove(tmpFile.Name())
	}

	if prefill != "" {
		_, err = tmpFile.WriteString(prefill)
		if err != nil {
			return nil, cleanup, fmt.Errorf("failed to write prefill to tmpFile: %w", err)
		}
	}

	e := exec.Command(editor, tmpFile.Name())
	e.Stdin = os.Stdin
	e.Stdout = os.Stdout
	err = e.Run()
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to get text from editor: %w", err)
	}

	fBytes, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to read file contents: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(fBytes))
	return scanner, cleanup, nil
}
