package vault

import (
	vaultApi "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestAuthenticate(t *testing.T) {
	clientConfig := &vaultApi.Config{}
	client, err := vaultApi.NewClient(clientConfig)
	if err != nil {
		log.Fatal(err.Error())

	}
	log.Println(client)
}
