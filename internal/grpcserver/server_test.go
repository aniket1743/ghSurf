package grpcserver

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"strings"
	"testing"

	pb "github.com/ghSurf/proto"
	"github.com/google/go-github/v71/github"
)

// mockGithubSearcher defines the interface for the part of the GitHub client we need to mock.
// mockGithubSearchService implements the GithubCodeSearcher interface defined in server.go.
type mockGithubSearchService struct {
	mockCodeSearchResult *github.CodeSearchResult
	mockResponse         *github.Response
	mockError            error
	calledWithQuery      string
	calledWithOptions    *github.SearchOptions
}

// Ensure this mock implements the GithubCodeSearcher interface from server.go
var _ GithubCodeSearcher = (*mockGithubSearchService)(nil)

// Code is the mock implementation of the Search.Code method.
func (m *mockGithubSearchService) Code(ctx context.Context, query string, opts *github.SearchOptions) (*github.CodeSearchResult, *github.Response, error) {
	m.calledWithQuery = query
	m.calledWithOptions = opts
	return m.mockCodeSearchResult, m.mockResponse, m.mockError
}

// --- Test Setup Helper ---
func newMockGithubResponse(statusCode int) *github.Response {
	return &github.Response{
		Response: &http.Response{
			StatusCode: statusCode,
		},
	}
}
func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
func boolPtr(b bool) *bool    { return &b }

