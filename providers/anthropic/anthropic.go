// Package anthropic provides an Anthropic API implementation of the Provider interface.
package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/lex00/wetwire-core-go/providers"
)

// DefaultModel is the default model used by the Anthropic provider.
const DefaultModel = "claude-sonnet-4-20250514"

// Provider implements the providers.Provider interface using the Anthropic API.
type Provider struct {
	client anthropic.Client
}

// Config contains configuration for the Anthropic provider.
type Config struct {
	// APIKey for Anthropic (defaults to ANTHROPIC_API_KEY env var)
	APIKey string
}

// New creates a new Anthropic provider.
func New(config Config) (*Provider, error) {
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	return &Provider{
		client: client,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "anthropic"
}

// CreateMessage sends a message request and returns the complete response.
func (p *Provider) CreateMessage(ctx context.Context, req providers.MessageRequest) (*providers.MessageResponse, error) {
	params := p.buildParams(req)

	resp, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	return p.convertResponse(resp), nil
}

// StreamMessage sends a message request and streams the response via the handler.
func (p *Provider) StreamMessage(ctx context.Context, req providers.MessageRequest, handler providers.StreamHandler) (*providers.MessageResponse, error) {
	params := p.buildParams(req)

	stream := p.client.Messages.NewStreaming(ctx, params)

	// Accumulate the full response
	var message *anthropic.Message
	var contentBlocks []anthropic.ContentBlockUnion
	currentTextContent := make(map[int64]*strings.Builder)
	currentToolInput := make(map[int64]*strings.Builder)

	for stream.Next() {
		event := stream.Current()

		switch event.Type {
		case "message_start":
			startEvent := event.AsMessageStart()
			message = &startEvent.Message
			contentBlocks = nil
			currentTextContent = make(map[int64]*strings.Builder)

		case "content_block_start":
			startEvent := event.AsContentBlockStart()

			if startEvent.ContentBlock.Type == "text" {
				currentTextContent[startEvent.Index] = &strings.Builder{}
			} else if startEvent.ContentBlock.Type == "tool_use" {
				currentToolInput[startEvent.Index] = &strings.Builder{}
			}

			block := anthropic.ContentBlockUnion{
				Type: startEvent.ContentBlock.Type,
				ID:   startEvent.ContentBlock.ID,
				Name: startEvent.ContentBlock.Name,
				Text: startEvent.ContentBlock.Text,
			}
			contentBlocks = append(contentBlocks, block)

		case "content_block_delta":
			deltaEvent := event.AsContentBlockDelta()

			if deltaEvent.Delta.Type == "text_delta" && deltaEvent.Delta.Text != "" {
				handler(deltaEvent.Delta.Text)

				if builder, ok := currentTextContent[deltaEvent.Index]; ok {
					builder.WriteString(deltaEvent.Delta.Text)
				}
			}

			if deltaEvent.Delta.Type == "input_json_delta" && deltaEvent.Delta.PartialJSON != "" {
				if builder, ok := currentToolInput[deltaEvent.Index]; ok {
					builder.WriteString(deltaEvent.Delta.PartialJSON)
				}
			}

		case "content_block_stop":
			stopEvent := event.AsContentBlockStop()
			idx := int(stopEvent.Index)
			if idx < len(contentBlocks) {
				if builder, ok := currentTextContent[stopEvent.Index]; ok {
					contentBlocks[idx].Text = builder.String()
				}
				if builder, ok := currentToolInput[stopEvent.Index]; ok {
					contentBlocks[idx].Input = json.RawMessage(builder.String())
				}
			}

		case "message_delta":
			deltaEvent := event.AsMessageDelta()
			if message != nil {
				message.StopReason = deltaEvent.Delta.StopReason
				message.StopSequence = deltaEvent.Delta.StopSequence
			}
		}
	}

	if err := stream.Err(); err != nil {
		return nil, err
	}

	if message != nil {
		message.Content = contentBlocks
	}

	return p.convertResponse(message), nil
}

// buildParams converts a MessageRequest to Anthropic API parameters.
func (p *Provider) buildParams(req providers.MessageRequest) anthropic.MessageNewParams {
	model := req.Model
	if model == "" {
		model = DefaultModel
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: int64(maxTokens),
	}

	if req.System != "" {
		params.System = []anthropic.TextBlockParam{{Text: req.System}}
	}

	params.Messages = p.convertMessages(req.Messages)
	params.Tools = p.convertTools(req.Tools)

	return params
}

// convertMessages converts provider messages to Anthropic message params.
func (p *Provider) convertMessages(msgs []providers.Message) []anthropic.MessageParam {
	result := make([]anthropic.MessageParam, 0, len(msgs))

	for _, msg := range msgs {
		var blocks []anthropic.ContentBlockParamUnion

		for _, block := range msg.Content {
			switch block.Type {
			case "text":
				blocks = append(blocks, anthropic.NewTextBlock(block.Text))
			case "tool_use":
				// Assistant's tool use - will be included via ToParam()
				blocks = append(blocks, anthropic.ContentBlockParamUnion{
					OfToolUse: &anthropic.ToolUseBlockParam{
						ID:    block.ID,
						Name:  block.Name,
						Input: block.Input,
					},
				})
			case "tool_result":
				blocks = append(blocks, anthropic.NewToolResultBlock(
					block.ToolUseID,
					block.Content,
					block.IsError,
				))
			}
		}

		if msg.Role == "user" {
			result = append(result, anthropic.NewUserMessage(blocks...))
		} else {
			result = append(result, anthropic.NewAssistantMessage(blocks...))
		}
	}

	return result
}

// convertTools converts provider tools to Anthropic tool params.
func (p *Provider) convertTools(tools []providers.Tool) []anthropic.ToolUnionParam {
	result := make([]anthropic.ToolUnionParam, 0, len(tools))

	for _, tool := range tools {
		result = append(result, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: tool.InputSchema.Properties,
					Required:   tool.InputSchema.Required,
				},
			},
		})
	}

	return result
}

// convertResponse converts an Anthropic response to a provider response.
func (p *Provider) convertResponse(resp *anthropic.Message) *providers.MessageResponse {
	if resp == nil {
		return &providers.MessageResponse{}
	}

	result := &providers.MessageResponse{
		StopReason: convertStopReason(resp.StopReason),
	}

	for _, block := range resp.Content {
		cb := providers.ContentBlock{
			Type: string(block.Type),
		}

		switch block.Type {
		case "text":
			cb.Text = block.Text
		case "tool_use":
			cb.ID = block.ID
			cb.Name = block.Name
			cb.Input = block.Input
		}

		result.Content = append(result.Content, cb)
	}

	return result
}

// convertStopReason converts Anthropic stop reason to provider stop reason.
func convertStopReason(reason anthropic.StopReason) providers.StopReason {
	switch reason {
	case anthropic.StopReasonEndTurn:
		return providers.StopReasonEndTurn
	case anthropic.StopReasonToolUse:
		return providers.StopReasonToolUse
	case anthropic.StopReasonMaxTokens:
		return providers.StopReasonMaxTokens
	case anthropic.StopReasonStopSequence:
		return providers.StopReasonStopSequence
	default:
		return providers.StopReason(string(reason))
	}
}
