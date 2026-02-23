//nolint:goconst
package github_helper

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// CallGithubAPI
// ---------------------------------------------------------------------------

func TestCallGithubAPI_GetSuccess(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Errorf("unexpected accept header: %s", r.Header.Get("Accept"))
		}

		if r.Header.Get("X-GitHub-Api-Version") != "2022-11-28" {
			t.Errorf("unexpected api version header: %s", r.Header.Get("X-GitHub-Api-Version"))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"key":"value"}`))
	}))

	resp, err := CallGithubAPI("test-token", "GET", "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp != `{"key":"value"}` {
		t.Errorf("unexpected response: %s", resp)
	}
}

func TestCallGithubAPI_PostWithBody(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var data map[string]string
		_ = json.Unmarshal(body, &data)

		if data["name"] != "test" {
			t.Errorf("unexpected body: %s", string(body))
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"sha":"abc123"}`))
	}))

	resp, err := CallGithubAPI("test-token", "POST", "/test", map[string]string{"name": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(resp, "abc123") {
		t.Errorf("unexpected response: %s", resp)
	}
}

func TestCallGithubAPI_404Error(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	}))

	_, err := CallGithubAPI("test-token", "GET", "/not-found", nil)
	if err == nil {
		t.Fatal("expected error for 404")
	}

	if !IsNotFound(err) {
		t.Errorf("expected 404 error, got: %v", err)
	}
}

func TestCallGithubAPI_500Error(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"Internal Server Error"}`))
	}))

	_, err := CallGithubAPI("test-token", "GET", "/error", nil)
	if err == nil {
		t.Fatal("expected error for 500")
	}

	apiErr, ok := err.(*GitHubAPIError)
	if !ok {
		t.Fatalf("expected GitHubAPIError, got %T", err)
	}

	if apiErr.StatusCode != 500 {
		t.Errorf("expected status 500, got %d", apiErr.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// GenerateToken
// ---------------------------------------------------------------------------

func TestGenerateToken_ValidKey(t *testing.T) {
	pemBytes := generateTestRSAKey(t)

	token, err := GenerateToken("12345", pemBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Error("expected non-empty token")
	}

	// JWT has 3 parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("expected 3 JWT parts, got %d", len(parts))
	}
}

func TestGenerateToken_InvalidPEM(t *testing.T) {
	_, err := GenerateToken("12345", []byte("not-a-valid-pem"))
	if err == nil {
		t.Fatal("expected error for invalid PEM")
	}
}

// ---------------------------------------------------------------------------
// GenerateInstallationAccessToken
// ---------------------------------------------------------------------------

func TestGenerateInstallationAccessToken_Success(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/app/installations/42/access_tokens") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"token":"ghs_test123","expires_at":"2024-01-01T00:00:00Z"}`))
	}))

	token, err := GenerateInstallationAccessToken("jwt-token", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token != "ghs_test123" {
		t.Errorf("expected ghs_test123, got %s", token)
	}
}

func TestGenerateInstallationAccessToken_APIError(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Bad credentials"}`))
	}))

	_, err := GenerateInstallationAccessToken("bad-token", 42)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGenerateInstallationAccessToken_InvalidJSON(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`not-json`))
	}))

	_, err := GenerateInstallationAccessToken("jwt-token", 42)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// ---------------------------------------------------------------------------
// GetReference
// ---------------------------------------------------------------------------

func TestGetReference_Success(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, http.StatusOK, GithubRefResponse{
			Ref:    "refs/heads/main",
			Object: RefObject{Sha: "abc123", Type: "commit"},
		})
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	ref, err := GetReference("refs/heads/main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ref.Object.Sha != "abc123" {
		t.Errorf("expected sha abc123, got %s", ref.Object.Sha)
	}
}

func TestGetReference_NilToken(t *testing.T) {
	resetGlobals()

	_, err := GetReference("refs/heads/main")
	if err == nil {
		t.Fatal("expected error for nil token")
	}
}

