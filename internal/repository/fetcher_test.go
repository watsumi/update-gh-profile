package repository

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v76/github"
)

// TestFetchUserRepositories_InvalidUsername verifies that an error is returned for an invalid username
func TestFetchUserRepositories_InvalidUsername(t *testing.T) {
	ctx := context.Background()
	client := github.NewClient(nil) // Dummy client

	// Call with invalid username (empty string)
	_, err := FetchUserRepositories(ctx, client, "", true, true)
	if err == nil {
		t.Errorf("Expected error for empty username, but got nil")
	}
	expectedError := "username is empty"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', but got '%s'", expectedError, err.Error())
	}
}

// TestHandleRateLimit_NoRateLimit verifies handling when there is no rate limit
func TestHandleRateLimit_NoRateLimit(t *testing.T) {
	ctx := context.Background()
	resp := &github.Response{
		Response: &http.Response{
			StatusCode: 200,
		},
		Rate: github.Rate{
			Limit:     5000,
			Remaining: 4999,
			Reset:     github.Timestamp{Time: time.Now().Add(1 * time.Hour)},
		},
	}

	err := HandleRateLimit(ctx, resp)
	if err != nil {
		t.Errorf("Expected no error when rate limit is not hit, but got: %v", err)
	}
}

// TestValidateOwnerAndRepo verifies validation of owner and repo
func TestValidateOwnerAndRepo(t *testing.T) {
	tests := []struct {
		name    string
		owner   string
		repo    string
		wantErr bool
	}{
		{"valid", "owner", "repo", false},
		{"empty owner", "", "repo", true},
		{"empty repo", "owner", "", true},
		{"both empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOwnerAndRepo(tt.owner, tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOwnerAndRepo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateUsername verifies validation of username
func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"valid", "username", false},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDetectLanguageFromFilename verifies language detection from filename
func TestDetectLanguageFromFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{"Go file", "main.go", "Go"},
		{"Python file", "script.py", "Python"},
		{"TypeScript file", "app.ts", "TypeScript"},
		{"JavaScript file", "app.js", "JavaScript"},
		{"Dockerfile", "Dockerfile", "Dockerfile"},
		{"Makefile", "Makefile", "Makefile"},
		{"unknown extension", "file.unknown", ""},
		{"no extension", "file", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectLanguageFromFilename(tt.filename)
			if got != tt.want {
				t.Errorf("DetectLanguageFromFilename(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

// TestCheckPagination verifies pagination determination
func TestCheckPagination(t *testing.T) {
	tests := []struct {
		name         string
		nextPage     int
		currentCount int
		perPage      int
		wantHasNext  bool
	}{
		{"has next page from header", 2, 100, 100, true},
		{"no next page, count < perPage", 0, 50, 100, false},
		{"no next page, count = perPage", 0, 100, 100, true},
		{"no next page, count = default", 0, 30, 100, true},
		{"no next page, count = 0", 0, 0, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &github.Response{
				NextPage: tt.nextPage,
			}
			result := CheckPagination(resp, tt.currentCount, tt.perPage)
			if result.HasNextPage != tt.wantHasNext {
				t.Errorf("CheckPagination() HasNextPage = %v, want %v", result.HasNextPage, tt.wantHasNext)
			}
			if tt.nextPage != 0 && result.NextPageNum != tt.nextPage {
				t.Errorf("CheckPagination() NextPageNum = %d, want %d", result.NextPageNum, tt.nextPage)
			}
		})
	}
}
