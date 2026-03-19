package file

type ListInput struct {
	Domain       *string `json:"domain,omitempty" jsonschema:"Pulp domain to query. Defaults to 'default' if not provided."`
	Name         *string `json:"name,omitempty" jsonschema:"Filter contents by name. Leave empty to list all."`
	API          any     // RepositoriesFileAPI,DistributionsFileAPI,RemotesFileAPI,etc
	ListMethod   string  // RepositoriesFileFileList,DistributionsFileFileList,RemotesFileFileList,etc
	FilterMethod string  // NameRegex,RelativePathRegex,Name,etc
}

type ListResourceInput struct {
	API          any     // RepositoriesFileAPI,DistributionsFileAPI,RemotesFileAPI,etc
	ListMethod   string  // RepositoriesFileFileList,DistributionsFileFileList,RemotesFileFileList,etc
	Domain       *string `json:"domain,omitempty" jsonschema:"Pulp domain. Defaults to 'default'."`
	Name         *string `json:"name,omitempty" jsonschema:"Filter by name. Leave empty to list all."`
	ContentType  string  `json:"content_type" jsonschema:"The Pulp content plugin:python,rpm,container,mvn,npm"`
	ResourceType string  `json:"resource_type" jsonschema:"The resource to list:repository,remote,distribution,content"`
	ActionType   string  `json:"action_type jsonschema:"The verb will be send to pulp: create,read/list,update,delete"`
}
