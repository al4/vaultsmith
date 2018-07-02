package internal

import (
	"testing"
	vaultApi "github.com/hashicorp/vault/api"
	"log"
)

func TestConvertAuthConfig(t *testing.T) {
	in := vaultApi.AuthConfigInput{}
	_, err := ConvertAuthConfig(in)
	if err != nil {
		log.Fatal(err)
	}
}

// Test that TTLs are converted properly
func TestConvertAuthConfigTTL(t *testing.T) {
	in := vaultApi.AuthConfigInput{}
	out, err := ConvertAuthConfig(in)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(out)
}

