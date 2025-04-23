# ghSurf - GitHub Code Search gRPC Service

A simple gRPC service written in Go that allows searching for code snippets on GitHub based on a search term and an optional user/organization filter. It utilizes the official `go-github` library to interact with the GitHub API v3.

## Features

*   **gRPC Endpoint:** Provides a `Search` RPC method to perform code searches.
*   **Basic Search:** Search for code containing a specific term.
*   **User/Org Filtering:** Optionally limit search results to a specific GitHub user or organization.
*   **Configuration:** Loads configuration (Port, Log Level, GitHub Token) from environment variables or a specified `.env` file.
*   **GitHub Token Required:** Requires a GitHub Personal Access Token (PAT) for API authentication.
*   **Error Handling:** Basic handling for invalid input (empty search term) and GitHub API errors (Rate Limiting, Generic Errors), mapping them to appropriate gRPC status codes.
*   **Unit Tested:** Includes unit tests for configuration loading and gRPC server logic (using mocks for the GitHub client interaction).

## Getting Started

### Prerequisites

*   **Go:** Version 1.23 or later (see `go.mod`).
*   **GitHub Personal Access Token (PAT):** You need a GitHub PAT with permissions to search code (e.g., `public_repo` scope might suffice for public repositories). Create a token here.
*   **(Optional) `protoc`:** The Protocol Buffer Compiler, if you ever need to modify and regenerate the Go code from `proto/ghsurf.proto`.

### Configuration

The service requires configuration, primarily the GitHub token.

1.  **Environment Variables:** Set the following environment variables directly:
    *   `GHSURF_GITHUB_TOKEN` ( **Required**): Your GitHub PAT.
    *   `PORT`: The port for the gRPC server (defaults to `8080`).
    *   `LOG_LEVEL`: Logging level (defaults to `INFO`, not currently used extensively).

2.  **`.env` File:**
    *   Create a `.env` file at a specific location (e.g., `/Users/John/envFiles/.env` as currently hardcoded in `cmd/server/main.go` - **you should modify this path in `main.go` to your preferred location**).
    *   Add the variables to the file.

### Future Improvements

1. API-Level Pagination: The most significant limitation. If the proto definition could be changed, adding `page`, `per_page` to `SearchRequest` and `total_count`, `current_page`, `total_pages` to `SearchResponse` would be the top priority to enable proper pagination.
2. Caching: Implement caching (e.g., using Redis) to store results for common queries, reduce GitHub API calls, avoid rate limits, and improve response times.
3. Enhanced Error Handling: Provide more granular error details or potentially map more GitHub errors to specific gRPC codes.
4. Structured Logging: Implement structured logging (e.g., using `log/slog` or libraries like `Zap/Logrus`) for better log analysis.
5. Observability (Metrics & Tracing): Add Prometheus metrics and distributed tracing (e.g., OpenTelemetry) using gRPC interceptors.
6. gRPC Security: Implement TLS encryption for the gRPC connection. Consider adding authentication/authorization mechanisms if needed.
7. More Search Qualifiers: If the proto could be changed, expose more of GitHub's search qualifiers (language, repository, path, filename, etc.) in the `SearchRequest`.
8. Configuration Flexibility: Make the .env file path configurable via a flag or environment variable instead of hardcoding it in `main.go`.
9. CI/CD Pipeline: Set up automated testing, building, and deployment using GitHub Actions or similar.
10. Containerization: Add a `Dockerfile` to easily build and run the service in a container.
