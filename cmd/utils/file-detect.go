package utils

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"mira/cmd/config"
)

const (
	// Git clone settings
	cloneTimeoutSeconds = 60
	maxFileSizeBytes    = 1024 * 1024 // 1MB
)

// BuildInfo contains build command and directory information for frameworks
type BuildInfo struct {
	Command string `json:"command"`
	Dir     string `json:"dir"`
}

// Framework build information mapping
var FrameworkBuildInfo = map[string]BuildInfo{
	"react":            {"build", "build"},
	"create-react-app": {"build", "build"},
	"next.js":          {"build", ".next"},
	"vue.js":           {"build", "dist"},
	"angular":          {"ng build", "dist"},
	"nuxt.js":          {"build", ".nuxt/dist"},
	"svelte":           {"build", "build"},
	"sveltekit":        {"build", "build"},
	"gatsby":           {"gatsby build", "public"},
	"remix":            {"build", "build"},
	"solidjs":          {"build", "dist"},
	"preact":           {"build", "build"},
	"vite":             {"build", "dist"},
	"webpack":          {"build", "dist"},
	"rollup":           {"rollup -c", "dist"},
	"parcel":           {"build", "dist"},
	"snowpack":         {"build", "build"},
}

// FrameworkInfo contains detailed information about detected frameworks
type FrameworkInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version,omitempty"`
	Confidence  string `json:"confidence"` // "high", "medium", "low"
	DetectedIn  string `json:"detected_in"`
	Description string `json:"description,omitempty"`
}

// FileDetectionRule represents a rule for detecting frameworks in files
type FileDetectionRule struct {
	FilePattern    string            `json:"file_pattern"`
	Dependencies   map[string]string `json:"dependencies,omitempty"`
	ScriptCommands map[string]string `json:"script_commands,omitempty"`
	ContentChecks  []string          `json:"content_checks,omitempty"`
	Framework      string            `json:"framework"`
	Confidence     string            `json:"confidence"`
	Description    string            `json:"description"`
	BuildCommand   string            `json:"build_command,omitempty"`
	BuildDir       string            `json:"build_dir,omitempty"`
}

// JavaScript framework detection rules
var JSFrameworkRules = []FileDetectionRule{
	// Package.json based detections for JavaScript frameworks
	{
		FilePattern: "package.json",
		Dependencies: map[string]string{
			"react":               "react",
			"@types/react":        "react",
			"react-dom":           "react",
			"next":                "next.js",
			"vue":                 "vue.js",
			"@vue/cli":            "vue.js",
			"nuxt":                "nuxt.js",
			"svelte":              "svelte",
			"@angular/core":       "angular",
			"@angular/cli":        "angular",
			"gatsby":              "gatsby",
			"@remix-run/node":     "remix",
			"@remix-run/react":    "remix",
			"solid-js":            "solidjs",
			"@solidjs/router":     "solidjs",
			"preact":              "preact",
			"@preact/preset-vite": "preact",
		},
		ScriptCommands: map[string]string{
			"vite":               "vite",
			"vite build":         "vite",
			"vite dev":           "vite",
			"vite preview":       "vite",
			"webpack":            "webpack",
			"webpack-dev-server": "webpack",
			"ng ":                "angular",
			"ng build":           "angular",
			"ng serve":           "angular",
			"react-scripts":      "create-react-app",
			"gatsby develop":     "gatsby",
			"gatsby build":       "gatsby",
			"next build":         "next.js",
			"next dev":           "next.js",
			"next start":         "next.js",
			"nuxt build":         "nuxt.js",
			"nuxt dev":           "nuxt.js",
			"nuxt generate":      "nuxt.js",
			"svelte-kit":         "sveltekit",
			"remix build":        "remix",
			"remix dev":          "remix",
			"parcel":             "parcel",
			"rollup":             "rollup",
		},
		Confidence:  "high",
		Description: "Major JavaScript framework detected from package.json dependencies and scripts",
	},
	// Configuration file based detections
	{
		FilePattern: "vite.config.js",
		Framework:   "vite",
		Confidence:  "high",
		Description: "Vite build tool configuration present",
	},
	{
		FilePattern: "vite.config.ts",
		Framework:   "vite",
		Confidence:  "high",
		Description: "Vite TypeScript configuration present",
	},
	{
		FilePattern: "webpack.config.js",
		Framework:   "webpack",
		Confidence:  "medium",
		Description: "Webpack bundler configuration present",
	},
	{
		FilePattern: "angular.json",
		Framework:   "angular",
		Confidence:  "high",
		Description: "Angular workspace configuration present",
	},
	{
		FilePattern: "svelte.config.js",
		Framework:   "svelte",
		Confidence:  "high",
		Description: "Svelte configuration file present",
	},
	{
		FilePattern: "nuxt.config.js",
		Framework:   "nuxt.js",
		Confidence:  "high",
		Description: "Nuxt.js configuration file present",
	},
	{
		FilePattern: "nuxt.config.ts",
		Framework:   "nuxt.js",
		Confidence:  "high",
		Description: "Nuxt.js TypeScript configuration present",
	},
	{
		FilePattern: "next.config.js",
		Framework:   "next.js",
		Confidence:  "high",
		Description: "Next.js configuration file present",
	},
	{
		FilePattern: "gatsby-config.js",
		Framework:   "gatsby",
		Confidence:  "high",
		Description: "Gatsby configuration file present",
	},
	{
		FilePattern: "vue.config.js",
		Framework:   "vue.js",
		Confidence:  "high",
		Description: "Vue.js configuration file present",
	},
	{
		FilePattern: "remix.config.js",
		Framework:   "remix",
		Confidence:  "high",
		Description: "Remix configuration file present",
	},
	{
		FilePattern: "solid.config.js",
		Framework:   "solidjs",
		Confidence:  "high",
		Description: "SolidJS configuration file present",
	},
	// Additional build tools and frameworks
	{
		FilePattern: "rollup.config.js",
		Framework:   "rollup",
		Confidence:  "high",
		Description: "Rollup bundler configuration present",
	},
	{
		FilePattern: "snowpack.config.js",
		Framework:   "snowpack",
		Confidence:  "high",
		Description: "Snowpack build tool configuration present",
	},
}

