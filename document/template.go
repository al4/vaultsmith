package document

import (
	"regexp"
	"os"
	"fmt"
	"io/ioutil"
	"log"
	"bytes"
	"io"
	"strings"
	"encoding/json"
)

// This seems a little unnecessary
type Renderer interface {
	Render(map[string][]string, string) error
}

// Our Template is a document that contains placeholders, which can be rendered into a valid Vault
// json document when provided with a ValueMapList
type Template struct {
	Path         string
	Content      string
	ValueMapList []TemplateConfig	// List of "instances" of the document, mapping the key-values for each one
	matcher      *regexp.Regexp		// Regex to find placeholders
}

// The structure of our json (within the array)
type TemplateConfig struct {
	Name 		string 				`json:"name"`
	Variables	map[string]string 	`json:"variables"`
}

// A rendered template which we can write to vault
type RenderedTemplate struct {
	Name string
	Content string
}

func NewTemplate(filepath string, mappingFile string) (t *Template, err error) {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s does not exist", filepath)
	}

	var jsonMapping []TemplateConfig
	if mappingFile != "" {
		file, err := ioutil.ReadFile(mappingFile)
		if err != nil {
			return &Template{}, fmt.Errorf("could not read mapping file %s: %s", mappingFile, err)
		}

		if err := json.Unmarshal(file, &jsonMapping); err != nil {
			return &Template{}, fmt.Errorf("could not unmarshall %s: %s", mappingFile, err)
		}
	}
	return &Template{
		Path:         filepath,
		matcher:      regexp.MustCompile(`{{\s*([^ }]*)?\s*}}`),
		ValueMapList: jsonMapping,
	}, nil
}

// Return slice containing all "versions" of the document, with template placeholders replaced
func (t *Template) Render() (renderedTemplates []RenderedTemplate, err error) {
	// No need to read if Content already defined
	if t.Content == "" {
		_, err = t.read()
		if err != nil {
			return
		}
	}

	renderedTemplates = []RenderedTemplate{}

	placeholders, err := t.findPlaceholders()
	if err != nil {
		return renderedTemplates, fmt.Errorf("error finding placeholders: %s", err)
	}

	if len(placeholders) == 0 {  // no placeholders in this file, just return content
		return []RenderedTemplate{
			{Content: t.Content},
		},nil
	}

	for _, templateConfig := range t.ValueMapList {
		var c = t.Content
		for k, v := range placeholders {
			c = strings.Replace(c, v, templateConfig.Variables[k], -1)
		}
		renderedTemplates = append(renderedTemplates, RenderedTemplate{
			Name: templateConfig.Name,
			Content: c,
		})
	}

	return renderedTemplates, nil
}

func (t *Template) read() (string, error) {
	file, err := os.Open(t.Path)
	if err != nil {
		err = fmt.Errorf("error opening file: %s", err)
		return "", err
	}
	defer file.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)

	if err != nil {
		log.Fatal(fmt.Sprintf("error reading from buffer: %s", err))
	}

	data := buf.String()
	t.Content = data

	return data, nil
}

// Map the keys to the actual placeholder string in the template
// For example, given template file:
//		Hello, {{ foo }}.
// We would return:
// 		map[string]string{"foo": "{{ foo }}"}
func (t *Template) findPlaceholders() (placeholders map[string]string, err error) {
	matches := t.matcher.FindAllStringSubmatch(t.Content, -1)
	placeholders = make(map[string]string)

	for _, m := range matches {
		placeholders[m[1]] = m[0]
	}

	return placeholders, nil
}

