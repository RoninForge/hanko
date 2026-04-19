// Package jsonpointer builds RFC 6901 JSON Pointer strings from mixed
// string and integer segments. Two separate packages (rules and
// validator) need this, and a circular-import-free home keeps the
// implementations from drifting.
package jsonpointer

import (
	"strconv"
	"strings"
)

// Escape implements RFC 6901 section 3 token escaping. Order matters:
// `~` must be escaped first, or the replacement `~1` from `/` → `~1`
// would itself get rewritten.
func Escape(s string) string {
	if !strings.ContainsAny(s, "~/") {
		return s
	}
	s = strings.ReplaceAll(s, "~", "~0")
	s = strings.ReplaceAll(s, "/", "~1")
	return s
}

// Build constructs a JSON Pointer from mixed string (object keys) and int
// (array indices) segments. Empty input yields "" (the root pointer per
// RFC 6901), not "/".
//
// Examples:
//
//	Build("plugins", 2, "name") -> "/plugins/2/name"
//	Build("a/b")                 -> "/a~1b"
//	Build("a~b")                 -> "/a~0b"
func Build(segments ...any) string {
	if len(segments) == 0 {
		return ""
	}
	var b strings.Builder
	for _, s := range segments {
		b.WriteByte('/')
		switch v := s.(type) {
		case string:
			b.WriteString(Escape(v))
		case int:
			b.WriteString(strconv.Itoa(v))
		}
	}
	return b.String()
}
