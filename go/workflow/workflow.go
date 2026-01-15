package workflow

import (
	"github.com/leftbin/stigmer-sdk/go/environment"
	"github.com/leftbin/stigmer-sdk/go/internal/registry"
)

// Workflow represents a workflow orchestration definition.
//
// Workflows are orchestration definitions that execute a series of tasks sequentially
// or in parallel. They support various task types including HTTP calls, gRPC calls,
// conditional logic, loops, error handling, and more.
//
// Use New() with functional options to create a workflow:
//
//	workflow, err := workflow.New(
//	    workflow.WithNamespace("my-org"),
//	    workflow.WithName("data-pipeline"),
//	    workflow.WithVersion("1.0.0"),
//	    workflow.WithDescription("Process data from external API"),
//	)
type Workflow struct {
	// Workflow metadata (namespace, name, version, description)
	Document Document

	// Human-readable description for UI and marketplace display
	Description string

	// Ordered list of tasks that make up this workflow
	Tasks []*Task

	// Environment variables required by the workflow
	EnvironmentVariables []environment.Variable

	// Organization that owns this workflow (optional)
	Org string
}

// Option is a functional option for configuring a Workflow.
type Option func(*Workflow) error

// New creates a new Workflow with the given options.
//
// The workflow is automatically registered in the global registry for synthesis.
// When the program exits and defer stigmeragent.Complete() is called, this workflow
// will be converted to a manifest proto and written to disk.
//
// Required options:
//   - WithNamespace: workflow namespace
//   - WithName: workflow name
//
// Optional (with defaults):
//   - WithVersion: workflow version (defaults to "0.1.0" if not provided)
//   - WithDescription: human-readable description
//   - WithOrg: organization identifier
//
// Example:
//
//	workflow, err := workflow.New(
//	    workflow.WithNamespace("data-processing"),
//	    workflow.WithName("daily-sync"),
//	    workflow.WithVersion("1.0.0"),  // Optional
//	    workflow.WithDescription("Sync data from external API daily"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
func New(opts ...Option) (*Workflow, error) {
	w := &Workflow{
		Document: Document{
			DSL: "1.0.0", // Default DSL version
		},
		Tasks:                []*Task{},
		EnvironmentVariables: []environment.Variable{},
	}

	// Apply all options
	for _, opt := range opts {
		if err := opt(w); err != nil {
			return nil, err
		}
	}

	// Auto-generate version if not provided
	if w.Document.Version == "" {
		w.Document.Version = "0.1.0" // Default version for development
	}

	// Validate the workflow
	if err := validate(w); err != nil {
		return nil, err
	}

	// Register in global registry for synthesis
	registry.Global().RegisterWorkflow(w)

	return w, nil
}

// WithNamespace sets the workflow namespace.
//
// The namespace is used for organization/categorization.
// This is a required field.
//
// Example:
//
//	workflow.WithNamespace("data-processing")
func WithNamespace(namespace string) Option {
	return func(w *Workflow) error {
		w.Document.Namespace = namespace
		return nil
	}
}

// WithName sets the workflow name.
//
// The name must be unique within the namespace.
// This is a required field.
//
// Example:
//
//	workflow.WithName("daily-sync")
func WithName(name string) Option {
	return func(w *Workflow) error {
		w.Document.Name = name
		return nil
	}
}

// WithVersion sets the workflow version.
//
// The version must be valid semver (e.g., "1.0.0").
// This is a required field.
//
// Example:
//
//	workflow.WithVersion("1.0.0")
func WithVersion(version string) Option {
	return func(w *Workflow) error {
		w.Document.Version = version
		return nil
	}
}

// WithDescription sets the workflow description.
//
// Description is displayed in UI and marketplace.
// This is an optional field.
//
// Example:
//
//	workflow.WithDescription("Process data from external API")
func WithDescription(description string) Option {
	return func(w *Workflow) error {
		w.Description = description
		w.Document.Description = description
		return nil
	}
}

// WithOrg sets the organization that owns this workflow.
//
// This is an optional field.
//
// Example:
//
//	workflow.WithOrg("my-org")
func WithOrg(org string) Option {
	return func(w *Workflow) error {
		w.Org = org
		return nil
	}
}

