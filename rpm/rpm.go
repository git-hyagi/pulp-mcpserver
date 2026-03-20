package rpm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"pulp-mcpserver/pulp"
	"strings"

	"github.com/git-hyagi/pulp-bindings-go/bindings"
)

func RPMTools(ctx context.Context, pulpClient pulp.PulpClient) (any, any, error) {
	resource := strings.ToLower(pulpClient.Pulp.Resource)
	action := strings.ToLower(pulpClient.Pulp.Action)
	domain := pulp.GetPulpDomain(pulpClient.Pulp.Domain)
	client := pulpClient.Client

	var name string
	if pulpClient.Pulp.Name != nil {
		name = *pulpClient.Pulp.Name
	}

	switch resource {
	case "contents", "content", "packages", "package":
		switch action {
		case "list":
			request := client.ContentPackagesAPI.ContentRpmPackagesList(ctx, domain)
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
		case "label", "labels":
			request := client.ContentPackagesAPI.ContentRpmPackagesList(ctx, domain)
			if name != "" {
				request = request.Name(name)
			}
			result, httpResp, err := request.Execute()
			if err != nil {
				return nil, httpResp, err
			}
			return client.ContentPackagesAPI.ContentRpmPackagesSetLabel(ctx, *result.Results[0].PulpHref).Execute()
		}
	case "repositories", "repository":
		clientAPI := client.RepositoriesRpmAPI
		switch action {
		case "list":
			request := clientAPI.RepositoriesRpmRpmList(ctx, domain)
			if name != "" {
				request = request.NameIexact(name)
			}
			return request.Execute()

		case "label", "labels":
			request := clientAPI.RepositoriesRpmRpmList(ctx, domain)
			if name != "" {
				request = request.NameIexact(name)
			}
			repos, httpResp, err := request.Execute()
			if err != nil {
				return nil, httpResp, err
			}
			repoHref := pulp.TrimHrefPrefix(*repos.Results[0].PulpHref)
			setLabel := bindings.SetLabel{
				Key:   pulpClient.Pulp.Label.Key,
				Value: pulpClient.Pulp.Label.Value,
			}
			return clientAPI.RepositoriesRpmRpmSetLabel(ctx, repoHref).SetLabel(setLabel).Execute()
		case "unlabel":
			request := clientAPI.RepositoriesRpmRpmList(ctx, domain)
			if name != "" {
				request = request.NameIexact(name)
			}
			repos, httpResp, err := request.Execute()
			if err != nil {
				return nil, httpResp, err
			}
			repoHref := pulp.TrimHrefPrefix(*repos.Results[0].PulpHref)
			unsetLabel := bindings.UnsetLabel{
				Key: pulpClient.Pulp.Label.Key,
			}
			return clientAPI.RepositoriesRpmRpmUnsetLabel(ctx, repoHref).UnsetLabel(unsetLabel).Execute()

		case "create":
			request := clientAPI.RepositoriesRpmRpmCreate(ctx, domain)
			if name == "" {
				return nil, nil, fmt.Errorf("ERROR! Failed to create a repository, at least a name must be provided")
			}
			repo := bindings.RpmRpmRepository{
				Name:        name,
				Description: pulpClient.Pulp.Description,
				Remote:      pulpClient.Pulp.Remote,
				Autopublish: pulpClient.Pulp.Autopublish,
			}
			return request.RpmRpmRepository(repo).Execute()
		case "update":
			repoRequest := clientAPI.RepositoriesRpmRpmList(ctx, domain)
			repo, _, err := repoRequest.Name(name).Execute()
			if err != nil {
				return nil, nil, err
			}
			request := clientAPI.RepositoriesRpmRpmPartialUpdate(ctx, pulp.TrimHrefPrefix(*repo.Results[0].PulpHref))

			repository := bindings.PatchedrpmRpmRepository{
				Description: pulpClient.Pulp.Description,
				Autopublish: pulpClient.Pulp.Autopublish,
			}
			if pulpClient.Pulp.Remote != nil {
				remote := *pulpClient.Pulp.Remote
				// if the remote is not a pulp_href, resolve the name to its href
				if !strings.HasPrefix(remote, "/") {
					remoteRequest := client.RemotesRpmAPI.RemotesRpmRpmList(ctx, domain)
					remoteResult, _, err := remoteRequest.NameIexact(remote).Execute()
					if err != nil {
						return nil, nil, err
					}
					if len(remoteResult.Results) == 0 {
						return nil, nil, fmt.Errorf("ERROR! Remote '%s' not found", remote)
					}
					remote = *remoteResult.Results[0].PulpHref
				}
				repository.Remote = &remote
			}
			result, httpResp, err := request.PatchedrpmRpmRepository(repository).Execute()
			return handleAsyncResponse(result, httpResp, err)
		case "delete":
			repoRequest := clientAPI.RepositoriesRpmRpmList(ctx, domain)
			repo, _, err := repoRequest.Name(name).Execute()
			if err != nil {
				return nil, nil, err
			}
			request := clientAPI.RepositoriesRpmRpmDelete(ctx, pulp.TrimHrefPrefix(*repo.Results[0].PulpHref))
			return request.Execute()
		case "sync":
			repoRequest := clientAPI.RepositoriesRpmRpmList(ctx, domain)
			repo, _, err := repoRequest.Name(name).Execute()
			if err != nil {
				return nil, nil, err
			}
			syncRequest := clientAPI.RepositoriesRpmRpmSync(ctx, pulp.TrimHrefPrefix(*repo.Results[0].PulpHref))
			if pulpClient.Pulp.Remote != nil {
				remote := *pulpClient.Pulp.Remote
				// if the remote is not a pulp_href, resolve the name to its href
				if !strings.HasPrefix(remote, "/") {
					remoteRequest := client.RemotesRpmAPI.RemotesRpmRpmList(ctx, domain)
					remoteResult, _, err := remoteRequest.Name(remote).Execute()
					if err != nil {
						return nil, nil, err
					}
					if len(remoteResult.Results) == 0 {
						return nil, nil, fmt.Errorf("ERROR! Remote '%s' not found", remote)
					}
					remote = *remoteResult.Results[0].PulpHref
				}
				syncURL := bindings.RpmRepositorySyncURL{Remote: &remote}
				syncRequest = syncRequest.RpmRepositorySyncURL(syncURL)
			}
			return syncRequest.Execute()
		}

	case "distributions", "distribution":
		clientAPI := client.DistributionsRpmAPI
		switch action {
		case "list":
			request := clientAPI.DistributionsRpmRpmList(ctx, domain)
			if name != "" {
				request = request.NameIexact(name)
			}
			result, httpResp, err := request.Execute()
			return handleAsyncResponse(result, httpResp, err)

		case "label", "labels":
			href, err := getRpmDistributionHref(ctx, clientAPI, domain, name)
			if err != nil {
				return nil, nil, err
			}
			distHref := pulp.TrimHrefPrefix(href)
			setLabel := bindings.SetLabel{
				Key:   pulpClient.Pulp.Label.Key,
				Value: pulpClient.Pulp.Label.Value,
			}
			return clientAPI.DistributionsRpmRpmSetLabel(ctx, distHref).SetLabel(setLabel).Execute()
		case "unlabel":
			href, err := getRpmDistributionHref(ctx, clientAPI, domain, name)
			if err != nil {
				return nil, nil, err
			}
			distHref := pulp.TrimHrefPrefix(href)
			unsetLabel := bindings.UnsetLabel{
				Key: pulpClient.Pulp.Label.Key,
			}
			return clientAPI.DistributionsRpmRpmUnsetLabel(ctx, distHref).UnsetLabel(unsetLabel).Execute()

		case "create":
			request := clientAPI.DistributionsRpmRpmCreate(ctx, domain)
			if name == "" {
				return nil, nil, fmt.Errorf("ERROR! Failed to create a distribution, at least a name must be provided")
			}
			var basePath string
			if pulpClient.Pulp.BasePath != nil {
				basePath = *pulpClient.Pulp.BasePath
			}
			//TODO: the distribution expects the href of repository and/or publication
			// update the code to handle cases where users passed repo name or remote name instead of href
			dist := bindings.RpmRpmDistribution{
				Name:        name,
				BasePath:    basePath,
				Repository:  pulpClient.Pulp.Repository,
				Publication: pulpClient.Pulp.Publication,
			}
			return request.RpmRpmDistribution(dist).Execute()
		case "update":
			href, err := getRpmDistributionHref(ctx, clientAPI, domain, name)
			if err != nil {
				return nil, nil, err
			}
			request := clientAPI.DistributionsRpmRpmPartialUpdate(ctx, pulp.TrimHrefPrefix(href))

			distribution := bindings.PatchedrpmRpmDistribution{
				BasePath:     pulpClient.Pulp.BasePath,
				ContentGuard: pulpClient.Pulp.ContentGuard,
				Publication:  pulpClient.Pulp.Publication,
			}
			if pulpClient.Pulp.Repository != nil {
				repository := *pulpClient.Pulp.Repository
				// if the repository is not a pulp_href, resolve the name to its href
				if !strings.HasPrefix(repository, "/") {
					repoRequest := client.RepositoriesRpmAPI.RepositoriesRpmRpmList(ctx, domain)
					repoResult, _, err := repoRequest.NameIexact(repository).Execute()
					if err != nil {
						return nil, nil, err
					}
					if len(repoResult.Results) == 0 {
						return nil, nil, fmt.Errorf("ERROR! Repository '%s' not found", repository)
					}
					repository = *repoResult.Results[0].PulpHref
				}
				distribution.Repository = &repository
			}
			result, httpResp, err := request.PatchedrpmRpmDistribution(distribution).Execute()
			return handleAsyncResponse(result, httpResp, err)
		case "delete":
			href, err := getRpmDistributionHref(ctx, clientAPI, domain, name)
			if err != nil {
				return nil, nil, err
			}
			request := clientAPI.DistributionsRpmRpmDelete(ctx, pulp.TrimHrefPrefix(href))
			return request.Execute()
		}
	case "remotes", "remote":
		clientAPI := client.RemotesRpmAPI
		switch action {
		case "list":
			request := clientAPI.RemotesRpmRpmList(ctx, domain)
			if name != "" {
				request = request.NameIexact(name)
			}
			return request.Execute()

		case "label", "labels":
			request := clientAPI.RemotesRpmRpmList(ctx, domain)
			if name != "" {
				request = request.NameIexact(name)
			}
			remote, httpResp, err := request.Execute()
			if err != nil {
				return nil, httpResp, err
			}
			distHref := pulp.TrimHrefPrefix(*remote.Results[0].PulpHref)
			setLabel := bindings.SetLabel{
				Key:   pulpClient.Pulp.Label.Key,
				Value: pulpClient.Pulp.Label.Value,
			}
			return clientAPI.RemotesRpmRpmSetLabel(ctx, distHref).SetLabel(setLabel).Execute()
		case "unlabel":
			request := clientAPI.RemotesRpmRpmList(ctx, domain)
			if name != "" {
				request = request.NameIexact(name)
			}
			dist, httpResp, err := request.Execute()
			if err != nil {
				return nil, httpResp, err
			}
			remoteHref := pulp.TrimHrefPrefix(*dist.Results[0].PulpHref)
			unsetLabel := bindings.UnsetLabel{
				Key: pulpClient.Pulp.Label.Key,
			}
			return clientAPI.RemotesRpmRpmUnsetLabel(ctx, remoteHref).UnsetLabel(unsetLabel).Execute()

		case "create":
			request := clientAPI.RemotesRpmRpmCreate(ctx, domain)
			if name == "" {
				return nil, nil, fmt.Errorf("ERROR! Falied to create a remote, at least a name must be provided")
			}
			var url string
			if pulpClient.Pulp.Url != nil {
				url = *pulpClient.Pulp.Url
			}
			remote := bindings.RpmRpmRemote{
				Name: name,
				Url:  url,
			}
			if pulpClient.Pulp.Policy != nil {
				policy, err := bindings.NewPolicy692EnumFromValue(*pulpClient.Pulp.Policy)
				if err != nil {
					return nil, nil, err
				}
				remote.Policy = policy
			}
			return request.RpmRpmRemote(remote).Execute()
		case "update":
			remoteRequest := clientAPI.RemotesRpmRpmList(ctx, domain)
			remote, _, err := remoteRequest.Name(name).Execute()
			if err != nil {
				return nil, nil, err
			}
			request := clientAPI.RemotesRpmRpmPartialUpdate(ctx, pulp.TrimHrefPrefix(*remote.Results[0].PulpHref))
			remoteChange := bindings.PatchedrpmRpmRemote{
				Url:    pulpClient.Pulp.Url,
				Policy: (*bindings.Policy692Enum)(pulpClient.Pulp.Policy),
			}
			result, httpResp, err := request.PatchedrpmRpmRemote(remoteChange).Execute()
			return handleAsyncResponse(result, httpResp, err)
		case "delete":
			remoteRequest := clientAPI.RemotesRpmRpmList(ctx, domain)
			remote, _, err := remoteRequest.Name(name).Execute()
			if err != nil {
				return nil, nil, err
			}
			request := clientAPI.RemotesRpmRpmDelete(ctx, pulp.TrimHrefPrefix(*remote.Results[0].PulpHref))
			return request.Execute()
		}
	}
	return nil, nil, nil
}

