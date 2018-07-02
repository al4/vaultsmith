package internal

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"io/ioutil"
	"os"
	"github.com/stretchr/testify/assert"
	"log"
)

type PathHandlerTestSuite struct {
	suite.Suite
	handler *BasePathHandler
}

func (suite *PathHandlerTestSuite) SetupTest() {
	ph := &BasePathHandler{}
	suite.handler = ph
}

func (suite *PathHandlerTestSuite) TearDownTest() {
}

func (suite *PathHandlerTestSuite) TestReadFile() {
	file, _ := ioutil.TempFile(".", "test-PathHandler-")
	err := ioutil.WriteFile(file.Name(), []byte("foo"), os.FileMode(int(0664)))
	if err != nil {
		log.Fatalf("Could not create file %s: %s", file.Name(), err)
	}
	defer os.Remove(file.Name())

	data, err := suite.handler.readFile(file.Name())
	if err != nil {
		log.Fatal(err)
	}
	assert.Contains(suite.T(), data, "foo")
}

func TestPathHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(PathHandlerTestSuite))
}
