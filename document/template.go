package document

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"regexp"
	"strings"
)

// This seems a little unnecessary
type Renderer interface {
	Render(map[string][]string, string) error
}

// Our Template is a document that contains placeholders, which can be rendered into a valid Vault
// json document when provided with a Instances
type Template struct {
	Path         string
	Content      string
	Instances    []TemplateParams // List of "instances" of the document, mapping the key-values for each one
	matcher      *regexp.Regexp   // Regex to find placeholders
	placeHolders map[string]string
}

// A rendered template which we can write to vault
type RenderedTemplate struct {
	Name    string
	Content string
}

func NewTemplate(filepath string, instances []TemplateParams) (t *Template) {
	return &Template{
		Path:      filepath,
		matcher:   regexp.MustCompile(`{{\s*([^ }]*)?\s*}}`),
		Instances: instances,
	}
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

	placeholders, err := t.findPlaceholders()
	if err != nil {
		return renderedTemplates, fmt.Errorf("error finding placeholders: %s", err)
	}

	if len(placeholders) == 0 || len(t.Instances) == 0 {
		// no placeholders or values to map, return a single result with the original content
		return []RenderedTemplate{
			{Content: t.Content},
		}, nil
	}

	// Avoid writing duplicate documents when all the placeholder values are the same
	if t.hasMultiple(placeholders) {
		for _, params := range t.Instances {
			rendered, err := t.createRenderedTemplate(params)
			if err != nil {
				return renderedTemplates, err

			}
			renderedTemplates = append(renderedTemplates, rendered)
		}
	} else {
		rendered, err := t.createRenderedTemplate(t.Instances[0])
		if err != nil {
			return renderedTemplates, err
		}

		// Avoid adding to the vault document path; there is no need to append instance name to the
		// path when we only have one instance
		rendered.Name = ""

		renderedTemplates = append(renderedTemplates, rendered)
	}
	return renderedTemplates, err
}

func (t *Template) createRenderedTemplate(params TemplateParams) (rt RenderedTemplate, err error) {
	placeholders, err := t.findPlaceholders()
	if err != nil {
		return rt, err
	}

	content := t.Content
	for pk, placeholderText := range placeholders {
		if value, ok := params.Variables[pk]; ok {
			content = strings.Replace(content, placeholderText, value, -1)
		} else {
			log.WithFields(log.Fields{"placeholder": pk, "path": t.Path}).Warn("Placeholder has no values")
		}
	}

	return RenderedTemplate{Name: params.Name, Content: content}, err
}

// Determine whether we have multiple values for any variable
func (t *Template) hasMultiple(placeholders map[string]string) (hasMultiple bool) {
	for key := range placeholders {
		seen := make(map[string]bool)
		for _, templateConfig := range t.Instances {
			if value, ok := templateConfig.Variables[key]; ok {
				seen[value] = true
			}
		}

		if len(seen) > 1 {
			hasMultiple = true
		}
	}

	return hasMultiple
}

func (t *Template) read() (string, error) {
	file, err := os.Open(t.Path)
	if err != nil {
		return "", fmt.Errorf("error opening file: %s", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)

	if err != nil {
		return "", fmt.Errorf("error reading from buffer: %s", err)
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
	if t.placeHolders != nil {
		// avoid re-reading file if already done as we call this in a few places
		return t.placeHolders, nil
	}

	matches := t.matcher.FindAllStringSubmatch(t.Content, -1)
	placeholders = make(map[string]string)

	for _, m := range matches {
		placeholders[m[1]] = m[0]
	}

	return placeholders, nil
}
