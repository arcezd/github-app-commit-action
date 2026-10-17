package github_helper

import (
	"fmt"
	"testing"
)

const notFoundBody = "not found"

func TestGitHubAPIError_Error(t *testing.T) {
	err := &GitHubAPIError{StatusCode: 404, Body: notFoundBody}
	expected := "error calling github api, status code: 404, response: not found"

	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestIsNotFound_404(t *testing.T) {
	err := &GitHubAPIError{StatusCode: 404, Body: notFoundBody}
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
	apiErr := &GitHubAPIError{StatusCode: 404, Body: notFoundBody}
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

func TestIsReferenceNotFound(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"404", &GitHubAPIError{StatusCode: 404, Body: `{"message":"Not Found"}`}, true},
		{"422 missing ref", &GitHubAPIError{StatusCode: 422, Body: `{"message":"Reference does not exist"}`}, true},
		{"422 not fast forward", &GitHubAPIError{StatusCode: 422, Body: `{"message":"Update is not a fast forward"}`}, false},
		{"403", &GitHubAPIError{StatusCode: 403, Body: `{"message":"Forbidden"}`}, false},
		{"non-API error", fmt.Errorf("boom"), false},
		{"nil", nil, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsReferenceNotFound(tc.err); got != tc.want {
				t.Errorf("IsReferenceNotFound() = %v, want %v", got, tc.want)
			}
		})
	}
}
