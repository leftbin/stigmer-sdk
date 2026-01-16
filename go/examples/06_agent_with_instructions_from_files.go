//go:build ignore
// Example 06: Agent with Instructions from Files + Synthesis
//
// This example demonstrates:
// 1. Loading agent instructions and skill content from external files
// 2. Auto-synthesis pattern using defer stigmer.Complete()
//
// Benefits of loading from files:
// 1. Better organization - keep large instructions separate from code
// 2. Easy to edit - use your favorite markdown editor
// 3. Version control - track instruction changes independently
// 4. Reusability - share instruction files across multiple agents
// 5. Maintainability - easier to review and update long instructions
//
// Synthesis Model:
// - SDK collects agent configuration in memory
// - On program exit, defer stigmer.Complete() writes manifest.pb
// - CLI reads manifest.pb and deploys to platform
//
// Directory structure:
//
//	examples/
//	├── 06_agent_with_instructions_from_files.go
//	└── instructions/
//	    ├── code-reviewer.md          (agent instructions)
//	    ├── security-guidelines.md    (skill content)
//	    └── testing-best-practices.md (skill content)
//
// Run modes:
// - Dry-run: go run examples/06_agent_with_instructions_from_files.go
// - Synthesis: STIGMER_OUT_DIR=/tmp go run examples/06_agent_with_instructions_from_files.go
//
package main

import (
	"fmt"
	"log"

	"github.com/leftbin/stigmer-sdk/go/stigmer"
	"github.com/leftbin/stigmer-sdk/go/agent"
	"github.com/leftbin/stigmer-sdk/go/mcpserver"
	"github.com/leftbin/stigmer-sdk/go/skill"
	"github.com/leftbin/stigmer-sdk/go/subagent"
)

func main() {
	// IMPORTANT: Complete() enables synthesis and runs on exit
	// This is the synthesis model where manifest.pb is automatically written
	// - If STIGMER_OUT_DIR is not set: Dry-run mode (prints success message)
	// - If STIGMER_OUT_DIR is set: Synthesis mode (writes manifest.pb to that directory)
	defer stigmer.Complete()

	fmt.Println("=== Example 06: Agent with Instructions from Files + Synthesis ===\n")

	// Example 1: Basic agent with instructions from file
	basicAgent := createBasicAgentFromFile()
	printAgent("1. Basic Agent with Instructions from File", basicAgent)

	// Example 2: Agent with inline skills loading markdown from files
	agentWithFileSkills := createAgentWithFileSkills()
	printAgent("2. Agent with Skills Loaded from Files", agentWithFileSkills)

	// Example 3: Complex agent with everything from files
	complexAgent := createComplexAgentFromFiles()
	printAgent("3. Complex Agent with All Content from Files", complexAgent)

	// Example 4: Sub-agent with instructions from file
	agentWithFileSubAgent := createAgentWithFileSubAgent()
	printAgent("4. Agent with Sub-Agent Instructions from File", agentWithFileSubAgent)
}

// Example 1: Basic agent with instructions from file
func createBasicAgentFromFile() *agent.Agent {
	ag, err := agent.New(
		agent.WithName("code-reviewer"),
		// Load instructions from external file instead of inline string
		agent.WithInstructionsFromFile("instructions/code-reviewer.md"),
		agent.WithDescription("AI code reviewer with comprehensive guidelines"),
	)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}
	return ag
}

// Example 2: Agent with inline skills loading markdown from files
func createAgentWithFileSkills() *agent.Agent {
	// Create inline skills with content loaded from files
	securitySkill, err := skill.New(
		skill.WithName("security-guidelines"),
		skill.WithDescription("Comprehensive security review guidelines"),
		// Load skill markdown from external file
		skill.WithMarkdownFromFile("instructions/security-guidelines.md"),
	)
	if err != nil {
		log.Fatalf("Failed to create security skill: %v", err)
	}

	testingSkill, err := skill.New(
		skill.WithName("testing-best-practices"),
		skill.WithDescription("Testing standards and best practices"),
		// Load skill markdown from external file
		skill.WithMarkdownFromFile("instructions/testing-best-practices.md"),
	)
	if err != nil {
		log.Fatalf("Failed to create testing skill: %v", err)
	}

	ag, err := agent.New(
		agent.WithName("senior-reviewer"),
		agent.WithInstructionsFromFile("instructions/code-reviewer.md"),
		agent.WithDescription("Senior code reviewer with security and testing expertise"),
		agent.WithSkills(*securitySkill, *testingSkill),
	)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}
	return ag
}

