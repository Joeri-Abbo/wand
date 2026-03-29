package main

import (
	"encoding/json"
	"testing"
)

func TestItemFilterValue(t *testing.T) {
	i := item("test-server")
	if i.FilterValue() != "test-server" {
		t.Errorf("FilterValue() = %q, want %q", i.FilterValue(), "test-server")
	}
}

func TestItemDelegateHeight(t *testing.T) {
	d := itemDelegate{}
	if d.Height() != 1 {
		t.Errorf("Height() = %d, want 1", d.Height())
	}
}

func TestItemDelegateSpacing(t *testing.T) {
	d := itemDelegate{}
	if d.Spacing() != 0 {
		t.Errorf("Spacing() = %d, want 0", d.Spacing())
	}
}

func TestConfigJSONParsing(t *testing.T) {
	raw := `[{"group":"prod","name":"web01","host":"10.0.0.1","user":"admin","type":"ssh"}]`
	var data []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(data))
	}
	if data[0]["name"] != "web01" {
		t.Errorf("name = %v, want web01", data[0]["name"])
	}
}
