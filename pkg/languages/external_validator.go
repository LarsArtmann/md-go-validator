// Package languages provides language detection and validation support.
package languages

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ExternalValidator validates code using external command-line tools.
type ExternalValidator struct {
	language     Language
	command      string
	args         []string
	timeout      time.Duration
	needsFile    bool
	fileExt      string
	checkCmd     []string // Optional: command to check if tool is available
}

// Language returns the language this validator handles.
func (v *ExternalValidator) Language() Language {
	return v.language
}

// IsAvailable checks if the external tool is installed.
func (v *ExternalValidator) IsAvailable() bool {
	// If checkCmd is provided, use it
	if len(v.checkCmd) > 0 {
		cmd := exec.CommandContext(context.Background(), v.checkCmd[0], v.checkCmd[1:]...)
		return cmd.Run() == nil
	}

	// Otherwise, try to find the main command
	_, err := exec.LookPath(v.command)
	return err == nil
}

// Validate validates code using the external tool.
func (v *ExternalValidator) Validate(ctx context.Context, code string) error {
	if v.needsFile {
		return v.validateWithFile(ctx, code)
	}
	return v.validateWithStdin(ctx, code)
}

// validateWithStdin validates code by piping it to the command via stdin.
func (v *ExternalValidator) validateWithStdin(ctx context.Context, code string) error {
	ctx, cancel := context.WithTimeout(ctx, v.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, v.command, v.args...)
	cmd.Stdin = strings.NewReader(code)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return &ValidationError{
			Message: fmt.Sprintf("%s validation error: %s", v.language, stderr.String()),
			Line:    0,
			Column:  0,
		}
	}
	return nil
}

// validateWithFile validates code by writing to a temp file and running the command.
func (v *ExternalValidator) validateWithFile(ctx context.Context, code string) error {
	ctx, cancel := context.WithTimeout(ctx, v.timeout)
	defer cancel()

	// Create temp file with appropriate extension
	tmpFile, err := os.CreateTemp("", "validate-*"+v.fileExt)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	if _, err := tmpFile.WriteString(code); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	// Build command args, replacing {{FILE}} placeholder if present
	args := make([]string, len(v.args))
	for i, arg := range v.args {
		if arg == "{{FILE}}" {
			args[i] = tmpFile.Name()
		} else {
			args[i] = arg
		}
	}

	cmd := exec.CommandContext(ctx, v.command, args...)
	cmd.Dir = filepath.Dir(tmpFile.Name())

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return &ValidationError{
			Message: fmt.Sprintf("%s validation error: %s", v.language, stderr.String()),
			Line:    0,
			Column:  0,
		}
	}
	return nil
}

// NewExternalValidator creates a new external validator.
func NewExternalValidator(
	language Language,
	command string,
	args []string,
	timeout time.Duration,
	needsFile bool,
	fileExt string,
	checkCmd []string,
) *ExternalValidator {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &ExternalValidator{
		language:  language,
		command:   command,
		args:      args,
		timeout:   timeout,
		needsFile: needsFile,
		fileExt:   fileExt,
		checkCmd:  checkCmd,
	}
}
