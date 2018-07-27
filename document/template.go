package document

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Regexp which defines how to find "placeholder" values
var matcher = regexp.MustCompile(`{{\s*([^ }]*)?\s*}}`)

// This seems a little unnecessary
type Renderer interface {
	Render(map[string][]string, string) error
}

// Our Template is a document that contains placeholders, which can be rendered into a valid Vault
// json document when provided with a Params
type Template struct {
	FileName     string
	Content      string
	Params       TemplateParams // List of "instances" of the document, mapping the key-values for each one
	placeHolders map[string]string
}

// A rendered template which we can write to vault
type RenderedTemplate struct {
	Name    string
	Content string
}

// Return slice containing all "versions" of the document, with template placeholders replaced
func (t *Template) Render() (renderedTemplates []RenderedTemplate, err error) {
	if t.placeHolders == nil {
		t.placeHolders, err = t.findPlaceholders(t.Content)
		if err != nil {
			return renderedTemplates, fmt.Errorf("error finding placeholders: %s", err)
		}
	}

	// if the filename has no placeholders, we only render it once
	fileNamePlaceholders, err := t.findPlaceholders(t.FileName)
	if len(fileNamePlaceholders) == 0 {
		name := strings.TrimSuffix(t.FileName, filepath.Ext(t.FileName))
		template, err := t.createRenderedTemplate(name, t.Params)
		if err != nil {
			return renderedTemplates, err
		}
		return []RenderedTemplate{template}, err
	}

	// could potentially support combinations... but doesn't seem useful
	if len(fileNamePlaceholders) > 1 {
		return renderedTemplates, fmt.Errorf("multiple placeholders in a filename is not supported")
	}

	// we have a placeholder, thus we have to render all instances of it
	var fileNamePlaceHolderKey, fileNamePlaceHolderValue string
	for k, v := range fileNamePlaceholders {
		fileNamePlaceHolderKey = k
		fileNamePlaceHolderValue = v
	}

	// else we have to iterate of all instances of it
	var instances []string
	if i, ok := t.Params.Instances[fileNamePlaceHolderKey]; ok {
		instances = i
	} else {
		return renderedTemplates, fmt.Errorf("tried to render file name %s, but there are no instances of it in the template config", t.FileName)
	}

	for _, instance := range instances {
		name := strings.TrimSuffix(t.FileName, filepath.Ext(t.FileName))
		name = strings.Replace(name, fileNamePlaceHolderValue, instance, -1)

		rendered, err := t.createRenderedTemplate(name, t.Params)
		if err != nil {
			return renderedTemplates, err
		}

		renderedTemplates = append(renderedTemplates, rendered)
	}

	return renderedTemplates, err
}

func (t *Template) createRenderedTemplate(name string, params TemplateParams) (rt RenderedTemplate, err error) {
	placeHolders, err := t.findPlaceholders(t.Content)
	if err != nil {
		return rt, fmt.Errorf("error finding placeholders: %s", err)
	}

	content := t.Content
	for pk, placeholderText := range placeHolders {
		if value, ok := params.Variables[pk]; ok {
			content = strings.Replace(content, placeholderText, value, -1)
		} else if _, ok := params.Instances[pk]; ok {
			// If the placeholder key is in instances, then we can assume this placeholder should be
			// replaced with the instance name. Making a variable with the same key would screw
			// this up, but intent would be ambiguous in that case.
			//
			// One weird side effect is that if you created a template with no placeholder in the
			// filename and a placeholder inside the file that matches an instances key, you'd get
			// the file name of the template. Could be worse.
			content = strings.Replace(content, placeholderText, name, -1)
		} else {
			log.WithFields(log.Fields{"placeholder": pk, "fileName": t.FileName}).Warn("Placeholder has no values")
		}
	}

	return RenderedTemplate{Name: name, Content: content}, err
}

func (t *Template) replaceText(initialText string, params TemplateParams) (output string, err error) {

	return
}

// Map the keys to the actual placeholder string in the template
// For example, given template file:
//		Hello, {{ foo }}.
// We would return:
// 		map[string]string{"foo": "{{ foo }}"}
func (t *Template) findPlaceholders(text string) (placeholders map[string]string, err error) {
	matches := matcher.FindAllStringSubmatch(text, -1)
	placeholders = make(map[string]string)

	for _, m := range matches {
		placeholders[m[1]] = m[0]
	}

	return placeholders, nil
}

func Read(filePath string) (content string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening file: %s", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)

	if err != nil {
		return "", fmt.Errorf("error reading from buffer: %s", err)
	}

	content = buf.String()

	return content, err
}
