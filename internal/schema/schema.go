// Package schema exposes the embedded Claude Code plugin and marketplace
// JSON schemas.
//
// The schemas are a synthesis of Anthropic's published plugin docs
// (code.claude.com/docs/en/plugins-reference and plugin-marketplaces),
// the unofficial hesreallyhim/claude-code-json-schema project, and real
// plugin.json files in the wild. They are deliberately a touch stricter
// than the official CLI validator in some places (semver pattern on
// version) and stricter than docs in others (additionalProperties: false
// at the top level). Runtime rules in internal/rules handle cases that
// do not round-trip cleanly through JSON Schema.
package schema

import _ "embed"

//go:embed plugin.schema.json
var pluginSchema []byte

//go:embed marketplace.schema.json
var marketplaceSchema []byte

// Kind identifies which embedded schema a caller wants.
type Kind string

const (
	// KindPlugin is the `.claude-plugin/plugin.json` schema.
	KindPlugin Kind = "plugin"
	// KindMarketplace is the `.claude-plugin/marketplace.json` schema.
	KindMarketplace Kind = "marketplace"
)

// PluginSchema returns the bytes of the plugin.json schema.
func PluginSchema() []byte { return pluginSchema }

// MarketplaceSchema returns the bytes of the marketplace.json schema.
func MarketplaceSchema() []byte { return marketplaceSchema }
