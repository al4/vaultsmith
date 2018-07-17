package path_handlers

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestReadFile(t *testing.T) {
	var expectStr = "foo"
	ph := &BaseHandler{}
	file, _ := ioutil.TempFile(".", "test-PathHandler-")
	err := ioutil.WriteFile(file.Name(), []byte(expectStr), os.FileMode(int(0664)))
	if err != nil {
		log.Fatalf("Could not create file %s: %s", file.Name(), err)
	}
	defer os.Remove(file.Name())

	data, err := ph.readFile(file.Name())
	if err != nil {
		log.Fatal(err)
	}
	if data != expectStr {
		log.Fatalf("Got %s, expected %s", data, expectStr)
	}
}
