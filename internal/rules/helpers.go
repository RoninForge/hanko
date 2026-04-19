package rules

import (
	"strconv"

	"github.com/RoninForge/hanko/internal/jsonpointer"
)

// itoa is a short alias for strconv.Itoa used where only a decimal int
// is needed (not a full JSON Pointer).
func itoa(i int) string { return strconv.Itoa(i) }

// jsonPointer is a thin wrapper around jsonpointer.Build that exists so
// rule code can keep calling `jsonPointer(...)` without importing the
// jsonpointer package at every call site. Validator code uses the
// package directly; both routes land on the same implementation.
func jsonPointer(segments ...any) string {
	return jsonpointer.Build(segments...)
}
