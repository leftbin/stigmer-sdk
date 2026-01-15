// Package examples demonstrates how to create workflows using the Stigmer SDK.
package main

import (
	"log"

	"github.com/leftbin/stigmer-sdk/go/environment"
	"github.com/leftbin/stigmer-sdk/go/synthesis"
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

// This example demonstrates creating a basic workflow with SET and HTTP_CALL tasks.
//
// The workflow:
// 1. Initializes variables
// 2. Makes an HTTP GET request to fetch data
// 3. Processes the response
func main() {
	// Enable auto-synthesis - workflows will be written to manifest.pb on exit
	defer synthesis.AutoSynth()

	// Create environment variable for API token
	apiToken, err := environment.New(
		environment.WithName("API_TOKEN"),
		environment.WithSecret(true),
		environment.WithDescription("Authentication token for the API"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create the workflow
	wf, err := workflow.New(
		// Workflow metadata
		workflow.WithNamespace("data-processing"),
		workflow.WithName("basic-data-fetch"),
		workflow.WithVersion("1.0.0"),
		workflow.WithDescription("Fetch data from an external API"),
		workflow.WithOrg("my-org"),

		// Environment variables
		workflow.WithEnvironmentVariable(apiToken),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Task 1: Initialize variables
	wf.AddTask(workflow.SetTask("initialize",
		workflow.SetVar("apiURL", "https://api.example.com"),
		workflow.SetVar("retryCount", "0"),
	))

	// Task 2: Fetch data from API
	wf.AddTask(workflow.HttpCallTask("fetchData",
		workflow.WithMethod("GET"),
		workflow.WithURI("${apiURL}/data"),
		workflow.WithHeader("Authorization", "Bearer ${API_TOKEN}"),
		workflow.WithHeader("Content-Type", "application/json"),
		workflow.WithTimeout(30),
	).Export("${.}")) // Export entire response to context

	// Task 3: Process the response
	wf.AddTask(workflow.SetTask("processResponse",
		workflow.SetVar("dataCount", "${.count}"),
		workflow.SetVar("status", "success"),
	))

	log.Printf("Created workflow: %s", wf)
	log.Println("Workflow will be written to manifest.pb on exit")
}
