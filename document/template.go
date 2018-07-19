package document

import (
	"bytes"
	"fmt"
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
// json document when provided with a ValueMapList
type Template struct {
	Path         string
	Content      string
	ValueMapList []TemplateParams // List of "instances" of the document, mapping the key-values for each one
	matcher      *regexp.Regexp   // Regex to find placeholders
	placeHolders map[string]string
}

// A rendered template which we can write to vault
type RenderedTemplate struct {
	Name    string
	Content string
}

func NewTemplate(filepath string, params []TemplateParams) (t *Template) {
	return &Template{
		Path:         filepath,
		matcher:      regexp.MustCompile(`{{\s*([^ }]*)?\s*}}`),
		ValueMapList: params,
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

	if len(placeholders) == 0 || len(t.ValueMapList) == 0 {
		// no placeholders or values to map, return a single result with the original content
		return []RenderedTemplate{
			{Content: t.Content},
		}, nil
	}

	// Avoid writing duplicate documents when all the placeholder values are the same
	hasMultiple, err := t.hasMultiple()
	if err != nil {
		return
	}
	if hasMultiple {
		return t.renderMultiple()
	} else {
		return t.renderOne()
	}

}

// Render only one template. Typically used after hasMultiple() returned false
// Still returns a slice as that's what's expected by the caller of Render()
func (t *Template) renderOne() (renderedTemplates []RenderedTemplate, err error) {
	renderedTemplates = []RenderedTemplate{}
	placeholders, err := t.findPlaceholders()
	if err != nil {
		return
	}

	templateConfig := t.ValueMapList[0]

	var c = t.Content
	for k, v := range placeholders {
		c = strings.Replace(c, v, templateConfig.Variables[k], -1)
	}

	return []RenderedTemplate{
		{Name: "", Content: c},
	}, nil
}

// Render an array of templates, presumably with different content!
func (t *Template) renderMultiple() (renderedTemplates []RenderedTemplate, err error) {
	renderedTemplates = []RenderedTemplate{}
	placeholders, err := t.findPlaceholders()
	if err != nil {
		return
	}

	for _, templateConfig := range t.ValueMapList {
		var c = t.Content
		for k, v := range placeholders {
			c = strings.Replace(c, v, templateConfig.Variables[k], -1)
		}
		renderedTemplates = append(renderedTemplates, RenderedTemplate{
			Name:    templateConfig.Name,
			Content: c,
		})
	}

	return renderedTemplates, nil
}

// Determine whether we have multiple values for any variable
func (t *Template) hasMultiple() (hasMultiple bool, err error) {
	placeholders, err := t.findPlaceholders()
	if err != nil {
		return hasMultiple, fmt.Errorf("error finding placeholders: %s", err)
	}

	for key := range placeholders {
		seen := make(map[string]bool)
		for _, templateConfig := range t.ValueMapList {
			seen[templateConfig.Variables[key]] = true
		}

		if len(seen) > 1 {
			return true, nil
		}
	}

	return false, nil
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
