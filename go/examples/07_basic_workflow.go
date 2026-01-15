//go:build ignore

// Package examples demonstrates how to create workflows using the Stigmer SDK.
package main

import (
	"log"

	stigmeragent "github.com/leftbin/stigmer-sdk/go"
	"github.com/leftbin/stigmer-sdk/go/environment"
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

// This example demonstrates creating a basic workflow with SET and HTTP_CALL tasks.
//
// The workflow:
// 1. Initializes variables
// 2. Makes an HTTP GET request to fetch data
// 3. Processes the response
//
// Key improvements demonstrated:
// - Optional version (can omit for development workflows)
// - Type-safe task references with .ThenRef()
// - Type-safe setters (SetInt, SetString, SetBool)
// - High-level helpers (ExportAll, VarRef, FieldRef, Interpolate)
func main() {
	// Enable auto-synthesis - workflows will be written to manifest.pb on exit
	defer stigmeragent.Complete()

	// Create environment variable for API token
	apiToken, err := environment.New(
		environment.WithName("API_TOKEN"),
		environment.WithSecret(true),
		environment.WithDescription("Authentication token for the API"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Task 1: Initialize variables using type-safe setters
	initTask := workflow.SetTask("initialize",
		workflow.SetString("apiURL", "https://api.example.com"),
		workflow.SetInt("retryCount", 0), // Type-safe integer instead of string "0"
	)

	// Task 2: Fetch data from API using variable interpolation
	fetchTask := workflow.HttpCallTask("fetchData",
		workflow.WithHTTPGet(), // Type-safe HTTP method
		workflow.WithURI(workflow.Interpolate(workflow.VarRef("apiURL"), "/data")), // Clean interpolation
		workflow.WithHeader("Authorization", workflow.Interpolate("Bearer ", workflow.VarRef("API_TOKEN"))),
		workflow.WithHeader("Content-Type", "application/json"),
		workflow.WithTimeout(30),
	).ExportAll() // High-level helper instead of Export("${.}")

	// Task 3: Process the response using field references
	processTask := workflow.SetTask("processResponse",
		workflow.SetVar("dataCount", workflow.FieldRef("count")), // Clean field reference instead of "${.count}"
		workflow.SetString("status", "success"),
	)

	// Connect tasks using type-safe references (refactoring-safe!)
	initTask.ThenRef(fetchTask)
	fetchTask.ThenRef(processTask)

	// Create the workflow with tasks (version is optional)
	wf, err := workflow.New(
		// Required metadata
		workflow.WithNamespace("data-processing"),
		workflow.WithName("basic-data-fetch"),

		// Optional fields
		workflow.WithVersion("1.0.0"), // Optional - defaults to "0.1.0" if omitted
		workflow.WithDescription("Fetch data from an external API"),
		workflow.WithOrg("my-org"),
		workflow.WithEnvironmentVariable(apiToken),

		// Tasks
		workflow.WithTasks(initTask, fetchTask, processTask),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Created workflow: %s", wf)
	log.Println("Workflow will be written to manifest.pb on exit")
}
