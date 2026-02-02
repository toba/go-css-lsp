package lsp

import "testing"

func TestIDUnmarshalJSON_EmptyData(t *testing.T) {
	var id ID
	err := id.UnmarshalJSON([]byte{})
	if err == nil {
		t.Fatal("expected error for empty data")
	}
}

func TestIDUnmarshalJSON_Number(t *testing.T) {
	var id ID
	if err := id.UnmarshalJSON([]byte("42")); err != nil {
		t.Fatal(err)
	}
	if id != 42 {
		t.Fatalf("got %d, want 42", id)
	}
}

func TestIDUnmarshalJSON_QuotedNumber(t *testing.T) {
	var id ID
	if err := id.UnmarshalJSON([]byte(`"7"`)); err != nil {
		t.Fatal(err)
	}
	if id != 7 {
		t.Fatalf("got %d, want 7", id)
	}
}

func TestIDUnmarshalJSON_NonNumeric(t *testing.T) {
	var id ID
	err := id.UnmarshalJSON([]byte(`"abc"`))
	if err == nil {
		t.Fatal("expected error for non-numeric string")
	}
}

func TestMakeInternalError(t *testing.T) {
	out := MakeInternalError("2.0", 1, "something broke")
	if out == nil {
		t.Fatal("expected non-nil response")
	}
	s := string(out)
	if !contains(s, `"code":-32603`) {
		t.Fatalf("expected internal error code, got: %s", s)
	}
	if !contains(s, "something broke") {
		t.Fatalf("expected error message, got: %s", s)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
