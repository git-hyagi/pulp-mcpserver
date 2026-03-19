package file

import (
	"context"
	"fmt"
	"reflect"

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

func (p PulpClient) init() any {
	rpmConfig := ListInput{
		API:          p.Client.RepositoriesFileAPI,
		ListMethod:   "RepositoriesRPMRPMList",
		FilterMethod: "Name",
	}
	return rpmConfig
}

func (p *PulpClient) _ListResource(ctx context.Context, req *mcp.CallToolRequest, in ListInput) (
	*mcp.CallToolResult,
	any,
	error,
) {
	domain := getPulpDomain(in.Domain)
	resourceAPI := reflect.ValueOf(in.API).String()
	// bindings.APIClient.RepositoriesFileFileList(p.authCtx(ctx),domain)
	request := reflect.ValueOf(p.Client).
		Elem().FieldByName(resourceAPI).
		MethodByName(in.ListMethod).
		Call([]reflect.Value{
			reflect.ValueOf(p.authCtx(ctx)), reflect.ValueOf(domain),
		})
	if in.Name != nil {
		//request[0] = <Resource><PulpType>API<Resources><PulpType><PulpType>ListRequest
		request = request[0].MethodByName(in.FilterMethod).Call([]reflect.Value{
			reflect.ValueOf(*in.Name),
		})
	}
	list := request[0].MethodByName("Execute").Call(nil)
	//list[0] = bindings.Paginated<pulpType><PulpType><Resource>ResponseList
	//list[1] = *http.Response
	//list[2] = error
	if !list[2].IsNil() {
		return nil, RepositoryResponse{}, list[2].Interface().(error)
	}

	results := list[0].Elem().FieldByName("Results").Interface()
	return nil, results, nil
}

/* func (p *PulpClient) ListResource(ctx context.Context, req *mcp.CallToolRequest, in ListResourceInput) (
	*mcp.CallToolResult,
	any,
	error,
) {
	domain := getPulpDomain(in.Domain)
	var request interface{}
	switch in.ContentType {
	case "python":
		switch in.ResourceType {
		case "repository":
			request = p.Client.RepositoriesPythonAPI.RepositoriesPythonPythonList(p.authCtx(ctx), domain)
		case "distribution":
			request = p.Client.DistributionsPypiAPI.DistributionsPythonPypiList(p.authCtx(ctx), domain)
		case "remote":
			request = p.Client.RemotesPythonAPI.RemotesPythonPythonList(p.authCtx(ctx), domain)
		case "content":
			request = p.Client.ContentPackagesAPI.ContentPythonPackagesList(p.authCtx(ctx), domain)
		}
	}
	if in.Name != nil {
		//request = request.NameRegex(*in.Name)
		request = reflect.ValueOf(request).MethodByName("NameRegex").Call([]reflect.Value{reflect.ValueOf(*in.Name)})
	}

	if err != nil {
		return nil, RepositoryResponse{}, err
	}

	return nil, RepositoryResponse{list.Results}, nil
} */

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
