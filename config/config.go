package config

type VaultsmithConfig struct {
	DocumentPath   string
	Dry            bool
	VaultRole      string
	TemplateFile   string
	TemplateParams []string
}
