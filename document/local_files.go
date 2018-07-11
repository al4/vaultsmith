package document


// LocalFiles effectively serves as a dummy implementation, as the rest of vaultsmith is designed
// to operate on a directory of files.

// Implements document.Set
type LocalFiles struct {
	WorkDir	string
	Directory string
}

func (l *LocalFiles) Get() (err error){
	// nothing to do here, they are already on the file system
	return nil
}

// Return the path to the documents
func (l *LocalFiles) Path() (path string){
	return l.Directory
}

func (l *LocalFiles) CleanUp() {
	// NOOP, should probably not remove files that existed before execution
	return
}

