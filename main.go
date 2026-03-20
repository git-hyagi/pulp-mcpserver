package main

import (
	"context"
	"log"
	"os"

	"pulp-mcpserver/tools"

	"github.com/git-hyagi/pulp-bindings-go/bindings"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	pulpURL := os.Getenv("PULP_URL")
	if pulpURL == "" {
		pulpURL = "http://localhost"
	}
	pulpUser := os.Getenv("PULP_USERNAME")
	if pulpUser == "" {
		pulpUser = "admin"
	}
	pulpPassword := os.Getenv("PULP_PASSWORD")
	if pulpPassword == "" {
		pulpPassword = "password"
	}

	cfg := &bindings.Configuration{
		DefaultHeader: make(map[string]string),
		UserAgent:     "pulp-mcp/1.0",
		Debug:         false,
		Servers: bindings.ServerConfigurations{
			{
				URL:         pulpURL,
				Description: "Pulp server",
			},
		},
		OperationServers: map[string]bindings.ServerConfigurations{},
	}

	pulpClientTools := tools.NewPulpClient(
		bindings.NewAPIClient(cfg),
		bindings.BasicAuth{UserName: pulpUser, Password: pulpPassword},
	)

	server := mcp.NewServer(&mcp.Implementation{Name: "pulp-mcp", Version: "v0.0.1"}, nil)

	mcp.AddTool(server, &mcp.Tool{Name: "pulp_manage_resources", Description: "CRUD for all pulp plugins (rpm,python) resources (repositories,remotes,distributions,contents,packages)"}, pulpClientTools.PulpTool)
	mcp.AddTool(server, &mcp.Tool{Name: "pulp_onboarding", Description: "Prepare the environment to onboard new users"}, pulpClientTools.PulpOnboardingTool)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
