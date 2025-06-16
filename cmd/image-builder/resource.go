package imagebuilder

import (
	cTypes "github.com/open-ug/conveyor/pkg/types"
)

var ImageBuilder = &cTypes.ResourceDefinition{
	Name:        "image-builder",
	Description: "Image Builder Resource for MIRA",
	Version:     "v1",
	Schema: map[string]interface{}{
		"required": []interface{}{"name", "spec"},
		"properties": map[string]interface{}{
			"spec": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"source": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"type": map[string]interface{}{
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
								"required": []interface{}{"url", "branch", "revision"},
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
		},
	},
}
