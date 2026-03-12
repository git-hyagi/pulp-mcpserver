package file

import (
	"github.com/git-hyagi/pulp-bindings-go/bindings"
)

type ListReposInput struct {
	Domain *string `json:"domain,omitempty" jsonschema:"Pulp domain to query. Defaults to 'default' if not provided."`
	Name   *string `json:"name,omitempty" jsonschema:"Filter repositories by name. Leave empty to list all."`
}

type RepositoryResponse struct {
	Response []bindings.FileFileRepositoryResponse `json:"repository" jsonschema:"pulp file repository"`
}
