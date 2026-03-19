package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/git-hyagi/pulp-bindings-go/bindings"
)

// PulpClient wraps the Pulp API client and auth credentials.
type PulpClient struct {
	Client *bindings.APIClient
	Auth   bindings.BasicAuth
}

type PulpResource struct {
	Domain   *string `json:"domain,omitempty" jsonschema:"Pulp domain to query. Defaults to 'default' if not provided."`
	Name     *string `json:"name,omitempty" jsonschema:"Filter contents by name. Leave empty to list all."`
	Plugin   string  `json:"plugin,omitempty" jsonschema:"Pulp plugin. Can be one of: python, rpm"`
	Resource string  `json:"resource,omitempty" jsonschema:"Pulp resource. Cane be one of: distribution,remote,repository,content,package"`
	Action   string  `json:"action,omitempty" jsonschema:"Action to take. Can be one of: create, read, list, update, delete, sync"`

	// Fields for remotes
	Url         *string  `json:"url,omitempty" jsonschema:"URL of the remote source (required for remote create)."`
	Policy      *string  `json:"policy,omitempty" jsonschema:"Download policy: immediate, on_demand, or streamed."`
	Includes    []string `json:"includes,omitempty" jsonschema:"Package specifiers to include (for Python remotes)."`
	Excludes    []string `json:"excludes,omitempty" jsonschema:"Package specifiers to exclude (for Python remotes)."`
	Prereleases *bool    `json:"prereleases,omitempty" jsonschema:"Include pre-release packages (for Python remotes)."`

	// Fields for repositories
	Description *string `json:"description,omitempty" jsonschema:"Description of the repository."`
	Remote      *string `json:"remote,omitempty" jsonschema:"Remote to associate (pulp_href or name). Used by repositories and distributions."`
	Autopublish *bool   `json:"autopublish,omitempty" jsonschema:"Auto-publish after content changes."`

	// Fields for distributions
	BasePath   *string `json:"base_path,omitempty" jsonschema:"Base path for the distribution URL (required for distribution create)."`
	Repository *string `json:"repository,omitempty" jsonschema:"Repository to serve (pulp_href or name). Used by distributions."`
}

// trimHrefPrefix strips the leading slash from a pulp_href to work around
// the Go OpenAPI bindings using url.PathEscape on href path parameters,
// which causes a double-slash in the request URL (e.g. //api/pulp/...).
func trimHrefPrefix(href string) string {
	return strings.TrimPrefix(href, "/")
}

func (p PulpResource) PulpFunction(ctx context.Context, client *bindings.APIClient, domain string) (any, any, error) {
	plugin := strings.ToLower(p.Plugin)
	resource := strings.ToLower(p.Resource)
	action := strings.ToLower(p.Action)

	var name string
	if p.Name != nil {
		name = *p.Name
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
					Description: p.Description,
					Remote:      p.Remote,
					Autopublish: p.Autopublish,
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
				if p.Remote != nil {
					remote := *p.Remote
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
				if p.BasePath != nil {
					basePath = *p.BasePath
				}
				//TODO: the distribution expects the href of repository and/or remote
				// update the code to handle cases where users passed repo name or remote name instead of href
				dist := bindings.PythonPythonDistribution{
					Name:       name,
					BasePath:   basePath,
					Repository: p.Repository,
					Remote:     p.Remote,
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
				if p.Url != nil {
					url = *p.Url
				}
				remote := bindings.PythonPythonRemote{
					Name:        name,
					Url:         url,
					Includes:    p.Includes,
					Excludes:    p.Excludes,
					Prereleases: p.Prereleases,
				}
				if p.Policy != nil {
					policy, err := bindings.NewPolicy692EnumFromValue(*p.Policy)
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