// CloneRepository clones a GitHub repository to a directory
func CloneRepository(ctx context.Context, owner, repo string) (string, func(), error) {
	// Create repository directory
	repoPath := filepath.Join(config.GIT_REPOS_DIR, fmt.Sprintf("%s-%s", owner, repo))

	// Check if directory already exists, if so remove it first
	if _, err := os.Stat(repoPath); err == nil {
		if err := os.RemoveAll(repoPath); err != nil {
			return "", nil, fmt.Errorf("failed to remove existing directory: %v", err)
		}
	}

	err := os.MkdirAll(repoPath, os.ModePerm)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create repository directory: %v", err)
	}

	// Cleanup function
	cleanup := func() {
		if err := os.RemoveAll(repoPath); err != nil {
			log.Printf("Warning: failed to cleanup repository directory %s: %v", repoPath, err)
		}
	}

	// Construct clone URL
	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)

	// Create git clone command with timeout
	cloneCtx, cancel := context.WithTimeout(ctx, cloneTimeoutSeconds*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cloneCtx, "git", "clone", "--depth", "1", cloneURL, repoPath)

	// Set up environment for authentication if token is available
	if config.GITHUB_ACCESS_TOKEN != "" {
		// Use token for authentication
		authenticatedURL := fmt.Sprintf("https://%s@github.com/%s/%s.git", config.GITHUB_ACCESS_TOKEN, owner, repo)
		cmd = exec.CommandContext(cloneCtx, "git", "clone", "--depth", "1", authenticatedURL, repoPath)
	}

	// Execute clone command
	output, err := cmd.CombinedOutput()
	if err != nil {
		cleanup() // Clean up on failure
		return "", nil, fmt.Errorf("failed to clone repository: %v, output: %s", err, string(output))
	}

	log.Printf("Successfully cloned repository %s/%s to %s", owner, repo, repoPath)
	return repoPath, cleanup, nil
}

// IsJavaScriptProjectLocal checks if the repository is a JavaScript project by looking for local files
func IsJavaScriptProjectLocal(repoDir string) (bool, error) {
	// Check for package.json first (most reliable indicator)
	packageJSONPath := filepath.Join(repoDir, "package.json")
	if _, err := os.Stat(packageJSONPath); err == nil {
		return true, nil
	}

	// Check for other JavaScript indicators
	jsFiles := []string{
		"index.js", "app.js", "main.js",
		"index.ts", "app.ts", "main.ts",
		"src/index.js", "src/app.js", "src/main.js",
		"src/index.ts", "src/app.ts", "src/main.ts",
	}

	for _, file := range jsFiles {
		filePath := filepath.Join(repoDir, file)
		if _, err := os.Stat(filePath); err == nil {
			return true, nil
		}
	}

	return false, nil
}

// GetFrameworkBuildInfo returns build command and directory for a framework
func GetFrameworkBuildInfo(frameworkName string) BuildInfo {
	if buildInfo, exists := FrameworkBuildInfo[frameworkName]; exists {
		return buildInfo
	}
	// Return defaults if framework not found
	return BuildInfo{Command: "npm run build", Dir: "build"}
}

