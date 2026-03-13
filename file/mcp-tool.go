package file

import (
	"context"
	"fmt"

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

func getPulpDomain(domain *string) string {
	if domain != nil && *domain != "" {
		return *domain
	}
	return "default"
}

func (p *PulpClient) ListFileRepositories(ctx context.Context, req *mcp.CallToolRequest, in ListReposInput) (
	*mcp.CallToolResult,
	RepositoryResponse,
	error,
) {
	domain := getPulpDomain(in.Domain)
	request := p.Client.RepositoriesFileAPI.RepositoriesFileFileList(p.authCtx(ctx), domain)
	if in.Name != nil {
		request = request.NameRegex(*in.Name)
	}
	list, _, err := request.Execute()
	if err != nil {
		return nil, RepositoryResponse{}, err
	}

	return nil, RepositoryResponse{list.Results}, nil
}

func (p *PulpClient) ListFileDistributions(ctx context.Context, req *mcp.CallToolRequest, in ListDistInput) (
	*mcp.CallToolResult,
	DistributionsResponse,
	error,
) {
	domain := getPulpDomain(in.Domain)
	request := p.Client.DistributionsFileAPI.DistributionsFileFileList(p.authCtx(ctx), domain)
	if in.Name != nil {
		request = request.NameRegex(*in.Name)
	}
	list, _, err := request.Execute()
	if err != nil {
		return nil, DistributionsResponse{}, err
	}

	return nil, DistributionsResponse{list.Results}, nil
}

func (p *PulpClient) ListFileRemotes(ctx context.Context, req *mcp.CallToolRequest, in ListRemotesInput) (
	*mcp.CallToolResult,
	RemotesResponse,
	error,
) {
	domain := getPulpDomain(in.Domain)
	request := p.Client.RemotesFileAPI.RemotesFileFileList(p.authCtx(ctx), domain)
	if in.Name != nil {
		request = request.NameRegex(*in.Name)
	}
	list, _, err := request.Execute()
	if err != nil {
		return nil, RemotesResponse{}, err
	}

	return nil, RemotesResponse{list.Results}, nil
}

func (p *PulpClient) ListFileContents(ctx context.Context, req *mcp.CallToolRequest, in ListContentsInput) (
	*mcp.CallToolResult,
	ContentsResponse,
	error,
) {
	domain := getPulpDomain(in.Domain)
	request := p.Client.ContentFilesAPI.ContentFileFilesList(p.authCtx(ctx), domain)
	if in.Name != nil {
		request = request.RelativePathRegex(*in.Name)
	}
	list, _, err := request.Execute()
	if err != nil {
		return nil, ContentsResponse{}, err
	}

	return nil, ContentsResponse{list.Results}, nil
}

func (p *PulpClient) ListFileRepositoryContents(ctx context.Context, req *mcp.CallToolRequest, in ListReposInput) (
	*mcp.CallToolResult,
	ContentsResponse,
	error,
) {
	if in.Name == nil {
		return nil, ContentsResponse{}, fmt.Errorf("ERROR! REPOSITORY NAME NOT PROVIDED")
	}

	domain := getPulpDomain(in.Domain)
	repoName := *in.Name
	authCtx := p.authCtx(ctx)

	// get repo HREF
	fileListRequest := p.Client.RepositoriesFileAPI.RepositoriesFileFileList(authCtx, domain)
	fileList, _, err := fileListRequest.Name(repoName).Execute()
	if err != nil || fileList.Count == 0 {
		return nil, ContentsResponse{}, fmt.Errorf("ERROR! FILE REPOSITORY %v NOT FOUND", repoName)
	}
	repoHREF := *fileList.Results[0].PulpHref

	// get repoVersion HREF
	repoVersionHREF := *fileList.Results[0].LatestVersionHref
	if in.Version != nil {
		repoVersionHREF = fmt.Sprintf("%sversions/%s/", repoHREF, *in.Version)
	}

	// get content from repoversion
	contents, _, err := p.Client.ContentFilesAPI.ContentFileFilesList(authCtx, domain).RepositoryVersion(repoVersionHREF).Execute()
	if err != nil {
		return nil, ContentsResponse{}, err
	}

	return nil, ContentsResponse{contents.Results}, nil
}
