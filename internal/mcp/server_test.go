package mcp

import (
	"encoding/json"
	"testing"
)

func TestHandleInitialize(t *testing.T) {
	s := New("1.0.0")
	RegisterAllTools(s)

	req := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}

	resp := s.handle(req)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("expected map result")
	}
	info := result["serverInfo"].(map[string]any)
	if info["name"] != "quill" {
		t.Errorf("expected server name 'quill', got %v", info["name"])
	}
}

func TestHandleToolsList(t *testing.T) {
	s := New("1.0.0")
	RegisterAllTools(s)

	req := Request{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	resp := s.handle(req)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	result := resp.Result.(map[string]any)
	tools := result["tools"].([]map[string]any)
	if len(tools) != 8 {
		t.Errorf("expected 8 tools, got %d", len(tools))
	}

	// Check that quill_bench is registered
	found := false
	for _, tool := range tools {
		if tool["name"] == "quill_bench" {
			found = true
			break
		}
	}
	if !found {
		t.Error("quill_bench tool not found in tools list")
	}
}

func TestHandleToolsCall_Status(t *testing.T) {
	s := New("1.0.0")
	RegisterAllTools(s)

	params, _ := json.Marshal(map[string]any{
		"name":      "quill_status",
		"arguments": map[string]any{},
	})

	req := Request{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params:  params,
	}

	resp := s.handle(req)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	result := resp.Result.(map[string]any)
	content := result["content"].([]map[string]any)
	if len(content) == 0 {
		t.Fatal("expected content in response")
	}
	if content[0]["type"] != "text" {
		t.Errorf("expected text content type, got %v", content[0]["type"])
	}
}

func TestHandleToolsCall_Unknown(t *testing.T) {
	s := New("1.0.0")
	RegisterAllTools(s)

	params, _ := json.Marshal(map[string]any{
		"name":      "quill_nonexistent",
		"arguments": map[string]any{},
	})

	req := Request{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "tools/call",
		Params:  params,
	}

	resp := s.handle(req)
	if resp.Error == nil {
		t.Fatal("expected error for unknown tool")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("expected error code -32602, got %d", resp.Error.Code)
	}
}

func TestHandleUnknownMethod(t *testing.T) {
	s := New("1.0.0")

	req := Request{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "unknown/method",
	}

	resp := s.handle(req)
	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
}
