package document

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// calculate path to test fixtures (example/)
func examplePath() string {
	wd, _ := os.Getwd()
	pathArray := strings.Split(wd, string(os.PathSeparator))
	pathArray = pathArray[:len(pathArray)-1] // trims "internal"
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
		{Name: "foo", Variables: map[string]string{"foo": "bar"}},
		{Name: "baz", Variables: map[string]string{"baz": "boz"}},
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
	if _, ok := rv["foo"]; !ok {
		log.Fatalf("Expected %q in map. Got %+v", "foo", rv)
	} else if rv["foo"] != "{{ foo }}" {
		log.Fatalf("Expected %q to be %q. Got %+v", "foo", "{{ foo }}", rv)
	}

	if _, ok := rv["baz"]; !ok {
		log.Fatalf("Expected %q in map. Got %+v", "baz", rv)
	} else if rv["baz"] != "{{ baz }}" {
		log.Fatalf("Expected %q to be %q. Got %+v", "baz", "{{ baz }}", rv)
	}

}

func TestTemplatedDocument_Render(t *testing.T) {
	mapping := []TemplateConfig{
		{Name: "one", Variables: map[string]string{"foo": "A", "bar": "A"}},
		{Name: "two", Variables: map[string]string{"foo": "B", "bar": "B"}},
	}

	tf := Template{
		Path:         "",
		matcher:      regexp.MustCompile(`{{\s*([^ }]*)?\s*}}`),
		ValueMapList: mapping,
		Content:      "foo is {{ foo }} bar is {{ bar }}.",
	}
	renderedTemplates, err := tf.Render()
	if err != nil {
		log.Fatal(err)
	}

	var exp string
	exp = "foo is A bar is A."
	if renderedTemplates[0].Content != exp {
		log.Fatalf("Expected %q, got %q", exp, renderedTemplates[0].Content)
	}

	exp = "foo is B bar is B."
	if renderedTemplates[1].Content != exp {
		log.Fatalf("Expected %q, got %q", exp, renderedTemplates[1].Content)
	}
}

func TestTemplatedDocument_Render_MultipleFoo(t *testing.T) {
	mapping := []TemplateConfig{
		{Name: "test", Variables: map[string]string{"foo": "A", "bar": "A"}},
	}

	tf := Template{
		Path:         "",
		matcher:      regexp.MustCompile(`{{\s*([^ }]*)?\s*}}`),
		ValueMapList: mapping,
		Content:      "foo is {{ foo }} bar is {{ bar }}. And foo is {{ foo }}",
	}
	renderedTemplates, err := tf.Render()
	if err != nil {
		log.Fatal(err)
	}

	var exp string
	exp = "foo is A bar is A. And foo is A"
	if renderedTemplates[0].Content != exp {
		log.Fatalf("Expected %q, got %q", exp, renderedTemplates[0].Content)
	}
}

// When there are identical documents we should not duplicate
func TestTemplate_Render_DoesNotDuplicate(t *testing.T) {
	mapping := []TemplateConfig{
		{Name: "one", Variables: map[string]string{"foo": "A"}},
		{Name: "two", Variables: map[string]string{"foo": "A"}},
	}

	tf := Template{
		Path:         "",
		matcher:      regexp.MustCompile(`{{\s*([^ }]*)?\s*}}`),
		ValueMapList: mapping,
		Content:      "foo is {{ foo }} bar is {{ bar }}.",
	}
	renderedTemplates, err := tf.Render()
	if err != nil {
		log.Fatal(err)
	}
	if len(renderedTemplates) != 1 {
		log.Printf("%+v", renderedTemplates)
		log.Fatalf("Expected 1 entry in rendered templates, got %v", len(renderedTemplates))
	}
}

func TestTemplate_hasMultiple_false(t *testing.T) {
	mapping := []TemplateConfig{
		{Name: "one", Variables: map[string]string{"foo": "A"}},
		{Name: "two", Variables: map[string]string{"foo": "A"}},
	}

	tf := Template{
		Path:         "",
		matcher:      regexp.MustCompile(`{{\s*([^ }]*)?\s*}}`),
		ValueMapList: mapping,
		Content:      "foo is {{ foo }} bar is {{ bar }}.",
	}

	rv, err := tf.hasMultiple()
	if err != nil {
		log.Fatalln(err)
	} else if rv == true {
		log.Fatalln("Expected hasMultiple call to be false")
	}

}

func TestTemplate_hasMultiple_true(t *testing.T) {
	mapping := []TemplateConfig{
		{Name: "one", Variables: map[string]string{"foo": "A"}},
		{Name: "two", Variables: map[string]string{"foo": "B"}},
	}

	tf := Template{
		Path:         "",
		matcher:      regexp.MustCompile(`{{\s*([^ }]*)?\s*}}`),
		ValueMapList: mapping,
		Content:      "foo is {{ foo }} bar is {{ bar }}.",
	}

	rv, err := tf.hasMultiple()
	if err != nil {
		log.Fatalln(err)
	} else if rv == false {
		log.Fatalln("Expected hasMultiple call to be true")
	}

}
