package models

// BuildResponse represents the response when starting a containerization build
type BuildResponse struct {
	Message string    `json:"message" example:"Image generation started"`
	Data    BuildData `json:"data"`
}

// BuildData contains build-specific information
type BuildData struct {
	Name    string `json:"name" example:"my-app"`
	BuildID string `json:"build_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	LogsURL string `json:"logs_url" example:"ws://localhost:3000/api/logs/550e8400-e29b-41d4-a716-446655440000"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request format"`
}

// FrameworkDetectionResponse represents the response from framework detection
type FrameworkDetectionResponse struct {
	Message    string                 `json:"message" example:"Framework detection completed"`
	Frameworks map[string]interface{} `json:"frameworks"`
}

// GitHubRepository represents a GitHub repository
type GitHubRepository struct {
	ID       int    `json:"id" example:"12345"`
	Name     string `json:"name" example:"my-repo"`
	FullName string `json:"full_name" example:"user/my-repo"`
	Private  bool   `json:"private" example:"false"`
	HTMLURL  string `json:"html_url" example:"https://github.com/user/my-repo"`
	CloneURL string `json:"clone_url" example:"https://github.com/user/my-repo.git"`
}

// GitLabProject represents a GitLab project
type GitLabProject struct {
	ID                int    `json:"id" example:"67890"`
	Name              string `json:"name" example:"my-project"`
	NameWithNamespace string `json:"name_with_namespace" example:"user/my-project"`
	Visibility        string `json:"visibility" example:"private"`
	WebURL            string `json:"web_url" example:"https://gitlab.com/user/my-project"`
	HTTPURLToRepo     string `json:"http_url_to_repo" example:"https://gitlab.com/user/my-project.git"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status" example:"OK"`
}
