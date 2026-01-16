package scenario

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/lex00/wetwire-core-go/agent/results"
	"github.com/lex00/wetwire-core-go/mcp"
	"github.com/lex00/wetwire-core-go/providers"
)

// ScenarioAgent is an autonomous agent that executes multi-domain scenarios using MCP tools.
type ScenarioAgent struct {
	provider  providers.Provider
	mcpServer *mcp.Server
	output    io.Writer
	model     string
	session   *results.Session
}

// ScenarioAgentConfig configures the ScenarioAgent.
type ScenarioAgentConfig struct {
	Provider  providers.Provider
	MCPServer *mcp.Server
	Output    io.Writer
	Model     string           // Optional, defaults to claude-sonnet-4-20250514
	Session   *results.Session // Optional, for result tracking
}

// NewScenarioAgent creates a new ScenarioAgent.
func NewScenarioAgent(config ScenarioAgentConfig) *ScenarioAgent {
	model := config.Model
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	output := config.Output
	if output == nil {
		output = io.Discard
	}

	return &ScenarioAgent{
		provider:  config.Provider,
		mcpServer: config.MCPServer,
		output:    output,
		model:     model,
		session:   config.Session,
	}
}

// Session returns the session for result tracking.
func (a *ScenarioAgent) Session() *results.Session {
	return a.session
}

// Run executes the agent with the given prompt.
// The agent runs autonomously without developer interaction.
func (a *ScenarioAgent) Run(ctx context.Context, prompt string) error {
	systemPrompt := `You are an autonomous infrastructure code generator for multi-domain scenarios.

Your job is to generate infrastructure code across multiple domains using the wetwire framework.

You have access to MCP tools provided by domain packages. Use these tools to:
1. Initialize packages
2. Write infrastructure code
3. Run linters
4. Build outputs
5. Validate cross-domain references

Work autonomously to complete the scenario requirements. Do not ask questions - make reasonable decisions based on the scenario description and validation criteria.

Follow the execution plan provided in the user prompt, respecting domain dependencies and validation requirements.`

	// Track initial prompt in session
	if a.session != nil {
		a.session.InitialPrompt = prompt
		a.session.AddMessage("user", prompt)
	}

	// Get MCP tools from the server
	tools := a.getMCPTools()

	messages := []providers.Message{
		providers.NewUserMessage(prompt),
	}

	// Agentic loop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req := providers.MessageRequest{
			Model:     a.model,
			MaxTokens: 4096,
			System:    systemPrompt,
			Messages:  messages,
			Tools:     tools,
		}

		resp, err := a.provider.CreateMessage(ctx, req)
		if err != nil {
			return fmt.Errorf("API call failed: %w", err)
		}

		// Add assistant response to messages
		messages = append(messages, providers.NewAssistantMessage(resp.Content))

		// Track assistant response in session
		if a.session != nil {
			// Extract text content from response
			var textContent string
			var toolCalls []results.ToolCall
			for _, block := range resp.Content {
				switch block.Type {
				case "text":
					textContent += block.Text
				case "tool_use":
					toolCalls = append(toolCalls, results.ToolCall{
						Name:  block.Name,
						Input: string(block.Input),
					})
				}
			}
			msg := results.Message{
				Role:      "runner",
				Content:   textContent,
				Timestamp: time.Now(),
				ToolCalls: toolCalls,
			}
			a.session.Messages = append(a.session.Messages, msg)
		}

		// Check for stop reason
		if resp.StopReason == providers.StopReasonEndTurn {
			// Agent is done
			break
		}

		// Process tool calls
		if resp.StopReason == providers.StopReasonToolUse {
			var toolResults []providers.ContentBlock

			for _, block := range resp.Content {
				if block.Type == "tool_use" {
					result, err := a.executeMCPTool(ctx, block.Name, block.Input)
					if err != nil {
						result = fmt.Sprintf("Error: %v", err)
					}
					toolResults = append(toolResults, providers.NewToolResult(
						block.ID,
						result,
						err != nil,
					))
				}
			}

			messages = append(messages, providers.NewToolResultMessage(toolResults))
		}
	}

	return nil
}

// getMCPTools converts MCP server tools to provider tools.
func (a *ScenarioAgent) getMCPTools() []providers.Tool {
	mcpTools := a.mcpServer.GetTools()
	tools := make([]providers.Tool, 0, len(mcpTools))

	for _, mcpTool := range mcpTools {
		tools = append(tools, providers.Tool{
			Name:        mcpTool.Name,
			Description: mcpTool.Description,
			InputSchema: providers.ToolInputSchema{
				Properties: a.convertSchema(mcpTool.InputSchema),
				Required:   a.extractRequired(mcpTool.InputSchema),
			},
		})
	}

	return tools
}

// executeMCPTool executes an MCP tool and returns the result.
func (a *ScenarioAgent) executeMCPTool(ctx context.Context, name string, input json.RawMessage) (string, error) {
	var params map[string]any
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("failed to parse tool input: %w", err)
	}

	result, err := a.mcpServer.ExecuteTool(ctx, name, params)
	if err != nil {
		return "", err
	}

	// Log tool execution to output
	_, _ = fmt.Fprintf(a.output, "[Tool: %s] %s\n", name, result)

	// Track generated files in session
	if a.session != nil {
		// Check if this is a write tool and track the file
		if name == "wetwire_write" {
			if path, ok := params["path"].(string); ok {
				a.session.GeneratedFiles = append(a.session.GeneratedFiles, path)
			}
		}
	}

	return result, nil
}

// convertSchema converts MCP input schema to provider schema format.
func (a *ScenarioAgent) convertSchema(inputSchema map[string]any) map[string]any {
	// MCP schema is already in JSON Schema format
	// Just extract properties if they exist
	if props, ok := inputSchema["properties"].(map[string]any); ok {
		return props
	}
	return make(map[string]any)
}

// extractRequired extracts required fields from MCP input schema.
func (a *ScenarioAgent) extractRequired(inputSchema map[string]any) []string {
	if req, ok := inputSchema["required"].([]any); ok {
		result := make([]string, 0, len(req))
		for _, r := range req {
			if s, ok := r.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}
