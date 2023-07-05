package main

import (
	"errors"
	"os"
	"os/exec"

	tfjson "github.com/hashicorp/terraform-json"
)

type providerIndex struct {
	Name        string
	Resources   []string
	DataSources []string
}

// loadProviderSchemas handles fetching configured provider schemas. If
// a schemafile is specified, the schema is read from disk, otherwise,
// the schema is loaded on-demand by executing `terraform providers schema -json`.
// The resource and data source names are also indexed for each provider.
func loadProviderSchemas(schemafile string) (tfjson.ProviderSchemas, []providerIndex, error) {
	var (
		ps  tfjson.ProviderSchemas
		b   []byte
		err error
	)

	if schemafile == "" {
		b, err = exec.Command("terraform", "providers", "schema", "-json").Output()
	} else {
		b, err = os.ReadFile(schemafile)
	}
	if err != nil {
		return ps, nil, err
	}

	if err := ps.UnmarshalJSON(b); err != nil {
		return ps, nil, err
	}
	if len(ps.Schemas) == 0 {
		return ps, nil, errors.New("no provider schemas found")
	}

	var index []providerIndex
	for k, v := range ps.Schemas {
		pi := providerIndex{Name: k}
		for r := range v.ResourceSchemas {
			pi.Resources = append(pi.Resources, r)
		}
		for ds := range v.DataSourceSchemas {
			pi.DataSources = append(pi.DataSources, ds)
		}
		index = append(index, pi)
	}

	return ps, index, nil
}
