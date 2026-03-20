package onboarding

import (
	"context"
	"fmt"
	"io"
	"pulp-mcpserver/pulp"
	"pulp-mcpserver/python"
	"pulp-mcpserver/rpm"
	"strings"

	"github.com/git-hyagi/pulp-bindings-go/bindings"
)

func OnboardingTools(ctx context.Context, pulpClient pulp.PulpClient) (any, any, error) {
	domain := pulp.GetPulpDomain(pulpClient.Pulp.Domain)
	client := pulpClient.Client
	plugin := strings.ToLower(pulpClient.Pulp.Plugin)

	/* defaults */
	description := "onboarding to Pulp"
	pythonRepo := bindings.PythonPythonRepository{
		Name:        "python-repo",
		Description: &description,
	}
	pythonDistro := bindings.PythonPythonDistribution{
		Name:     "python-distro",
		BasePath: "pypi",
	}
	rpmRepo := bindings.RpmRpmRepository{
		Name:        "rpm-repo",
		Description: &description,
	}
	rpmDistro := bindings.RpmRpmDistribution{
		Name:     "rpm-distro",
		BasePath: domain,
	}
	// create domain
	if domain == "default" {
		return nil, nil, fmt.Errorf("ERROR! The name of the new domain must be defined")
	}

	if plugin != "rpm" && plugin != "python" {
		return nil, nil, fmt.Errorf("ERROR! A plugin must be specified (rpm|python)")
	}

	var created []string
	domainObj := bindings.NewDomain(domain, bindings.STORAGECLASSENUM_PULPCORE_APP_MODELS_STORAGE_FILE_SYSTEM, map[string]interface{}{"location": "/var/lib/pulp/media/"})
	domainObj.Description = &description
	newDomain, httpResponse, err := client.DomainsAPI.DomainsCreate(ctx, "default").Domain(*domainObj).Execute()
	if err != nil {
		if httpResponse != nil && httpResponse.Body != nil {
			body, readErr := io.ReadAll(httpResponse.Body)
			if readErr == nil {
				return nil, nil, fmt.Errorf("%w: %s", err, string(body))
			}
		}
		return nil, httpResponse, err
	}
	created = append(created, *newDomain.PulpHref)

	switch plugin {
	case "rpm":
		rpmRepoRequest := client.RepositoriesRpmAPI.RepositoriesRpmRpmCreate(ctx, domain)
		newRPMRepo, httpResponse, err := rpmRepoRequest.RpmRpmRepository(rpmRepo).Execute()
		if err != nil {
			return nil, httpResponse, err
		}
		created = append(created, *newRPMRepo.PulpHref)

		rpmDistroRequest := client.DistributionsRpmAPI.DistributionsRpmRpmCreate(ctx, domain)
		_, httpResponse, err = rpmDistroRequest.RpmRpmDistribution(rpmDistro).Execute()
		if err != nil {
			return nil, httpResponse, err
		}
		rpmDistroHref, err := rpm.GetRpmDistributionHref(ctx, client.DistributionsRpmAPI, domain, rpmDistro.Name)
		if err != nil {
			return nil, nil, err
		}
		created = append(created, rpmDistroHref)

	case "python":
		pythonRepoRequest := client.RepositoriesPythonAPI.RepositoriesPythonPythonCreate(ctx, domain)
		newPythonRepo, httpResponse, err := pythonRepoRequest.PythonPythonRepository(pythonRepo).Execute()
		if err != nil {
			return nil, httpResponse, err
		}
		created = append(created, *newPythonRepo.PulpHref)

		pythonDistroRequest := client.DistributionsPypiAPI.DistributionsPythonPypiCreate(ctx, domain)
		_, httpResponse, err = pythonDistroRequest.PythonPythonDistribution(pythonDistro).Execute()
		if err != nil {
			return nil, httpResponse, err
		}
		pythonDistroHref, err := python.GetPythonDistributionHref(ctx, client.DistributionsPypiAPI, domain, pythonDistro.Name)
		if err != nil {
			return nil, nil, err
		}
		created = append(created, pythonDistroHref)
	}
	return nil, created, nil
}
