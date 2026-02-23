//nolint:goconst
package github_helper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// UploadFileToGitHubBlob
// ---------------------------------------------------------------------------

func TestUploadFileToGitHubBlob_Success(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, http.StatusCreated, GithubBlobResponse{Sha: "blob-sha", Url: "http://test"})
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test-blob-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, _ = tmpFile.WriteString("test content")
	tmpFile.Close()

	resp, err := UploadFileToGitHubBlob(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Sha != "blob-sha" {
		t.Errorf("expected blob-sha, got %s", resp.Sha)
	}
}

func TestUploadFileToGitHubBlob_NonExistentFile(t *testing.T) {
	resetGlobals()

	_, err := UploadFileToGitHubBlob("/nonexistent/path/file.txt")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestUploadFileToGitHubBlob_APIError(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"error"}`))
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	tmpFile, err := os.CreateTemp("", "test-blob-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, _ = tmpFile.WriteString("test content")
	tmpFile.Close()

	_, err = UploadFileToGitHubBlob(tmpFile.Name())
	if err == nil {
		t.Fatal("expected error for API error")
	}
}

// ---------------------------------------------------------------------------
// UploadFilesToGitHubBlob
// ---------------------------------------------------------------------------

func TestUploadFilesToGitHubBlob_MixedFiles(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, http.StatusCreated, GithubBlobResponse{Sha: "blob-sha", Url: "http://test"})
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	// Create a real file
	tmpFile, err := os.CreateTemp("", "test-upload-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, _ = tmpFile.WriteString("content")
	tmpFile.Close()

	files := []string{tmpFile.Name(), "/nonexistent/deleted-file.go"}

	gitFiles, err := UploadFilesToGitHubBlob(files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(gitFiles) != 2 {
		t.Fatalf("expected 2 git files, got %d", len(gitFiles))
	}

	if gitFiles[0].WasDeleted {
		t.Error("expected first file to not be deleted")
	}

	if !gitFiles[1].WasDeleted {
		t.Error("expected second file to be marked as deleted")
	}
}

func TestUploadFilesToGitHubBlob_EmptyList(t *testing.T) {
	resetGlobals()

	gitFiles, err := UploadFilesToGitHubBlob([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(gitFiles) != 0 {
		t.Errorf("expected 0 files, got %d", len(gitFiles))
	}
}

func TestUploadFilesToGitHubBlob_APIErrorOnExistingFile(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"error"}`))
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	tmpFile, err := os.CreateTemp("", "test-upload-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, _ = tmpFile.WriteString("content")
	tmpFile.Close()

	_, err = UploadFilesToGitHubBlob([]string{tmpFile.Name()})
	if err == nil {
		t.Fatal("expected error for API error")
	}
}

// ---------------------------------------------------------------------------
// SignJWTAppToken
// ---------------------------------------------------------------------------

