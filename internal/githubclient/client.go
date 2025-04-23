package githubclient

import (
	"context"

	"golang.org/x/oauth2" // Used for GitHub token authentication
	// Use a recent version of go-github, e.g., v62. Check for the latest stable version if needed.
	"github.com/google/go-github/v71/github"
)

// New creates and returns a new GitHub API client authenticated using the provided personal access token (PAT).
func New(token string) *github.Client {
	// Create a background context. In real handlers, you'd use the request's context.
	ctx := context.Background()

	// Create a token source using the static token provided.
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	// Create an authenticated HTTP client using the token source.
	tc := oauth2.NewClient(ctx, ts)

	// Create the GitHub API client using the authenticated HTTP client.
	client := github.NewClient(tc)

	return client
}
