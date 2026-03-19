package tools

import (
	"context"
	"fmt"

	"github.com/git-hyagi/pulp-bindings-go/bindings"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewPulpClient creates a new PulpClient with the given API client and credentials.
func NewPulpClient(client *bindings.APIClient, auth bindings.BasicAuth) *PulpClient {
	return &PulpClient{Client: client, Auth: auth}
}

func getPulpDomain(domain *string) string {
	if domain != nil && *domain != "" {
		return *domain
	}
	return "default"
}

func (p *PulpClient) authCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, bindings.ContextBasicAuth, p.Auth)
}

func (p *PulpClient) ResourceFactory(ctx context.Context, req *mcp.CallToolRequest, in PulpResource) (
	*mcp.CallToolResult,
	any,
	error,
) {
	authCtx := p.authCtx(ctx)
	domain := getPulpDomain(in.Domain)

	pulpFunc, _, err := in.PulpFunction(authCtx, p.Client, domain)
	if err != nil {
		return nil, nil, fmt.Errorf("ERROR! Failed to get resource function %w", err)
	}
	return nil, pulpFunc, nil
}
