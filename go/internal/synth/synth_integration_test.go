package synth

import (
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	agentv1 "buf.build/gen/go/leftbin/stigmer/protocolbuffers/go/ai/stigmer/agentic/agent/v1"

	"github.com/leftbin/stigmer-sdk/go/agent"
	"github.com/leftbin/stigmer-sdk/go/environment"
	"github.com/leftbin/stigmer-sdk/go/internal/registry"
	"github.com/leftbin/stigmer-sdk/go/mcpserver"
	"github.com/leftbin/stigmer-sdk/go/skill"
	"github.com/leftbin/stigmer-sdk/go/subagent"
)

// TestAutoSynth_DryRunMode verifies dry-run mode (no STIGMER_OUT_DIR).
func TestAutoSynth_DryRunMode(t *testing.T) {
	// Reset registry
	defer registry.Global().Clear()
	registry.Global().Clear()

	// Ensure STIGMER_OUT_DIR is not set
	os.Unsetenv("STIGMER_OUT_DIR")

	// Create agent (auto-registered)
	testAgent, err := agent.New(
		agent.WithName("test-agent"),
		agent.WithInstructions("Test instructions for dry-run mode"),
	)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Verify agent is registered
	if !registry.Global().HasAgent() {
		t.Fatal("Agent should be registered in global registry")
	}

	if registry.Global().Count() != 1 {
		t.Errorf("Registry count = %d, want 1", registry.Global().Count())
	}

	// Call AutoSynth (should print message, not write file)
	AutoSynth()

	// In dry-run mode, no files should be written
	// This is a successful dry-run
	_ = testAgent // Use the agent to avoid unused variable warning
}

