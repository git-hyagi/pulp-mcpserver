package file

import (
	"context"

	"github.com/git-hyagi/pulp-bindings-go/bindings"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PulpClient wraps the Pulp API client and auth credentials.
type PulpClient struct {
	Client *bindings.APIClient
	Auth   bindings.BasicAuth
}

// NewPulpClient creates a new PulpClient with the given API client and credentials.
func NewPulpClient(client *bindings.APIClient, auth bindings.BasicAuth) *PulpClient {
	return &PulpClient{Client: client, Auth: auth}
}

func (p *PulpClient) authCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, bindings.ContextBasicAuth, p.Auth)
}

func getPulpDomain(in ListReposInput) string {
	if in.Domain != nil && *in.Domain != "" {
		return *in.Domain
	}
	return "default"
}

func (p *PulpClient) ListFileRepositories(ctx context.Context, req *mcp.CallToolRequest, in ListReposInput) (
	*mcp.CallToolResult,
	RepositoryResponse,
	error,
) {
	domain := getPulpDomain(in)
	request := p.Client.RepositoriesFileAPI.RepositoriesFileFileList(p.authCtx(ctx), domain)
	if in.Name != nil {
		request = request.NameRegex(*in.Name)
	}
	list, _, err := request.Limit(25).Execute()
	if err != nil {
		return nil, RepositoryResponse{}, err
	}

	return nil, RepositoryResponse{list.Results}, nil
}
