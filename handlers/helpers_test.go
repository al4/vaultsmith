package handlers

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

func TestIsTtlEquivalent(t *testing.T) {
	tests := []struct {
		name string
		ttlA interface{}
		ttlB interface{}
		expected bool
	}{
		{name: "strings", ttlA: "1m", ttlB: "1m", expected: true},
		{name: "ints", ttlA: 60, ttlB: 60, expected: true},
		{name: "string + int", ttlA: "1m", ttlB: 60, expected: true},

		{name: "unequal ints", ttlA: 10, ttlB: 20, expected: false},
		{name: "unequal strings", ttlA: "1m", ttlB: "2m", expected: false},
		{name: "unequal strings + int", ttlA: "1m", ttlB: 120, expected: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rv := IsTtlEquivalent(test.ttlA, test.ttlB)
			if rv != test.expected {
				log.Fatalf("Test case %s failed. Expected %v, got %v. ttlA: %s, ttlB: %s",
					test.name, test.expected, rv, test.ttlA, test.ttlB)
			}

		})
	}
}

func TestIsSliceEquivalent(t *testing.T) {
	tests := []struct {
		name string
		valueA interface{}
		valueB interface{}
		expected bool
	}{
		{ name: "equal str", valueA: "foo", valueB: "foo", expected: true },
		{ name: "equal str arr", valueA: []string{"foo"}, valueB: []string{"foo"}, expected: true },
		{ name: "equal str + array", valueA: "foo", valueB: []string{"foo"}, expected: true },

		{ name: "unequal str", valueA: "foo", valueB: "bar", expected: false },
		{ name: "unequal str arr", valueA: []string{"foo"}, valueB: []string{"bar"}, expected: false },
		{ name: "unequal str + str array", valueA: "foo", valueB: []string{"bar"}, expected: false },

		{ name: "equal int arr", valueA: []int{99}, valueB: []int{99}, expected: true },
		{ name: "unequal int arr", valueA: []int{99}, valueB: []int{1}, expected: false },
		{ name: "unequal int + int arr", valueA: []int{99}, valueB: 1, expected: false },
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rv := IsSliceEquivalent(test.valueA, test.valueB)
			if rv != test.expected {
				log.Fatalf("Test case %q failed. A: %s, B: %s. Expected %b",
					test.name, test.valueA, test.valueB, test.expected)
			}

		})
	}
}

