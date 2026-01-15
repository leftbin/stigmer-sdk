package registry

import (
	"sync"
)

// globalRegistry is the singleton instance accessed via Global().
var (
	globalRegistry *Registry
	once           sync.Once
)

// Registry is a thread-safe global singleton that holds all agents and workflows being configured.
//
// When a user creates agents using agent.New() or workflows using workflow.New(), each resource
// automatically registers itself with this global registry. The synthesis layer (internal/synth)
// reads from this registry to convert all resources to a manifest proto.
//
// This enables the "synthesis model" where:
//  1. User defines multiple agents and workflows (via agent.New(), workflow.New())
//  2. Each resource registers in global registry (automatic)
//  3. On exit, synth reads ALL resources from registry and writes manifest.pb (automatic)
//  4. CLI reads manifest.pb and deploys all resources (separate process)
//
// Note: Registry stores []interface{} to avoid import cycles.
type Registry struct {
	mu        sync.RWMutex
	agents    []interface{} // Stores []*agent.Agent, but uses []interface{} to avoid import cycle
	workflows []interface{} // Stores []*workflow.Workflow, but uses []interface{} to avoid import cycle
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

// Clear removes all registered agents and workflows from the registry.
//
// This is primarily useful for testing to reset state between test cases.
// Should not be called in production code.
//
// Thread-safe: Uses write lock.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents = nil
	r.workflows = nil
}

// RegisterWorkflow registers a workflow in the global registry.
//
// This is called automatically by workflow.New() and should not be called directly by users.
// Multiple workflows can be registered - each call appends to the list.
//
// Thread-safe: Uses write lock to ensure concurrent calls are safe.
func (r *Registry) RegisterWorkflow(w interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.workflows = append(r.workflows, w)
}

// GetWorkflows returns all registered workflows.
//
// Returns []interface{} which should be type-asserted to []*workflow.Workflow by the caller.
//
// Thread-safe: Uses read lock to ensure concurrent reads are safe.
func (r *Registry) GetWorkflows() []interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Return a copy to prevent external modification
	result := make([]interface{}, len(r.workflows))
	copy(result, r.workflows)
	return result
}

// GetWorkflow returns the first registered workflow (for backward compatibility).
//
// Returns interface{} which should be type-asserted to *workflow.Workflow by the caller.
// Returns nil if no workflows have been registered.
//
// Deprecated: Use GetWorkflows() instead to get all workflows.
//
// Thread-safe: Uses read lock to ensure concurrent reads are safe.
func (r *Registry) GetWorkflow() interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.workflows) == 0 {
		return nil
	}
	return r.workflows[0]
}

// HasWorkflow returns true if at least one workflow has been registered.
//
// Thread-safe: Uses read lock.
func (r *Registry) HasWorkflow() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.workflows) > 0
}

// WorkflowCount returns the number of registered workflows.
//
// Thread-safe: Uses read lock.
func (r *Registry) WorkflowCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.workflows)
}

// HasAny returns true if at least one agent or workflow has been registered.
//
// Thread-safe: Uses read lock.
func (r *Registry) HasAny() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents) > 0 || len(r.workflows) > 0
}
