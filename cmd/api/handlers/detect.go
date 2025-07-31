package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	_ "mira/cmd/api/models"

	"github.com/gofiber/fiber/v2"
)

// RequestBody represents the request body for framework detection
type RequestBody struct {
	RepoURL string `json:"repo_url" example:"https://github.com/user/repo" validate:"required" doc:"GitHub repository URL"`
}

type GitHubContentResponse struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

var detectionMap = map[string]interface{}{
	"package.json": map[string]string{
		"react":         "React",
		"next":          "Next.js",
		"vue":           "Vue.js",
		"nuxt":          "Nuxt.js",
		"svelte":        "Svelte",
		"@angular/core": "Angular",
		"vite":          "Vite",
		"webpack":       "Webpack",
		"parcel":        "Parcel",
	},
	"vite.config.js":    "Vite",
	"webpack.config.js": "Webpack",
	"angular.json":      "Angular",
	"svelte.config.js":  "Svelte",
	"nuxt.config.js":    "Nuxt.js",
	"next.config.js":    "Next.js",
}

// DetectFramework analyzes a GitHub repository to detect frameworks and technologies
// @Summary Detect framework from repository
// @Description Analyzes package.json and other configuration files to detect the technology stack
// @Tags images
// @Accept json
// @Produce json
// @Param request body RequestBody true "Repository URL"
// @Success 200 {object} models.FrameworkDetectionResponse "Detected frameworks and technologies"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /images/detect [post]
func DetectFramework(c *fiber.Ctx) error {
	fmt.Println("Got Request")
	var req RequestBody
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	owner, repo := parseGitHubURL(req.RepoURL)
	if owner == "" || repo == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid GitHub URL"})
	}

	detected := make(map[string]bool)

	for file, indicators := range detectionMap {
		content := fetchFileContent(owner, repo, file)
		if content == "" {
			continue
		}

		switch v := indicators.(type) {
		case string:
			detected[v] = true
		case map[string]string:
			for keyword, framework := range v {
				if strings.Contains(strings.ToLower(content), strings.ToLower(keyword)) {
					detected[framework] = true
				}
			}
		}
	}

	result := []string{}
	for fw := range detected {
		result = append(result, fw)
	}

	return c.JSON(fiber.Map{"detected": result})
}

// Helper to parse GitHub repo URL
func parseGitHubURL(url string) (string, string) {
	fmt.Println("Pursing url")
	re := regexp.MustCompile(`https://github.com/([^/]+)/([^/]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) == 3 {
		return matches[1], matches[2]
	}
	return "", ""
}

// Fetch file content from GitHub using GitHub API
func fetchFileContent(owner, repo, filepath string) string {
	fmt.Println("fetching file content")
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, filepath)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return ""
	}
	fmt.Println("Response:", resp)
	fmt.Println("Error:", err)
	defer resp.Body.Close()

	var data GitHubContentResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return ""
	}

	if data.Encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(data.Content)
		if err != nil {
			return ""
		}
		return string(decoded)
	}

	return ""
}
