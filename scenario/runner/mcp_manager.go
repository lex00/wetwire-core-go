// Package runner provides a reusable scenario execution engine.
package runner

import (
	"context"
	"fmt"
	"sync"

	"github.com/lex00/wetwire-core-go/mcp"
	"github.com/lex00/wetwire-core-go/scenario"
)

// MCPManager manages MCP server connections for multiple domains.
type MCPManager struct {
	clients map[string]*mcp.Client // domain name -> client
	tools   map[string][]mcp.ToolInfo // domain name -> available tools
	mu      sync.RWMutex
	workDir string
	debug   bool
}

// NewMCPManager creates a new MCP manager.
func NewMCPManager(workDir string, debug bool) *MCPManager {
	return &MCPManager{
		clients: make(map[string]*mcp.Client),
		tools:   make(map[string][]mcp.ToolInfo),
		workDir: workDir,
		debug:   debug,
	}
}

// Start launches MCP servers for all domains in the scenario.
func (m *MCPManager) Start(ctx context.Context, domains []scenario.DomainSpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, domain := range domains {
		if domain.CLI == "" {
			continue // Skip domains without CLI
		}

		// Start MCP server: {cli} mcp
		client, err := mcp.NewClient(ctx, mcp.ClientConfig{
			Command: domain.CLI,
			Args:    []string{"mcp"},
			WorkDir: m.workDir,
			Debug:   m.debug,
		})
		if err != nil {
			// Clean up already started clients
			_ = m.closeAllLocked()
			return fmt.Errorf("failed to start MCP server for domain %s: %w", domain.Name, err)
		}

		m.clients[domain.Name] = client

		// List available tools
		tools, err := client.ListTools(ctx)
		if err != nil {
			_ = m.closeAllLocked()
			return fmt.Errorf("failed to list tools for domain %s: %w", domain.Name, err)
		}
		m.tools[domain.Name] = tools
	}

	return nil
}

// Stop terminates all MCP server connections.
func (m *MCPManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closeAllLocked()
}

// closeAllLocked closes all clients (must be called with lock held).
func (m *MCPManager) closeAllLocked() error {
	var firstErr error
	for name, client := range m.clients {
		if err := client.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("failed to close %s: %w", name, err)
		}
	}
	m.clients = make(map[string]*mcp.Client)
	m.tools = make(map[string][]mcp.ToolInfo)
	return firstErr
}

// GetTools returns available tools for a domain.
func (m *MCPManager) GetTools(domain string) []mcp.ToolInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tools[domain]
}

// GetAllTools returns all tools from all domains, with domain prefix.
// Tools are returned as "{domain}.{tool_name}" to avoid collisions.
func (m *MCPManager) GetAllTools() []mcp.ToolInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var allTools []mcp.ToolInfo
	for domain, tools := range m.tools {
		for _, tool := range tools {
			prefixedTool := mcp.ToolInfo{
				Name:        fmt.Sprintf("%s.%s", domain, tool.Name),
				Description: fmt.Sprintf("[%s] %s", domain, tool.Description),
				InputSchema: tool.InputSchema,
			}
			allTools = append(allTools, prefixedTool)
		}
	}
	return allTools
}

// CallTool invokes a tool on a specific domain.
// The toolName should be the unprefixed tool name (e.g., "wetwire_build").
func (m *MCPManager) CallTool(ctx context.Context, domain, toolName string, arguments map[string]any) (*mcp.ToolCallResult, error) {
	m.mu.RLock()
	client, ok := m.clients[domain]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no MCP client for domain: %s", domain)
	}

	return client.CallTool(ctx, toolName, arguments)
}

// CallPrefixedTool invokes a tool using the prefixed name (e.g., "aws.wetwire_build").
// It parses the prefix to route to the correct domain.
func (m *MCPManager) CallPrefixedTool(ctx context.Context, prefixedName string, arguments map[string]any) (*mcp.ToolCallResult, error) {
	domain, toolName, err := parsePrefixedTool(prefixedName)
	if err != nil {
		return nil, err
	}
	return m.CallTool(ctx, domain, toolName, arguments)
}

// Domains returns the list of connected domain names.
func (m *MCPManager) Domains() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	domains := make([]string, 0, len(m.clients))
	for name := range m.clients {
		domains = append(domains, name)
	}
	return domains
}

// IsConnected returns true if the domain has an active MCP connection.
func (m *MCPManager) IsConnected(domain string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.clients[domain]
	return ok
}

// parsePrefixedTool splits "domain.toolName" into its components.
func parsePrefixedTool(prefixed string) (domain, toolName string, err error) {
	for i := 0; i < len(prefixed); i++ {
		if prefixed[i] == '.' {
			if i == 0 || i == len(prefixed)-1 {
				return "", "", fmt.Errorf("invalid prefixed tool name: %s", prefixed)
			}
			return prefixed[:i], prefixed[i+1:], nil
		}
	}
	return "", "", fmt.Errorf("tool name must be prefixed with domain: %s (expected domain.toolName)", prefixed)
}
