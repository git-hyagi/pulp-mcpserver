package pulp

import (
	"context"

	"github.com/git-hyagi/pulp-bindings-go/bindings"
)

// PulpClient wraps the Pulp API client and auth credentials.
type PulpClient struct {
	Client *bindings.APIClient
	Auth   bindings.BasicAuth
	Pulp   PulpResource
}

// NewPulpClient creates a new PulpClient with the given API client and credentials.
func NewPulpClient(client *bindings.APIClient, auth bindings.BasicAuth) *PulpClient {
	return &PulpClient{Client: client, Auth: auth}
}

func (p *PulpClient) AuthCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, bindings.ContextBasicAuth, p.Auth)
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