func TestSignJWTAppToken_Success(t *testing.T) {
	resetGlobals()

	pemBytes := generateTestRSAKey(t)

	err := SignJWTAppToken("12345", pemBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSignJWTAppToken_InvalidPEM(t *testing.T) {
	resetGlobals()

	err := SignJWTAppToken("12345", []byte("invalid"))
	if err == nil {
		t.Fatal("expected error for invalid PEM")
	}
}

// ---------------------------------------------------------------------------
// SignJWTAppTokenWithFilename
// ---------------------------------------------------------------------------

func TestSignJWTAppTokenWithFilename_Success(t *testing.T) {
	resetGlobals()

	pemBytes := generateTestRSAKey(t)

	tmpFile, err := os.CreateTemp("", "test-pem-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, _ = tmpFile.Write(pemBytes)
	tmpFile.Close()

	err = SignJWTAppTokenWithFilename("12345", tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSignJWTAppTokenWithFilename_EmptyFilename(t *testing.T) {
	resetGlobals()

	err := SignJWTAppTokenWithFilename("12345", "")
	if err == nil {
		t.Fatal("expected error for empty filename")
	}
}

func TestSignJWTAppTokenWithFilename_MissingFile(t *testing.T) {
	resetGlobals()

	err := SignJWTAppTokenWithFilename("12345", "/nonexistent/file.pem")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// ---------------------------------------------------------------------------
// GenerateInstallationAppToken
// ---------------------------------------------------------------------------

func TestGenerateInstallationAppToken_Success(t *testing.T) {
	resetGlobals()

	pemBytes := generateTestRSAKey(t)
	_ = initAppToken("12345", pemBytes)

	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// access_tokens path also contains "installation", so check it first
		if strings.Contains(r.URL.Path, "/access_tokens") {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"token":"ghs_install_token","expires_at":"2024-01-01T00:00:00Z"}`))

			return
		}

		if strings.Contains(r.URL.Path, "/installation") {
			jsonResponse(w, http.StatusOK, GithubAppInstallationResponse{Id: 42})
			return
		}

		http.NotFound(w, r)
	}))

	token, err := GenerateInstallationAppToken(GitHubRepo{Owner: "owner", Repo: "repo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token.Token != "ghs_install_token" {
		t.Errorf("expected ghs_install_token, got %s", token.Token)
	}
}

// ---------------------------------------------------------------------------
// CommitAndPush
// ---------------------------------------------------------------------------

func TestCommitAndPush_HappyPath(t *testing.T) {
	resetGlobals()

	// Mock exec for git commands
	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		if command == "git" && len(args) > 0 {
			if args[0] == "add" {
				return []byte(""), nil
			}

			if args[0] == "diff" {
				return []byte("file1.go\n"), nil
			}
		}

		return []byte(""), nil
	})

	// Create a temp file that UploadFileToGitHubBlob can find
	tmpFile, err := os.CreateTemp("", "file1.go")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, _ = tmpFile.WriteString("package main")
	tmpFile.Close()

	// Mock git diff to return the temp file name
	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		if command == "git" && len(args) > 0 {
			if args[0] == "add" {
				return []byte(""), nil
			}

			if args[0] == "diff" {
				return []byte(tmpFile.Name() + "\n"), nil
			}
		}

		return []byte(""), nil
	})

	// Set up summary/output env
	summaryFile, _ := os.CreateTemp("", "gh-summary-*")
	defer os.Remove(summaryFile.Name())

	summaryFile.Close()

	outputFile, _ := os.CreateTemp("", "gh-output-*")
	defer os.Remove(outputFile.Name())

	outputFile.Close()
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile.Name())
	t.Setenv("GITHUB_OUTPUT", outputFile.Name())

	setupTestServer(t, apiRouter(map[string]http.HandlerFunc{
		"GET /repos/owner/repo/git/refs/heads/main": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusOK, GithubRefResponse{
				Ref:    "refs/heads/main",
				Object: RefObject{Sha: "base-sha"},
			})
		},
		"POST /repos/owner/repo/git/blobs": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubBlobResponse{Sha: "blob-sha"})
		},
		"POST /repos/owner/repo/git/trees": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubTreeResponse{Sha: "tree-sha"})
		},
		"POST /repos/owner/repo/git/commits": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubCommitResponse{Sha: "commit-sha"})
		},
		"PATCH /repos/owner/repo/git/refs/heads/main": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusOK, GithubRefResponse{
				Ref:    "refs/heads/main",
				Object: RefObject{Sha: "commit-sha"},
			})
		},
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	sha, err := CommitAndPush(
		GitHubRepo{Owner: "owner", Repo: "repo"},
		GitCommit{
			Branch:  "main",
			Message: "test commit",
			Options: CommitOptions{AddNewFiles: true},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sha != "commit-sha" {
		t.Errorf("expected commit-sha, got %s", sha)
	}
}

func TestCommitAndPush_WithCoauthors(t *testing.T) {
	resetGlobals()

	tmpFile, err := os.CreateTemp("", "file1.go")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, _ = tmpFile.WriteString("package main")
	tmpFile.Close()

	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		if args[0] == "add" {
			return []byte(""), nil
		}

		return []byte(tmpFile.Name() + "\n"), nil
	})

	t.Setenv("GITHUB_ACTIONS", "false")

	var capturedCommitMsg string

	setupTestServer(t, apiRouter(map[string]http.HandlerFunc{
		"GET /repos/owner/repo/git/refs/heads/main": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusOK, GithubRefResponse{Object: RefObject{Sha: "base-sha"}})
		},
		"POST /repos/owner/repo/git/blobs": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubBlobResponse{Sha: "blob-sha"})
		},
		"POST /repos/owner/repo/git/trees": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubTreeResponse{Sha: "tree-sha"})
		},
		"POST /repos/owner/repo/git/commits": func(w http.ResponseWriter, r *http.Request) {
			var req GithubCommitRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			capturedCommitMsg = req.Message

			jsonResponse(w, http.StatusCreated, GithubCommitResponse{Sha: "commit-sha"})
		},
		"PATCH /repos/owner/repo/git/refs/heads/main": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusOK, GithubRefResponse{Object: RefObject{Sha: "commit-sha"}})
		},
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	coauthors := []GitUser{{Name: "Alice", Email: "alice@test.com"}}

	_, err = CommitAndPush(
		GitHubRepo{Owner: "owner", Repo: "repo"},
		GitCommit{
			Branch:    "main",
			Message:   "test commit",
			Coauthors: &coauthors,
			Options:   CommitOptions{AddNewFiles: true},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedCommitMsg, "Co-authored-by: Alice <alice@test.com>") {
		t.Errorf("expected coauthor in commit message, got: %s", capturedCommitMsg)
	}
}

func TestCommitAndPush_BranchNotFound_CreatesBranch(t *testing.T) {
	resetGlobals()

	tmpFile, err := os.CreateTemp("", "file1.go")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, _ = tmpFile.WriteString("content")
	tmpFile.Close()

	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		if args[0] == "add" {
			return []byte(""), nil
		}

		return []byte(tmpFile.Name() + "\n"), nil
	})

	t.Setenv("GITHUB_ACTIONS", "false")

	headBranch := "main"

	setupTestServer(t, apiRouter(map[string]http.HandlerFunc{
		"GET /repos/owner/repo/git/refs/heads/main": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusOK, GithubRefResponse{Object: RefObject{Sha: "base-sha"}})
		},
		"POST /repos/owner/repo/git/blobs": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubBlobResponse{Sha: "blob-sha"})
		},
		"POST /repos/owner/repo/git/trees": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubTreeResponse{Sha: "tree-sha"})
		},
		"POST /repos/owner/repo/git/commits": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubCommitResponse{Sha: "commit-sha"})
		},
		"PATCH /repos/owner/repo/git/refs/heads/new-branch": func(w http.ResponseWriter, _ *http.Request) {
			// 404 - branch doesn't exist
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"Not Found"}`))
		},
		"POST /repos/owner/repo/git/refs": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubRefResponse{
				Ref:    "refs/heads/new-branch",
				Object: RefObject{Sha: "commit-sha"},
			})
		},
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	sha, err := CommitAndPush(
		GitHubRepo{Owner: "owner", Repo: "repo"},
		GitCommit{
			Branch:     "new-branch",
			HeadBranch: &headBranch,
			Message:    "test commit",
			Options:    CommitOptions{AddNewFiles: true},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sha != "commit-sha" {
		t.Errorf("expected commit-sha, got %s", sha)
	}
}

func TestCommitAndPush_GetReferenceError(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"error"}`))
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	_, err := CommitAndPush(
		GitHubRepo{Owner: "owner", Repo: "repo"},
		GitCommit{Branch: "main", Message: "test"},
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "error getting head reference") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCommitAndPush_ForceFlag(t *testing.T) {
	resetGlobals()

	tmpFile, err := os.CreateTemp("", "file1.go")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, _ = tmpFile.WriteString("content")
	tmpFile.Close()

	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		if args[0] == "add" {
			return []byte(""), nil
		}

		return []byte(tmpFile.Name() + "\n"), nil
	})

	t.Setenv("GITHUB_ACTIONS", "false")

	var capturedForce bool

	setupTestServer(t, apiRouter(map[string]http.HandlerFunc{
		"GET /repos/owner/repo/git/refs/heads/main": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusOK, GithubRefResponse{Object: RefObject{Sha: "base-sha"}})
		},
		"POST /repos/owner/repo/git/blobs": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubBlobResponse{Sha: "blob-sha"})
		},
		"POST /repos/owner/repo/git/trees": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubTreeResponse{Sha: "tree-sha"})
		},
		"POST /repos/owner/repo/git/commits": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubCommitResponse{Sha: "commit-sha"})
		},
		"PATCH /repos/owner/repo/git/refs/heads/main": func(w http.ResponseWriter, r *http.Request) {
			var req GithubRefRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			capturedForce = req.Force

			jsonResponse(w, http.StatusOK, GithubRefResponse{Object: RefObject{Sha: "commit-sha"}})
		},
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	_, err = CommitAndPush(
		GitHubRepo{Owner: "owner", Repo: "repo"},
		GitCommit{
			Branch:  "main",
			Message: "test commit",
			Options: CommitOptions{AddNewFiles: true, Force: true},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !capturedForce {
		t.Error("expected force flag to be true")
	}
}

func TestCommitAndPush_HeadBranchDefaultsToBranch(t *testing.T) {
	resetGlobals()
	t.Setenv("GITHUB_ACTIONS", "false")

	var capturedRefPath string

	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/git/refs/heads/") {
			capturedRefPath = r.URL.Path

			jsonResponse(w, http.StatusOK, GithubRefResponse{Object: RefObject{Sha: "base-sha"}})

			return
		}
		// Return errors for subsequent calls to stop the flow early
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"stop"}`))
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		return nil, fmt.Errorf("stop here")
	})

	// HeadBranch is nil, should default to Branch="develop"
	_, _ = CommitAndPush(
		GitHubRepo{Owner: "owner", Repo: "repo"},
		GitCommit{Branch: "develop", Message: "test"},
	)

	if !strings.Contains(capturedRefPath, "heads/develop") {
		t.Errorf("expected head branch to default to 'develop', path was: %s", capturedRefPath)
	}
}

