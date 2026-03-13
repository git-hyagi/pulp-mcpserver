package file

import "github.com/git-hyagi/pulp-bindings-go/bindings"

type ListDistInput struct {
	Domain *string `json:"domain,omitempty" jsonschema:"Pulp domain to query. Defaults to 'default' if not provided."`
	Name   *string `json:"name,omitempty" jsonschema:"Filter distributions by name. Leave empty to list all."`
}

type DistributionsResponse struct {
	Response []bindings.FileFileDistributionResponse `json:"distribution" jsonschema:"pulp file distribution"`
}
