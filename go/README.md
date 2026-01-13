# Stigmer Agent SDK - Go

A Go SDK for defining AI agent blueprints for the Stigmer platform.

**Repository**: [github.com/leftbin/stigmer-sdk](https://github.com/leftbin/stigmer-sdk)  
**Go Package**: `github.com/leftbin/stigmer-sdk/go`

## Features

- **Proto-agnostic SDK**: Pure Go library with no proto dependencies
- **File-based content**: Load instructions and skills from markdown files
- **Inline resources**: Define skills and sub-agents directly in your repository
- **Go-idiomatic API**: Functional options and builder patterns for flexible configuration
- **Type-safe**: Leverage Go's type system for compile-time safety
- **Feature parity**: 1:1 compatibility with Python SDK
- **Well-tested**: Comprehensive unit and integration tests

## Installation

```bash
go get github.com/leftbin/stigmer-sdk/go
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/leftbin/stigmer-sdk/go/agent"
    "github.com/leftbin/stigmer-sdk/go/skill"
    "github.com/leftbin/stigmer-sdk/go/mcpserver"
    "github.com/leftbin/stigmer-sdk/go/internal/synth"
)

func main() {
    // Enable auto-synthesis (writes manifest.pb on exit)
    defer synth.AutoSynth()
    
    // Create inline skill from markdown file
    securitySkill, _ := skill.New(
        skill.WithName("security-guidelines"),
        skill.WithDescription("Security review guidelines"),
        skill.WithMarkdownFromFile("skills/security.md"),
    )

    // Create MCP server
    githubMCP, _ := mcpserver.Stdio(
        mcpserver.WithName("github"),
        mcpserver.WithCommand("npx"),
        mcpserver.WithArgs("-y", "@modelcontextprotocol/server-github"),
        mcpserver.WithEnvPlaceholder("GITHUB_TOKEN", "${GITHUB_TOKEN}"),
    )

    // Create agent with instructions from file
    myAgent, err := agent.New(
        agent.WithName("code-reviewer"),
        agent.WithInstructionsFromFile("instructions/reviewer.md"),
        agent.WithDescription("AI code reviewer with security expertise"),
        agent.WithIconURL("https://example.com/icon.png"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Use builder methods to add components
    myAgent.
        AddSkill(*securitySkill).                    // Inline skill
        AddSkill(skill.Platform("coding-standards")). // Platform skill
        AddMCPServer(githubMCP)
    
    fmt.Printf("Agent created: %s\n", myAgent.Name)
    // On exit, defer synth.AutoSynth() automatically writes manifest.pb
}
```

## Core Concepts

### Agent

The `Agent` is the main blueprint that defines:
- Name and instructions (required) - load from files with `WithInstructionsFromFile()`
- Description and icon (optional)
- Skills (knowledge references) - inline or platform/org references
- MCP servers (tool providers)
- Sub-agents (delegatable agents)
- Environment variables (configuration)

**Key Features:**
- **File-based instructions**: Load from markdown files instead of inline strings
- **Builder pattern**: Add components after creation with `AddSkill()`, `AddMCPServer()`, etc.
- **Proto-agnostic**: No proto types or conversion - just pure Go

### Skills

Skills provide knowledge to agents. Three ways to use them:

#### 1. Inline Skills (Defined in Repository)
Create skills with markdown content from files:

```go
// Define skill in your repository
securitySkill, _ := skill.New(
    skill.WithName("security-guidelines"),
    skill.WithDescription("Security review guidelines"),
    skill.WithMarkdownFromFile("skills/security.md"),
)

// Add to agent
myAgent.AddSkill(*securitySkill)
```

**Benefits:**
- ✅ Version controlled with your agent code
- ✅ Easy to edit and update
- ✅ Sharable across agents in your repository

#### 2. Platform Skills (Shared)
Reference skills available platform-wide:

```go
myAgent.AddSkill(skill.Platform("coding-best-practices"))
```

#### 3. Organization Skills (Private)
Reference skills private to your organization:

```go
myAgent.AddSkill(skill.Organization("my-org", "internal-standards"))
```

### MCP Servers

MCP (Model Context Protocol) servers provide tools to agents. Three types:

#### 1. Stdio Servers
Subprocess-based servers (most common):

```go
agent.WithMCPServer(mcpserver.Stdio(
    mcpserver.WithName("github"),
    mcpserver.WithCommand("npx"),
    mcpserver.WithArgs("-y", "@modelcontextprotocol/server-github"),
    mcpserver.WithEnvPlaceholder("GITHUB_TOKEN", "${GITHUB_TOKEN}"),
))
```

#### 2. HTTP Servers
Remote HTTP + SSE servers:

```go
agent.WithMCPServer(mcpserver.HTTP(
    mcpserver.WithName("remote-mcp"),
    mcpserver.WithURL("https://mcp.example.com/github"),
    mcpserver.WithHeader("Authorization", "Bearer ${API_TOKEN}"),
    mcpserver.WithTimeout(30),
))
```

#### 3. Docker Servers
Containerized MCP servers:

```go
agent.WithMCPServer(mcpserver.Docker(
    mcpserver.WithName("custom-mcp"),
    mcpserver.WithImage("ghcr.io/org/custom-mcp:latest"),
    mcpserver.WithVolumeMount("/host/path", "/container/path", false),
    mcpserver.WithPortMapping(8080, 80, "tcp"),
))
```

### Sub-Agents

Sub-agents allow delegation to specialized agents:

#### Inline Sub-Agents
Defined within the parent agent:

```go
agent.WithSubAgent(subagent.Inline(
    subagent.WithName("code-analyzer"),
    subagent.WithInstructions("Analyze code quality"),
    subagent.WithMCPServer("github"),
    subagent.WithSkill(skill.Platform("static-analysis")),
))
```

#### Referenced Sub-Agents
Reference existing agents:

```go
agent.WithSubAgent(subagent.Reference("agent-instance-id"))
```

### Environment Variables

Define configuration and secret requirements for agents.

#### Secret Variables
Required secrets are encrypted at rest:

```go
apiKey, _ := environment.New(
    environment.WithName("API_KEY"),
    environment.WithSecret(true),
    environment.WithDescription("API key for external service"),
)
agent.WithEnvironmentVariable(apiKey)
```

#### Configuration with Defaults
Optional configuration values with sensible defaults:

```go
region, _ := environment.New(
    environment.WithName("AWS_REGION"),
    environment.WithDefaultValue("us-east-1"),
    environment.WithDescription("AWS deployment region"),
)
agent.WithEnvironmentVariable(region)
```

#### Key Features
- **Secrets**: Encrypted at rest, redacted in logs (use `WithSecret(true)`)
- **Configuration**: Plaintext values for non-sensitive data
- **Defaults**: Variables with defaults are automatically optional
- **Validation**: Names must be uppercase with underscores (e.g., `GITHUB_TOKEN`)
- **Required/Optional**: Control whether values must be provided

## Architecture

The SDK follows a **proto-agnostic architecture**:

```
User Repository (Your Code)
    ↓ uses
SDK (Pure Go, No Proto)
    ↓ reads
CLI (stigmer-cli)
    ↓ converts to proto
Platform (Stigmer API)
```

**Key Principles:**
- ✅ SDK is proto-ignorant - no proto dependencies
- ✅ Users write pure Go code
- ✅ CLI handles all proto conversion and deployment
- ✅ SDK and proto can evolve independently

See [docs/references/proto-mapping.md](docs/references/proto-mapping.md) for how CLI converts SDK types to proto messages.

## Validation

All inputs are validated at construction time:

- **Name**: lowercase alphanumeric + hyphens, max 63 chars
- **Instructions**: min 10 chars, max 10,000 chars
- **Description**: max 500 chars
- **URLs**: valid URL format

Validation errors provide clear, actionable feedback:

```go
agent, err := agent.New(agent.WithName("Invalid Name!"))
// err: validation failed for field "name": name must be lowercase...
```

## Examples

See the [examples/](examples/) directory for complete examples:

1. **Basic Agent** (`01_basic_agent.go`) - Simple agent with name and instructions
2. **Agent with Skills** (`02_agent_with_skills.go`) - Platform, organization, and inline skills
3. **Agent with MCP Servers** (`03_agent_with_mcp_servers.go`) - Full MCP server configuration (stdio, http, docker)
4. **Agent with Sub-Agents** (`04_agent_with_subagents.go`) - Inline and referenced sub-agents
5. **Agent with Environment Variables** (`05_agent_with_environment_variables.go`) - Secrets, configs, and validation
6. **Agent with Instructions from Files** (`06_agent_with_instructions_from_files.go`) - **Recommended pattern** - Load all content from files

**🌟 Start with Example 06** - it demonstrates the recommended pattern of loading instructions and skill content from external markdown files.

## Development

### Prerequisites

- Go 1.21 or higher
- golangci-lint (for linting)

### Build

```bash
make build
```

### Test

```bash
make test              # Run all tests
make test-coverage     # Generate coverage report
```

### Lint

```bash
make lint              # Run golangci-lint
```

### Verify

```bash
make verify            # Run fmt, vet, lint, and test
```

## API Documentation

Full API documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/leftbin/stigmer-sdk/go).