// getRpmDistributionHref resolves an RPM distribution name to its pulp_href.
// Falls back to raw JSON parsing when the Go bindings reject valid Pulp
// responses containing fields not present in the generated struct
// (e.g. repository_version).
func getRpmDistributionHref(ctx context.Context, clientAPI *bindings.DistributionsRpmAPIService, domain, name string) (string, error) {
	request := clientAPI.DistributionsRpmRpmList(ctx, domain)
	result, httpResp, err := request.Name(name).Execute()
	if err == nil {
		if len(result.Results) == 0 {
			return "", fmt.Errorf("distribution '%s' not found", name)
		}
		return *result.Results[0].PulpHref, nil
	}

	if httpResp != nil && httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
		body, readErr := io.ReadAll(httpResp.Body)
		if readErr != nil {
			return "", err
		}
		var raw struct {
			Results []struct {
				PulpHref string `json:"pulp_href"`
			} `json:"results"`
		}
		if jsonErr := json.Unmarshal(body, &raw); jsonErr != nil {
			return "", err
		}
		if len(raw.Results) == 0 {
			return "", fmt.Errorf("distribution '%s' not found", name)
		}
		return raw.Results[0].PulpHref, nil
	}

	return "", err
}

// handleAsyncResponse works around the Go bindings' strict JSON
// deserialization. Two known cases:
//   - Pulp returns 202 with an AsyncOperationResponse, but the bindings
//     expect the resource-specific response type.
//   - Pulp returns 200 with fields (e.g. repository_version) that the
//     bindings' response struct doesn't define, and DisallowUnknownFields
//     rejects them.
//
// In both cases the HTTP operation succeeded; we re-read the response body
// and return it in a usable form.
func handleAsyncResponse(result any, httpResp *http.Response, err error) (any, *http.Response, error) {
	if err == nil || httpResp == nil {
		return result, httpResp, err
	}

	body, readErr := io.ReadAll(httpResp.Body)
	if readErr != nil {
		return nil, httpResp, fmt.Errorf("failed to read response: %w", readErr)
	}

	if httpResp.StatusCode == 202 {
		var asyncResp bindings.AsyncOperationResponse
		if jsonErr := json.Unmarshal(body, &asyncResp); jsonErr != nil {
			return nil, httpResp, fmt.Errorf("failed to parse async response: %w", jsonErr)
		}
		return asyncResp, httpResp, nil
	}

	if httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
		var raw map[string]interface{}
		if jsonErr := json.Unmarshal(body, &raw); jsonErr != nil {
			return nil, httpResp, fmt.Errorf("failed to parse response: %w", jsonErr)
		}
		return raw, httpResp, nil
	}

	return result, httpResp, err
}