// DetectPackageManager determines if project uses yarn or npm
func DetectPackageManager(repoDir string) string {
	yarnLockPath := filepath.Join(repoDir, "yarn.lock")
	if _, err := os.Stat(yarnLockPath); err == nil {
		return "yarn"
	}
	return "npm run"
}

// DetermineBuildInfo determines build command and directory based on detected frameworks and package manager
// Priority: Non-React frameworks > React > defaults
func DetermineBuildInfo(frameworks []FrameworkInfo, repoDir string) (string, string, string) {
	defaultBuildDir := "build"
	packageManager := DetectPackageManager(repoDir)
	defaultCommand := packageManager + " build"

	// First, look for any non-React framework (these take precedence)
	for _, framework := range frameworks {
		if framework.Name != "react" {
			buildInfo := GetFrameworkBuildInfo(framework.Name)
			// Replace npm run with detected package manager
			command := buildInfo.Command
			if strings.HasPrefix(command, "npm run ") {
				command = strings.Replace(command, "npm run ", packageManager+" ", 1)
			} else {
				command = packageManager + " " + buildInfo.Command
			}
			return command, buildInfo.Dir, packageManager
		}
	}

	// If no non-React framework found, check for React
	for _, framework := range frameworks {
		if framework.Name == "React" {
			buildInfo := GetFrameworkBuildInfo(framework.Name)
			// Replace npm run with detected package manager
			command := buildInfo.Command
			if strings.HasPrefix(command, "npm run ") {
				command = strings.Replace(command, "npm run ", packageManager+" ", 1)
			}
			return command, buildInfo.Dir, packageManager
		}
	}

	// Return defaults if no frameworks found
	return defaultCommand, defaultBuildDir, packageManager
}

// DetectJavaScriptFrameworksLocal performs JavaScript framework detection from local repository
func DetectJavaScriptFrameworksLocal(repoDir string) ([]FrameworkInfo, error) {
	detectedFrameworks := make(map[string]FrameworkInfo)

	for _, rule := range JSFrameworkRules {
		frameworks, err := analyzeFileWithRuleLocal(repoDir, rule)
		if err != nil {
			log.Printf("Error analyzing file %s with rule: %v", rule.FilePattern, err)
			continue // Continue with other rules
		}

		// Add detected frameworks to results
		for _, framework := range frameworks {
			// Use framework name as key to avoid duplicates (prioritize higher confidence)
			if existing, exists := detectedFrameworks[framework.Name]; !exists ||
				GetConfidenceScore(framework.Confidence) > GetConfidenceScore(existing.Confidence) {
				detectedFrameworks[framework.Name] = framework
			} else if GetConfidenceScore(framework.Confidence) == GetConfidenceScore(existing.Confidence) {
				// If same confidence, combine detection sources
				if !strings.Contains(existing.DetectedIn, framework.DetectedIn) {
					existing.DetectedIn = existing.DetectedIn + ", " + framework.DetectedIn
					existing.Description = existing.Description + "; " + framework.Description
					detectedFrameworks[framework.Name] = existing
				}
			}
		}
	}

	// Convert map to slice
	var result []FrameworkInfo
	for _, framework := range detectedFrameworks {
		result = append(result, framework)
	}

	return result, nil
}

// analyzeFileWithRuleLocal analyzes a single local file with a detection rule
func analyzeFileWithRuleLocal(repoDir string, rule FileDetectionRule) ([]FrameworkInfo, error) {
	var frameworks []FrameworkInfo

	filePath := filepath.Join(repoDir, rule.FilePattern)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return frameworks, nil // File not found is not an error
	}

	// Read file content
	content, err := readLocalFile(filePath)
	if err != nil {
		return frameworks, err
	}

	if content == "" {
		return frameworks, nil
	}

	// If rule has a direct framework mapping
	if rule.Framework != "" {
		frameworks = append(frameworks, FrameworkInfo{
			Name:        rule.Framework,
			Confidence:  rule.Confidence,
			DetectedIn:  rule.FilePattern,
			Description: rule.Description,
		})
	}

	// Check dependencies in file content
	if rule.Dependencies != nil {
		detectedDeps := analyzeFileDependencies(content, rule)
		frameworks = append(frameworks, detectedDeps...)
	}

	// Check script commands in package.json
	if rule.ScriptCommands != nil && rule.FilePattern == "package.json" {
		scriptFrameworks := analyzeScriptCommands(content, rule)
		frameworks = append(frameworks, scriptFrameworks...)
	}

	// Check content patterns
	if rule.ContentChecks != nil {
		contentFrameworks := analyzeContentPatterns(content, rule)
		frameworks = append(frameworks, contentFrameworks...)
	}

	return frameworks, nil
}

