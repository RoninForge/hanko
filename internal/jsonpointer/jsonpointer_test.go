package jsonpointer

import "testing"

func TestBuild(t *testing.T) {
	tests := []struct {
		name     string
		segments []any
		want     string
	}{
		{"empty", nil, ""},
		{"single string", []any{"name"}, "/name"},
		{"nested string and index", []any{"plugins", 2, "name"}, "/plugins/2/name"},
		{"escaped slash", []any{"a/b"}, "/a~1b"},
		{"escaped tilde", []any{"a~b"}, "/a~0b"},
		{"tilde before slash ordering", []any{"~/"}, "/~0~1"},
		{"plain key fast path", []any{"foo"}, "/foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Build(tt.segments...); got != tt.want {
				t.Errorf("Build(%v) = %q, want %q", tt.segments, got, tt.want)
			}
		})
	}
}

func TestEscapeFastPath(t *testing.T) {
	// A key with no reserved characters should return the same string
	// (the implementation returns early to avoid any allocation).
	if got := Escape("plain-key"); got != "plain-key" {
		t.Errorf("Escape(plain) = %q, want same string back", got)
	}
}
