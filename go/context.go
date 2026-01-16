package stigmeragent

import (
	"fmt"
	"sync"

	"github.com/leftbin/stigmer-sdk/go/agent"
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

// Context is the central orchestration context for Stigmer SDK.
// It provides type-safe variable management and tracks all workflows and agents
// created within its scope.
//
// Context follows the Pulumi pattern where all resources are created within
// an explicit context that manages their lifecycle.
//
// Example:
//
//	stigmeragent.Run(func(ctx *stigmeragent.Context) error {
//	    apiURL := ctx.SetString("apiURL", "https://api.example.com")
//	    
//	    wf, _ := workflow.New(ctx, ...)
//	    ag, _ := agent.New(ctx, ...)
//	    
//	    return nil
//	})
type Context struct {
	// variables stores all context variables by name
	variables map[string]Ref

	// workflows tracks all workflows created in this context
	workflows []*workflow.Workflow

	// agents tracks all agents created in this context
	agents []*agent.Agent

	// mu protects concurrent access to context state
	mu sync.RWMutex

	// synthesized tracks whether synthesis has been performed
	synthesized bool
}

// newContext creates a new Context instance.
// This is internal - users should use Run() instead.
func newContext() *Context {
	return &Context{
		variables: make(map[string]Ref),
		workflows: make([]*workflow.Workflow, 0),
		agents:    make([]*agent.Agent, 0),
	}
}

// =============================================================================
// Variable Management - Typed Setters
// =============================================================================

// SetString creates a string variable in the context and returns a typed reference.
// The variable can be used in workflows and agents, and will be resolved at runtime.
//
// Example:
//
//	apiURL := ctx.SetString("apiURL", "https://api.example.com")
//	// Use apiURL in task builders: task.WithURI(apiURL)
func (c *Context) SetString(name, value string) *StringRef {
	c.mu.Lock()
	defer c.mu.Unlock()

	ref := &StringRef{
		baseRef: baseRef{
			name:     name,
			isSecret: false,
		},
		value: value,
	}
	c.variables[name] = ref
	return ref
}

// SetSecret creates a secret string variable in the context.
// Secrets are marked as sensitive and may be handled differently during synthesis.
//
// Example:
//
//	apiKey := ctx.SetSecret("apiKey", "secret-key-123")
//	// Use in headers: task.WithHeader("Authorization", apiKey.Prepend("Bearer "))
func (c *Context) SetSecret(name, value string) *StringRef {
	c.mu.Lock()
	defer c.mu.Unlock()

	ref := &StringRef{
		baseRef: baseRef{
			name:     name,
			isSecret: true,
		},
		value: value,
	}
	c.variables[name] = ref
	return ref
}

// SetInt creates an integer variable in the context and returns a typed reference.
//
// Example:
//
//	retries := ctx.SetInt("retries", 3)
//	// Use in task builders: task.WithMaxRetries(retries)
func (c *Context) SetInt(name string, value int) *IntRef {
	c.mu.Lock()
	defer c.mu.Unlock()

	ref := &IntRef{
		baseRef: baseRef{
			name:     name,
			isSecret: false,
		},
		value: value,
	}
	c.variables[name] = ref
	return ref
}

// SetBool creates a boolean variable in the context and returns a typed reference.
//
// Example:
//
//	isProd := ctx.SetBool("isProd", true)
//	// Use in conditionals: workflow.If(isProd, ...)
func (c *Context) SetBool(name string, value bool) *BoolRef {
	c.mu.Lock()
	defer c.mu.Unlock()

	ref := &BoolRef{
		baseRef: baseRef{
			name:     name,
			isSecret: false,
		},
		value: value,
	}
	c.variables[name] = ref
	return ref
}

// SetObject creates an object (map) variable in the context and returns a typed reference.
//
// Example:
//
//	config := ctx.SetObject("config", map[string]interface{}{
//	    "database": map[string]interface{}{
//	        "host": "localhost",
//	        "port": 5432,
//	    },
//	})
//	dbHost := config.FieldAsString("database", "host")
func (c *Context) SetObject(name string, value map[string]interface{}) *ObjectRef {
	c.mu.Lock()
	defer c.mu.Unlock()

	ref := &ObjectRef{
		baseRef: baseRef{
			name:     name,
			isSecret: false,
		},
		value: value,
	}
	c.variables[name] = ref
	return ref
}

// =============================================================================
// Variable Retrieval
// =============================================================================

// Get retrieves a variable by name and returns its reference.
// Returns nil if the variable doesn't exist.
//
// Example:
//
//	ref := ctx.Get("apiURL")
//	if ref != nil {
//	    // Use the reference
//	}
func (c *Context) Get(name string) Ref {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.variables[name]
}

// GetString retrieves a string variable by name.
// Returns nil if the variable doesn't exist or is not a StringRef.
func (c *Context) GetString(name string) *StringRef {
	ref := c.Get(name)
	if stringRef, ok := ref.(*StringRef); ok {
		return stringRef
	}
	return nil
}

// GetInt retrieves an integer variable by name.
// Returns nil if the variable doesn't exist or is not an IntRef.
func (c *Context) GetInt(name string) *IntRef {
	ref := c.Get(name)
	if intRef, ok := ref.(*IntRef); ok {
		return intRef
	}
	return nil
}

// GetBool retrieves a boolean variable by name.
// Returns nil if the variable doesn't exist or is not a BoolRef.
func (c *Context) GetBool(name string) *BoolRef {
	ref := c.Get(name)
	if boolRef, ok := ref.(*BoolRef); ok {
		return boolRef
	}
	return nil
}

// GetObject retrieves an object variable by name.
// Returns nil if the variable doesn't exist or is not an ObjectRef.
func (c *Context) GetObject(name string) *ObjectRef {
	ref := c.Get(name)
	if objRef, ok := ref.(*ObjectRef); ok {
		return objRef
	}
	return nil
}

// =============================================================================
// Resource Registration
// =============================================================================

// RegisterWorkflow registers a workflow with this context.
// This is typically called automatically by workflow.New() when passed a context.
func (c *Context) RegisterWorkflow(wf *workflow.Workflow) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.workflows = append(c.workflows, wf)
}