// Example 3: Complex agent with everything from files
func createComplexAgentFromFiles() *agent.Agent {
	// Create MCP server
	github, err := mcpserver.Stdio(
		mcpserver.WithName("github"),
		mcpserver.WithCommand("npx"),
		mcpserver.WithArgs("-y", "@modelcontextprotocol/server-github"),
		mcpserver.WithEnvPlaceholder("GITHUB_TOKEN", "${GITHUB_TOKEN}"),
	)
	if err != nil {
		log.Fatalf("Failed to create GitHub MCP server: %v", err)
	}

	// Create skills from files
	securitySkill, err := skill.New(
		skill.WithName("security-guidelines"),
		skill.WithDescription("Security review guidelines"),
		skill.WithMarkdownFromFile("instructions/security-guidelines.md"),
	)
	if err != nil {
		log.Fatalf("Failed to create security skill: %v", err)
	}

	testingSkill, err := skill.New(
		skill.WithName("testing-best-practices"),
		skill.WithDescription("Testing best practices"),
		skill.WithMarkdownFromFile("instructions/testing-best-practices.md"),
	)
	if err != nil {
		log.Fatalf("Failed to create testing skill: %v", err)
	}

	// Create agent with everything from files
	ag, err := agent.New(
		agent.WithName("github-reviewer"),
		agent.WithInstructionsFromFile("instructions/code-reviewer.md"),
		agent.WithDescription("GitHub PR reviewer with comprehensive guidelines"),
		agent.WithMCPServer(github),
		agent.WithSkills(*securitySkill, *testingSkill),
		// Also reference platform skills
		agent.WithSkill(skill.Platform("coding-best-practices")),
	)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}
	return ag
}

// Example 4: Sub-agent with instructions from file
func createAgentWithFileSubAgent() *agent.Agent {
	// Create sub-agent with instructions loaded from file
	securitySpecialist, err := subagent.Inline(
		subagent.WithName("security-specialist"),
		subagent.WithInstructionsFromFile("instructions/security-guidelines.md"),
		subagent.WithDescription("Security-focused code analyzer"),
	)
	if err != nil {
		log.Fatalf("Failed to create security specialist: %v", err)
	}

	ag, err := agent.New(
		agent.WithName("orchestrator"),
		agent.WithInstructionsFromFile("instructions/code-reviewer.md"),
		agent.WithDescription("Main orchestrator with specialized sub-agents"),
		agent.WithSubAgent(securitySpecialist),
	)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}
	return ag
}

// Helper function to print agent information
func printAgent(title string, ag *agent.Agent) {
	fmt.Printf("\n%s\n", title)
	fmt.Println("=" + string(make([]byte, len(title))))
	fmt.Printf("Agent Name: %s\n", ag.Name)
	fmt.Printf("Description: %s\n", ag.Description)
	fmt.Printf("Instructions Length: %d characters\n", len(ag.Instructions))
	fmt.Printf("Skills: %d\n", len(ag.Skills))
	fmt.Printf("MCP Servers: %d\n", len(ag.MCPServers))
	fmt.Printf("Sub-Agents: %d\n", len(ag.SubAgents))

	// Show first 100 chars of instructions to verify they were loaded
	if len(ag.Instructions) > 0 {
		preview := ag.Instructions
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		fmt.Printf("Instructions Preview: %s\n", preview)
	}

	fmt.Println("\n✅ Files loaded successfully!")
	fmt.Println("\nℹ️  Synthesis Mode:")
	fmt.Println("   - Dry-run: No STIGMER_OUT_DIR set (current)")
	fmt.Println("   - Synthesis: Set STIGMER_OUT_DIR=/tmp to write manifest.pb")
}
