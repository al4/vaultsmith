package document

import (
	"log"
	"path/filepath"
	"testing"
)

func TestGenerateTemplateParams(t *testing.T) {
	filePath := filepath.Join(examplePath(), "_vaultsmith.json")
	tp, err := GenerateTemplateParams(filePath, []string{})
	if err != nil {
		t.Errorf("GenerateTemplateParams returned err: %s", err)
	}
	log.Printf("%+v", tp)
}

func TestGenerateTemplateParams_returnsMultipleTemplateConfigs(t *testing.T) {
	filePath := filepath.Join(examplePath(), "_vaultsmith.json")
	templateParams, err := GenerateTemplateParams(filePath, []string{"option=overridden"})
	if err != nil {
		t.Errorf("GenerateTemplateParams returned err: %s", err)
	}
	log.Printf("%+v", templateParams)
	if templateParams.Variables["option"] != "overridden" {
		t.Errorf("TemplateParams object %q has option %q, should be %q",
			templateParams.Instances, templateParams.Variables["option"], "overridden")
	}
}

func TestSetParams_addsParam(t *testing.T) {
	r := setParams(
		TemplateParams{
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
			Variables: map[string]string{},
		}, []string{"foo=bar=boz"})
	if r.Variables["foo"] != "bar=boz" {
		t.Errorf("Expected 'bar=boz', got %q", r.Variables["foo"])
	}
}
