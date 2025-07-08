package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type RequestBody struct {
	RepoURL string `json:"repo_url"`
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

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return ""
	}
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
