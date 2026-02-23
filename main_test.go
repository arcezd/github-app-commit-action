package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"strings"
	"testing"
)

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

func TestRun_HelpFlag(t *testing.T) {
	err := run([]string{"-help"}, func(string) string { return "" })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRun_VersionFlag(t *testing.T) {
	err := run([]string{"-version"}, func(string) string { return "" })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRun_MissingRepository(t *testing.T) {
	err := run([]string{}, func(string) string { return "" })
	if err == nil {
		t.Fatal("expected error for missing repository")
	}
	if !strings.Contains(err.Error(), "repository flag is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_InvalidRepoFormat(t *testing.T) {
	err := run([]string{"-r", "invalid"}, func(string) string { return "" })
	if err == nil {
		t.Fatal("expected error for invalid repo format")
	}
	if !strings.Contains(err.Error(), "invalid repository format") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_ValidRepoFormat_MissingAppId(t *testing.T) {
	err := run([]string{"-r", "owner/repo"}, func(string) string { return "" })
	if err == nil {
		t.Fatal("expected error for missing app id")
	}
	if !strings.Contains(err.Error(), "GitHub app id is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_NoPrivateKey(t *testing.T) {
	err := run([]string{"-r", "owner/repo", "-i", "12345"}, func(string) string { return "" })
	if err == nil {
		t.Fatal("expected error for no private key")
	}
	if !strings.Contains(err.Error(), "provide a private key") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_PEMFromEnvVar(t *testing.T) {
	pemBytes := generateTestRSAKey(t)
	// This will fail at GenerateInstallationAppToken since there's no server,
	// but it validates PEM parsing succeeds
	err := run([]string{"-r", "owner/repo", "-i", "12345"}, func(key string) string {
		if key == "GH_APP_PRIVATE_KEY" {
			return string(pemBytes)
		}
		return ""
	})
	// Should fail later (at API call), not at PEM parsing
	if err == nil {
		t.Fatal("expected error (API call should fail)")
	}
	if strings.Contains(err.Error(), "failed to decode PEM") {
		t.Errorf("PEM parsing should have succeeded, got: %v", err)
	}
}

func TestRun_Base64EncodedPEM(t *testing.T) {
	pemBytes := generateTestRSAKey(t)
	encoded := base64.StdEncoding.EncodeToString(pemBytes)

	err := run([]string{"-r", "owner/repo", "-i", "12345"}, func(key string) string {
		if key == "GH_APP_PRIVATE_KEY" {
			return encoded
		}
		return ""
	})
	// Should fail later (at API call), not at PEM/base64 parsing
	if err == nil {
		t.Fatal("expected error (API call should fail)")
	}
	if strings.Contains(err.Error(), "failed to decode PEM") {
		t.Errorf("base64 PEM parsing should have succeeded, got: %v", err)
	}
}

func TestRun_InvalidPEM(t *testing.T) {
	err := run([]string{"-r", "owner/repo", "-i", "12345"}, func(key string) string {
		if key == "GH_APP_PRIVATE_KEY" {
			return "not-a-valid-pem"
		}
		return ""
	})
	if err == nil {
		t.Fatal("expected error for invalid PEM")
	}
	if !strings.Contains(err.Error(), "failed to decode PEM") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_CoauthorParsing_Valid(t *testing.T) {
	pemBytes := generateTestRSAKey(t)

	err := run([]string{"-r", "owner/repo", "-i", "12345", "-c", "Alice <alice@test.com>"}, func(key string) string {
		if key == "GH_APP_PRIVATE_KEY" {
			return string(pemBytes)
		}
		return ""
	})
	// Should fail at API call, not at coauthor parsing
	if err != nil && strings.Contains(err.Error(), "invalid coauthor format") {
		t.Errorf("coauthor parsing should have succeeded, got: %v", err)
	}
}

func TestRun_CoauthorParsing_Invalid(t *testing.T) {
	pemBytes := generateTestRSAKey(t)

	err := run([]string{"-r", "owner/repo", "-i", "12345", "-c", "invalid-format"}, func(key string) string {
		if key == "GH_APP_PRIVATE_KEY" {
			return string(pemBytes)
		}
		return ""
	})
	if err == nil {
		t.Fatal("expected error for invalid coauthor")
	}
	if !strings.Contains(err.Error(), "invalid coauthor format") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_DefaultCommitMessage(t *testing.T) {
	pemBytes := generateTestRSAKey(t)

	err := run([]string{"-r", "owner/repo", "-i", "12345"}, func(key string) string {
		if key == "GH_APP_PRIVATE_KEY" {
			return string(pemBytes)
		}
		return ""
	})
	// Just verify it doesn't fail at message parsing
	if err != nil && strings.Contains(err.Error(), "commit message") {
		t.Errorf("unexpected commit message error: %v", err)
	}
}

func TestRun_PEMFromFile_EmptyFilename(t *testing.T) {
	// -p flag with empty value + no env var -> should hit "provide a private key" error
	err := run([]string{"-r", "owner/repo", "-i", "12345", "-p", ""}, func(string) string { return "" })
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "provide a private key") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_PEMFromFile_MissingFile(t *testing.T) {
	err := run([]string{"-r", "owner/repo", "-i", "12345", "-p", "/nonexistent/key.pem"}, func(string) string { return "" })
	if err == nil {
		t.Fatal("expected error for missing PEM file")
	}
	if !strings.Contains(err.Error(), "failed to sign JWT token from file") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_PEMFromFile_ValidFile(t *testing.T) {
	pemBytes := generateTestRSAKey(t)
	tmpFile, err := createTempPEMFile(t, pemBytes)
	if err != nil {
		t.Fatal(err)
	}

	err = run([]string{"-r", "owner/repo", "-i", "12345", "-p", tmpFile}, func(string) string { return "" })
	// Should fail at API call, not at PEM file reading
	if err != nil && strings.Contains(err.Error(), "failed to sign JWT token from file") {
		t.Errorf("PEM file should have been read successfully, got: %v", err)
	}
}

func TestRun_HeadBranchFlag(t *testing.T) {
	pemBytes := generateTestRSAKey(t)

	err := run([]string{"-r", "owner/repo", "-i", "12345", "-h", "develop", "-b", "main"}, func(key string) string {
		if key == "GH_APP_PRIVATE_KEY" {
			return string(pemBytes)
		}
		return ""
	})
	// Should fail at API level, not at flag parsing
	if err != nil && strings.Contains(err.Error(), "head branch") {
		t.Errorf("unexpected head branch error: %v", err)
	}
}

func TestRun_CustomCommitMessage(t *testing.T) {
	pemBytes := generateTestRSAKey(t)

	err := run([]string{"-r", "owner/repo", "-i", "12345", "-m", "custom message"}, func(key string) string {
		if key == "GH_APP_PRIVATE_KEY" {
			return string(pemBytes)
		}
		return ""
	})
	// Should fail at API level, not at message parsing
	if err != nil && strings.Contains(err.Error(), "commit message") {
		t.Errorf("unexpected error: %v", err)
	}
}

func createTempPEMFile(t *testing.T, pemBytes []byte) (string, error) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "test-pem-*")
	if err != nil {
		return "", err
	}
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })
	tmpFile.Write(pemBytes)
	tmpFile.Close()
	return tmpFile.Name(), nil
}