func TestGetReference_APIError(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	_, err := GetReference("refs/heads/nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// CreateTree
// ---------------------------------------------------------------------------

func TestCreateTree_Success(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, http.StatusCreated, GithubTreeResponse{Sha: "tree-sha"})
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	resp, err := CreateTree(GithubTreeRequest{BaseTree: "base", Tree: []TreeItem{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Sha != "tree-sha" {
		t.Errorf("expected tree-sha, got %s", resp.Sha)
	}
}

func TestCreateTree_NilToken(t *testing.T) {
	resetGlobals()

	_, err := CreateTree(GithubTreeRequest{})
	if err == nil {
		t.Fatal("expected error for nil token")
	}
}

// ---------------------------------------------------------------------------
// CreateCommit
// ---------------------------------------------------------------------------

func TestCreateCommit_Success(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, http.StatusCreated, GithubCommitResponse{Sha: "commit-sha"})
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	resp, err := CreateCommit(GithubCommitRequest{Message: "test", Tree: "tree-sha", Parents: []string{"parent"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Sha != "commit-sha" {
		t.Errorf("expected commit-sha, got %s", resp.Sha)
	}
}

func TestCreateCommit_NilToken(t *testing.T) {
	resetGlobals()

	_, err := CreateCommit(GithubCommitRequest{})
	if err == nil {
		t.Fatal("expected error for nil token")
	}
}

// ---------------------------------------------------------------------------
// CreateReference
// ---------------------------------------------------------------------------

func TestCreateReference_Success(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, http.StatusCreated, GithubRefResponse{
			Ref:    "refs/heads/new-branch",
			Object: RefObject{Sha: "ref-sha"},
		})
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	resp, err := CreateReference(GithubRefRequest{Ref: "refs/heads/new-branch", Sha: "ref-sha"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Object.Sha != "ref-sha" {
		t.Errorf("expected ref-sha, got %s", resp.Object.Sha)
	}
}

func TestCreateReference_NilToken(t *testing.T) {
	resetGlobals()

	_, err := CreateReference(GithubRefRequest{})
	if err == nil {
		t.Fatal("expected error for nil token")
	}
}

// ---------------------------------------------------------------------------
// UpdateReference
// ---------------------------------------------------------------------------

func TestUpdateReference_PatchSuccess(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		jsonResponse(w, http.StatusOK, GithubRefResponse{
			Ref:    "refs/heads/main",
			Object: RefObject{Sha: "updated-sha"},
		})
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	resp, err := UpdateReference(GithubRefRequest{Sha: "updated-sha"}, "heads/main", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Object.Sha != "updated-sha" {
		t.Errorf("expected updated-sha, got %s", resp.Object.Sha)
	}
}

func TestUpdateReference_CreateBranch(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST for create branch, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var req GithubRefRequest
		_ = json.Unmarshal(body, &req)

		if !strings.HasPrefix(req.Ref, "refs/heads/") {
			t.Errorf("expected refs/heads/ prefix, got %s", req.Ref)
		}

		jsonResponse(w, http.StatusCreated, GithubRefResponse{
			Ref:    req.Ref,
			Object: RefObject{Sha: "new-sha"},
		})
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	resp, err := UpdateReference(GithubRefRequest{Sha: "new-sha"}, "heads/new-branch", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Object.Sha != "new-sha" {
		t.Errorf("expected new-sha, got %s", resp.Object.Sha)
	}
}

func TestUpdateReference_NilToken(t *testing.T) {
	resetGlobals()

	_, err := UpdateReference(GithubRefRequest{}, "heads/main", false)
	if err == nil {
		t.Fatal("expected error for nil token")
	}
}

// ---------------------------------------------------------------------------
// CreateBlob
// ---------------------------------------------------------------------------

func TestCreateBlob_Success(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, http.StatusCreated, GithubBlobResponse{Sha: "blob-sha", Url: "http://test"})
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	resp, err := CreateBlob(GithubBlobRequest{Content: "dGVzdA==", Encoding: "base64"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Sha != "blob-sha" {
		t.Errorf("expected blob-sha, got %s", resp.Sha)
	}
}

func TestCreateBlob_NilToken(t *testing.T) {
	resetGlobals()

	_, err := CreateBlob(GithubBlobRequest{})
	if err == nil {
		t.Fatal("expected error for nil token")
	}
}

// ---------------------------------------------------------------------------
// CreateTag
// ---------------------------------------------------------------------------

func TestCreateTag_Success(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, http.StatusCreated, GithubTagResponse{Sha: "tag-sha", Tag: "v1.0.0"})
	}))
	setTestAppToken(t, &GitHubAppToken{
		Token: "test-token",
		Repo:  GitHubRepo{Owner: "owner", Repo: "repo"},
	})

	resp, err := CreateTag(GithubTagRequest{Tag: "v1.0.0", Message: "release", Object: "commit-sha", Type: "commit"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Sha != "tag-sha" {
		t.Errorf("expected tag-sha, got %s", resp.Sha)
	}
}

func TestCreateTag_NilToken(t *testing.T) {
	resetGlobals()

	_, err := CreateTag(GithubTagRequest{})
	if err == nil {
		t.Fatal("expected error for nil token")
	}
}

// ---------------------------------------------------------------------------
// GetAppInstallationDetails
// ---------------------------------------------------------------------------

func TestGetAppInstallationDetails_Success(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/repos/owner/repo/installation") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		jsonResponse(w, http.StatusOK, GithubAppInstallationResponse{Id: 42, AppId: 12345})
	}))

	resp, err := GetAppInstallationDetails("jwt-token", GitHubRepo{Owner: "owner", Repo: "repo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Id != 42 {
		t.Errorf("expected installation id 42, got %d", resp.Id)
	}
}

func TestGetAppInstallationDetails_APIError(t *testing.T) {
	resetGlobals()
	setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	}))

	_, err := GetAppInstallationDetails("jwt-token", GitHubRepo{Owner: "owner", Repo: "repo"})
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// initAppToken
// ---------------------------------------------------------------------------

func TestInitAppToken_Success(t *testing.T) {
	resetGlobals()

	pemBytes := generateTestRSAKey(t)

	err := initAppToken("12345", pemBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ghAppSignedToken == "" {
		t.Error("expected non-empty signed token")
	}
}

func TestInitAppToken_InvalidPEM(t *testing.T) {
	resetGlobals()

	err := initAppToken("12345", []byte("invalid"))
	if err == nil {
		t.Fatal("expected error for invalid PEM")
	}
}

func TestInitAppToken_OnlyRunsOnce(t *testing.T) {
	resetGlobals()

	pemBytes := generateTestRSAKey(t)

	err := initAppToken("12345", pemBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	firstToken := ghAppSignedToken

	// Second call should be no-op (sync.Once)
	err = initAppToken("99999", pemBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ghAppSignedToken != firstToken {
		t.Error("expected token to remain unchanged after second call")
	}
}

// ---------------------------------------------------------------------------
// SetGithubAppToken
// ---------------------------------------------------------------------------

func TestSetGithubAppToken_Success(t *testing.T) {
	resetGlobals()

	token := &GitHubAppToken{Token: "test", Repo: GitHubRepo{Owner: "o", Repo: "r"}}

	err := SetGithubAppToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ghAppToken == nil {
		t.Fatal("expected token to be set")
	}

	if ghAppToken.Token != "test" {
		t.Errorf("expected token 'test', got '%s'", ghAppToken.Token)
	}
}

func TestSetGithubAppToken_Nil(t *testing.T) {
	resetGlobals()

	err := SetGithubAppToken(nil)
	if err == nil {
		t.Fatal("expected error for nil token")
	}
}

func TestSetGithubAppToken_OnlyRunsOnce(t *testing.T) {
	resetGlobals()

	token1 := &GitHubAppToken{Token: "first", Repo: GitHubRepo{Owner: "o", Repo: "r"}}
	token2 := &GitHubAppToken{Token: "second", Repo: GitHubRepo{Owner: "o", Repo: "r"}}

	_ = SetGithubAppToken(token1)
	_ = SetGithubAppToken(token2)

	if ghAppToken.Token != "first" {
		t.Errorf("expected 'first' (sync.Once), got '%s'", ghAppToken.Token)
	}
}