// TestAutoSynth_SynthesisMode verifies synthesis mode (with STIGMER_OUT_DIR).
func TestAutoSynth_SynthesisMode(t *testing.T) {
	// Reset registry
	defer registry.Global().Clear()
	registry.Global().Clear()

	// Create temp directory for output
	tempDir := t.TempDir()
	os.Setenv("STIGMER_OUT_DIR", tempDir)
	defer os.Unsetenv("STIGMER_OUT_DIR")

	// Create a comprehensive agent
	securitySkill, err := skill.New(
		skill.WithName("security-review"),
		skill.WithDescription("Security review guidelines"),
		skill.WithMarkdown("# Security Review\n\nCheck for vulnerabilities."),
	)
	if err != nil {
		t.Fatalf("Failed to create skill: %v", err)
	}

	githubMCP, err := mcpserver.Stdio(
		mcpserver.WithName("github"),
		mcpserver.WithCommand("npx"),
		mcpserver.WithArgs("-y", "@modelcontextprotocol/server-github"),
		mcpserver.WithEnvPlaceholder("GITHUB_TOKEN", "${GITHUB_TOKEN}"),
	)
	if err != nil {
		t.Fatalf("Failed to create MCP server: %v", err)
	}

	securitySub, err := subagent.Inline(
		subagent.WithName("security-specialist"),
		subagent.WithInstructions("Focus on security vulnerabilities"),
	)
	if err != nil {
		t.Fatalf("Failed to create subagent: %v", err)
	}

	githubToken, err := environment.New(
		environment.WithName("GITHUB_TOKEN"),
		environment.WithSecret(true),
		environment.WithDescription("GitHub API token"),
	)
	if err != nil {
		t.Fatalf("Failed to create environment variable: %v", err)
	}

	// Create agent with all components
	testAgent, err := agent.New(
		agent.WithName("comprehensive-reviewer"),
		agent.WithInstructions("Review code comprehensively with security focus"),
		agent.WithDescription("A comprehensive code reviewer"),
		agent.WithIconURL("https://example.com/icon.png"),
		agent.WithSkills(*securitySkill, skill.Platform("coding-standards")),
		agent.WithMCPServer(githubMCP),
		agent.WithSubAgent(securitySub),
		agent.WithEnvironmentVariable(githubToken),
	)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Verify agent is registered
	if !registry.Global().HasAgent() {
		t.Fatal("Agent should be registered in global registry")
	}

	// Call AutoSynth (should write manifest.pb)
	AutoSynth()

	// Verify manifest.pb was created
	manifestPath := filepath.Join(tempDir, "manifest.pb")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Fatalf("manifest.pb was not created at %s", manifestPath)
	}

	// Read and verify manifest.pb content
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("Failed to read manifest.pb: %v", err)
	}

	// Deserialize manifest
	var manifest agentv1.AgentManifest
	if err := proto.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("Failed to unmarshal manifest.pb: %v", err)
	}

	// Verify manifest content
	if manifest.SdkMetadata == nil {
		t.Fatal("SdkMetadata is nil")
	}
	if manifest.SdkMetadata.Language != "go" {
		t.Errorf("SDK language = %v, want go", manifest.SdkMetadata.Language)
	}
	if manifest.SdkMetadata.Version != "0.1.0" {
		t.Errorf("SDK version = %v, want 0.1.0", manifest.SdkMetadata.Version)
	}

	// Verify we have exactly 1 agent
	if len(manifest.Agents) != 1 {
		t.Fatalf("Manifest contains %d agents, want 1", len(manifest.Agents))
	}

	agentProto := manifest.Agents[0]
	if agentProto.Name != testAgent.Name {
		t.Errorf("Agent name = %v, want %v", agentProto.Name, testAgent.Name)
	}
	if agentProto.Instructions != testAgent.Instructions {
		t.Errorf("Agent instructions mismatch")
	}
	if agentProto.Description != testAgent.Description {
		t.Errorf("Agent description = %v, want %v", agentProto.Description, testAgent.Description)
	}

	// Verify skills
	if len(agentProto.Skills) != 2 {
		t.Errorf("Skills count = %d, want 2", len(agentProto.Skills))
	}

	// Verify MCP servers
	if len(agentProto.McpServers) != 1 {
		t.Errorf("MCP servers count = %d, want 1", len(agentProto.McpServers))
	}
	if len(agentProto.McpServers) > 0 {
		if agentProto.McpServers[0].Name != "github" {
			t.Errorf("MCP server name = %v, want github", agentProto.McpServers[0].Name)
		}
	}

	// Verify sub-agents
	if len(agentProto.SubAgents) != 1 {
		t.Errorf("Sub-agents count = %d, want 1", len(agentProto.SubAgents))
	}

	// Verify environment variables
	if len(agentProto.EnvironmentVariables) != 1 {
		t.Errorf("Environment variables count = %d, want 1", len(agentProto.EnvironmentVariables))
	}
	if len(agentProto.EnvironmentVariables) > 0 {
		envVar := agentProto.EnvironmentVariables[0]
		if envVar.Name != "GITHUB_TOKEN" {
			t.Errorf("Env var name = %v, want GITHUB_TOKEN", envVar.Name)
		}
		if !envVar.IsSecret {
			t.Error("GITHUB_TOKEN should be marked as secret")
		}
	}

	t.Logf("✅ Manifest successfully written to: %s", manifestPath)
	t.Logf("✅ Manifest contains all agent components:")
	t.Logf("   - Agent: %s", agentProto.Name)
	t.Logf("   - Skills: %d", len(agentProto.Skills))
	t.Logf("   - MCP Servers: %d", len(agentProto.McpServers))
	t.Logf("   - Sub-Agents: %d", len(agentProto.SubAgents))
	t.Logf("   - Environment Variables: %d", len(agentProto.EnvironmentVariables))
}

