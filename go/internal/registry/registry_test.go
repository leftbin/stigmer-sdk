package registry

import (
	"sync"
	"testing"
)

// mockAgent is a simple test struct that mimics agent.Agent without importing it
// This avoids import cycle: agent imports registry, registry_test shouldn't import agent
type mockAgent struct {
	Name         string
	Instructions string
}

// Helper to create a test agent
func newTestAgent(name string) *mockAgent {
	return &mockAgent{
		Name:         name,
		Instructions: "Test agent instructions",
	}
}

func TestGlobal(t *testing.T) {
	// Reset registry before test
	defer Global().Clear()

	r1 := Global()
	r2 := Global()

	if r1 != r2 {
		t.Error("Global() should return the same singleton instance")
	}
}

func TestRegisterAgent(t *testing.T) {
	// Reset registry before and after test
	defer Global().Clear()
	Global().Clear()

	testAgent := newTestAgent("test-agent")
	Global().RegisterAgent(testAgent)

	retrieved := Global().GetAgent()
	if retrieved == nil {
		t.Fatal("GetAgent() returned nil after RegisterAgent")
	}

	// Type assert to mockAgent
	agent, ok := retrieved.(*mockAgent)
	if !ok {
		t.Fatalf("GetAgent() returned wrong type: %T", retrieved)
	}

	if agent.Name != "test-agent" {
		t.Errorf("Agent name = %v, want test-agent", agent.Name)
	}
}

func TestGetAgent_NoAgentRegistered(t *testing.T) {
	// Reset registry before and after test
	defer Global().Clear()
	Global().Clear()

	retrieved := Global().GetAgent()
	if retrieved != nil {
		t.Errorf("GetAgent() = %v, want nil when no agent registered", retrieved)
	}
}

func TestHasAgent(t *testing.T) {
	// Reset registry before and after test
	defer Global().Clear()
	Global().Clear()

	// Initially no agent
	if Global().HasAgent() {
		t.Error("HasAgent() = true, want false initially")
	}

	// Register agent
	testAgent := newTestAgent("test-agent")
	Global().RegisterAgent(testAgent)

	if !Global().HasAgent() {
		t.Error("HasAgent() = false, want true after RegisterAgent")
	}

	// Clear registry
	Global().Clear()

	if Global().HasAgent() {
		t.Error("HasAgent() = true, want false after Clear")
	}
}

func TestRegisterMultipleAgents(t *testing.T) {
	// Reset registry before and after test
	defer Global().Clear()
	Global().Clear()

	agent1 := newTestAgent("agent-1")
	agent2 := newTestAgent("agent-2")
	agent3 := newTestAgent("agent-3")

	// Register first agent
	Global().RegisterAgent(agent1)
	if Global().Count() != 1 {
		t.Errorf("Count after 1 registration = %d, want 1", Global().Count())
	}

	// Register second agent (should append, not replace)
	Global().RegisterAgent(agent2)
	if Global().Count() != 2 {
		t.Errorf("Count after 2 registrations = %d, want 2", Global().Count())
	}

	// Register third agent
	Global().RegisterAgent(agent3)
	if Global().Count() != 3 {
		t.Errorf("Count after 3 registrations = %d, want 3", Global().Count())
	}

	// Verify all agents are retrievable
	agents := Global().GetAgents()
	if len(agents) != 3 {
		t.Fatalf("GetAgents() returned %d agents, want 3", len(agents))
	}

	// Verify order is preserved
	expectedNames := []string{"agent-1", "agent-2", "agent-3"}
	for i, agentInterface := range agents {
		agent, ok := agentInterface.(*mockAgent)
		if !ok {
			t.Errorf("Agent[%d] wrong type: %T", i, agentInterface)
			continue
		}
		if agent.Name != expectedNames[i] {
			t.Errorf("Agent[%d] name = %v, want %v", i, agent.Name, expectedNames[i])
		}
	}
}

func TestClear(t *testing.T) {
	// Reset registry before and after test
	defer Global().Clear()
	Global().Clear()

	testAgent1 := newTestAgent("test-agent-1")
	testAgent2 := newTestAgent("test-agent-2")
	Global().RegisterAgent(testAgent1)
	Global().RegisterAgent(testAgent2)

	// Verify agents are registered
	if Global().Count() != 2 {
		t.Fatalf("Count should be 2, got %d", Global().Count())
	}

	// Clear registry
	Global().Clear()

	// Verify all agents are removed
	if Global().Count() != 0 {
		t.Errorf("Count should be 0 after Clear(), got %d", Global().Count())
	}
	if Global().HasAgent() {
		t.Error("HasAgent() should return false after Clear()")
	}
}

// TestConcurrentAccess tests thread safety of the registry with multiple agents.
func TestConcurrentAccess(t *testing.T) {
	// Reset registry before and after test
	defer Global().Clear()
	Global().Clear()

	var wg sync.WaitGroup
	iterations := 100

	// Create test agents
	agents := make([]*mockAgent, iterations)
	for i := 0; i < iterations; i++ {
		agents[i] = newTestAgent("agent-" + string(rune(i)))
	}

	// Concurrent writes (all agents should be registered)
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			Global().RegisterAgent(agents[idx])
		}(i)
	}

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = Global().GetAgents()
			_ = Global().HasAgent()
			_ = Global().Count()
		}()
	}

	wg.Wait()

	// Verify registry has all agents after concurrent access
	if !Global().HasAgent() {
		t.Error("Registry should have agents after concurrent writes")
	}

	count := Global().Count()
	if count != iterations {
		t.Errorf("Registry count = %d, want %d", count, iterations)
	}

	allAgents := Global().GetAgents()
	if len(allAgents) != iterations {
		t.Errorf("GetAgents() returned %d agents, want %d", len(allAgents), iterations)
	}
}

// TestRegistrySingletonAcrossGoroutines ensures singleton is truly global.
func TestRegistrySingletonAcrossGoroutines(t *testing.T) {
	// Reset registry before and after test
	defer Global().Clear()
	Global().Clear()

	var wg sync.WaitGroup
	registries := make([]*Registry, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			registries[idx] = Global()
		}(i)
	}

	wg.Wait()

	// All should point to same instance
	first := registries[0]
	for i := 1; i < len(registries); i++ {
		if registries[i] != first {
			t.Errorf("Registry[%d] is different instance than Registry[0]", i)
		}
	}
}

// TestRegistryIsolation ensures Clear() properly isolates test cases.
func TestRegistryIsolation(t *testing.T) {
	tests := []struct {
		name      string
		agentName string
	}{
		{"test1", "agent-1"},
		{"test2", "agent-2"},
		{"test3", "agent-3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear before each subtest
			Global().Clear()

			// Register agent
			testAgent := newTestAgent(tt.agentName)
			Global().RegisterAgent(testAgent)

			// Verify correct agent
			retrieved := Global().GetAgent()
			if retrieved == nil {
				t.Fatal("GetAgent() returned nil")
			}
			
			agent, ok := retrieved.(*mockAgent)
			if !ok {
				t.Fatalf("GetAgent() returned wrong type: %T", retrieved)
			}
			
			if agent.Name != tt.agentName {
				t.Errorf("Agent name = %v, want %v", agent.Name, tt.agentName)
			}
		})
	}
}
