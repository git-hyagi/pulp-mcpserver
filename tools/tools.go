package tools

import (
	"context"
	"fmt"
	"pulp-mcpserver/pulp"
	"pulp-mcpserver/python"
	"pulp-mcpserver/rpm"
	"strings"

	"github.com/git-hyagi/pulp-bindings-go/bindings"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PulpClient embeds pulp.PulpClient and adds MCP tool integration.
type PulpClient struct {
	*pulp.PulpClient
}

func PulpFunction(ctx context.Context, pulpClient pulp.PulpClient) (any, any, error) {
	plugin := strings.ToLower(pulpClient.Pulp.Plugin)

	switch plugin {
	case "python":
		return python.PythonTools(ctx, pulpClient)
	case "rpm":
		return rpm.RPMTools(ctx, pulpClient)
	}
	return nil, nil, nil
}

// NewPulpClient creates a new PulpClient with the given API client and credentials.
func NewPulpClient(client *bindings.APIClient, auth bindings.BasicAuth) *PulpClient {
	return &PulpClient{PulpClient: pulp.NewPulpClient(client, auth)}
}

func (p *PulpClient) PulpTool(ctx context.Context, req *mcp.CallToolRequest, in pulp.PulpResource) (
	*mcp.CallToolResult,
	any,
	error,
) {
	p.Pulp = in
	authCtx := p.AuthCtx(ctx)

	pulpFunc, _, err := PulpFunction(authCtx, *p.PulpClient)
	if err != nil {
		return nil, nil, fmt.Errorf("ERROR! Failed to get resource function: %w", err)
	}
	return nil, pulpFunc, nil
}
