package file

import "github.com/git-hyagi/pulp-bindings-go/bindings"

type ListContentsInput struct {
	Domain *string `json:"domain,omitempty" jsonschema:"Pulp domain to query. Defaults to 'default' if not provided."`
	Name   *string `json:"name,omitempty" jsonschema:"Filter contents by name. Leave empty to list all."`
}

type ContentsResponse struct {
	Response []bindings.FileFileContentResponse `json:"content" jsonschema:"pulp file contents"`
}
