package types

type ImageRoutePayload struct {
	Name       string `json:"name"`
	SourceType string `json:"source_type"`
	SourceURL  string `json:"source_url"`
}

type ImageBuild struct {
	Name     string           `json:"name"`
	Resource string           `json:"resource"`
	Spec     ImageBuilderSpec `json:"spec"`
}

type ImageBuilderSpec struct {
	Source ImageBuilderSource `json:"source"`
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
