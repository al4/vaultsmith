package document

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Set of parameters to apply to our Template document
// Matches the structure of the template.json file (within the array)
type TemplateParams struct {
	Name      string            `json:"name"`
	Variables map[string]string `json:"variables"`
}

// Build template configurations from a template file and slice of overrides, for passing to Template
func GenerateTemplateParams(templateFile string, overrides []string) (tp []TemplateParams, err error) {
	if _, err := os.Stat(templateFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s does not exist", templateFile)
	}

	var templateConfigs []TemplateParams
	if templateFile != "" {
		file, err := ioutil.ReadFile(templateFile)
		if err != nil {
			return tp, fmt.Errorf("could not read mapping file %s: %s", templateFile, err)
		}

		if err := json.Unmarshal(file, &templateConfigs); err != nil {
			return tp, fmt.Errorf("could not unmarshall %s: %s", templateFile, err)
		}
	} else {
		// no file, create a no-name default (blank name means no suffix is added to the path)
		templateConfigs = append(templateConfigs, TemplateParams{
			Variables: map[string]string{}, // need to initialise for setParams()
		})
	}

	// override/add variables from overrideParams in each TemplateParams
	var templateConfigsNew []TemplateParams
	for _, tc := range templateConfigs {
		templateConfigsNew = append(templateConfigsNew, setParams(tc, overrides))
	}
	return templateConfigsNew, nil
}

// Set params from a slice in TemplateParams.Variables
func setParams(tp TemplateParams, params []string) TemplateParams {
	for _, v := range params {
		split := strings.Split(v, "=")
		tp.Variables[split[0]] = strings.Join(split[1:], "=") // in case more than 1 "=" in string
	}
	return tp
}
