package models

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestPromptUnmarshalJSON(t *testing.T) {
	input := `{"name":"test","content":"c","category":"x","order_num":0,"fields":[],"subcategories":[],"card_type":"normal","password":null}`
	var p Prompt
	if err := json.Unmarshal([]byte(input), &p); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if p.Name != "test" {
		t.Errorf("Name = %q, want %q", p.Name, "test")
	}
	if p.Fields == nil || *p.Fields != "[]" {
		t.Errorf("Fields = %v, want '[]'", p.Fields)
	}
	if p.Subcategories == nil || *p.Subcategories != "[]" {
		t.Errorf("Subcategories = %v, want '[]'", p.Subcategories)
	}
	if p.Password != nil {
		t.Errorf("Password = %v, want nil", p.Password)
	}
	t.Logf("OK: %+v", p)
}

func TestBookUnmarshalJSON(t *testing.T) {
	input := `{"title":"测试","description":"简介","tags":["玄幻","仙侠"],"status":"draft","is_public":false}`
	var b Book
	if err := json.Unmarshal([]byte(input), &b); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	wantTags := StringList{"玄幻", "仙侠"}
	if !slices.Equal(b.Tags, wantTags) {
		t.Errorf("Tags = %v, want %v", b.Tags, wantTags)
	}
	if b.IsPublic != 0 {
		t.Errorf("IsPublic = %d, want 0", b.IsPublic)
	}
	t.Logf("OK: %+v", b)
}