// readLocalFile reads content from a local file with size limits
func readLocalFile(filePath string) (string, error) {
	// Check file size first
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %v", err)
	}

	if fileInfo.Size() > maxFileSizeBytes {
		return "", fmt.Errorf("file too large: %d bytes (max: %d)", fileInfo.Size(), maxFileSizeBytes)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	return string(content), nil
}

// analyzeFileDependencies checks for framework dependencies in file content
func analyzeFileDependencies(content string, rule FileDetectionRule) []FrameworkInfo {
	var frameworks []FrameworkInfo
	contentLower := strings.ToLower(content)

	for dependency, frameworkName := range rule.Dependencies {
		if strings.Contains(contentLower, strings.ToLower(dependency)) {
			// Try to extract version if possible
			version := extractVersion(content, dependency)

			frameworks = append(frameworks, FrameworkInfo{
				Name:        frameworkName,
				Version:     version,
				Confidence:  rule.Confidence,
				DetectedIn:  rule.FilePattern,
				Description: rule.Description,
			})
		}
	}

	return frameworks
}

// analyzeScriptCommands checks for framework indicators in package.json scripts section
func analyzeScriptCommands(content string, rule FileDetectionRule) []FrameworkInfo {
	var frameworks []FrameworkInfo

	// Extract scripts section from package.json
	scriptsContent := extractScriptsSection(content)
	if scriptsContent == "" {
		return frameworks
	}

	scriptsLower := strings.ToLower(scriptsContent)

	for command, frameworkName := range rule.ScriptCommands {
		if strings.Contains(scriptsLower, strings.ToLower(command)) {
			frameworks = append(frameworks, FrameworkInfo{
				Name:        frameworkName,
				Confidence:  "medium", // Scripts have medium confidence
				DetectedIn:  "package.json scripts",
				Description: fmt.Sprintf("Detected from script command: %s", command),
			})
		}
	}

	return frameworks
}

// extractScriptsSection extracts the scripts section from package.json content
func extractScriptsSection(content string) string {
	// Find the scripts section in the JSON
	re := regexp.MustCompile(`"scripts"\s*:\s*\{([^}]*(?:\{[^}]*\}[^}]*)*)\}`)
	matches := re.FindStringSubmatch(content)

	if len(matches) >= 2 {
		return matches[1]
	}

	return ""
}

// analyzeContentPatterns checks for specific content patterns
func analyzeContentPatterns(content string, rule FileDetectionRule) []FrameworkInfo {
	var frameworks []FrameworkInfo

	for _, pattern := range rule.ContentChecks {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			frameworks = append(frameworks, FrameworkInfo{
				Name:        rule.Framework,
				Confidence:  rule.Confidence,
				DetectedIn:  rule.FilePattern,
				Description: rule.Description,
			})
		}
	}

	return frameworks
}

// extractVersion attempts to extract version information from dependency declarations
func extractVersion(content, dependency string) string {
	// Pattern to match version in package.json format: "dependency": "^1.2.3"
	pattern := fmt.Sprintf(`"%s":\s*"([^"]+)"`, regexp.QuoteMeta(dependency))
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)

	if len(matches) >= 2 {
		version := strings.Trim(matches[1], "^~>=<")
		return version
	}

	return ""
}

// GetConfidenceScore returns numeric confidence score for comparison
func GetConfidenceScore(confidence string) int {
	switch strings.ToLower(confidence) {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// ParseRepositoryURL parses and validates repository URL
func ParseRepositoryURL(repoURL string) (owner, repo string, err error) {
	if repoURL == "" {
		return "", "", fmt.Errorf("repository URL is required")
	}

	// Parse URL
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid URL format: %v", err)
	}

	// Validate it's a GitHub URL
	if parsedURL.Host != "github.com" {
		return "", "", fmt.Errorf("only GitHub repositories are supported")
	}

	// Extract owner and repo from path
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 {
		return "", "", fmt.Errorf("invalid GitHub repository URL format")
	}

	owner = pathParts[0]
	repo = strings.TrimSuffix(pathParts[1], ".git")

	// Validate owner and repo names
	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("invalid repository owner or name")
	}

	// Basic validation for GitHub naming conventions
	validName := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validName.MatchString(owner) || !validName.MatchString(repo) {
		return "", "", fmt.Errorf("invalid repository owner or name format")
	}

	return owner, repo, nil
}
