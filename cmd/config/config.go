package config

import "os"

// Directories
var GIT_REPOS_DIR = "/tmp/git-repos"

var GITHUB_CLIENT_ID = os.Getenv("GITHUB_CLIENT_ID")

var GITHUB_ACCESS_TOKEN = os.Getenv("GITHUB_ACCESS_TOKEN")
var GITHUB_API_URL = "https://api.github.com"

var SECURE_SOCKET_URL = os.Getenv("SECURE_SOCKET_URL") == "false"
