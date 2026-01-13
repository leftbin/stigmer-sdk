// Package stigmeragent provides a Go SDK for defining Stigmer AI agent blueprints.
//
// This SDK allows you to define agent templates with skills, MCP servers, sub-agents,
// and environment variables that convert to proto messages for the Stigmer platform.
//
// # Quick Start
//
// Create a simple agent:
//
//	agent, err := agent.New(
//	    agent.WithName("code-reviewer"),
//	    agent.WithInstructions("Review code and suggest improvements"),
//	    agent.WithDescription("AI code reviewer"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Convert to proto
//	proto := agent.ToProto()
//
// # Architecture
//
// The SDK follows a proto-first architecture where Go types are designed
// to convert cleanly to protobuf messages. This ensures compatibility with
// the Stigmer platform while providing an idiomatic Go API.
//
// # Key Packages
//
//   - agent: Core agent builder with functional options pattern
//   - skill: Skill reference configuration
//   - mcpserver: MCP server definitions (stdio, HTTP, Docker)
//   - subagent: Sub-agent configuration (inline and referenced)
//   - environment: Environment variable configuration
//
// # Design Patterns
//
// The SDK uses the functional options pattern for flexible, type-safe configuration:
//
//	agent.New(
//	    agent.WithName("my-agent"),
//	    agent.WithInstructions("Do something useful"),
//	    agent.WithSkill(skill.Platform("coding-best-practices")),
//	    agent.WithMCPServer(mcpserver.Stdio(
//	        mcpserver.WithName("github"),
//	        mcpserver.WithCommand("npx"),
//	        mcpserver.WithArgs("-y", "@modelcontextprotocol/server-github"),
//	    )),
//	)
//
// # Proto Conversion
//
// All types implement ToProto() methods for conversion to protobuf messages
// that can be sent to the Stigmer platform API.
//
// For more examples, see the examples/ directory.
package stigmeragent
