//go:build ignore

// Package examples demonstrates how to create workflows using the Stigmer SDK with typed context.
package main

import (
	"log"

	stigmeragent "github.com/leftbin/stigmer-sdk/go"
	"github.com/leftbin/stigmer-sdk/go/environment"
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

// This example demonstrates creating a workflow with typed context variables.
//
// The workflow:
// 1. Initializes variables using typed context (SetString, SetInt)
// 2. Makes an HTTP GET request using context variables
// 3. Processes the response using field references
//
// Key features demonstrated:
// - stigmer.Run() pattern for automatic context management
// - Typed context variables (apiURL, retryCount)
// - Compile-time checked references (no string typos)
// - IDE autocomplete for context variables
// - Type-safe task builders accepting Ref types
// - Automatic synthesis on completion
func main() {
	// Use stigmer.Run() for automatic context and synthesis management
	err := stigmeragent.Run(func(ctx *stigmeragent.Context) error {
		// Create typed context variables (compile-time checked!)
		apiURL := ctx.SetString("apiURL", "https://jsonplaceholder.typicode.com")
		retryCount := ctx.SetInt("retryCount", 0)

		// Create environment variable for API token
		apiToken, err := environment.New(
			environment.WithName("API_TOKEN"),
			environment.WithSecret(true),
			environment.WithDescription("Authentication token for the API"),
		)
		if err != nil {
			return err
		}

		// Task 1: Initialize variables using context references
		// Note: We're using the typed references directly!
		initTask := workflow.SetTask("initialize",
			workflow.SetVar("currentURL", apiURL),         // Use typed reference
			workflow.SetVar("currentRetries", retryCount), // Use typed reference
		)

		// Task 2: Fetch data from API using typed context variable
		// The apiURL reference is compile-time checked - no string typos possible!
		endpoint := apiURL.Concat("/posts/1") // Type-safe string concatenation

		fetchTask := workflow.HttpCallTask("fetchData",
			workflow.WithHTTPGet(),
			workflow.WithURI(endpoint), // Use the typed reference
			workflow.WithHeader("Content-Type", "application/json"),
			workflow.WithTimeout(30),
		).ExportAll()

		// Task 3: Process the response using field references
		processTask := workflow.SetTask("processResponse",
			workflow.SetVar("postTitle", workflow.FieldRef("title")),
			workflow.SetVar("postBody", workflow.FieldRef("body")),
			workflow.SetString("status", "success"),
		)

		// Connect tasks using type-safe references (refactoring-safe!)
		initTask.ThenRef(fetchTask)
		fetchTask.ThenRef(processTask)

		// Create the workflow with typed context
		wf, err := workflow.NewWithContext(ctx,
			// Required metadata
			workflow.WithNamespace("data-processing"),
			workflow.WithName("basic-data-fetch"),

			// Optional fields
			workflow.WithVersion("1.0.0"),
			workflow.WithDescription("Fetch data from an external API using typed context"),
			workflow.WithOrg("my-org"),
			workflow.WithEnvironmentVariable(apiToken),

			// Tasks
			workflow.WithTasks(initTask, fetchTask, processTask),
		)
		if err != nil {
			return err
		}

		log.Printf("Created workflow: %s", wf)
		log.Println("Workflow will be synthesized automatically on completion")
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Println("✅ Workflow created and synthesized successfully!")
}
