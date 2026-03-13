package main

import (
	"context"
	"log"
	"os"

	"pulp-mcpserver/file"

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

	pulpClient := file.NewPulpClient(
		bindings.NewAPIClient(cfg),
		bindings.BasicAuth{UserName: pulpUser, Password: pulpPassword},
	)

	server := mcp.NewServer(&mcp.Implementation{Name: "pulp-mcp", Version: "v0.0.1"}, nil)

	mcp.AddTool(server, &mcp.Tool{Name: "pulp_list_file_repos", Description: "list pulp file repositories"}, pulpClient.ListFileRepositories)
	mcp.AddTool(server, &mcp.Tool{Name: "pulp_list_file_remotes", Description: "list pulp file remotes"}, pulpClient.ListFileRemotes)
	mcp.AddTool(server, &mcp.Tool{Name: "pulp_list_file_distributions", Description: "list pulp file distributions"}, pulpClient.ListFileDistributions)
	mcp.AddTool(server, &mcp.Tool{Name: "pulp_list_file_contents", Description: "list pulp file contents"}, pulpClient.ListFileContents)
	mcp.AddTool(server, &mcp.Tool{Name: "pulp_list_file_repository_contents", Description: "list pulp file repository contents"}, pulpClient.ListFileRepositoryContents)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
