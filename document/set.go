package document

// Set of methods to retrieve the configuration files that we want to apply to Vault
type Set interface {
	Get() error		// fetch the configuration
	Path() string	// the path to the configuration documents
	CleanUp()
}

