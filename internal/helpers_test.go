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
func TestConvertAuthConfigConvertsDefaultLeaseTTL(t *testing.T) {
	expected := 70
	in := vaultApi.AuthConfigInput{
		DefaultLeaseTTL: "1m10s",
	}
	out, err := ConvertAuthConfig(in)
	if err != nil {
		log.Fatal(err)
	}
	if out.DefaultLeaseTTL != expected {
		log.Fatalf("Wrong DefaultLeastTTL value %d, expected %d", out.DefaultLeaseTTL, expected)
	}
}

func TestConvertAuthConfigConvertsMaxLeaseTTL(t *testing.T) {
	expected := 70
	in := vaultApi.AuthConfigInput{
		MaxLeaseTTL: "1m10s",
	}
	out, err := ConvertAuthConfig(in)
	if err != nil {
		log.Fatal(err)
	}
	if out.MaxLeaseTTL != expected {
		log.Fatalf("Wrong MaxLeastTTL value %d, expected %d", out.MaxLeaseTTL, expected)
	}
}
