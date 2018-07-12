package document

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
	_, err := NewTemplate(
		filepath.Join(examplePath(), "sys/policy/read_service.json"),
		filepath.Join(examplePath(), "template.json"))
	if err != nil {
		log.Fatal(err)
	}
}

func TestTemplatedDocument_findPlaceholders(t *testing.T) {
	mapping := []TemplateConfig{
		{ Name: "foo", Variables: map[string]string{"foo": "bar"} },
		{ Name: "baz", Variables: map[string]string{"baz": "boz"} },
	}

	tf := Template{
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
	mapping := []TemplateConfig{
		{ Name: "one", Variables: map[string]string{"foo": "A", "bar": "A"} },
		{ Name: "two", Variables: map[string]string{"foo": "B", "bar": "B"} },
	}

	tf := Template{
		Path:         "",
		matcher:      regexp.MustCompile(`{{\s*([^ }]*)?\s*}}`),
		ValueMapList: mapping,
		Content:      "foo is {{ foo }} bar is {{ bar }}.",
	}
	contentMap, err := tf.Render()
	if err != nil {
		log.Fatal(err)
	}

	var exp string
	exp = "foo is A bar is A."
	if contentMap["one"] != exp {
		log.Fatalf("Expected %q, got %q", exp, contentMap["one"])
	}

	exp = "foo is B bar is B."
	if contentMap["two"] != exp {
		log.Fatalf("Expected %q, got %q", exp, contentMap["two"])
	}
}

func TestTemplatedDocument_Render_MultipleFoo(t *testing.T) {
	mapping := []TemplateConfig{
		{ Name: "test", Variables: map[string]string{"foo": "A", "bar": "A"} },
	}

	tf := Template{
		Path:         "",
		matcher:      regexp.MustCompile(`{{\s*([^ }]*)?\s*}}`),
		ValueMapList: mapping,
		Content:      "foo is {{ foo }} bar is {{ bar }}. And foo is {{ foo }}",
	}
	contentMap, err := tf.Render()
	if err != nil {
		log.Fatal(err)
	}

	var exp string
	exp = "foo is A bar is A. And foo is A"
	if contentMap["test"] != exp {
		log.Fatalf("Expected %q, got %q", exp, contentMap["test"])
	}
}
