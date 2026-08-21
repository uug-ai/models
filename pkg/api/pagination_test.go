package api

import (
	"encoding/json"
	"strings"
	"testing"
)

// A client that says nothing about the total must not be read as asking for
// one. The field is the difference between a page read and a whole-tenant
// count, so the zero value has to mean "no" on the wire as well as in Go.
func TestCursorPaginationOmitsIncludeTotalWhenUnset(t *testing.T) {
	encoded, err := json.Marshal(CursorPagination{Cursor: "abc", Limit: 25})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(encoded), "includeTotal") {
		t.Errorf("an unset IncludeTotal reached the wire: %s", encoded)
	}

	var decoded CursorPagination
	if err := json.Unmarshal([]byte(`{"cursor":"abc","limit":25}`), &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.IncludeTotal {
		t.Error("a request that omitted includeTotal decoded as asking for a total")
	}
}

func TestCursorPaginationCarriesIncludeTotalWhenRequested(t *testing.T) {
	encoded, err := json.Marshal(CursorPagination{Limit: 25, IncludeTotal: true})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(encoded), `"includeTotal":true`) {
		t.Errorf("IncludeTotal did not reach the wire: %s", encoded)
	}

	var decoded CursorPagination
	if err := json.Unmarshal([]byte(`{"limit":25,"includeTotal":true}`), &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !decoded.IncludeTotal {
		t.Error("includeTotal:true did not decode as asking for a total")
	}
}

// Total is omitempty, so a zero total is indistinguishable on the wire from a
// total that was never requested. That is tolerable only because a caller knows
// whether it asked; this pins the ambiguity rather than leaving it to be
// discovered by a client reading a missing total as "no results".
func TestCursorPaginationOmitsZeroTotal(t *testing.T) {
	encoded, err := json.Marshal(CursorPagination{HasMore: false, Total: 0})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(encoded), "total") {
		t.Errorf("a zero total reached the wire: %s", encoded)
	}
}
