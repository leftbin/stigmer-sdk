// Package templates provides canonical code templates used by the Stigmer CLI
// and documentation. These templates demonstrate proper SDK usage patterns and
// serve as the single source of truth for generated code.
//
// The CLI's `stigmer init` command imports and uses these templates to ensure
// generated code stays in sync with SDK capabilities and best practices.
package templates

// BasicAgent returns a complete, minimal example of creating an agent.
// This template demonstrates the simplest possible agent configuration with
// only required fields.
//
// Used by: stigmer init (when --template=agent flag is used)
// Demonstrates: agent.New(), stigmer.Run(), minimal configuration
func BasicAgent() string {
	return `package main

import (
	"log"

	"github.com/leftbin/stigmer-sdk/go/agent"
	"github.com/leftbin/stigmer-sdk/go/stigmer"
)

func main() {
	// Use stigmer.Run() for automatic context and synthesis management
	err := stigmer.Run(func(ctx *stigmer.Context) error {
		// Create a basic agent with required fields only
		jokeAgent, err := agent.New(ctx,
			agent.WithName("joke-buddy"),
			agent.WithInstructions(` + "`" + `You are a friendly AI that tells programming jokes and puns.
When someone interacts with you, respond with a light-hearted programming joke or pun.
Keep it fun, simple, and appropriate for all audiences.

Examples:
- Why do programmers prefer dark mode? Because light attracts bugs!
- How many programmers does it take to change a light bulb? None, that's a hardware problem.
- A SQL query walks into a bar, walks up to two tables and asks: "Can I join you?"` + "`" + `),
			agent.WithDescription("A friendly AI that tells programming jokes"),
		)
		if err != nil {
			return err
		}

		log.Println("✅ Created joke-telling agent:")
		log.Printf("   Name: %s\n", jokeAgent.Name)
		log.Printf("   Description: %s\n", jokeAgent.Description)
		log.Println("\n🚀 Agent created successfully!")
		log.Println("   Deploy with: stigmer apply")

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Println("✅ Resources synthesized successfully!")
}
`
}

// BasicWorkflow returns a complete, minimal example of creating a workflow.
// This template demonstrates a simple HTTP GET workflow with task dependencies.
//
// Used by: stigmer init (when --template=workflow flag is used)
// Demonstrates: workflow.New(), context config, HttpGet, SetVars, implicit dependencies
func BasicWorkflow() string {
	return `package main

import (
	"log"

	"github.com/leftbin/stigmer-sdk/go/stigmer"
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

func main() {
	// Use stigmer.Run() for automatic context and synthesis management
	err := stigmer.Run(func(ctx *stigmer.Context) error {
		// Use context for shared configuration (Pulumi-aligned pattern)
		apiBase := ctx.SetString("apiBase", "https://jsonplaceholder.typicode.com")

		// Create workflow with context
		wf, err := workflow.New(ctx,
			workflow.WithNamespace("demo"),
			workflow.WithName("basic-data-fetch"),
			workflow.WithVersion("1.0.0"),
			workflow.WithDescription("A simple workflow that fetches data from an API"),
		)
		if err != nil {
			return err
		}

		// Build endpoint URL using context config
		endpoint := apiBase.Concat("/posts/1")

		// Task 1: Fetch data from API (clean, one-liner!)
		// Dependencies are implicit - no ThenRef needed!
		fetchTask := wf.HttpGet("fetchData", endpoint,
			workflow.Header("Content-Type", "application/json"),
			workflow.Timeout(30),
		)

		// Task 2: Process response using direct task references
		// Dependencies are automatic through field references!
		processTask := wf.SetVars("processResponse",
			"postTitle", fetchTask.Field("title"),
			"postBody", fetchTask.Field("body"),
			"status", "completed",
		)

		log.Println("✅ Created data-fetching workflow:")
		log.Printf("   Name: %s\n", wf.Document.Name)
		log.Printf("   Description: %s\n", wf.Description)
		log.Printf("   Tasks: %d\n", len(wf.Tasks))
		log.Printf("     - %s (HTTP GET)\n", fetchTask.Name)
		log.Printf("     - %s (depends on %s implicitly)\n", processTask.Name, fetchTask.Name)
		log.Println("\n🚀 Workflow created successfully!")
		log.Println("   Deploy with: stigmer apply")

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Println("✅ Resources synthesized successfully!")
}
`
}

