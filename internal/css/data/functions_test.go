package data

import (
	"testing"
)

func TestLookupFunction(t *testing.T) {
	fn := LookupFunction("rgb")
	if fn == nil {
		t.Fatal("expected to find rgb function")
	}
	if len(fn.Signatures) == 0 {
		t.Error("expected rgb to have signatures")
	}
	if fn.MDN == "" {
		t.Error("expected rgb to have MDN URL")
	}
	if fn.Description == "" {
		t.Error("expected rgb to have description")
	}
}

func TestLookupFunctionUnknown(t *testing.T) {
	fn := LookupFunction("notafunction")
	if fn != nil {
		t.Error("expected nil for unknown function")
	}
}

func TestAllFunctionsCoverage(t *testing.T) {
	all := make(map[string]bool)
	for _, fn := range ColorFunctions {
		all[fn] = false
	}
	for _, fn := range CommonFunctions {
		all[fn] = false
	}

	for _, fn := range Functions {
		all[fn.Name] = true
	}

	for name, found := range all {
		if !found {
			t.Errorf(
				"function %q in ColorFunctions/CommonFunctions "+
					"has no Function entry", name,
			)
		}
	}
}

func TestAllFunctionsHaveSignatures(t *testing.T) {
	for _, fn := range Functions {
		if len(fn.Signatures) == 0 {
			t.Errorf(
				"function %q has no signatures", fn.Name,
			)
		}
		if fn.Description == "" {
			t.Errorf(
				"function %q has no description", fn.Name,
			)
		}
	}
}
