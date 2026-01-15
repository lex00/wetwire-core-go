# Recording Conversations

The scenario package can record user/agent conversations as animated SVG files using [termsvg](https://github.com/mrmarble/termsvg).

## Overview

Recordings show the conversation as it would appear in a terminal:
- **User messages** appear with typing simulation (green text)
- **Agent messages** appear line-by-line (white text)
- **Black background** for terminal aesthetic

This is useful for:
- Documentation and demos
- Sharing scenario examples
- Training materials

## Requirements

Install termsvg:

```bash
go install github.com/mrmarble/termsvg/cmd/termsvg@latest
```

The recorder automatically looks in `~/go/bin` if termsvg isn't in PATH.

## Recording a Session

The primary API is `RecordSession`, which takes a session's messages and creates an animated SVG:

```go
import "github.com/lex00/wetwire-core-go/scenario"

// After running a scenario through the orchestrator...
session, _ := orchestrator.Run(ctx)

// Create an adapter for the session
adapter := &SessionAdapter{session: session}

// Record to SVG
err := scenario.RecordSession(adapter, scenario.SessionRecordOptions{
    OutputDir: "./recordings",
})
// Creates: ./recordings/<session-name>.svg
```

### SessionMessages Interface

Your session must implement the `SessionMessages` interface:

```go
type SessionMessages interface {
    Name() string
    GetMessages() []SessionMessage
}

type SessionMessage struct {
    Role    string // "developer" (user) or "runner" (agent)
    Content string
}
```

## Recording Options

```go
opts := scenario.SessionRecordOptions{
    // Output directory (default: ./recordings)
    OutputDir: "./output",

    // Terminal dimensions in characters (default: 80x30)
    TermWidth:  80,
    TermHeight: 30,

    // Typing speed for user messages (default: 25ms per char)
    TypingSpeed: 25 * time.Millisecond,

    // Delay between lines for agent output (default: 100ms)
    LineDelay: 100 * time.Millisecond,

    // Pause between conversation turns (default: 500ms)
    MessageDelay: 500 * time.Millisecond,
}
```

## Visual Style

The recording uses:
- **Black background** (`#000000`)
- **Green user text** with `>` prompt (ANSI escape codes)
- **White agent text** (default terminal color)
- **Blank lines** between conversation turns for readability

## Animation Behavior

- Animation plays **once** and stops at the final frame
- No looping (patched after termsvg export)
- Final frame remains visible

## Example: Recording a Demo Conversation

```go
package main

import (
    "github.com/lex00/wetwire-core-go/scenario"
)

// Implement SessionMessages
type DemoSession struct {
    name     string
    messages []scenario.SessionMessage
}

func (s *DemoSession) Name() string                          { return s.name }
func (s *DemoSession) GetMessages() []scenario.SessionMessage { return s.messages }

func main() {
    session := &DemoSession{
        name: "s3_bucket_demo",
        messages: []scenario.SessionMessage{
            {Role: "developer", Content: "I need an S3 bucket for logs"},
            {Role: "runner", Content: "I'll create an S3 bucket...\n\n```yaml\nResources:\n  LogsBucket:\n    Type: AWS::S3::Bucket\n```"},
            {Role: "developer", Content: "Add encryption please"},
            {Role: "runner", Content: "Adding SSE-S3 encryption...\n\nDone!"},
        },
    }

    scenario.RecordSession(session, scenario.SessionRecordOptions{
        OutputDir: "./recordings",
    })
}
```

## Alternative: RunWithRecording

For simpler cases where you want to capture stdout from a function:

```go
err := scenario.RunWithRecording("demo", scenario.RecordOptions{
    Enabled:    true,
    OutputDir:  "./recordings",
    UserPrompt: "Create a VPC for my EKS cluster",
}, func() error {
    fmt.Println("Creating VPC template...")
    fmt.Println("```yaml")
    fmt.Println("Resources:")
    fmt.Println("  VPC:")
    fmt.Println("    Type: AWS::EC2::VPC")
    fmt.Println("```")
    return nil
})
```

This is less flexible than `RecordSession` but useful for quick recordings.

## Checking Availability

```go
if scenario.CanRecord() {
    // termsvg is available
} else {
    // Fall back or skip recording
}
```

## Output

The SVG file can be:
- Viewed in any web browser
- Embedded in documentation
- Shared as a standalone file

SVG dimensions are approximately:
- Width: `TermWidth * 12.5` pixels
- Height: `TermHeight * 27` pixels

For 80x30 characters: ~1000x810 pixels