// ---------------------------------------------------------------------------
// CreateTagAndPush
// ---------------------------------------------------------------------------

func TestCreateTagAndPush_NewTag(t *testing.T) {
	resetGlobals()
	t.Setenv("GITHUB_ACTIONS", "false")

	setupTestServer(t, apiRouter(map[string]http.HandlerFunc{
		"POST /repos/owner/repo/git/tags": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubTagResponse{Sha: "tag-sha", Tag: "v1.0.0"})
		},
		"GET /repos/owner/repo/git/refs/tags/v1.0.0": func(w http.ResponseWriter, _ *http.Request) {
			// 404 - tag doesn't exist
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"Not Found"}`))
		},
		"POST /repos/owner/repo/git/refs": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubRefResponse{
				Ref:    "refs/tags/v1.0.0",
				Object: RefObject{Sha: "tag-sha"},
			})
		},
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	err := CreateTagAndPush(GitTag{TagName: "v1.0.0", Message: "release", CommitSha: "commit-sha"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateTagAndPush_ExistingTag(t *testing.T) {
	resetGlobals()
	t.Setenv("GITHUB_ACTIONS", "false")

	setupTestServer(t, apiRouter(map[string]http.HandlerFunc{
		"POST /repos/owner/repo/git/tags": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubTagResponse{Sha: "tag-sha", Tag: "v1.0.0"})
		},
		"GET /repos/owner/repo/git/refs/tags/v1.0.0": func(w http.ResponseWriter, _ *http.Request) {
			// Tag exists
			jsonResponse(w, http.StatusOK, GithubRefResponse{
				Ref:    "refs/tags/v1.0.0",
				Object: RefObject{Sha: "old-tag-sha"},
			})
		},
		"PATCH /repos/owner/repo/git/refs/tags/v1.0.0": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusOK, GithubRefResponse{
				Ref:    "refs/tags/v1.0.0",
				Object: RefObject{Sha: "tag-sha"},
			})
		},
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	err := CreateTagAndPush(GitTag{TagName: "v1.0.0", Message: "release", CommitSha: "commit-sha"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateTagAndPush_CreateTagAPIError(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"error"}`))
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	err := CreateTagAndPush(GitTag{TagName: "v1.0.0", Message: "release", CommitSha: "commit-sha"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateTagAndPush_Non404GetReferenceError(t *testing.T) {
	resetGlobals()
	t.Setenv("GITHUB_ACTIONS", "false")

	setupTestServer(t, apiRouter(map[string]http.HandlerFunc{
		"POST /repos/owner/repo/git/tags": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubTagResponse{Sha: "tag-sha", Tag: "v1.0.0"})
		},
		"GET /repos/owner/repo/git/refs/tags/v1.0.0": func(w http.ResponseWriter, _ *http.Request) {
			// Non-404 error
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"message":"Forbidden"}`))
		},
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	err := CreateTagAndPush(GitTag{TagName: "v1.0.0", Message: "release", CommitSha: "commit-sha"})
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "error checking tag reference") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCommitAndPush_AddNewFilesFalse(t *testing.T) {
	resetGlobals()

	tmpFile, err := os.CreateTemp("", "file1.go")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, _ = tmpFile.WriteString("content")
	tmpFile.Close()

	var stageCmdArgs []string

	mockExecCommand(t, func(command string, args []string) ([]byte, error) {
		if args[0] == "add" {
			stageCmdArgs = args

			return []byte(""), nil
		}

		return []byte(tmpFile.Name() + "\n"), nil
	})

	t.Setenv("GITHUB_ACTIONS", "false")

	setupTestServer(t, apiRouter(map[string]http.HandlerFunc{
		"GET /repos/owner/repo/git/refs/heads/main": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusOK, GithubRefResponse{Object: RefObject{Sha: "base-sha"}})
		},
		"POST /repos/owner/repo/git/blobs": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubBlobResponse{Sha: "blob-sha"})
		},
		"POST /repos/owner/repo/git/trees": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubTreeResponse{Sha: "tree-sha"})
		},
		"POST /repos/owner/repo/git/commits": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusCreated, GithubCommitResponse{Sha: "commit-sha"})
		},
		"PATCH /repos/owner/repo/git/refs/heads/main": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, http.StatusOK, GithubRefResponse{Object: RefObject{Sha: "commit-sha"}})
		},
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	_, err = CommitAndPush(
		GitHubRepo{Owner: "owner", Repo: "repo"},
		GitCommit{
			Branch:  "main",
			Message: "test commit",
			Options: CommitOptions{AddNewFiles: false},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// When AddNewFiles=false, should use StageModifiedFiles which calls "git add -u"
	if len(stageCmdArgs) >= 2 && stageCmdArgs[1] != "-u" {
		t.Errorf("expected 'git add -u' for AddNewFiles=false, got args: %v", stageCmdArgs)
	}
}
