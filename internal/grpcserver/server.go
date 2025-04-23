package grpcserver

import (
	"context"
	"fmt"
	"log"

	pb "github.com/ghSurf/proto"
	"github.com/google/go-github/v71/github"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GithubCodeSearcher defines the interface for searching code on GitHub.
// This allows for mocking in tests.
type GithubCodeSearcher interface {
	Code(ctx context.Context, query string, opts *github.SearchOptions) (*github.CodeSearchResult, *github.Response, error)
}

// GrpcServer implements the generated GithubSearchServiceServer interface.
type GrpcServer struct {
	pb.UnimplementedGithubSearchServiceServer
	// Use the interface instead of the concrete client
	ghSearcher GithubCodeSearcher
}

// NewGrpcServer creates a new instance of GrpcServer.
func NewGrpcServer(searcher GithubCodeSearcher) *GrpcServer {
	if searcher == nil {
		log.Fatal("GitHub searcher cannot be nil in NewGrpcServer")
	}

	return &GrpcServer{
		ghSearcher: searcher, // Store the interface
	}
}

// Search implements the Search RPC method defined in the proto file.
func (s *GrpcServer) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	log.Printf("Received Search request: term='%s', user='%s'", req.SearchTerm, req.User)

	// --- 1. Validate Input ---
	if req.SearchTerm == "" {
		log.Println("Error: Search term is empty")
		return nil, status.Error(codes.InvalidArgument, "search_term cannot be empty")
	}

	// --- 2. Construct GitHub API Query ---
	query := req.SearchTerm
	if req.User != "" {
		query = fmt.Sprintf("%s user:%s", query, req.User)
	}
	log.Printf("Constructed GitHub code search query: '%s'", query)

	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{},
	}

	// --- 3. Call GitHub API (Use the interface field) ---
	// Use s.ghSearcher instead of s.ghClient.Search
	results, resp, err := s.ghSearcher.Code(ctx, query, opts) // <--- Use the interface field
	if err != nil {
		log.Printf("Error searching GitHub code: %v", err)
		if _, ok := err.(*github.RateLimitError); ok {
			log.Println("GitHub rate limit exceeded")
			return nil, status.Errorf(codes.ResourceExhausted, "GitHub API rate limit exceeded. Response: %v", resp)
		}
		if _, ok := err.(*github.AbuseRateLimitError); ok {
			log.Println("GitHub abuse detection triggered")
			return nil, status.Errorf(codes.ResourceExhausted, "GitHub API abuse detection triggered. Response: %v", resp)
		}
		return nil, status.Errorf(codes.Internal, "failed to search code on GitHub: %v", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("GitHub API returned non-success status code: %d", resp.StatusCode)
		return nil, status.Errorf(codes.Internal, "GitHub API error: status code %d", resp.StatusCode)
	}

	log.Printf("GitHub code search successful. Found %d total results. Processing %d results from this page.", safeTotal(results.Total), len(results.CodeResults)) // Use safeTotal helper

	// --- 4. Process Results and Map ---
	responseResults := make([]*pb.Result, 0, len(results.CodeResults))
	for _, codeResult := range results.CodeResults {
		if codeResult == nil {
			continue
		}
		fileURL := safeString(codeResult.HTMLURL)
		repoName := safeRepoFullName(codeResult.Repository)

		if fileURL != "" && repoName != "" {
			responseResults = append(responseResults, &pb.Result{
				FileUrl: fileURL,
				Repo:    repoName,
			})
		} else {
			log.Printf("Skipping result due to missing data: FileURL='%s', RepoName='%s', Path='%s'", fileURL, repoName, safeString(codeResult.Path))
		}
	}

	response := &pb.SearchResponse{
		Results: responseResults,
	}

	// --- 5. Return Response ---
	log.Printf("Returning %d processed results.", len(responseResults))
	return response, nil
}

// --- Helper functions to safely dereference pointers ---
func safeString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
func safeRepoFullName(repo *github.Repository) string {
	if repo == nil || repo.FullName == nil {
		return ""
	}
	return *repo.FullName
}

// Helper function to safely get total results count
func safeTotal(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}