// WithTask adds a task to the workflow.
//
// Tasks are executed in the order they are added.
//
// Example:
//
//	workflow.WithTask(workflow.SetTask("init", workflow.SetVar("x", "1")))
func WithTask(task *Task) Option {
	return func(w *Workflow) error {
		w.Tasks = append(w.Tasks, task)
		return nil
	}
}

// WithTasks adds multiple tasks to the workflow.
//
// Example:
//
//	workflow.WithTasks(
//	    workflow.SetTask("init", workflow.SetVar("x", "1")),
//	    workflow.HttpCallTask("fetch", workflow.WithHTTPGet(), workflow.WithURI("${.url}")),
//	)
func WithTasks(tasks ...*Task) Option {
	return func(w *Workflow) error {
		w.Tasks = append(w.Tasks, tasks...)
		return nil
	}
}

// WithEnvironmentVariable adds an environment variable to the workflow.
//
// Environment variables define what external configuration the workflow needs to run.
//
// Example:
//
//	apiToken, _ := environment.New(
//	    environment.WithName("API_TOKEN"),
//	    environment.WithSecret(true),
//	)
//	workflow.WithEnvironmentVariable(apiToken)
func WithEnvironmentVariable(variable environment.Variable) Option {
	return func(w *Workflow) error {
		w.EnvironmentVariables = append(w.EnvironmentVariables, variable)
		return nil
	}
}

// WithEnvironmentVariables adds multiple environment variables to the workflow.
//
// Example:
//
//	apiToken, _ := environment.New(environment.WithName("API_TOKEN"), environment.WithSecret(true))
//	apiURL, _ := environment.New(environment.WithName("API_URL"))
//	workflow.WithEnvironmentVariables(apiToken, apiURL)
func WithEnvironmentVariables(variables ...environment.Variable) Option {
	return func(w *Workflow) error {
		w.EnvironmentVariables = append(w.EnvironmentVariables, variables...)
		return nil
	}
}

// AddTask adds a task to the workflow after creation.
//
// This is a builder method that allows adding tasks after the workflow is created.
//
// Example:
//
//	wf, _ := workflow.New(workflow.WithNamespace("ns"), workflow.WithName("wf"), workflow.WithVersion("1.0.0"))
//	wf.AddTask(workflow.SetTask("init", workflow.SetVar("x", "1")))
func (w *Workflow) AddTask(task *Task) *Workflow {
	w.Tasks = append(w.Tasks, task)
	return w
}

// AddTasks adds multiple tasks to the workflow after creation.
//
// Example:
//
//	wf, _ := workflow.New(...)
//	wf.AddTasks(
//	    workflow.SetTask("init", workflow.SetVar("x", "1")),
//	    workflow.HttpCallTask("fetch", workflow.WithHTTPGet(), workflow.WithURI("${.url}")),
//	)
func (w *Workflow) AddTasks(tasks ...*Task) *Workflow {
	w.Tasks = append(w.Tasks, tasks...)
	return w
}

// AddEnvironmentVariable adds an environment variable to the workflow after creation.
//
// Example:
//
//	wf, _ := workflow.New(...)
//	apiToken, _ := environment.New(environment.WithName("API_TOKEN"))
//	wf.AddEnvironmentVariable(apiToken)
func (w *Workflow) AddEnvironmentVariable(variable environment.Variable) *Workflow {
	w.EnvironmentVariables = append(w.EnvironmentVariables, variable)
	return w
}

// AddEnvironmentVariables adds multiple environment variables to the workflow after creation.
//
// Example:
//
//	wf, _ := workflow.New(...)
//	wf.AddEnvironmentVariables(apiToken, apiURL)
func (w *Workflow) AddEnvironmentVariables(variables ...environment.Variable) *Workflow {
	w.EnvironmentVariables = append(w.EnvironmentVariables, variables...)
	return w
}

// String returns a string representation of the Workflow.
func (w *Workflow) String() string {
	return "Workflow(namespace=" + w.Document.Namespace + ", name=" + w.Document.Name + ", version=" + w.Document.Version + ")"
}