// --- Test Functions ---
func TestSearch_Success(t *testing.T) {
	mockSearcher := &mockGithubSearchService{
		mockCodeSearchResult: &github.CodeSearchResult{
			Total:             intPtr(1),
			IncompleteResults: boolPtr(false), // <-- Corrected field name
			CodeResults: []*github.CodeResult{
				{
					Name:    strPtr("main.go"),
					Path:    strPtr("cmd/server/main.go"),
					HTMLURL: strPtr("https://github.com/owner/repo/blob/main/cmd/server/main.go"),
					Repository: &github.Repository{
						FullName: strPtr("owner/repo"),
					},
				},
			},
		},
		mockResponse: newMockGithubResponse(http.StatusOK),
		mockError:    nil,
	}

	// Create the GrpcServer instance, injecting the mock searcher directly.
	server := NewGrpcServer(mockSearcher)
	req := &pb.SearchRequest{
		SearchTerm: "test query",
		User:       "",
	}

	// --- Act ---
	resp, err := server.Search(context.Background(), req)

	// --- Assert ---
	if err != nil {
		t.Fatalf("Search() returned an unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("Search() response should not be nil on success")
	}
	expectedQuery := "test query"
	if mockSearcher.calledWithQuery != expectedQuery {
		t.Errorf("Expected mock Search.Code to be called with query '%s', but got '%s'", expectedQuery, mockSearcher.calledWithQuery)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("Expected 1 result, but got %d", len(resp.Results))
	}
	expectedFileURL := "https://github.com/owner/repo/blob/main/cmd/server/main.go"
	expectedRepo := "owner/repo"
	if resp.Results[0].FileUrl != expectedFileURL {
		t.Errorf("Expected FileUrl '%s', but got '%s'", expectedFileURL, resp.Results[0].FileUrl)
	}
	if resp.Results[0].Repo != expectedRepo {
		t.Errorf("Expected Repo '%s', but got '%s'", expectedRepo, resp.Results[0].Repo)
	}
}

// TestSearch_Success_WithUserFilter tests the happy path with a user filter applied.
func TestSearch_Success_WithUserFilter(t *testing.T) {
	// --- Arrange ---
	mockSearcher := &mockGithubSearchService{
		mockCodeSearchResult: &github.CodeSearchResult{
			Total:             intPtr(1),
			IncompleteResults: boolPtr(false),
			CodeResults: []*github.CodeResult{
				{
					Name:    strPtr("client.go"),
					Path:    strPtr("internal/githubclient/client.go"),
					HTMLURL: strPtr("https://github.com/ghSurf/ghSurf/blob/main/internal/githubclient/client.go"),
					Repository: &github.Repository{
						FullName: strPtr("ghSurf/ghSurf"),
					},
				},
			},
		},
		mockResponse: newMockGithubResponse(http.StatusOK),
		mockError:    nil,
	}

	server := NewGrpcServer(mockSearcher)

	// Prepare the request with both search term and user
	req := &pb.SearchRequest{
		SearchTerm: "NewClient",
		User:       "ghSurf", // Add a user filter
	}

	// --- Act ---
	resp, err := server.Search(context.Background(), req)

	// --- Assert ---
	if err != nil {
		t.Fatalf("Search() returned an unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("Search() response should not be nil on success")
	}

	// *** Assert the query passed to the mock includes the user filter ***
	expectedQuery := "NewClient user:ghSurf" // Check for correct format
	if mockSearcher.calledWithQuery != expectedQuery {
		t.Errorf("Expected mock Search.Code to be called with query '%s', but got '%s'", expectedQuery, mockSearcher.calledWithQuery)
	}

	// Check the content of the response (similar to the previous test)
	if len(resp.Results) != 1 {
		t.Fatalf("Expected 1 result, but got %d", len(resp.Results))
	}
	expectedFileURL := "https://github.com/ghSurf/ghSurf/blob/main/internal/githubclient/client.go"
	expectedRepo := "ghSurf/ghSurf"
	if resp.Results[0].FileUrl != expectedFileURL {
		t.Errorf("Expected FileUrl '%s', but got '%s'", expectedFileURL, resp.Results[0].FileUrl)
	}
	if resp.Results[0].Repo != expectedRepo {
		t.Errorf("Expected Repo '%s', but got '%s'", expectedRepo, resp.Results[0].Repo)
	}
}

// TestSearch_Error_EmptySearchTerm tests the case where the search term is empty.
func TestSearch_Error_EmptySearchTerm(t *testing.T) {
	// --- Arrange ---
	// The mock's return values don't matter here, as the function should error out before calling it.
	mockSearcher := &mockGithubSearchService{}

	server := NewGrpcServer(mockSearcher)

	// Prepare the request with an empty search term
	req := &pb.SearchRequest{
		SearchTerm: "", // Empty search term
		User:       "anyuser",
	}

	// --- Act ---
	resp, err := server.Search(context.Background(), req)

	// --- Assert ---

	// 1. Check that the response is nil
	if resp != nil {
		t.Errorf("Expected response to be nil for an empty search term, but got %v", resp)
	}

	// 2. Check that an error was returned
	if err == nil {
		t.Fatal("Expected an error for an empty search term, but got nil")
	}

	// 3. Check the gRPC status code of the error
	st, ok := status.FromError(err)
	if !ok {
		// This should not happen for errors generated by status.Error
		t.Fatalf("Expected error to be a gRPC status error, but it wasn't: %v", err)
	}

	// 4. Assert the code is InvalidArgument
	expectedCode := codes.InvalidArgument
	if st.Code() != expectedCode {
		t.Errorf("Expected gRPC status code %v, but got %v", expectedCode, st.Code())
	}

	// 5. Optional: Assert that the mock was NOT called
	if mockSearcher.calledWithQuery != "" {
		t.Errorf("Expected mock Search.Code NOT to be called, but it was called with query '%s'", mockSearcher.calledWithQuery)
	}
}

// TestSearch_Error_GithubRateLimit tests handling of a GitHub rate limit error.
func TestSearch_Error_GithubRateLimit(t *testing.T) {
	// --- Arrange ---

	// 1. Create a mock HTTP response simulating a rate limit (403 Forbidden)
	// Although the RateLimitError struct doesn't strictly *require* the response,
	// the go-github library typically includes it, and our server code logs it.
	mockHttpResponse := &http.Response{
		StatusCode: http.StatusForbidden, // 403 is common for rate limits
		Header:     make(http.Header),
	}
	// Add rate limit headers for realism (optional but good)
	mockHttpResponse.Header.Set("X-RateLimit-Limit", "60")
	mockHttpResponse.Header.Set("X-RateLimit-Remaining", "0")
	mockHttpResponse.Header.Set("X-RateLimit-Reset", "1678886400")

	mockGithubResp := &github.Response{Response: mockHttpResponse}

	// 2. Create the specific RateLimitError
	rateLimitErr := &github.RateLimitError{
		Rate:     github.Rate{Limit: 60, Remaining: 0, Reset: github.Timestamp{}}, // Populate Rate struct
		Response: mockHttpResponse,                                                // Include the response
		Message:  "API rate limit exceeded for user ID 12345.",                    // Example message
	}

	// 3. Configure the mock searcher to return this error
	mockSearcher := &mockGithubSearchService{
		mockCodeSearchResult: nil,            // No result on error
		mockResponse:         mockGithubResp, // The response associated with the error
		mockError:            rateLimitErr,   // The specific error to return
	}

	server := NewGrpcServer(mockSearcher)

	// 4. Prepare a valid request (the error happens during the API call)
	req := &pb.SearchRequest{
		SearchTerm: "valid term",
		User:       "validuser",
	}

	// --- Act ---
	resp, err := server.Search(context.Background(), req)

	// --- Assert ---

	// 1. Check that the response is nil
	if resp != nil {
		t.Errorf("Expected response to be nil on rate limit error, but got %v", resp)
	}

	// 2. Check that an error was returned
	if err == nil {
		t.Fatal("Expected an error on rate limit error, but got nil")
	}

	// 3. Check the gRPC status code of the error
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("Expected error to be a gRPC status error, but it wasn't: %v", err)
	}

	// 4. Assert the code is ResourceExhausted
	expectedCode := codes.ResourceExhausted
	if st.Code() != expectedCode {
		t.Errorf("Expected gRPC status code %v for rate limit, but got %v", expectedCode, st.Code())
	}

	// 5. Assert that the mock *was* called (error happened after the call)
	expectedQuery := "valid term user:validuser"
	if mockSearcher.calledWithQuery != expectedQuery {
		t.Errorf("Expected mock Search.Code to be called with query '%s', but got '%s'", expectedQuery, mockSearcher.calledWithQuery)
	}

	// 6. Optional: Check the error message contains expected text
	if !strings.Contains(st.Message(), "GitHub API rate limit exceeded") {
		t.Errorf("Expected error message to contain 'GitHub API rate limit exceeded', but got: %s", st.Message())
	}
}

// TestSearch_Error_GithubGeneric tests handling of a generic error from the GitHub client.
func TestSearch_Error_GithubGeneric(t *testing.T) {
	// --- Arrange ---

	// 1. Define the generic error to be returned by the mock
	genericErr := errors.New("simulated generic GitHub API error")

	// 2. Optionally, create a mock response (e.g., 500) associated with the error
	mockHttpResponse := &http.Response{
		StatusCode: http.StatusInternalServerError, // 500 Internal Server Error
		Header:     make(http.Header),
	}
	mockGithubResp := &github.Response{Response: mockHttpResponse}

	// 3. Configure the mock searcher to return the generic error
	mockSearcher := &mockGithubSearchService{
		mockCodeSearchResult: nil,            // No result on error
		mockResponse:         mockGithubResp, // Response associated with the error
		mockError:            genericErr,     // The generic error
	}

	server := NewGrpcServer(mockSearcher)

	// 4. Prepare a valid request
	req := &pb.SearchRequest{
		SearchTerm: "another valid term",
		User:       "",
	}

	// --- Act ---
	resp, err := server.Search(context.Background(), req)

	// --- Assert ---

	// 1. Check that the response is nil
	if resp != nil {
		t.Errorf("Expected response to be nil on generic error, but got %v", resp)
	}

	// 2. Check that an error was returned
	if err == nil {
		t.Fatal("Expected an error on generic error, but got nil")
	}

	// 3. Check the gRPC status code of the error
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("Expected error to be a gRPC status error, but it wasn't: %v", err)
	}

	// 4. Assert the code is Internal
	expectedCode := codes.Internal // Expect Internal for unhandled GitHub errors
	if st.Code() != expectedCode {
		t.Errorf("Expected gRPC status code %v for generic error, but got %v", expectedCode, st.Code())
	}

	// 5. Assert that the mock *was* called
	expectedQuery := "another valid term"
	if mockSearcher.calledWithQuery != expectedQuery {
		t.Errorf("Expected mock Search.Code to be called with query '%s', but got '%s'", expectedQuery, mockSearcher.calledWithQuery)
	}

	// 6. Optional: Check the error message contains the original error text
	if !strings.Contains(st.Message(), genericErr.Error()) {
		t.Errorf("Expected error message to contain '%s', but got: %s", genericErr.Error(), st.Message())
	}
}
