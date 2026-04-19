package schema

import (
	"encoding/json"
	"testing"
)

func TestEmbedNotEmpty(t *testing.T) {
	if len(PluginSchema()) == 0 {
		t.Fatal("plugin.schema.json was not embedded")
	}
	if len(MarketplaceSchema()) == 0 {
		t.Fatal("marketplace.schema.json was not embedded")
	}
}

func TestEmbedsAreValidJSON(t *testing.T) {
	var v any
	if err := json.Unmarshal(PluginSchema(), &v); err != nil {
		t.Errorf("plugin.schema.json is not valid JSON: %v", err)
	}
	if err := json.Unmarshal(MarketplaceSchema(), &v); err != nil {
		t.Errorf("marketplace.schema.json is not valid JSON: %v", err)
	}
}
