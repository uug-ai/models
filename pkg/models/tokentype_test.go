package models

import "testing"

func TestParseTokenType(t *testing.T) {
	tests := []struct {
		value any
		want  TokenType
		ok    bool
	}{
		{"session", TokenTypeSession, true},
		{"access", TokenTypeAccess, true},
		{"service", TokenTypeService, true},
		{TokenTypeService, TokenTypeService, true},
		{nil, "", false},
		{"", "", false},
		{"Session", "", false},
		{"refresh", "", false},
		{42, "", false},
	}
	for _, test := range tests {
		got, ok := ParseTokenType(test.value)
		if ok != test.ok || (ok && got != test.want) {
			t.Errorf("ParseTokenType(%#v) = %q, %v; want %q, %v", test.value, got, ok, test.want, test.ok)
		}
	}
}

func TestTokenTypeBehaviour(t *testing.T) {
	if TokenTypeSession.HasFixedScopes() || TokenTypeSession.IsBound() {
		t.Error("session tokens must resolve current authorization and not be bound")
	}
	for _, typ := range []TokenType{TokenTypeAccess, TokenTypeService} {
		if !typ.HasFixedScopes() || !typ.IsBound() {
			t.Errorf("%s tokens must have fixed scopes and be bound", typ)
		}
	}
	if TokenType("").Valid() || TokenType("legacy").HasFixedScopes() {
		t.Error("unknown token types must not be valid")
	}
	if len(TokenTypes()) != 3 {
		t.Errorf("TokenTypes() = %v", TokenTypes())
	}
}
