package templateDocument

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

type Template interface {
	FindPlaceholders() ([]string, error)
	Render(map[string][]string, string) error
}

type TemplatedDocument struct {
	Path         string
	Content      string
	ValueMapList []map[string]string
	matcher      *regexp.Regexp
}

func NewTemplatedDocument(filepath string, mappingFile string) (t *TemplatedDocument, err error) {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s does not exist", filepath)
	}

	var jsonMapping []map[string]string
	if mappingFile != "" {
		file, err := ioutil.ReadFile(mappingFile)
		if err != nil {
			return &TemplatedDocument{}, fmt.Errorf("could not read mapping file %s: %s", mappingFile, err)
		}

		if err := json.Unmarshal(file, &jsonMapping); err != nil {
			return &TemplatedDocument{}, fmt.Errorf("could not unmarshall %s: %s", mappingFile, err)
		}
	}
	return &TemplatedDocument{
		Path:         filepath,
		matcher:      regexp.MustCompile(`{{\s*([^ }]*)?\s*}}`),
		ValueMapList: jsonMapping,
	}, nil
}

// Return slice containing all "versions" of the document, with template placeholders replaced
func (t *TemplatedDocument) Render() (versions []string, err error) {
	// No need to read if Content already defined
	if t.Content == "" {
		_, err = t.read()
		if err != nil {
			return []string{}, err
		}
	}

	placeholders, err := t.findPlaceholders()
	if err != nil {
		return []string{}, fmt.Errorf("error finding placeholders: %s", err)
	}

	var contentSlice []string

	for _, valueMap := range t.ValueMapList {
		var c = t.Content
		for k, v := range placeholders {
			c = strings.Replace(c, v, valueMap[k], -1)
		}
		contentSlice = append(contentSlice, c)
	}

	return contentSlice, nil
}


func (t *TemplatedDocument) read() (string, error) {
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
func (t *TemplatedDocument) findPlaceholders() (placeholders map[string]string, err error) {
	matches := t.matcher.FindAllStringSubmatch(t.Content, -1)
	placeholders = make(map[string]string)

	for _, m := range matches {
		placeholders[m[1]] = m[0]
	}

	return placeholders, nil
}

