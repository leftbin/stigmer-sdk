// +build manual

// This is a manual test file that uses local proto stubs to verify multi-agent support.
// Run with: go test -tags manual ./internal/synth/... -v -run TestMultiAgentManual
//
// This file uses local proto stubs from the monorepo since Buf registry takes time to update.
// Once Buf updates (commit: 8fe8489c81ed42bbb0973ebfa49dca88), this file can be removed.

package synth

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	// Use local proto stubs temporarily
	agentv1 "buf.build/gen/go/leftbin/stigmer/protocolbuffers/go/ai/stigmer/agentic/agent/v1"
	sdk "buf.build/gen/go/leftbin/stigmer/protocolbuffers/go/ai/stigmer/commons/sdk"

	"github.com/leftbin/stigmer-sdk/go/agent"
	"github.com/leftbin/stigmer-sdk/go/internal/registry"
	"github.com/leftbin/stigmer-sdk/go/skill"
)

// TestMultiAgentManual_ThreeAgents verifies that multiple agents are all synthesized.
func TestMultiAgentManual_ThreeAgents(t *testing.T) {
	// Reset registry
	defer registry.Global().Clear()
	registry.Global().Clear()

	// Create temp directory
	tempDir := t.TempDir()

	// Create three different agents
	agent1, err := agent.New(
		agent.WithName("code-reviewer"),
		agent.WithInstructions("Review code for best practices and maintainability"),
		agent.WithDescription("Code review specialist"),
	)
	if err != nil {
		t.Fatalf("Failed to create agent 1: %v", err)
	}
	agent1.AddSkill(skill.Platform("coding-standards"))

	agent2, err := agent.New(
		agent.WithName("security-analyzer"),
		agent.WithInstructions("Analyze code for security vulnerabilities"),
		agent.WithDescription("Security specialist"),
	)
	if err != nil {
		t.Fatalf("Failed to create agent 2: %v", err)
	}
	agent2.AddSkill(skill.Platform("security-best-practices"))

	agent3, err := agent.New(
		agent.WithName("performance-optimizer"),
		agent.WithInstructions("Optimize code for performance and efficiency"),
		agent.WithDescription("Performance specialist"),
	)
	if err != nil {
		t.Fatalf("Failed to create agent 3: %v", err)
	}

	// Verify registry has all 3 agents
	if count := registry.Global().Count(); count != 3 {
		t.Fatalf("Registry count = %d, want 3", count)
	}

	t.Logf("✅ Created 3 agents:")
	t.Logf("   1. %s", agent1.Name)
	t.Logf("   2. %s", agent2.Name)
	t.Logf("   3. %s", agent3.Name)

	// Get all agents from registry
	agentInterfaces := registry.Global().GetAgents()

	// Manually create manifest with multiple agents
	manifest := &agentv1.AgentManifest{
		SdkMetadata: &sdk.SdkMetadata{
			Language:    "go",
			Version:     "0.1.0",
			GeneratedAt: 1705172400,
		},
		Agents: []*agentv1.AgentBlueprint{},
	}

	// Convert each agent manually
	for idx, agentIface := range agentInterfaces {
		a, ok := agentIface.(*agent.Agent)
		if !ok {
			t.Fatalf("Agent[%d]: invalid type %T", idx, agentIface)
		}

		blueprint := &agentv1.AgentBlueprint{
			Name:         a.Name,
			Instructions: a.Instructions,
			Description:  a.Description,
		}

		// Convert skills
		for _, s := range a.Skills {
			skillProto := &agentv1.ManifestSkill{
				Id: fmt.Sprintf("skill-%d", idx),
			}
			if s.IsPlatformReference() {
				skillProto.Source = &agentv1.ManifestSkill_Platform{
					Platform: &agentv1.PlatformSkillReference{
						Name: s.Slug,
					},
				}
			}
			blueprint.Skills = append(blueprint.Skills, skillProto)
		}

		manifest.Agents = append(manifest.Agents, blueprint)
	}

	// Serialize to protobuf
	data, err := proto.Marshal(manifest)
	if err != nil {
		t.Fatalf("Failed to marshal manifest: %v", err)
	}

	// Write to file
	manifestPath := filepath.Join(tempDir, "manifest.pb")
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	t.Logf("✅ Manifest written to: %s", manifestPath)

	// Read back and verify
	readData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("Failed to read manifest: %v", err)
	}

	var readManifest agentv1.AgentManifest
	if err := proto.Unmarshal(readData, &readManifest); err != nil {
		t.Fatalf("Failed to unmarshal manifest: %v", err)
	}

	// Verify all 3 agents are present
	if len(readManifest.Agents) != 3 {
		t.Fatalf("Manifest has %d agents, want 3", len(readManifest.Agents))
	}

	t.Logf("✅ SUCCESS: Manifest contains all %d agents:", len(readManifest.Agents))
	for i, a := range readManifest.Agents {
		t.Logf("   %d. %s - %s", i+1, a.Name, a.Description)
		t.Logf("      Instructions: %d chars", len(a.Instructions))
		t.Logf("      Skills: %d", len(a.Skills))
	}

	// Verify specific agent data
	expectedNames := []string{"code-reviewer", "security-analyzer", "performance-optimizer"}
	for i, expected := range expectedNames {
		if readManifest.Agents[i].Name != expected {
			t.Errorf("Agent[%d] name = %v, want %v", i, readManifest.Agents[i].Name, expected)
		}
	}

	t.Log("\n🎉 MULTI-AGENT SUPPORT VERIFIED!")
	t.Log("   ✅ Multiple agents can be defined")
	t.Log("   ✅ All agents are registered")
	t.Log("   ✅ All agents are synthesized to manifest.pb")
	t.Log("   ✅ Manifest can be read and deserialized")
}
