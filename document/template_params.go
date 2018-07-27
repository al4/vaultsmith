package document

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

// Set of parameters to apply to our Template document
// Defines the structure of the _vaultsmith.json file
type TemplateParams struct {
	Instances map[string][]string `json:"instances"`
	Variables map[string]string   `json:"variables"`
}

// Build template configurations from a template file and slice of overrides, for passing to Template
func GenerateTemplateParams(templateFile string, overrides []string) (tp TemplateParams, err error) {
	var templateConfig TemplateParams

	if templateFile != "" {
		file, err := ioutil.ReadFile(templateFile)
		if err != nil {
			return tp, fmt.Errorf("could not read template file %s: %s", templateFile, err)
		}

		if err := json.Unmarshal(file, &templateConfig); err != nil {
			return tp, fmt.Errorf("could not unmarshall %s: %s", templateFile, err)
		}
	} else {
		// no file, create a no-name default (blank name means no suffix is added to the path)
		templateConfig = TemplateParams{
			Variables: map[string]string{}, // need to initialise for setParams()
		}
	}

	// override/add variables from overrideParams in each TemplateParams
	var templateConfigNew TemplateParams
	templateConfigNew = setParams(templateConfig, overrides)
	return templateConfigNew, nil
}

// Set params from a slice in TemplateParams.Variables
func setParams(tp TemplateParams, params []string) TemplateParams {
	for _, v := range params {
		split := strings.Split(v, "=")
		tp.Variables[split[0]] = strings.Join(split[1:], "=") // in case more than 1 "=" in string
	}
	return tp
}
