package main

import "testing"

func TestBuildVersion_NonEmpty(t *testing.T) {
	if BuildVersion == "" {
		t.Error("expected BuildVersion to be non-empty")
	}
}
