package config

import (
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
	"golang.org/x/oauth2/gitlab"
	"os"
)

func GitHubOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GITHUB_REDIRECT_URI"),
		Scopes:       []string{"read:user", "repo"},
		Endpoint:     githuboauth.Endpoint,
	}
}

func GitLabOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GITLAB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITLAB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GITLAB_REDIRECT_URI"),
		Scopes:       []string{"read_user", "read_api"},
		Endpoint:     gitlab.Endpoint,
	}
}
