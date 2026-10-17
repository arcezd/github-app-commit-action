//nolint:goconst
package github_helper

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// resetGlobals resets all package-level state so tests are isolated.
func resetGlobals() {
	initJwt = sync.Once{}
	initToken = sync.Once{}
	ghAppSignedToken = ""
	ghAppToken = nil
	initJwtErr = nil
	githubBaseURL = "https://api.github.com"
}

// setupTestServer creates an httptest.Server, points githubBaseURL at it, and
// registers cleanup to restore the original URL.
func setupTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()

	ts := httptest.NewServer(handler)
	githubBaseURL = ts.URL

	t.Cleanup(func() {
		ts.Close()

		githubBaseURL = "https://api.github.com"
	})

	return ts
}

// generateTestRSAKey generates a 2048-bit RSA private key PEM for testing.
func generateTestRSAKey(t *testing.T) []byte {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	return pemBytes
}

// apiRouter maps "METHOD /path" strings to handlers.
func apiRouter(routes map[string]http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		if h, ok := routes[key]; ok {
			h(w, r)
			return
		}

		// try without trailing slash
		path := strings.TrimRight(r.URL.Path, "/")

		key = fmt.Sprintf("%s %s", r.Method, path)
		if h, ok := routes[key]; ok {
			h(w, r)
			return
		}

		http.NotFound(w, r)
	}
}

// mockExecCommand swaps execCommandFunc for the duration of the test.
func mockExecCommand(t *testing.T, fn func(command string, args []string) ([]byte, error)) {
	t.Helper()

	original := execCommandFunc
	execCommandFunc = fn

	t.Cleanup(func() {
		execCommandFunc = original
	})
}

// jsonResponse writes a JSON response with the given status code.
func jsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

// setTestAppToken sets ghAppToken for tests that need it.
func setTestAppToken(t *testing.T, token *GitHubAppToken) {
	t.Helper()

	ghAppToken = token

	t.Cleanup(func() {
		ghAppToken = nil
	})
}