// AgentAndWorkflow returns a combined example with both agent and workflow.
// This is the default template used by `stigmer init` to demonstrate both
// major SDK capabilities in a single project.
//
// Used by: stigmer init (default template)
// Demonstrates: agent.New(), workflow.New(), stigmer.Run(), combined resources
func AgentAndWorkflow() string {
	return `package main

import (
	"log"

	"github.com/leftbin/stigmer-sdk/go/agent"
	"github.com/leftbin/stigmer-sdk/go/stigmer"
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

func main() {
	// Use stigmer.Run() for automatic context and synthesis management
	err := stigmer.Run(func(ctx *stigmer.Context) error {
		// ============================================
		// AGENT: Create a simple joke-telling agent
		// ============================================
		jokeAgent, err := agent.New(ctx,
			agent.WithName("joke-buddy"),
			agent.WithInstructions(` + "`" + `You are a friendly AI that tells programming jokes and puns.
When someone interacts with you, respond with a light-hearted programming joke or pun.
Keep it fun, simple, and appropriate for all audiences.

Examples:
- Why do programmers prefer dark mode? Because light attracts bugs!
- How many programmers does it take to change a light bulb? None, that's a hardware problem.
- A SQL query walks into a bar, walks up to two tables and asks: "Can I join you?"` + "`" + `),
			agent.WithDescription("A friendly AI that tells programming jokes"),
		)
		if err != nil {
			return err
		}

		log.Println("✅ Created joke-telling agent:")
		log.Printf("   Name: %s\n", jokeAgent.Name)
		log.Printf("   Description: %s\n", jokeAgent.Description)

		// ============================================
		// WORKFLOW: Create a basic data-fetching workflow
		// ============================================

		// Use context for shared configuration (Pulumi-aligned pattern)
		apiBase := ctx.SetString("apiBase", "https://jsonplaceholder.typicode.com")

		// Create workflow with context
		wf, err := workflow.New(ctx,
			workflow.WithNamespace("demo"),
			workflow.WithName("basic-data-fetch"),
			workflow.WithVersion("1.0.0"),
			workflow.WithDescription("A simple workflow that fetches data from an API"),
		)
		if err != nil {
			return err
		}

		// Build endpoint URL using context config
		endpoint := apiBase.Concat("/posts/1")

		// Task 1: Fetch data from API (clean, one-liner!)
		// Dependencies are implicit - no ThenRef needed!
		fetchTask := wf.HttpGet("fetchData", endpoint,
			workflow.Header("Content-Type", "application/json"),
			workflow.Timeout(30),
		)

		// Task 2: Process response using direct task references
		// Dependencies are automatic through field references!
		processTask := wf.SetVars("processResponse",
			"postTitle", fetchTask.Field("title"),
			"postBody", fetchTask.Field("body"),
			"status", "completed",
		)

		log.Println("\n✅ Created data-fetching workflow:")
		log.Printf("   Name: %s\n", wf.Document.Name)
		log.Printf("   Description: %s\n", wf.Description)
		log.Printf("   Tasks: %d\n", len(wf.Tasks))
		log.Printf("     - %s (HTTP GET)\n", fetchTask.Name)
		log.Printf("     - %s (depends on %s implicitly)\n", processTask.Name, fetchTask.Name)

		log.Println("\n🚀 Resources created successfully!")
		log.Println("   Deploy with: stigmer apply")
		log.Println("   Both agent and workflow will be deployed to your organization")

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Println("✅ Resources synthesized successfully!")
}
`
}
