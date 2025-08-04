package imagebuilder

import (
	common "mira/cmd/common"
)

// ImageBuild represents a build request for the image builder
// This is compatible with common.BuildRequest but maintains backwards compatibility
type ImageBuild struct {
	Name string                  `json:"name"`
	Spec common.ImageBuilderSpec `json:"spec"`
}

// ImageBuilderSpec is an alias for common.ImageBuilderSpec for backwards compatibility
type ImageBuilderSpec = common.ImageBuilderSpec

// ImageBuilderSource is an alias for common.ImageBuilderSource for backwards compatibility
type ImageBuilderSource = common.ImageBuilderSource

// ImageBuilderGitRepo is an alias for common.ImageBuilderGitRepo for backwards compatibility
type ImageBuilderGitRepo = common.ImageBuilderGitRepo

// ImageBuilderBlobFile is an alias for common.ImageBuilderBlobFile for backwards compatibility
type ImageBuilderBlobFile = common.ImageBuilderBlobFile
