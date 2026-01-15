// Package workflow provides types and builders for creating Stigmer workflows.
//
// Workflows are orchestration definitions that execute a series of tasks sequentially
// or in parallel. They support various task types including HTTP calls, gRPC calls,
// conditional logic, loops, error handling, and more.
//
// # Creating Workflows
//
// Use New() with functional options to create a workflow:
//
//	workflow, err := workflow.New(
//	    workflow.WithNamespace("my-org"),
//	    workflow.WithName("data-pipeline"),
//	    workflow.WithVersion("1.0.0"),
//	    workflow.WithDescription("Process data from external API"),
//	)
//
// # Adding Tasks
//
// Tasks can be added using task builder functions:
//
//	// Set variables
//	workflow.AddTask(workflow.SetTask("init",
//	    workflow.SetVar("apiURL", "https://api.example.com"),
//	    workflow.SetVar("count", "0"),
//	))
//
//	// HTTP call
//	workflow.AddTask(workflow.HttpCallTask("fetchData",
//	    workflow.WithMethod("GET"),
//	    workflow.WithURI("${apiURL}/data"),
//	    workflow.WithHeader("Authorization", "Bearer ${TOKEN}"),
//	))
//
//	// Conditional logic
//	workflow.AddTask(workflow.SwitchTask("processResponse",
//	    workflow.WithCase("${.status == 200}", "processSuccess"),
//	    workflow.WithDefault("handleError"),
//	))
//
// # Task Types
//
// The workflow package supports all Zigflow DSL task types:
//
//   - SET: Assign variables in workflow state
//   - HTTP_CALL: Make HTTP requests (GET, POST, PUT, DELETE, PATCH)
//   - GRPC_CALL: Make gRPC calls
//   - SWITCH: Conditional branching
//   - FOR: Iterate over collections
//   - FORK: Parallel task execution
//   - TRY: Error handling with catch blocks
//   - LISTEN: Wait for external events
//   - WAIT: Pause execution for a duration
//   - CALL_ACTIVITY: Execute Temporal activities
//   - RAISE: Throw errors
//   - RUN: Execute sub-workflows
//
// # Environment Variables
//
// Workflows can declare required environment variables:
//
//	import "github.com/leftbin/stigmer-sdk/go/environment"
//
//	apiToken, _ := environment.New(
//	    environment.WithName("TOKEN"),
//	    environment.WithSecret(true),
//	    environment.WithDescription("API authentication token"),
//	)
//	workflow.AddEnvironmentVariable(apiToken)
//
// # Flow Control
//
// Control task execution flow using export and flow directives:
//
//	task := workflow.HttpCallTask("fetchData",
//	    workflow.WithMethod("GET"),
//	    workflow.WithURI("${apiURL}/data"),
//	)
//	task.Export("${.}") // Export entire response
//	task.Then("processData") // Jump to task named "processData"
//
// # Registration and Synthesis
//
// Workflows are automatically registered in the global registry when created.
// When the program exits (with defer synth.AutoSynth()), workflows are converted
// to manifest protos and written to disk.
//
//	import "github.com/leftbin/stigmer-sdk/go/synthesis"
//
//	func main() {
//	    defer synthesis.AutoSynth()
//
//	    workflow, _ := workflow.New(...)
//	    // ... add tasks
//	}
//
// # Validation
//
// Workflows are validated when created and when tasks are added:
//
//   - Document: namespace, name, and version are required (semver)
//   - Tasks: must have at least one task
//   - Task names: must be unique within workflow
//   - Task configs: validated based on task type
//
// # Example
//
// Complete workflow example:
//
//	package main
//
//	import (
//	    "log"
//	    "github.com/leftbin/stigmer-sdk/go/workflow"
//	    "github.com/leftbin/stigmer-sdk/go/environment"
//	    "github.com/leftbin/stigmer-sdk/go/synthesis"
//	)
//
//	func main() {
//	    defer synthesis.AutoSynth()
//
//	    // Create environment variable
//	    apiToken, _ := environment.New(
//	        environment.WithName("API_TOKEN"),
//	        environment.WithSecret(true),
//	    )
//
//	    // Create workflow
//	    wf, err := workflow.New(
//	        workflow.WithNamespace("data-processing"),
//	        workflow.WithName("daily-sync"),
//	        workflow.WithVersion("1.0.0"),
//	        workflow.WithDescription("Sync data from external API daily"),
//	        workflow.WithEnvironmentVariable(apiToken),
//	    )
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // Add tasks
//	    wf.AddTask(workflow.SetTask("init",
//	        workflow.SetVar("apiURL", "https://api.example.com"),
//	    ))
//
//	    wf.AddTask(workflow.HttpCallTask("fetchData",
//	        workflow.WithMethod("GET"),
//	        workflow.WithURI("${apiURL}/data"),
//	        workflow.WithHeader("Authorization", "Bearer ${API_TOKEN}"),
//	    )).Export("${.}").Then("processData")
//
//	    wf.AddTask(workflow.SetTask("processData",
//	        workflow.SetVar("processed", "${.count}"),
//	    ))
//	}
package workflow
