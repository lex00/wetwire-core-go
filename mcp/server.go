// Package mcp provides MCP (Model Context Protocol) server infrastructure
// for exposing tools to Claude Code and other MCP clients.
//
// This package is the server-side complement to the kiro package, which
// handles launching agents that connect to MCP servers.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

// ToolHandler processes tool invocations and returns results.
type ToolHandler func(ctx context.Context, args map[string]any) (string, error)

// Tool represents a registered tool that can be invoked by MCP clients.
type Tool struct {
	Name        string
	Description string
	Handler     ToolHandler
	InputSchema map[string]any // JSON Schema for input parameters
}

// Config configures the MCP server.
type Config struct {
	// Name is the server name (e.g., "wetwire-azure")
	Name string

	// Version is the server version (e.g., "1.0.0")
	Version string

	// Debug enables debug logging to stderr
	Debug bool
}

// Server implements the MCP protocol over stdio.
type Server struct {
	config Config
	tools  map[string]*Tool
	mu     sync.RWMutex

	// For testing - allow injection of reader/writer
	reader io.Reader
	writer io.Writer
}

// NewServer creates a new MCP server with the given configuration.
func NewServer(config Config) *Server {
	return &Server{
		config: config,
		tools:  make(map[string]*Tool),
		reader: os.Stdin,
		writer: os.Stdout,
	}
}

// RegisterTool adds a tool that MCP clients can invoke.
func (s *Server) RegisterTool(name, description string, handler ToolHandler) {
	s.RegisterToolWithSchema(name, description, handler, nil)
}

// RegisterToolWithSchema adds a tool with a JSON Schema for input validation.
func (s *Server) RegisterToolWithSchema(name, description string, handler ToolHandler, inputSchema map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tools[name] = &Tool{
		Name:        name,
		Description: description,
		Handler:     handler,
		InputSchema: inputSchema,
	}

	s.debugf("Registered tool: %s", name)
}

// Name returns the server name.
func (s *Server) Name() string {
	return s.config.Name
}

// Start begins listening for MCP requests on stdio.
// This method blocks until the connection is closed or an error occurs.
func (s *Server) Start(ctx context.Context) error {
	s.debugf("Starting MCP server: %s", s.config.Name)

	scanner := bufio.NewScanner(s.reader)
	// Set a larger buffer for potentially large JSON-RPC messages
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		s.debugf("Received: %s", line)

		response, err := s.handleMessage(ctx, []byte(line))
		if err != nil {
			s.debugf("Error handling message: %v", err)
			continue
		}

		if response != nil {
			responseBytes, err := json.Marshal(response)
			if err != nil {
				s.debugf("Error marshaling response: %v", err)
				continue
			}
			if _, err := fmt.Fprintln(s.writer, string(responseBytes)); err != nil {
				s.debugf("Error writing response: %v", err)
				continue
			}
			s.debugf("Sent: %s", string(responseBytes))
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

// handleMessage processes a single JSON-RPC message.
func (s *Server) handleMessage(ctx context.Context, data []byte) (*JSONRPCResponse, error) {
	var req JSONRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    ParseError,
				Message: "Parse error",
			},
			ID: nil,
		}, nil
	}

	s.debugf("Handling method: %s", req.Method)

	switch req.Method {
	case "initialize":
		return s.handleInitialize(&req)
	case "tools/list":
		return s.handleToolsList(&req)
	case "tools/call":
		return s.handleToolsCall(ctx, &req)
	case "notifications/initialized":
		// Notification, no response needed
		return nil, nil
	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    MethodNotFound,
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			},
			ID: req.ID,
		}, nil
	}
}

// handleInitialize handles the initialize request.
func (s *Server) handleInitialize(req *JSONRPCRequest) (*JSONRPCResponse, error) {
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		ServerInfo: ServerInfo{
			Name:    s.config.Name,
			Version: s.config.Version,
		},
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{},
		},
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      req.ID,
	}, nil
}

// handleToolsList handles the tools/list request.
func (s *Server) handleToolsList(req *JSONRPCRequest) (*JSONRPCResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]ToolInfo, 0, len(s.tools))
	for _, tool := range s.tools {
		info := ToolInfo{
			Name:        tool.Name,
			Description: tool.Description,
		}
		if tool.InputSchema != nil {
			info.InputSchema = tool.InputSchema
		} else {
			// Default empty schema
			info.InputSchema = map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			}
		}
		tools = append(tools, info)
	}

	result := ToolsListResult{
		Tools: tools,
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      req.ID,
	}, nil
}

// handleToolsCall handles the tools/call request.
func (s *Server) handleToolsCall(ctx context.Context, req *JSONRPCRequest) (*JSONRPCResponse, error) {
	var params ToolCallParams
	if req.Params != nil {
		paramsBytes, err := json.Marshal(req.Params)
		if err != nil {
			return &JSONRPCResponse{
				JSONRPC: "2.0",
				Error: &JSONRPCError{
					Code:    InvalidParams,
					Message: "Invalid params",
				},
				ID: req.ID,
			}, nil
		}
		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			return &JSONRPCResponse{
				JSONRPC: "2.0",
				Error: &JSONRPCError{
					Code:    InvalidParams,
					Message: "Invalid params structure",
				},
				ID: req.ID,
			}, nil
		}
	}

	s.mu.RLock()
	tool, exists := s.tools[params.Name]
	s.mu.RUnlock()

	if !exists {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    InvalidParams,
				Message: fmt.Sprintf("Tool not found: %s", params.Name),
			},
			ID: req.ID,
		}, nil
	}

	// Execute the tool handler
	result, err := tool.Handler(ctx, params.Arguments)
	if err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Result: ToolCallResult{
				Content: []ContentBlock{
					{
						Type: "text",
						Text: fmt.Sprintf("Error: %v", err),
					},
				},
				IsError: true,
			},
			ID: req.ID,
		}, nil
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Result: ToolCallResult{
			Content: []ContentBlock{
				{
					Type: "text",
					Text: result,
				},
			},
			IsError: false,
		},
		ID: req.ID,
	}, nil
}

// ExecuteTool executes a tool directly without going through stdio.
// This enables in-process tool execution for agent workflows.
func (s *Server) ExecuteTool(ctx context.Context, name string, args map[string]any) (string, error) {
	s.mu.RLock()
	tool, exists := s.tools[name]
	s.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("tool not found: %s", name)
	}

	return tool.Handler(ctx, args)
}

// GetTools returns the list of registered tools for provider integration.
// This converts MCP tools to the ToolInfo format for the Agent interface.
func (s *Server) GetTools() []ToolInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]ToolInfo, 0, len(s.tools))
	for _, tool := range s.tools {
		info := ToolInfo{
			Name:        tool.Name,
			Description: tool.Description,
		}
		if tool.InputSchema != nil {
			info.InputSchema = tool.InputSchema
		} else {
			// Default empty schema
			info.InputSchema = map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			}
		}
		tools = append(tools, info)
	}

	return tools
}

// debugf logs a debug message if debug mode is enabled.
func (s *Server) debugf(format string, args ...any) {
	if s.config.Debug || os.Getenv("WETWIRE_MCP_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[MCP:%s] "+format+"\n", append([]any{s.config.Name}, args...)...)
	}
}

// GetInstallInstructions returns Claude Code configuration instructions.
func GetInstallInstructions(serverName, binaryName string) string {
	return fmt.Sprintf(`To use %s with Claude Code, add the following to your Claude Code settings:

{
  "mcpServers": {
    "%s": {
      "command": "%s"
    }
  }
}

Then restart Claude Code to enable the MCP server.`, serverName, serverName, binaryName)
}
