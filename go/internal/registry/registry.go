package registry

import (
	"sync"
)

// globalRegistry is the singleton instance accessed via Global().
var (
	globalRegistry *Registry
	once           sync.Once
)

// Registry is a thread-safe global singleton that holds all agents being configured.
//
// When a user creates agents using agent.New(), each agent automatically registers
// itself with this global registry. The synthesis layer (internal/synth) reads from
// this registry to convert all agents to a manifest proto.
//
// This enables the "synthesis model" where:
//  1. User defines multiple agents (via agent.New())
//  2. Each agent registers in global registry (automatic)
//  3. On exit, synth reads ALL agents from registry and writes manifest.pb (automatic)
//  4. CLI reads manifest.pb and deploys all agents (separate process)
//
// Note: Registry stores []interface{} to avoid import cycles. The actual type is []*agent.Agent.
type Registry struct {
	mu     sync.RWMutex
	agents []interface{} // Stores []*agent.Agent, but uses []interface{} to avoid import cycle
}

// Global returns the global registry singleton.
//
// The singleton is initialized once using sync.Once to ensure thread safety.
// All subsequent calls return the same instance.
func Global() *Registry {
	once.Do(func() {
		globalRegistry = &Registry{}
	})
	return globalRegistry
}

// RegisterAgent registers an agent in the global registry.
//
// This is called automatically by agent.New() and should not be called directly by users.
// Multiple agents can be registered - each call appends to the list.
//
// Thread-safe: Uses write lock to ensure concurrent calls are safe.
func (r *Registry) RegisterAgent(a interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents = append(r.agents, a)
}

// GetAgents returns all registered agents.
//
// Returns []interface{} which should be type-asserted to []*agent.Agent by the caller.
//
// Thread-safe: Uses read lock to ensure concurrent reads are safe.
func (r *Registry) GetAgents() []interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Return a copy to prevent external modification
	result := make([]interface{}, len(r.agents))
	copy(result, r.agents)
	return result
}

// GetAgent returns the first registered agent (for backward compatibility).
//
// Returns interface{} which should be type-asserted to *agent.Agent by the caller.
// Returns nil if no agents have been registered.
//
// Deprecated: Use GetAgents() instead to get all agents.
//
// Thread-safe: Uses read lock to ensure concurrent reads are safe.
func (r *Registry) GetAgent() interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.agents) == 0 {
		return nil
	}
	return r.agents[0]
}

// HasAgent returns true if at least one agent has been registered.
//
// Thread-safe: Uses read lock.
func (r *Registry) HasAgent() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents) > 0
}

// Count returns the number of registered agents.
//
// Thread-safe: Uses read lock.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents)
}

// Clear removes all registered agents from the registry.
//
// This is primarily useful for testing to reset state between test cases.
// Should not be called in production code.
//
// Thread-safe: Uses write lock.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents = nil
}
