package document

import (
	"log"
	"path/filepath"
	"testing"
)

func TestGenerateTemplateParams(t *testing.T) {
	filePath := filepath.Join(examplePath(), "template.json")
	tp, err := GenerateTemplateParams(filePath, []string{})
	if err != nil {
		t.Errorf("GenerateTemplateParams returned err: %s", err)
	}
	if len(tp) != 4 {
		t.Error("Expected array of 4 TemplateParams (based on template.json)")
	}
}

func TestGenerateTemplateParams_returnsMultipleTemplateConfigs(t *testing.T) {
	filePath := filepath.Join(examplePath(), "template.json")
	templateParams, err := GenerateTemplateParams(filePath, []string{"option=overridden"})
	if err != nil {
		t.Errorf("GenerateTemplateParams returned err: %s", err)
	}
	log.Printf("%+v", templateParams)
	for _, tp := range templateParams {
		// All objects should have the overridden parameter set
		if tp.Variables["option"] != "overridden" {
			t.Errorf("TemplateParams object %q has option %q, should be %q",
				tp.Name, tp.Variables["option"], "overridden")
		}
	}
}

func TestSetParams(t *testing.T) {
	r := setParams(TemplateParams{Name: "test"}, []string{})
	if r.Name != "test" {
		t.Errorf("Fail")
	}
}

func TestSetParams_addsParam(t *testing.T) {
	r := setParams(
		TemplateParams{
			Name:      "test",
			Variables: map[string]string{},
		}, []string{"foo=bar"})
	if r.Variables["foo"] != "bar" {
		t.Errorf("Expected 'bar', got %q", r.Variables["foo"])
	}
}

// not sure we even need this ¯\_(ツ)_/¯
func TestSetParams_addsParamMultipleEquals(t *testing.T) {
	r := setParams(
		TemplateParams{
			Name:      "test",
			Variables: map[string]string{},
		}, []string{"foo=bar=boz"})
	if r.Variables["foo"] != "bar=boz" {
		t.Errorf("Expected 'bar=boz', got %q", r.Variables["foo"])
	}
}