// RegisterAgent registers an agent with this context.
// This is typically called automatically by agent.New() when passed a context.
func (c *Context) RegisterAgent(ag *agent.Agent) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.agents = append(c.agents, ag)
}

// =============================================================================
// Synthesis
// =============================================================================

// Synthesize converts all registered workflows and agents to their proto representations
// and writes them to disk. This is called automatically by Run() when the function completes.
func (c *Context) Synthesize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.synthesized {
		return fmt.Errorf("context already synthesized")
	}

	// TODO: Implement synthesis logic
	// This will:
	// 1. Convert all workflows to proto manifests
	// 2. Convert all agents to proto manifests
	// 3. Inject context variables into manifests
	// 4. Write manifests to disk

	c.synthesized = true
	return nil
}

// =============================================================================
// Context Lifecycle - Run Pattern
// =============================================================================

// Run executes a function with a new Context and automatically handles synthesis.
// This is the primary entry point for using the Stigmer SDK with typed context.
//
// The function is called with a fresh Context instance. Any workflows or agents
// created within the function are automatically registered and synthesized when
// the function completes successfully.
//
// Example:
//
//	func main() {
//	    err := stigmeragent.Run(func(ctx *stigmeragent.Context) error {
//	        apiURL := ctx.SetString("apiURL", "https://api.example.com")
//	        
//	        wf, err := workflow.New(ctx,
//	            workflow.WithName("data-pipeline"),
//	            workflow.WithNamespace("my-org"),
//	        )
//	        if err != nil {
//	            return err
//	        }
//	        
//	        task, _ := wf.AddHTTPTask(
//	            workflow.WithURI(apiURL.Append("/users")),
//	        )
//	        
//	        return nil
//	    })
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	}
func Run(fn func(*Context) error) error {
	ctx := newContext()

	// Execute the user function
	if err := fn(ctx); err != nil {
		return fmt.Errorf("context function failed: %w", err)
	}

	// Synthesize all resources
	if err := ctx.Synthesize(); err != nil {
		return fmt.Errorf("synthesis failed: %w", err)
	}

	return nil
}

// =============================================================================
// Inspection Methods (for debugging and testing)
// =============================================================================

// Variables returns a copy of all variables in the context.
// This is primarily useful for testing and debugging.
func (c *Context) Variables() map[string]Ref {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]Ref, len(c.variables))
	for k, v := range c.variables {
		result[k] = v
	}
	return result
}

// Workflows returns a copy of all workflows registered in the context.
// This is primarily useful for testing and debugging.
func (c *Context) Workflows() []*workflow.Workflow {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]*workflow.Workflow, len(c.workflows))
	copy(result, c.workflows)
	return result
}

// Agents returns a copy of all agents registered in the context.
// This is primarily useful for testing and debugging.
func (c *Context) Agents() []*agent.Agent {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]*agent.Agent, len(c.agents))
	copy(result, c.agents)
	return result
}
