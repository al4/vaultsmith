package document

// Retrieve the configuration files that we want to apply to Vault
type Set interface {
	Path() string // the path to the configuration documents. Should return nil if not present.
	Get() error   // fetch the configuration documents
	CleanUp()
}
