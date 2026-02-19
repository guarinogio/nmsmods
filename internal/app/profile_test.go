package app

import "testing"

func TestValidateProfileName(t *testing.T) {
	ok := []string{"default", "vanilla", "vr_2026", "a.b-c_d"}
	for _, n := range ok {
		if err := ValidateProfileName(n); err != nil {
			t.Fatalf("expected valid: %q, got %v", n, err)
		}
	}
	bad := []string{"", "..", "../x", "name with spaces", "*", "a/", "a\\b"}
	for _, n := range bad {
		if err := ValidateProfileName(n); err == nil {
			t.Fatalf("expected invalid: %q", n)
		}
	}
}
