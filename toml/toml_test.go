package toml_test

import (
	"testing"

	itoml "github.com/fkasper/sitrep-authentication/toml"
)

// Ensure that megabyte sizes can be parsed.
func TestSize_UnmarshalText_MB(t *testing.T) {
	var s itoml.Size
	if err := s.UnmarshalText([]byte("200m")); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if s != 200*(1<<20) {
		t.Fatalf("unexpected size: %d", s)
	}
}

// Ensure that gigabyte sizes can be parsed.
func TestSize_UnmarshalText_GB(t *testing.T) {
	var s itoml.Size
	if err := s.UnmarshalText([]byte("1g")); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if s != 1073741824 {
		t.Fatalf("unexpected size: %d", s)
	}
}
