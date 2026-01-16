package agents

import (
	"context"

	"github.com/lex00/wetwire-core-go/mcp"
)

// MCPServerAdapter adapts mcp.Server to the MCPServer interface.
// This allows mcp.Server to be used directly with the unified Agent.
type MCPServerAdapter struct {
	server *mcp.Server
}

// NewMCPServerAdapter creates an adapter that wraps an mcp.Server.
func NewMCPServerAdapter(server *mcp.Server) *MCPServerAdapter {
	return &MCPServerAdapter{server: server}
}

// ExecuteTool executes a tool via the underlying MCP server.
func (a *MCPServerAdapter) ExecuteTool(ctx context.Context, name string, args map[string]any) (string, error) {
	return a.server.ExecuteTool(ctx, name, args)
}

// GetTools returns the list of tools from the MCP server.
func (a *MCPServerAdapter) GetTools() []MCPToolInfo {
	mcpTools := a.server.GetTools()
	tools := make([]MCPToolInfo, len(mcpTools))
	for i, t := range mcpTools {
		tools[i] = MCPToolInfo{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		}
	}
	return tools
}
