package vaultClient

import (
	"testing"
	"log"
	vaultApi "github.com/hashicorp/vault/api"
)

func TestAuthenticate(t *testing.T) {
	clientConfig := &vaultApi.Config{}
	client, err := vaultApi.NewClient(clientConfig)
	if err != nil {
		log.Fatal(err.Error())

	}
	log.Println(client)
}
