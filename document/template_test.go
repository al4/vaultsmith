package document

import (
	log "github.com/sirupsen/logrus"
	"os"
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

func TestTemplatedDocument_findPlaceholders(t *testing.T) {
	tf := Template{}
	rv, err := tf.findPlaceholders("This is a test file; foo is {{ foo }} baz is {{ baz }}.")
	if err != nil {
		log.Fatal(err)
	}
	if _, ok := rv["foo"]; !ok {
		t.Errorf("Expected %q in map. Got %+v", "foo", rv)
	} else if rv["foo"] != "{{ foo }}" {
		t.Errorf("Expected %q to be %q. Got %+v", "foo", "{{ foo }}", rv)
	}

	if _, ok := rv["baz"]; !ok {
		t.Errorf("Expected %q in map. Got %+v", "baz", rv)
	} else if rv["baz"] != "{{ baz }}" {
		t.Errorf("Expected %q to be %q. Got %+v", "baz", "{{ baz }}", rv)
	}

}

func TestTemplatedDocument_Render(t *testing.T) {
	mapping := TemplateParams{
		Variables: map[string]string{
			"foo": "A", "bar": "B",
		},
	}

	tf := Template{
		Params:  mapping,
		Content: "foo is {{ foo }} bar is {{ bar }}.",
	}
	renderedTemplates, err := tf.Render()
	if err != nil {
		log.Fatal(err)
	}
	var exp string
	exp = "foo is A bar is B."
	if renderedTemplates[0].Content != exp {
		t.Errorf("Expected %q, got %q", exp, renderedTemplates[0].Content)
	}
}

func TestTemplatedDocument_Render_MultipleFoo(t *testing.T) {
	mapping := TemplateParams{
		Variables: map[string]string{
			"foo": "A", "bar": "B",
		},
	}

	tf := Template{
		Params:  mapping,
		Content: "foo is {{ foo }} bar is {{ bar }}. And foo is {{ foo }}",
	}
	renderedTemplates, err := tf.Render()
	if err != nil {
		log.Fatal(err)
	}

	var exp string
	exp = "foo is A bar is B. And foo is A"
	if renderedTemplates[0].Content != exp {
		t.Errorf("Expected %q, got %q", exp, renderedTemplates[0].Content)
	}
}

// Test that we render the templated filename
func TestTemplatedDocument_Render_FileName(t *testing.T) {
	mapping := TemplateParams{
		Instances: map[string][]string{
			"read_service": {"reader1", "reader2"},
		},
		Variables: map[string]string{
			"foo": "A", "bar": "B",
		},
	}
	tf := Template{
		FileName: "{{ read_service }}",
		Params:   mapping,
		Content:  "foo is {{ foo }} bar is {{ bar }}. And foo is {{ foo }}",
	}
	renderedTemplates, err := tf.Render()
	if err != nil {
		log.Fatal(err)
	}

	var exp string
	exp = "reader1"
	if renderedTemplates[0].Name != exp {
		t.Errorf("Expected %q, got %q", exp, renderedTemplates[0].Name)
	}
}
