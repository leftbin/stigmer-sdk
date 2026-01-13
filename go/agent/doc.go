// Package agent provides the core Agent builder for defining AI agent templates.
//
// The agent package implements the functional options pattern to provide a
// flexible, type-safe API for building agent configurations that convert to
// protobuf messages.
//
// # Basic Usage
//
//	agent, err := agent.New(
//	    agent.WithName("code-reviewer"),
//	    agent.WithInstructions("Review code and suggest improvements"),
//	    agent.WithDescription("AI code reviewer"),
//	)
//
// # Validation
//
// All agent fields are validated during construction:
//
//   - Name: lowercase alphanumeric + hyphens, max 63 characters
//   - Instructions: min 10 characters, max 10,000 characters
//   - Description: max 500 characters (optional)
//   - IconURL: valid URL format (optional)
//
// Validation errors are returned from New() and provide detailed context.
//
// # Proto Conversion
//
// Agents can be converted to protobuf messages:
//
//	proto := agent.ToProto()
//	// proto is *agentv1.AgentSpec
//
// The proto conversion is designed to be lossless - all information in the
// Go Agent struct is preserved in the protobuf message.
//
// # Configuration Options
//
// The following options are available:
//
//   - WithName: Set the agent name (required)
//   - WithInstructions: Set behavior instructions (required)
//   - WithDescription: Set human-readable description
//   - WithIconURL: Set icon URL for UI display
//   - WithOrg: Set organization owner
//   - WithSkill: Add a skill reference
//   - WithSkills: Add multiple skill references
//   - WithMCPServer: Add an MCP server definition
//   - WithMCPServers: Add multiple MCP server definitions
//   - WithSubAgent: Add a sub-agent
//   - WithSubAgents: Add multiple sub-agents
//   - WithEnvVar: Add an environment variable
//   - WithEnvVars: Add multiple environment variables
//
// # Error Handling
//
// The package provides specific error types for different failure modes:
//
//   - ValidationError: Field validation failures
//   - ConversionError: Proto conversion failures
//
// Common validation errors are also exported as sentinel errors:
//
//   - ErrInvalidName
//   - ErrInvalidInstructions
//   - ErrInvalidDescription
//   - ErrInvalidIconURL
package agent
