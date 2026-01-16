package main

import (
	"fmt"
	"log"

	"github.com/leftbin/stigmer-sdk/go/stigmer"
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

// Example demonstrating context variable injection
// This tests the automatic SET task generation for ctx.SetString(), ctx.SetInt(), etc.
func main() {
	err := stigmer.Run(func(ctx *stigmer.Context) error {
		// Set various types of context variables
		apiURL := ctx.SetString("apiURL", "https://api.example.com")
		apiVersion := ctx.SetString("apiVersion", "v1")
		retries := ctx.SetInt("retries", 3)
		timeout := ctx.SetInt("timeout", 30)
		isProd := ctx.SetBool("isProd", false)
		enableDebug := ctx.SetBool("enableDebug", true)

		// Set a complex object
		config := ctx.SetObject("config", map[string]interface{}{
			"database": map[string]interface{}{
				"host": "localhost",
				"port": 5432,
			},
			"cache": map[string]interface{}{
				"enabled": true,
				"ttl":     300,
			},
		})

		// Create workflow using context variables
		_, err := workflow.New(ctx,
			workflow.WithName("context-variables-test"),
			workflow.WithNamespace("examples"),
			workflow.WithDescription("Test workflow demonstrating automatic context variable injection"),
			workflow.WithTasks(
				// Task 1: Use string concatenation with context variables
				workflow.HttpCallTask("fetch_users",
					workflow.WithHTTPGet(),
					workflow.WithURI(apiURL.Concat("/", apiVersion, "/users").Expression()), // Uses context vars!
					workflow.WithTimeout(timeout.Value()),
				),

				// Task 2: Use conditional logic with context variables
				workflow.SwitchTask("environment_check",
					workflow.WithCase(isProd.Expression(), "production_flow"),
					workflow.WithCase(enableDebug.Expression(), "debug_flow"),
					workflow.WithDefault("default_flow"),
				),

				// Task 3: Use object field access
				workflow.SetTask("database_config",
					workflow.SetVar("db_host", config.FieldAsString("database", "host").Expression()),
					workflow.SetVar("db_port", config.FieldAsInt("database", "port").Expression()),
					workflow.SetVar("cache_enabled", config.FieldAsBool("cache", "enabled").Expression()),
				),

				// Task 4: Demonstrate all context variables are accessible
				workflow.SetTask("use_all_vars",
					workflow.SetVar("api_base", apiURL.Expression()),
					workflow.SetVar("version", apiVersion.Expression()),
					workflow.SetVar("max_retries", retries.Expression()),
					workflow.SetVar("is_production", isProd.Expression()),
				),
			),
		)
		if err != nil {
			return err
		}

		fmt.Println("✅ Workflow created successfully!")
		fmt.Println("📦 Context variables will be automatically injected as SET task")
		fmt.Println("🎯 Variables set:")
		fmt.Println("   - apiURL (string): https://api.example.com")
		fmt.Println("   - apiVersion (string): v1")
		fmt.Println("   - retries (int): 3")
		fmt.Println("   - timeout (int): 30")
		fmt.Println("   - isProd (bool): false")
		fmt.Println("   - enableDebug (bool): true")
		fmt.Println("   - config (object): {database: {...}, cache: {...}}")

		return nil
	})

	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("\n✅ Synthesis complete! Check output for generated manifest.")
}
