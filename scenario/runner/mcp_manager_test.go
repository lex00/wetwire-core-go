package runner

import (
	"context"
	"testing"
)

func TestParsePrefixedTool(t *testing.T) {
	tests := []struct {
		name       string
		prefixed   string
		wantDomain string
		wantTool   string
		wantErr    bool
	}{
		{
			name:       "valid domain-a tool",
			prefixed:   "domain-a.wetwire_build",
			wantDomain: "domain-a",
			wantTool:   "wetwire_build",
		},
		{
			name:       "valid domain-b tool",
			prefixed:   "domain-b.wetwire_lint",
			wantDomain: "domain-b",
			wantTool:   "wetwire_lint",
		},
		{
			name:       "valid tool with underscore in name",
			prefixed:   "domain-a.wetwire_init_package",
			wantDomain: "domain-a",
			wantTool:   "wetwire_init_package",
		},
		{
			name:     "missing prefix",
			prefixed: "wetwire_build",
			wantErr:  true,
		},
		{
			name:     "empty domain",
			prefixed: ".wetwire_build",
			wantErr:  true,
		},
		{
			name:     "empty tool",
			prefixed: "domain-a.",
			wantErr:  true,
		},
		{
			name:     "empty string",
			prefixed: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain, toolName, err := parsePrefixedTool(tt.prefixed)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parsePrefixedTool(%q) expected error, got nil", tt.prefixed)
				}
				return
			}
			if err != nil {
				t.Errorf("parsePrefixedTool(%q) unexpected error: %v", tt.prefixed, err)
				return
			}
			if domain != tt.wantDomain {
				t.Errorf("parsePrefixedTool(%q) domain = %q, want %q", tt.prefixed, domain, tt.wantDomain)
			}
			if toolName != tt.wantTool {
				t.Errorf("parsePrefixedTool(%q) toolName = %q, want %q", tt.prefixed, toolName, tt.wantTool)
			}
		})
	}
}

func TestNewMCPManager(t *testing.T) {
	mgr := NewMCPManager("/tmp/test", false)
	if mgr == nil {
		t.Fatal("NewMCPManager returned nil")
	}
	if mgr.workDir != "/tmp/test" {
		t.Errorf("workDir = %q, want %q", mgr.workDir, "/tmp/test")
	}
	if mgr.debug {
		t.Error("debug should be false")
	}
	if len(mgr.clients) != 0 {
		t.Error("clients should be empty initially")
	}
	if len(mgr.tools) != 0 {
		t.Error("tools should be empty initially")
	}
}

func TestMCPManager_NoClientsInitially(t *testing.T) {
	mgr := NewMCPManager("/tmp/test", false)

	// Check no domains connected
	domains := mgr.Domains()
	if len(domains) != 0 {
		t.Errorf("Domains() = %v, want empty", domains)
	}

	// Check IsConnected returns false
	if mgr.IsConnected("domain-a") {
		t.Error("IsConnected(domain-a) should be false")
	}

	// Check GetTools returns nil
	tools := mgr.GetTools("domain-a")
	if tools != nil {
		t.Errorf("GetTools(domain-a) = %v, want nil", tools)
	}

	// Check GetAllTools returns empty
	allTools := mgr.GetAllTools()
	if len(allTools) != 0 {
		t.Errorf("GetAllTools() = %v, want empty", allTools)
	}
}

func TestMCPManager_CallToolNotConnected(t *testing.T) {
	mgr := NewMCPManager("/tmp/test", false)

	// Calling tool on non-existent domain should error
	_, err := mgr.CallTool(context.Background(), "domain-a", "wetwire_build", nil)
	if err == nil {
		t.Error("CallTool on non-existent domain should error")
	}
}
