package file

import "github.com/git-hyagi/pulp-bindings-go/bindings"

type ListRemotesInput struct {
	Domain *string `json:"domain,omitempty" jsonschema:"Pulp domain to query. Defaults to 'default' if not provided."`
	Name   *string `json:"name,omitempty" jsonschema:"Filter remotes by name. Leave empty to list all."`
}

type RemotesResponse struct {
	Response []bindings.FileFileRemoteResponse `json:"remote" jsonschema:"pulp file remote"`
}
