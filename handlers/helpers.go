package handlers

import (
	"github.com/hashicorp/vault/api"
	"time"
	"fmt"
	"log"
	"reflect"
	"encoding/json"
)

// TODO this is small for now, but figure out how to organise these functions

// convert AuthConfigInput type to AuthConfigOutput type
// A potential problem with this is that the transformation doesn't use the same code that Vault
// uses internally, so bugs are possible; but ParseDuration is pretty standard (and vault
// does use this same method)
func ConvertAuthConfig(input api.AuthConfigInput) (api.AuthConfigOutput, error) {
	var output api.AuthConfigOutput
	var dur time.Duration
	var err error

	var DefaultLeaseTTL int // was string

	if input.DefaultLeaseTTL != "" {
		dur, err = time.ParseDuration(input.DefaultLeaseTTL)
		if err != nil {
			return output, fmt.Errorf("could not parse DefaultLeaseTTL value %s as seconds: %s", input.DefaultLeaseTTL, err)
		}
		DefaultLeaseTTL = int(dur.Seconds())
	}

	var MaxLeaseTTL int // was string
	if input.MaxLeaseTTL != "" {
		dur, err = time.ParseDuration(input.MaxLeaseTTL)
		if err != nil {
			return output, fmt.Errorf("could not parse MaxLeaseTTL value %s as seconds: %s", input.MaxLeaseTTL, err)
		}
		MaxLeaseTTL = int(dur.Seconds())
	}

	output = api.AuthConfigOutput{
		DefaultLeaseTTL:           DefaultLeaseTTL,
		MaxLeaseTTL:               MaxLeaseTTL,
		PluginName:                input.PluginName,
		AuditNonHMACRequestKeys:   input.AuditNonHMACRequestKeys,
		AuditNonHMACResponseKeys:  input.AuditNonHMACResponseKeys,
		ListingVisibility:         input.ListingVisibility,
		PassthroughRequestHeaders: input.PassthroughRequestHeaders,
	}

	return output, nil
}

// Determine whether a string ttl is equal to an int ttl
func IsTtlEquivalent(ttlA interface{}, ttlB interface{}) bool {
	durA, err := convertToDuration(ttlA)
	if err != nil {
		log.Printf("WARN: Error parsing %+v: %s", ttlA, err)
		return false
	}
	durB, err := convertToDuration(ttlB)
	if err != nil {
		log.Printf("WARN: Error converting %+v to duration: %s", ttlA, err)
		return false
	}

	if durA == durB {
		return true
	}

	return false
}

// convert x to time.Duration. if x is an integer, we assume it is in seconds
func convertToDuration(x interface{}) (time.Duration, error) {
	var duration time.Duration
	var err error

	switch x.(type) {
	case string:
		duration, err = time.ParseDuration(x.(string))
		if err != nil {
			return 0, fmt.Errorf("%q can't be parsed as duration", x)
		}
	case int64:
		duration = time.Duration(x.(int64)) * time.Second
	case int:
		duration = time.Duration(int64(x.(int))) * time.Second
	case json.Number:
		i, err := x.(json.Number).Int64()
		if err != nil {
			return 0, fmt.Errorf("could not parse %+v as json number: %s", x, err.Error())
		}
		duration = time.Duration(i) * time.Second
	default:
		return 0, fmt.Errorf("type of '%+v' not handled", reflect.TypeOf(x))
	}

	return duration, nil

}

// Determine whether an array is logically equivalent as far as Vault is concerned.
// e.g. [policy] == policy
func IsSliceEquivalent(a interface{}, b interface{}) (equivalent bool) {
	if reflect.TypeOf(a).Kind() == reflect.TypeOf(b).Kind() {
		// just compare directly if type is the same
		return reflect.DeepEqual(a, b)
	}

	if reflect.TypeOf(a).Kind() == reflect.Slice {
		// b must not be a slice, compare a[0] to it
		return firstElementEqual(a, b)
	}

	if reflect.TypeOf(b).Kind() == reflect.Slice {
		// a must not be a slice, compare b[0] to it
		return firstElementEqual(b, a)
	}

	return false
}

// Return true if value is equal to the first item in slice
func firstElementEqual(slice interface{}, value interface{}) bool {
	switch t := slice.(type) {
	case []string:
		if t[0] == value && len(t) == 1 {
			return true
		}
	case []int:
		if t[0] == value && len(t) == 1 {
			return true
		}
	case []interface{}:
		s := reflect.ValueOf(t)
		var val interface{}
		for i := 0; i < s.Len(); i++ {
			if i > 0 { // length > 1, cannot be equivalent
				return false
			}
			if i == 0 {
				// This is a little scary in a strongly typed context, as we're parsing everything
				// as a string. But in the context of vault API responses it should be OK...
				val = fmt.Sprintf("%v", s.Index(i))
			}
		}
		if val == value {
			return true
		}
	default:
		log.Fatalf("Unhandled type %T, please add this to the switch statement", t)
	}

	return false
}

