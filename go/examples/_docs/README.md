# Examples Documentation

This directory contains documentation for the Stigmer SDK Go examples.

## Current Examples (9 total)

### Agent Examples (6)
1. **01_basic_agent.go** - Basic agent creation with required and optional fields
2. **02_agent_with_skills.go** - Agent with inline, platform, and organization skills
3. **03_agent_with_mcp_servers.go** - Agent with MCP servers (Stdio, HTTP, Docker)
4. **04_agent_with_subagents.go** - Agent with sub-agents (inline and referenced)
5. **05_agent_with_environment_variables.go** - Agent with environment variables
6. **06_agent_with_instructions_from_files.go** - Agent loading instructions from files

### Workflow Examples (1)
7. **07_basic_workflow.go** - Basic workflow with HTTP GET and task field references

### Context Examples (2)
8. **08_agent_with_typed_context.go** - Agent using typed context variables
9. **09_workflow_and_agent_shared_context.go** - Workflow and agent sharing context

## API Pattern

All examples use the modern `stigmer.Run()` API:

```go
func main() {
    err := stigmer.Run(func(ctx *stigmer.Context) error {
        // Create agents
        agent, err := agent.New(ctx,
            agent.WithName("my-agent"),
            agent.WithInstructions("..."),
        )
        
        // Create workflows
        wf, err := workflow.New(ctx,
            workflow.WithNamespace("my-org"),
            workflow.WithName("my-workflow"),
        )
        
        return nil
    })
    
    if err != nil {
        log.Fatal(err)
    }
}
```

## Key Features

- **Clean API**: `agent.New(ctx, ...)` and `workflow.New(ctx, ...)` - context as first parameter
- **Automatic synthesis**: Manifests generated when `stigmer.Run()` completes
- **Type safety**: Compile-time checks and IDE autocomplete
- **Pulumi-aligned**: Familiar patterns for infrastructure-as-code developers

## Instructions Directory

The `instructions/` directory contains sample markdown files used by example 06:
- `code-reviewer.md` - Sample agent instructions
- `security-guidelines.md` - Sample skill content
- `testing-best-practices.md` - Sample skill content

These demonstrate loading content from external files.

---

*Last Updated: 2026-01-17*
