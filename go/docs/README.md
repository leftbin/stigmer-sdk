# Go SDK Documentation

Complete documentation for the Stigmer Go SDK.

## Quick Start

- **Main README**: [../README.md](../README.md) - Go SDK overview and quick start

## Architecture

- [Synthesis Model](./architecture/synthesis-model.md) - Why `defer stigmeragent.Complete()` is required and Go's limitations
- [Synthesis Architecture](./architecture/synthesis-architecture.md) - Auto-synthesis model with defer pattern and manifest.pb generation
- [Multi-Agent Support](./architecture/multi-agent-support.md) - Multiple agents in one file - implementation complete
- [Synthesis Behavior and Limitations](./architecture/synthesis-behavior-and-limitations.md) - How synthesis works, testing, and CLI session isolation

## Guides

### Getting Started
- [Migration Guide](./guides/migration-guide.md) - Migrating from proto-coupled to proto-agnostic design
- [Buf Dependency Guide](./guides/buf-dependency-guide.md) - Using Buf Schema Registry dependencies

## Implementation

- [Synthesis API Improvement](./implementation/synthesis-api-improvement.md) - Evolution from `synth.AutoSynth()` to `stigmeragent.Complete()`

## References

- [Proto Mapping Reference](./references/proto-mapping.md) - Proto field mapping reference for CLI developers

## Examples

See [../examples/](../examples/) for working code examples:

1. `01_basic_agent.go` - Basic agent creation
2. `02_agent_with_skills.go` - Adding skills
3. `03_agent_with_mcp_servers.go` - Adding MCP servers
4. `04_agent_with_subagents.go` - Adding sub-agents
5. `05_agent_with_environment_variables.go` - Environment variables
6. `06_agent_with_instructions_from_files.go` - Loading instructions from files
7. `task3-manifest-example.go` - Reference for Task 3 implementation

## Implementation Details

- **Proto-Agnostic Architecture**: SDK is now pure Go, CLI handles proto conversion
- **Buf Schema Registry**: Uses `buf.build/leftbin/stigmer` for manifest proto
- **Builder Pattern**: Functional options for configuration
- **File Loading**: Support for loading instructions and markdown from files
- **Inline Skills**: Create skills directly in repository code

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./agent
```

## Building

```bash
# Build all packages
go build ./...

# Build examples
go build ./examples/...
```

## Related Documentation

- [Root Documentation Index](../../docs/README.md) - Multi-language SDK documentation
- [Python SDK Documentation](../../python/README.md) - Python SDK docs