## Migration from Python SDK

If you're migrating from the Python SDK, see [docs/guides/migration-guide.md](docs/guides/migration-guide.md) for a side-by-side comparison and translation guide.

## Project Structure

```
sdk/go/
├── agent/           # Core agent builder
├── skill/           # Skill configuration
├── mcpserver/       # MCP server definitions
├── subagent/        # Sub-agent configuration
├── environment/     # Environment variables
├── examples/        # Usage examples
├── testdata/        # Test fixtures and golden files
└── Makefile         # Build targets
```

## Contributing

We welcome contributions! Please ensure:

1. All tests pass (`make test`)
2. Code is formatted (`make fmt`)
3. Linter passes (`make lint`)
4. Coverage remains high (90%+ target)

## License

Apache 2.0 - see [LICENSE](../LICENSE) for details.

## Support

For questions and support:
- GitHub Issues: [leftbin/stigmer-sdk](https://github.com/leftbin/stigmer-sdk/issues)
- Discussions: [GitHub Discussions](https://github.com/leftbin/stigmer-sdk/discussions)
- Documentation: [docs.stigmer.ai](https://docs.stigmer.ai)

## Version

**Current Version**: `v0.1.0` (Initial Public Release)

**Status**: ✅ Production Ready

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for detailed version history.

## Related Documentation

- **Multi-language SDK Overview**: [Main README](../README.md)
- **Complete Go SDK Documentation**: [docs/README.md](docs/README.md) - Full documentation index

### Architecture
- **Synthesis Architecture**: [docs/architecture/synthesis-architecture.md](docs/architecture/synthesis-architecture.md) - Auto-synthesis model with defer pattern

### Guides
- **Migration Guide**: [docs/guides/migration-guide.md](docs/guides/migration-guide.md) - Migrating from proto-coupled design
- **Buf Dependency Guide**: [docs/guides/buf-dependency-guide.md](docs/guides/buf-dependency-guide.md) - Using Buf Schema Registry

### References
- **Proto Mapping**: [docs/references/proto-mapping.md](docs/references/proto-mapping.md) - CLI conversion reference

### Contributing
- **Contributing**: [../CONTRIBUTING.md](../CONTRIBUTING.md)
