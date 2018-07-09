package templateDocument

import (
	"testing"
	"os"
	"strings"
	"path/filepath"
	"log"
	"regexp"
)

// calculate path to test fixtures (example/)
func examplePath() string {
	wd, _ := os.Getwd()
	pathArray := strings.Split(wd, string(os.PathSeparator))
	pathArray = pathArray[:len(pathArray) - 1]  // trims "internal"
	path := append(pathArray, "example")
	return strings.Join(path, string(os.PathSeparator))
}

func TestNewTemplatedDocument(t *testing.T) {
	_, err := NewTemplatedDocument(
		filepath.Join(examplePath(), "sys/policy/read_service.json"),
		filepath.Join(examplePath(), "template.json"))
	if err != nil {
		log.Fatal(err)
	}
}

func TestTemplatedDocument_findPlaceholders(t *testing.T) {
	mapping := []map[string]string{
		{"foo": "bar"},
		{"baz": "boz"},
	}

	tf := TemplatedDocument{
		Path:         "",
		matcher:      regexp.MustCompile(`{{\s*([^ }]*)?\s*}}`),
		ValueMapList: mapping,
		Content:      "This is a test file; foo is {{ foo }} baz is {{ baz }}.",
	}
	rv, err := tf.findPlaceholders()
	if err != nil {
		log.Fatal(err)
	}
	if _, ok := rv["foo"]; ! ok {
		log.Fatalf("Expected %q in map. Got %+v", "foo", rv)
	} else if rv["foo"] != "{{ foo }}" {
		log.Fatalf("Expected %q to be %q. Got %+v", "foo", "{{ foo }}", rv)
	}

	if _, ok := rv["baz"]; ! ok {
		log.Fatalf("Expected %q in map. Got %+v", "baz", rv)
	} else if rv["baz"] != "{{ baz }}" {
		log.Fatalf("Expected %q to be %q. Got %+v", "baz", "{{ baz }}", rv)
	}

}

func TestTemplatedDocument_Render(t *testing.T) {
	mapping := []map[string]string{
		{"foo": "A", "bar": "A"},
		{"foo": "B", "bar": "B"},
	}

	tf := TemplatedDocument{
		Path:         "",
		matcher:      regexp.MustCompile(`{{\s*([^ }]*)?\s*}}`),
		ValueMapList: mapping,
		Content:      "foo is {{ foo }} bar is {{ bar }}.",
	}
	contentSlice, err := tf.Render()
	if err != nil {
		log.Fatal(err)
	}

	var exp string
	exp = "foo is A bar is A."
	if contentSlice[0] != exp {
		log.Fatalf("Expected %q, got %q", exp, contentSlice[0])
	}

	exp = "foo is B bar is B."
	if contentSlice[1] != exp {
		log.Fatalf("Expected %q, got %q", exp, contentSlice[1])
	}
}

