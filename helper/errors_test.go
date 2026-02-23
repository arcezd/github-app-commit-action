package github_helper

import (
	"fmt"
	"testing"
)

func TestGitHubAPIError_Error(t *testing.T) {
	err := &GitHubAPIError{StatusCode: 404, Body: "not found"}
	expected := "error calling github api, status code: 404, response: not found"

	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestIsNotFound_404(t *testing.T) {
	err := &GitHubAPIError{StatusCode: 404, Body: "not found"}
	if !IsNotFound(err) {
		t.Error("expected IsNotFound to return true for 404")
	}
}

func TestIsNotFound_Non404(t *testing.T) {
	err := &GitHubAPIError{StatusCode: 403, Body: "forbidden"}
	if IsNotFound(err) {
		t.Error("expected IsNotFound to return false for 403")
	}
}

func TestIsNotFound_NonAPIError(t *testing.T) {
	err := fmt.Errorf("some other error")
	if IsNotFound(err) {
		t.Error("expected IsNotFound to return false for non-API error")
	}
}

func TestIsNotFound_WrappedError(t *testing.T) {
	apiErr := &GitHubAPIError{StatusCode: 404, Body: "not found"}
	wrapped := fmt.Errorf("wrapped: %w", apiErr)

	if !IsNotFound(wrapped) {
		t.Error("expected IsNotFound to return true for wrapped 404 error")
	}
}

func TestIsNotFound_Nil(t *testing.T) {
	if IsNotFound(nil) {
		t.Error("expected IsNotFound to return false for nil")
	}
}
