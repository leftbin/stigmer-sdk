//go:build ignore
// Example 01: Basic Agent
//
// This example demonstrates creating a simple agent with just the required fields.
package main

import (
	"fmt"
	"log"

	"github.com/leftbin/stigmer-sdk/go/agent"
)

func main() {
	// Create a basic agent with required fields only
	basicAgent, err := agent.New(
		agent.WithName("code-reviewer"),
		agent.WithInstructions("Review code and suggest improvements based on best practices"),
	)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	fmt.Println("✅ Created basic agent:")
	fmt.Printf("   Name: %s\n", basicAgent.Name)
	fmt.Printf("   Instructions: %s\n", basicAgent.Instructions)

	// Create an agent with optional fields
	fullAgent, err := agent.New(
		agent.WithName("code-reviewer-pro"),
		agent.WithInstructions("Review code and suggest improvements based on best practices and security considerations"),
		agent.WithDescription("Professional code reviewer with security focus"),
		agent.WithIconURL("https://example.com/icons/code-reviewer.png"),
		agent.WithOrg("my-org"),
	)
	if err != nil {
		log.Fatalf("Failed to create full agent: %v", err)
	}

	fmt.Println("\n✅ Created full agent:")
	fmt.Printf("   Name: %s\n", fullAgent.Name)
	fmt.Printf("   Instructions: %s\n", fullAgent.Instructions)
	fmt.Printf("   Description: %s\n", fullAgent.Description)
	fmt.Printf("   IconURL: %s\n", fullAgent.IconURL)
	fmt.Printf("   Org: %s\n", fullAgent.Org)

	// Note: Proto conversion is handled by the CLI.
	// The SDK provides user-friendly objects for defining agents.
	// When you run `stigmer deploy`, the CLI will convert these to proto messages.

	// Example of validation error
	fmt.Println("\n❌ Attempting to create invalid agent:")
	_, err = agent.New(
		agent.WithName("Invalid Name!"),
		agent.WithInstructions("Test"),
	)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	}

	fmt.Println("\n✅ Example completed successfully!")
}
