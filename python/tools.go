package python

import (
	"context"
	"fmt"
	"pulp-mcpserver/pulp"
	"strings"

	"github.com/git-hyagi/pulp-bindings-go/bindings"
)

// trimHrefPrefix strips the leading slash from a pulp_href to work around
// the Go OpenAPI bindings using url.PathEscape on href path parameters,
// which causes a double-slash in the request URL (e.g. //api/pulp/...).
func trimHrefPrefix(href string) string {
	return strings.TrimPrefix(href, "/")
}

func getPulpDomain(domain *string) string {
	if domain != nil && *domain != "" {
		return *domain
	}
	return "default"
}

func PythonTools(ctx context.Context, pulpClient pulp.PulpClient) (any, any, error) {
	plugin := strings.ToLower(pulpClient.Pulp.Plugin)
	resource := strings.ToLower(pulpClient.Pulp.Resource)
	action := strings.ToLower(pulpClient.Pulp.Action)
	domain := getPulpDomain(pulpClient.Pulp.Domain)
	client := pulpClient.Client

	var name string
	if pulpClient.Pulp.Name != nil {
		name = *pulpClient.Pulp.Name
	}

	switch plugin {
	case "python":
		switch resource {
		case "contents", "content", "packages", "package":
			switch action {
			case "list":
				request := client.ContentPackagesAPI.ContentPythonPackagesList(ctx, domain)
				if name != "" {
					request = request.Name(name)
				}
				result, httpResp, err := request.Execute()
				if err != nil {
					return nil, httpResp, err
				}
				seen := map[string]bool{}
				names := []string{}
				for _, pkg := range result.Results {
					if pkg.Name != nil && !seen[*pkg.Name] {
						seen[*pkg.Name] = true
						names = append(names, *pkg.Name)
					}
				}
				return struct {
					Count    int32    `json:"count"`
					Next     *string  `json:"next,omitempty"`
					Previous *string  `json:"previous,omitempty"`
					Names    []string `json:"names"`
				}{
					Count:    result.Count,
					Next:     result.Next,
					Previous: result.Previous,
					Names:    names,
				}, httpResp, nil
			}
		case "repositories", "repository":
			//TODO: missing update
			//TODO: missing labels
			switch action {
			case "list":
				request := client.RepositoriesPythonAPI.RepositoriesPythonPythonList(ctx, domain)
				if name != "" {
					request = request.NameIexact(name)
				}
				return request.Execute()
			case "create":
				request := client.RepositoriesPythonAPI.RepositoriesPythonPythonCreate(ctx, domain)
				if name == "" {
					return nil, nil, fmt.Errorf("ERROR! Failed to create a repository, at least a name must be provided")
				}
				repo := bindings.PythonPythonRepository{
					Name:        name,
					Description: pulpClient.Pulp.Description,
					Remote:      pulpClient.Pulp.Remote,
					Autopublish: pulpClient.Pulp.Autopublish,
				}
				return request.PythonPythonRepository(repo).Execute()
			case "delete":
				repoRequest := client.RepositoriesPythonAPI.RepositoriesPythonPythonList(ctx, domain)
				repo, _, err := repoRequest.Name(name).Execute()
				if err != nil {
					return nil, nil, err
				}
				request := client.RepositoriesPythonAPI.RepositoriesPythonPythonDelete(ctx, trimHrefPrefix(*repo.Results[0].PulpHref))
				return request.Execute()
			case "sync":
				repoRequest := client.RepositoriesPythonAPI.RepositoriesPythonPythonList(ctx, domain)
				repo, _, err := repoRequest.Name(name).Execute()
				if err != nil {
					return nil, nil, err
				}
				syncRequest := client.RepositoriesPythonAPI.RepositoriesPythonPythonSync(ctx, trimHrefPrefix(*repo.Results[0].PulpHref))
				if pulpClient.Pulp.Remote != nil {
					remote := *pulpClient.Pulp.Remote
					// if the remote is not a pulp_href, resolve the name to its href
					if !strings.HasPrefix(remote, "/") {
						remoteRequest := client.RemotesPythonAPI.RemotesPythonPythonList(ctx, domain)
						remoteResult, _, err := remoteRequest.Name(remote).Execute()
						if err != nil {
							return nil, nil, err
						}
						if len(remoteResult.Results) == 0 {
							return nil, nil, fmt.Errorf("ERROR! Remote '%s' not found", remote)
						}
						remote = *remoteResult.Results[0].PulpHref
					}
					syncURL := bindings.RepositorySyncURL{Remote: &remote}
					syncRequest = syncRequest.RepositorySyncURL(syncURL)
				}
				return syncRequest.Execute()
			}

		case "distributions", "distribution":
			//TODO: missing update
			//TODO: missing labels
			switch action {
			case "list":
				request := client.DistributionsPypiAPI.DistributionsPythonPypiList(ctx, domain)
				if name != "" {
					request = request.NameIexact(name)
				}
				return request.Execute()
			case "create":
				request := client.DistributionsPypiAPI.DistributionsPythonPypiCreate(ctx, domain)
				if name == "" {
					return nil, nil, fmt.Errorf("ERROR! Failed to create a distribution, at least a name must be provided")
				}
				var basePath string
				if pulpClient.Pulp.BasePath != nil {
					basePath = *pulpClient.Pulp.BasePath
				}
				//TODO: the distribution expects the href of repository and/or remote
				// update the code to handle cases where users passed repo name or remote name instead of href
				dist := bindings.PythonPythonDistribution{
					Name:       name,
					BasePath:   basePath,
					Repository: pulpClient.Pulp.Repository,
					Remote:     pulpClient.Pulp.Remote,
				}
				return request.PythonPythonDistribution(dist).Execute()
			case "delete":
				distRequest := client.DistributionsPypiAPI.DistributionsPythonPypiList(ctx, domain)
				dist, _, err := distRequest.Name(name).Execute()
				if err != nil {
					return nil, nil, err
				}
				request := client.DistributionsPypiAPI.DistributionsPythonPypiDelete(ctx, trimHrefPrefix(*dist.Results[0].PulpHref))
				return request.Execute()
			}
		case "remotes", "remote":
			//TODO: missing update
			//TODO: missing labels
			switch action {
			case "list":
				request := client.RemotesPythonAPI.RemotesPythonPythonList(ctx, domain)
				if name != "" {
					request = request.NameIexact(name)
				}
				return request.Execute()
			case "create":
				request := client.RemotesPythonAPI.RemotesPythonPythonCreate(ctx, domain)
				if name == "" {
					return nil, nil, fmt.Errorf("ERROR! Falied to create a remote, at least a name must be provided")
				}
				var url string
				if pulpClient.Pulp.Url != nil {
					url = *pulpClient.Pulp.Url
				}
				remote := bindings.PythonPythonRemote{
					Name:        name,
					Url:         url,
					Includes:    pulpClient.Pulp.Includes,
					Excludes:    pulpClient.Pulp.Excludes,
					Prereleases: pulpClient.Pulp.Prereleases,
				}
				if pulpClient.Pulp.Policy != nil {
					policy, err := bindings.NewPolicy692EnumFromValue(*pulpClient.Pulp.Policy)
					if err != nil {
						return nil, nil, err
					}
					remote.Policy = policy
				}
				return request.PythonPythonRemote(remote).Execute()
			case "delete":
				remoteRequest := client.RemotesPythonAPI.RemotesPythonPythonList(ctx, domain)
				remote, _, err := remoteRequest.Name(name).Execute()
				if err != nil {
					return nil, nil, err
				}
				request := client.RemotesPythonAPI.RemotesPythonPythonDelete(ctx, trimHrefPrefix(*remote.Results[0].PulpHref))
				return request.Execute()
			}
		}
	}
	return nil, nil, nil
}
