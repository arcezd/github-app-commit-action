//nolint:goconst
package github_helper

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// executeCommandDefault
// ---------------------------------------------------------------------------

func TestExecuteCommandDefault_Success(t *testing.T) {
	output, err := executeCommandDefault("echo", []string{"hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(string(output)) != "hello" {
		t.Errorf("expected 'hello', got '%s'", strings.TrimSpace(string(output)))
	}
}

func TestExecuteCommandDefault_NonExistent(t *testing.T) {
	_, err := executeCommandDefault("nonexistent-command-xyz", []string{})
	if err == nil {
		t.Fatal("expected error for non-existent command")
	}
}

// ---------------------------------------------------------------------------
// StageModifiedAndNewFiles
// ---------------------------------------------------------------------------

func TestStageModifiedAndNewFiles_Success(t *testing.T) {
	var calledCmd string
	var calledArgs []string

	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		calledCmd = command
		calledArgs = args

		return []byte(""), nil
	})

	err := StageModifiedAndNewFiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if calledCmd != "git" {
		t.Errorf("expected git, got %s", calledCmd)
	}

	if len(calledArgs) < 2 || calledArgs[0] != "add" || calledArgs[1] != "-A" {
		t.Errorf("expected 'add -A', got %v", calledArgs)
	}
}

func TestStageModifiedAndNewFiles_Error(t *testing.T) {
	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		return nil, fmt.Errorf("git error")
	})

	err := StageModifiedAndNewFiles()
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// StageModifiedFiles
// ---------------------------------------------------------------------------

func TestStageModifiedFiles_Success(t *testing.T) {
	var calledArgs []string

	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		calledArgs = args
		return []byte(""), nil
	})

	err := StageModifiedFiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(calledArgs) < 2 || calledArgs[0] != "add" || calledArgs[1] != "-u" {
		t.Errorf("expected 'add -u', got %v", calledArgs)
	}
}

func TestStageModifiedFiles_Error(t *testing.T) {
	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		return nil, fmt.Errorf("git error")
	})

	err := StageModifiedFiles()
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// GetModifiedAndNewFiles
// ---------------------------------------------------------------------------

func TestGetModifiedAndNewFiles_Success(t *testing.T) {
	callCount := 0

	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		callCount++
		if callCount == 1 {
			// StageModifiedAndNewFiles → git add -A
			return []byte(""), nil
		}
		// GetModifiedFilesFromGitDiff → git diff --name-only --cached
		return []byte("file1.go\nfile2.go\n"), nil
	})

	files, err := GetModifiedAndNewFiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestGetModifiedAndNewFiles_StageError(t *testing.T) {
	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		return nil, fmt.Errorf("stage error")
	})

	_, err := GetModifiedAndNewFiles()
	if err == nil {
		t.Fatal("expected error to propagate from staging")
	}
}

// ---------------------------------------------------------------------------
// GetModifiedFiles
// ---------------------------------------------------------------------------

func TestGetModifiedFiles_Success(t *testing.T) {
	callCount := 0

	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		callCount++
		if callCount == 1 {
			return []byte(""), nil
		}

		return []byte("modified.go\n"), nil
	})

	files, err := GetModifiedFiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 1 || files[0] != "modified.go" {
		t.Errorf("unexpected files: %v", files)
	}
}

// ---------------------------------------------------------------------------
// GetModifiedFilesFromGitDiff
// ---------------------------------------------------------------------------

func TestGetModifiedFilesFromGitDiff_Staged(t *testing.T) {
	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		if command != "git" {
			t.Errorf("expected git, got %s", command)
		}
		// Should have --cached for staged
		found := false

		for _, a := range args {
			if a == "--cached" {
				found = true
			}
		}

		if !found {
			t.Error("expected --cached flag for staged diff")
		}

		return []byte("a.go\nb.go\n"), nil
	})

	files, err := GetModifiedFilesFromGitDiff(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestGetModifiedFilesFromGitDiff_Unstaged(t *testing.T) {
	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		for _, a := range args {
			if a == "--cached" {
				t.Error("did not expect --cached flag for unstaged diff")
			}
		}

		return []byte("c.go\n"), nil
	})

	files, err := GetModifiedFilesFromGitDiff(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}
}

func TestGetModifiedFilesFromGitDiff_EmptyOutput(t *testing.T) {
	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		return []byte(""), nil
	})

	files, err := GetModifiedFilesFromGitDiff(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

// ---------------------------------------------------------------------------
// IsGitHubActions
// ---------------------------------------------------------------------------

func TestIsGitHubActions_True(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")

	if !IsGitHubActions() {
		t.Error("expected true when GITHUB_ACTIONS=true")
	}
}

func TestIsGitHubActions_Unset(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "")

	if IsGitHubActions() {
		t.Error("expected false when GITHUB_ACTIONS is not set")
	}
}

// ---------------------------------------------------------------------------
// AppendToGHActionsSummary
// ---------------------------------------------------------------------------

func TestAppendToGHActionsSummary_InActions(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "gh-summary-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.Close()

	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_STEP_SUMMARY", tmpFile.Name())

	AppendToGHActionsSummary("test summary")

	content, err := os.ReadFile(tmpFile.Name()) //nolint:gosec // test file
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "test summary") {
		t.Errorf("expected summary content, got: %s", string(content))
	}
}

func TestAppendToGHActionsSummary_NotInActions(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "false")
	// Should not panic or error
	AppendToGHActionsSummary("test summary")
}

// ---------------------------------------------------------------------------
// SendToGHActionsOutput
// ---------------------------------------------------------------------------

func TestSendToGHActionsOutput_InActions(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "gh-output-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.Close()

	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_OUTPUT", tmpFile.Name())

	SendToGHActionsOutput("sha", "abc123")

	content, err := os.ReadFile(tmpFile.Name()) //nolint:gosec // test file
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "sha=abc123") {
		t.Errorf("expected 'sha=abc123', got: %s", string(content))
	}
}

func TestSendToGHActionsOutput_NotInActions(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "false")
	// Should not panic or error
	SendToGHActionsOutput("sha", "abc123")
}

// ---------------------------------------------------------------------------
// writeToGHActionsVar
// ---------------------------------------------------------------------------

func TestWriteToGHActionsVar_EmptyName(t *testing.T) {
	// Should be no-op, no panic
	writeToGHActionsVar("", "some value")
}

func TestWriteToGHActionsVar_ValidFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "gh-var-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.Close()

	writeToGHActionsVar(tmpFile.Name(), "hello world")

	content, err := os.ReadFile(tmpFile.Name()) //nolint:gosec // test file
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "hello world") {
		t.Errorf("expected content, got: %s", string(content))
	}
}

// ---------------------------------------------------------------------------
// ListFiles
// ---------------------------------------------------------------------------

func TestListFiles_Success(t *testing.T) {
	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		return []byte("total 8\ndrwxr-xr-x  5 user group 160 Jan  1 12:00 .\ndrwxr-xr-x  3 user group  96 Jan  1 12:00 ..\n-rw-r--r--  1 user group  42 Jan  1 12:00 file.txt\n"), nil
	})

	files, err := ListFiles(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// should have 3 entries (., .., file.txt)
	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d", len(files))
	}
}

func TestListFiles_Error(t *testing.T) {
	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		return nil, fmt.Errorf("ls error")
	})

	_, err := ListFiles(nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListFiles_EmptyOutput(t *testing.T) {
	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		return []byte("total 0\n"), nil
	})

	files, err := ListFiles(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}
