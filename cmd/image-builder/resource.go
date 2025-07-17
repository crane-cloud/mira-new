package imagebuilder

import (
	cTypes "github.com/open-ug/conveyor/pkg/types"
)

var ImageBuilder = &cTypes.ResourceDefinition{
	Name:        "image-builder",
	Description: "Image Builder Resource for MIRA",
	Version:     "v1",
	Schema: map[string]interface{}{
		"properties": map[string]interface{}{
			"source": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"sourceType": map[string]interface{}{
						"type": "string",
					},
					"buildCommand": map[string]interface{}{
						"type": "string",
					},
					"outputDir": map[string]interface{}{
						"type": "string",
					},
					"projectId": map[string]interface{}{
						"type": "string",
					},
					"token": map[string]interface{}{
						"type": "string",
					},
					"gitRepo": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"url":      map[string]interface{}{"type": "string"},
							"branch":   map[string]interface{}{"type": "string"},
							"revision": map[string]interface{}{"type": "string"},
							"username": map[string]interface{}{"type": "string"},
							"password": map[string]interface{}{"type": "string"},
						},
						"required": []interface{}{"url"},
					},
					"blobFile": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"source": map[string]interface{}{"type": "string"},
						},
						"required": []interface{}{"source"},
					},
				},
			},
		},
	},
}

type ImageBuild struct {
	Name     string           `json:"name"`
	Resource string           `json:"resource"`
	Spec     ImageBuilderSpec `json:"spec"`
}

type ImageBuilderSpec struct {
	Source       ImageBuilderSource `json:"source"`
	BuildCommand string             `json:"buildCommand"`
	OutputDir    string             `json:"outputDir"`
	ProjectID    string             `json:"projectId,omitempty"`
	Token        string             `json:"token"`
	Port         int                `json:"port,omitempty"`
}

type ImageBuilderSource struct {
	Type     string               `json:"sourceType"`
	GitRepo  ImageBuilderGitRepo  `json:"gitRepo,omitempty"`
	BlobFile ImageBuilderBlobFile `json:"blobFile,omitempty"`
}

type ImageBuilderGitRepo struct {
	URL      string `json:"url"`
	Branch   string `json:"branch,omitempty"`
	Revision string `json:"revision,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type ImageBuilderBlobFile struct {
	Source string `json:"source"`
}
