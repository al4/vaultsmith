package document

// Retrieve the configuration files that we want to apply to Vault
type Set interface {
	Get() error		// fetch the configuration
	Path() string	// the path to the configuration documents
	CleanUp()
}