// TestMultipleAgents_AllSynthesized verifies multi-agent support.
func TestMultipleAgents_AllSynthesized(t *testing.T) {
	// Reset registry
	defer registry.Global().Clear()
	registry.Global().Clear()

	// Create temp directory for output
	tempDir := t.TempDir()
	os.Setenv("STIGMER_OUT_DIR", tempDir)
	defer os.Unsetenv("STIGMER_OUT_DIR")

	// Create first agent
	agent1, err := agent.New(
		agent.WithName("code-reviewer"),
		agent.WithInstructions("Review code for best practices"),
		agent.WithDescription("First agent - code reviewer"),
	)
	if err != nil {
		t.Fatalf("Failed to create agent 1: %v", err)
	}

	// Create second agent
	agent2, err := agent.New(
		agent.WithName("security-analyzer"),
		agent.WithInstructions("Analyze code for security vulnerabilities"),
		agent.WithDescription("Second agent - security analyzer"),
	)
	if err != nil {
		t.Fatalf("Failed to create agent 2: %v", err)
	}

	// Create third agent
	agent3, err := agent.New(
		agent.WithName("performance-optimizer"),
		agent.WithInstructions("Optimize code for performance"),
		agent.WithDescription("Third agent - performance optimizer"),
	)
	if err != nil {
		t.Fatalf("Failed to create agent 3: %v", err)
	}

	// Verify registry has all agents
	if count := registry.Global().Count(); count != 3 {
		t.Errorf("Registry count = %d, want 3", count)
	}

	// Call AutoSynth
	AutoSynth()

	// Read manifest
	manifestPath := filepath.Join(tempDir, "manifest.pb")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("Failed to read manifest.pb: %v", err)
	}

	var manifest agentv1.AgentManifest
	if err := proto.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("Failed to unmarshal manifest.pb: %v", err)
	}

	// Verify ALL agents were written
	if len(manifest.Agents) != 3 {
		t.Fatalf("Manifest contains %d agents, want 3", len(manifest.Agents))
	}

	// Verify agent names match
	expectedNames := []string{agent1.Name, agent2.Name, agent3.Name}
	for i, agentProto := range manifest.Agents {
		if agentProto.Name != expectedNames[i] {
			t.Errorf("Agent[%d] name = %v, want %v", i, agentProto.Name, expectedNames[i])
		}
		if agentProto.Instructions == "" {
			t.Errorf("Agent[%d] has empty instructions", i)
		}
	}

	t.Logf("✅ SUCCESS: All %d agents synthesized", len(manifest.Agents))
	t.Logf("   Created agents: %s, %s, %s", agent1.Name, agent2.Name, agent3.Name)
	t.Logf("   Synthesized agents:")
	for i, a := range manifest.Agents {
		t.Logf("      %d. %s", i+1, a.Name)
	}
}

// TestConcurrentSynthesis_IsolatedDirectories demonstrates CLI session isolation.
func TestConcurrentSynthesis_IsolatedDirectories(t *testing.T) {
	// Simulate two concurrent CLI sessions with different output directories

	// Session 1
	t.Run("session-1", func(t *testing.T) {
		registry.Global().Clear()
		defer registry.Global().Clear()

		tempDir1 := t.TempDir()
		os.Setenv("STIGMER_OUT_DIR", tempDir1)
		defer os.Unsetenv("STIGMER_OUT_DIR")

		agent1, _ := agent.New(
			agent.WithName("session-1-agent"),
			agent.WithInstructions("Agent for session 1"),
		)

		AutoSynth()

		// Verify manifest in session 1 directory
		manifestPath := filepath.Join(tempDir1, "manifest.pb")
		data, _ := os.ReadFile(manifestPath)
		var manifest agentv1.AgentManifest
		proto.Unmarshal(data, &manifest)

		if len(manifest.Agents) != 1 {
			t.Errorf("Session 1 has %d agents, want 1", len(manifest.Agents))
		}
		if len(manifest.Agents) > 0 && manifest.Agents[0].Name != agent1.Name {
			t.Errorf("Session 1 manifest = %v, want %v", manifest.Agents[0].Name, agent1.Name)
		}

		t.Logf("✅ Session 1 isolated: %s → %s", agent1.Name, tempDir1)
	})

	// Session 2 (separate registry state)
	t.Run("session-2", func(t *testing.T) {
		registry.Global().Clear()
		defer registry.Global().Clear()

		tempDir2 := t.TempDir()
		os.Setenv("STIGMER_OUT_DIR", tempDir2)
		defer os.Unsetenv("STIGMER_OUT_DIR")

		agent2, _ := agent.New(
			agent.WithName("session-2-agent"),
			agent.WithInstructions("Agent for session 2"),
		)

		AutoSynth()

		// Verify manifest in session 2 directory
		manifestPath := filepath.Join(tempDir2, "manifest.pb")
		data, _ := os.ReadFile(manifestPath)
		var manifest agentv1.AgentManifest
		proto.Unmarshal(data, &manifest)

		if len(manifest.Agents) != 1 {
			t.Errorf("Session 2 has %d agents, want 1", len(manifest.Agents))
		}
		if len(manifest.Agents) > 0 && manifest.Agents[0].Name != agent2.Name {
			t.Errorf("Session 2 manifest = %v, want %v", manifest.Agents[0].Name, agent2.Name)
		}

		t.Logf("✅ Session 2 isolated: %s → %s", agent2.Name, tempDir2)
	})
}
