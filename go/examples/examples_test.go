package examples_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	agentv1 "buf.build/gen/go/leftbin/stigmer/protocolbuffers/go/ai/stigmer/agentic/agent/v1"
	workflowv1 "buf.build/gen/go/leftbin/stigmer/protocolbuffers/go/ai/stigmer/agentic/workflow/v1"

	"github.com/leftbin/stigmer-sdk/go/stigmer"
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

// TestExample01_BasicAgent tests the basic agent example
func TestExample01_BasicAgent(t *testing.T) {
	runExampleTest(t, "01_basic_agent.go", func(t *testing.T, outputDir string) {
		// Verify agent-manifest.pb was created
		manifestPath := filepath.Join(outputDir, "agent-manifest.pb")
		assertFileExists(t, manifestPath)

		// Verify manifest content
		var manifest agentv1.AgentManifest
		readProtoManifest(t, manifestPath, &manifest)

		// Should have 2 agents (basicAgent and fullAgent)
		if len(manifest.Agents) != 2 {
			t.Errorf("Expected 2 agents, got %d", len(manifest.Agents))
		}

		// Verify first agent
		if len(manifest.Agents) > 0 {
			agent := manifest.Agents[0]
			if agent.Name != "code-reviewer" {
				t.Errorf("First agent name = %v, want code-reviewer", agent.Name)
			}
			if agent.Instructions == "" {
				t.Error("First agent instructions should not be empty")
			}
		}

		// Verify second agent has optional fields
		if len(manifest.Agents) > 1 {
			agent := manifest.Agents[1]
			if agent.Name != "code-reviewer-pro" {
				t.Errorf("Second agent name = %v, want code-reviewer-pro", agent.Name)
			}
			if agent.Description == "" {
				t.Error("Second agent should have description")
			}
			if agent.IconUrl == "" {
				t.Error("Second agent should have icon URL")
			}
		}
	})
}

// TestExample02_AgentWithSkills tests the agent with skills example
func TestExample02_AgentWithSkills(t *testing.T) {
	runExampleTest(t, "02_agent_with_skills.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "agent-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest agentv1.AgentManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Agents) < 1 {
			t.Fatal("Expected at least 1 agent")
		}

		agent := manifest.Agents[0]
		if len(agent.Skills) == 0 {
			t.Error("Agent should have skills")
		}
	})
}

// TestExample03_AgentWithMCPServers tests the agent with MCP servers example
func TestExample03_AgentWithMCPServers(t *testing.T) {
	runExampleTest(t, "03_agent_with_mcp_servers.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "agent-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest agentv1.AgentManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Agents) < 1 {
			t.Fatal("Expected at least 1 agent")
		}

		agent := manifest.Agents[0]
		if len(agent.McpServers) == 0 {
			t.Error("Agent should have MCP servers")
		}
	})
}

// TestExample04_AgentWithSubagents tests the agent with subagents example
func TestExample04_AgentWithSubagents(t *testing.T) {
	runExampleTest(t, "04_agent_with_subagents.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "agent-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest agentv1.AgentManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Agents) < 1 {
			t.Fatal("Expected at least 1 agent")
		}

		agent := manifest.Agents[0]
		if len(agent.SubAgents) == 0 {
			t.Error("Agent should have sub-agents")
		}
	})
}

// TestExample05_AgentWithEnvironmentVariables tests the agent with environment variables example
func TestExample05_AgentWithEnvironmentVariables(t *testing.T) {
	runExampleTest(t, "05_agent_with_environment_variables.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "agent-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest agentv1.AgentManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Agents) < 1 {
			t.Fatal("Expected at least 1 agent")
		}

		agent := manifest.Agents[0]
		if len(agent.EnvironmentVariables) == 0 {
			t.Error("Agent should have environment variables")
		}
	})
}

// TestExample06_AgentWithInstructionsFromFiles tests the agent with instructions from files example
func TestExample06_AgentWithInstructionsFromFiles(t *testing.T) {
	runExampleTest(t, "06_agent_with_instructions_from_files.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "agent-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest agentv1.AgentManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Agents) < 1 {
			t.Fatal("Expected at least 1 agent")
		}

		// Verify agent loaded instructions from files
		agent := manifest.Agents[0]
		if agent.Instructions == "" {
			t.Error("Agent should have instructions loaded from file")
		}
	})
}

// TestExample07_BasicWorkflow tests the basic workflow example
// This test also validates:
// 1. Compile-time variable resolution (no __stigmer_init_context task)
// 2. Auto-export functionality (tasks export when .Field() is called)
func TestExample07_BasicWorkflow(t *testing.T) {
	runExampleTest(t, "07_basic_workflow.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest workflowv1.WorkflowManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Workflows) != 1 {
			t.Fatalf("Expected 1 workflow, got %d", len(manifest.Workflows))
		}

		workflow := manifest.Workflows[0]
		if workflow.Spec == nil || workflow.Spec.Document == nil {
			t.Fatal("Workflow spec or document is nil")
		}

		if workflow.Spec.Document.Namespace != "data-processing" {
			t.Errorf("Workflow namespace = %v, want data-processing", workflow.Spec.Document.Namespace)
		}
		if workflow.Spec.Document.Name != "basic-data-fetch" {
			t.Errorf("Workflow name = %v, want basic-data-fetch", workflow.Spec.Document.Name)
		}

		// Verify workflow has tasks
		if len(workflow.Spec.Tasks) == 0 {
			t.Error("Workflow should have tasks")
		}

		// COMPILE-TIME VARIABLE RESOLUTION TEST:
		// Should have ONLY 2 user tasks (fetchData + processResponse)
		// NO __stigmer_init_context task (context variables resolved at compile-time)
		if len(workflow.Spec.Tasks) != 2 {
			t.Errorf("Expected 2 tasks (NO context init with compile-time resolution), got %d", len(workflow.Spec.Tasks))
		}

		// Verify NO __stigmer_init_context task exists
		for _, task := range workflow.Spec.Tasks {
			if task.Name == "__stigmer_init_context" {
				t.Error("Found __stigmer_init_context task - compile-time resolution should eliminate this task!")
			}
		}

		// Find fetchData task
		var fetchDataTask *workflowv1.WorkflowTask
		for _, task := range workflow.Spec.Tasks {
			if task.Name == "fetchData" {
				fetchDataTask = task
				break
			}
		}

		if fetchDataTask == nil {
			t.Fatal("fetchData task not found")
		}

		// SMART RESOLUTION TEST:
		// The example uses apiBase.Concat("/posts/1") where both parts are known at synthesis time.
		// The SDK should resolve this IMMEDIATELY to the final URL, not create a runtime expression.
		//
		// Expected: "https://jsonplaceholder.typicode.com/posts/1" (fully resolved)
		// NOT: "${ $context.apiBase + "/posts/1" }" (runtime expression)
		//
		// This is the core of compile-time resolution - resolve everything we can at synthesis time.
		if fetchDataTask.TaskConfig == nil {
			t.Fatal("fetchData task config is nil")
		}

		endpointField, ok := fetchDataTask.TaskConfig.Fields["endpoint"]
		if !ok {
			t.Fatal("fetchData should have 'endpoint' field")
		}

		endpointStruct := endpointField.GetStructValue()
		if endpointStruct == nil {
			t.Fatal("Endpoint should be a struct")
		}

		uri, ok := endpointStruct.Fields["uri"]
		if !ok {
			t.Fatal("Endpoint should have 'uri' field")
		}

		uriValue := uri.GetStringValue()
		t.Logf("URI value: %s", uriValue)

		// The URI should be FULLY RESOLVED at synthesis time
		// because .Concat() was called on known values
		expectedURI := "https://jsonplaceholder.typicode.com/posts/1"
		if uriValue != expectedURI {
			t.Errorf("URI = %v, want %v (compile-time resolution should resolve .Concat() on known values)", uriValue, expectedURI)
		}

		// AUTO-EXPORT FUNCTIONALITY TEST:
		// fetchData task should have auto-export set because processResponse uses fetchTask.Field()
		if fetchDataTask.Export == nil {
			t.Error("fetchData task should have auto-export (set when .Field() is called)")
		} else if fetchDataTask.Export.As != "${.}" {
			t.Errorf("fetchData export.as = %v, want ${.}", fetchDataTask.Export.As)
		}

		// Find processResponse task
		var processTask *workflowv1.WorkflowTask
		for _, task := range workflow.Spec.Tasks {
			if task.Name == "processResponse" {
				processTask = task
				break
			}
		}

		if processTask == nil {
			t.Fatal("processResponse task not found")
		}

		// Verify processResponse has variables that reference fetchData fields
		// This demonstrates the auto-export feature working
		if processTask.TaskConfig == nil {
			t.Fatal("processResponse task config is nil")
		}

		varsField, ok := processTask.TaskConfig.Fields["variables"]
		if !ok {
			t.Fatal("processResponse should have 'variables' field")
		}

		varsStruct := varsField.GetStructValue()
		if varsStruct == nil {
			t.Fatal("Variables should be a struct")
		}

		// Verify field references point to fetchData task
		postTitle, ok := varsStruct.Fields["postTitle"]
		if !ok {
			t.Fatal("Expected variable 'postTitle' not found")
		}

		// The reference should be to fetchData task output
		postTitleRef := postTitle.GetStringValue()
		if postTitleRef != "${ $context.fetchData.title }" {
			t.Errorf("postTitle reference = %v, want ${ $context.fetchData.title }", postTitleRef)
		}

		t.Log("✅ Compile-time variable resolution verified:")
		t.Log("   - NO __stigmer_init_context task generated")
		t.Log("   - .Concat() on known values resolved immediately")
		t.Log("   - URL fully resolved: https://jsonplaceholder.typicode.com/posts/1")
		t.Log("✅ Auto-export functionality verified: fetchData exports when .Field() is used")
	})
}

// TestExample08_WorkflowWithConditionals tests the workflow with conditionals example
// TODO: This test is currently skipped because Switch/Case workflow features are not yet implemented.
// Required implementations:
//   - wf.Switch() method
//   - workflow.SwitchOn() option
//   - workflow.Case() option
//   - workflow.Equals() condition builder
//   - workflow.DefaultCase() option
//   - task.DependsOn() method
func TestExample08_WorkflowWithConditionals(t *testing.T) {
	t.Skip("TODO: Switch/Case workflow features not yet implemented (post-MVP)")

	runExampleTest(t, "08_workflow_with_conditionals.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest workflowv1.WorkflowManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Workflows) != 1 {
			t.Fatalf("Expected 1 workflow, got %d", len(manifest.Workflows))
		}

		workflow := manifest.Workflows[0]
		if workflow.Spec == nil || workflow.Spec.Document == nil {
			t.Fatal("Workflow spec or document is nil")
		}

		if workflow.Spec.Document.Name != "conditional-deployment" {
			t.Errorf("Workflow name = %v, want conditional-deployment", workflow.Spec.Document.Name)
		}

		// Verify workflow has tasks (conditionals use SWITCH task kind)
		if len(workflow.Spec.Tasks) == 0 {
			t.Error("Workflow should have tasks")
		}
	})
}

// TestExample09_WorkflowWithLoops tests the workflow with loops example
// TODO: This test is currently skipped because ForEach/Loop workflow features are not yet implemented.
// Required implementations:
//   - wf.ForEach() method
//   - workflow.IterateOver() option
//   - workflow.WithLoopBody() option
//   - workflow.LoopVar type
//   - workflow.Body() helper (alias for WithBody)
func TestExample09_WorkflowWithLoops(t *testing.T) {
	t.Skip("TODO: ForEach/Loop workflow features not yet implemented (post-MVP)")

	runExampleTest(t, "09_workflow_with_loops.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest workflowv1.WorkflowManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Workflows) != 1 {
			t.Fatalf("Expected 1 workflow, got %d", len(manifest.Workflows))
		}

		workflow := manifest.Workflows[0]
		if workflow.Spec == nil || workflow.Spec.Document == nil {
			t.Fatal("Workflow spec or document is nil")
		}

		if workflow.Spec.Document.Name != "batch-processor" {
			t.Errorf("Workflow name = %v, want batch-processor", workflow.Spec.Document.Name)
		}

		// Verify workflow has tasks (loops use FOR task kind)
		if len(workflow.Spec.Tasks) == 0 {
			t.Error("Workflow should have tasks")
		}
	})
}

// TestExample10_WorkflowWithErrorHandling tests the workflow with error handling example
// TODO: This test is currently skipped because Try/Catch/Finally workflow features are not yet implemented.
// Required implementations:
//   - wf.Try() method
//   - workflow.TryBlock() option
//   - workflow.CatchBlock() option
//   - workflow.FinallyBlock() option
//   - workflow.ErrorRef type
func TestExample10_WorkflowWithErrorHandling(t *testing.T) {
	t.Skip("TODO: Try/Catch/Finally workflow features not yet implemented (post-MVP)")

	runExampleTest(t, "10_workflow_with_error_handling.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest workflowv1.WorkflowManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Workflows) != 1 {
			t.Fatalf("Expected 1 workflow, got %d", len(manifest.Workflows))
		}

		workflow := manifest.Workflows[0]
		if workflow.Spec == nil || workflow.Spec.Document == nil {
			t.Fatal("Workflow spec or document is nil")
		}

		if workflow.Spec.Document.Name != "resilient-api-call" {
			t.Errorf("Workflow name = %v, want resilient-api-call", workflow.Spec.Document.Name)
		}

		// Verify workflow has tasks (error handling uses TRY task kind)
		if len(workflow.Spec.Tasks) == 0 {
			t.Error("Workflow should have tasks")
		}
	})
}

// TestExample11_WorkflowWithParallelExecution tests the workflow with parallel execution example
// TODO: This test is currently skipped because Fork/Join parallel execution features are not yet implemented.
// Required implementations:
//   - wf.Fork() method
//   - workflow.ParallelBranches() option
//   - workflow.Branch() builder
//   - workflow.WaitForAll() option
//   - task.Branch() method to access branch results
func TestExample11_WorkflowWithParallelExecution(t *testing.T) {
	t.Skip("TODO: Fork/Join parallel execution features not yet implemented (post-MVP)")

	runExampleTest(t, "11_workflow_with_parallel_execution.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest workflowv1.WorkflowManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Workflows) != 1 {
			t.Fatalf("Expected 1 workflow, got %d", len(manifest.Workflows))
		}

		workflow := manifest.Workflows[0]
		if workflow.Spec == nil || workflow.Spec.Document == nil {
			t.Fatal("Workflow spec or document is nil")
		}

		if workflow.Spec.Document.Name != "parallel-data-fetch" {
			t.Errorf("Workflow name = %v, want parallel-data-fetch", workflow.Spec.Document.Name)
		}

		// Verify workflow has parallel execution constructs
		if len(workflow.Spec.Tasks) == 0 {
			t.Error("Workflow should have tasks")
		}
	})
}

// TestExample12_AgentWithTypedContext tests the agent with typed context example
func TestExample12_AgentWithTypedContext(t *testing.T) {
	runExampleTest(t, "12_agent_with_typed_context.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "agent-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest agentv1.AgentManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Agents) < 1 {
			t.Fatal("Expected at least 1 agent")
		}

		agent := manifest.Agents[0]
		if agent.Name != "code-reviewer" {
			t.Errorf("Agent name = %v, want code-reviewer", agent.Name)
		}

		// Verify agent has description (uses typed context variable)
		if agent.Description == "" {
			t.Error("Agent should have description from typed context")
		}

		// This example demonstrates typed context with agent
		// The key point is that typed context variables can be used for configuration
		if agent.Instructions == "" {
			t.Error("Agent should have instructions")
		}

		// Verify agent has skills (demonstrates complete agent configuration)
		if len(agent.Skills) == 0 {
			t.Error("Agent should have skills")
		}

		// Verify agent has MCP servers
		if len(agent.McpServers) == 0 {
			t.Error("Agent should have MCP servers")
		}
	})
}

// TestExample13_WorkflowAndAgentSharedContext tests the workflow and agent with shared context example
func TestExample13_WorkflowAndAgentSharedContext(t *testing.T) {
	runExampleTest(t, "13_workflow_and_agent_shared_context.go", func(t *testing.T, outputDir string) {
		// This example creates BOTH workflow and agent manifests
		workflowManifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
		agentManifestPath := filepath.Join(outputDir, "agent-manifest.pb")

		// Verify both manifests were created
		assertFileExists(t, workflowManifestPath)
		assertFileExists(t, agentManifestPath)

		// Validate workflow manifest
		var workflowManifest workflowv1.WorkflowManifest
		readProtoManifest(t, workflowManifestPath, &workflowManifest)

		if len(workflowManifest.Workflows) != 1 {
			t.Fatalf("Expected 1 workflow, got %d", len(workflowManifest.Workflows))
		}

		workflow := workflowManifest.Workflows[0]
		if workflow.Spec == nil || workflow.Spec.Document == nil {
			t.Fatal("Workflow spec or document is nil")
		}

		if workflow.Spec.Document.Name != "fetch-and-analyze" {
			t.Errorf("Workflow name = %v, want fetch-and-analyze", workflow.Spec.Document.Name)
		}

		// Validate agent manifest
		var agentManifest agentv1.AgentManifest
		readProtoManifest(t, agentManifestPath, &agentManifest)

		if len(agentManifest.Agents) != 1 {
			t.Fatalf("Expected 1 agent, got %d", len(agentManifest.Agents))
		}

		agent := agentManifest.Agents[0]
		if agent.Name != "data-analyzer" {
			t.Errorf("Agent name = %v, want data-analyzer", agent.Name)
		}

		// Verify workflow has tasks (demonstrating shared context usage)
		if len(workflow.Spec.Tasks) == 0 {
			t.Error("Workflow should have tasks using shared context variables")
		}

		// Verify agent has instructions (demonstrating shared context usage)
		if agent.Instructions == "" {
			t.Error("Agent should have instructions")
		}

		// Both workflow and agent should be configured with shared context
		// This example demonstrates that both can use the same context for configuration
		// The key point is that both manifests are created successfully from the same context
	})
}

// TestCompileTimeVariableResolution tests that context variables are resolved at compile-time
// This integration test verifies:
// 1. NO __stigmer_init_context SET task is generated
// 2. .Concat() on known values resolves immediately (no runtime JQ expressions)
// 3. Multiple concatenations work correctly
func TestCompileTimeVariableResolution(t *testing.T) {
	// Create temporary output directory
	outputDir := t.TempDir()

	// Set output directory environment variable
	originalEnv := os.Getenv("STIGMER_OUT_DIR")
	os.Setenv("STIGMER_OUT_DIR", outputDir)
	defer os.Setenv("STIGMER_OUT_DIR", originalEnv)

	// Create a workflow with context variables
	err := stigmer.Run(func(ctx *stigmer.Context) error {
		// Define context variables that will be used with .Concat()
		baseURL := ctx.SetString("baseURL", "https://api.example.com")
		apiVersion := ctx.SetString("version", "v1")
		timeout := ctx.SetInt("timeout", 30)

		// Create workflow
		wf, err := workflow.New(ctx,
			workflow.WithNamespace("test"),
			workflow.WithName("compile-time-test"),
			workflow.WithVersion("1.0.0"),
		)
		if err != nil {
			return err
		}

		// Create an HTTP task using .Concat() - should resolve immediately
		// because all parts are known at synthesis time
		endpoint := baseURL.Concat("/v1/users")

		wf.HttpGet("fetchAPI",
			endpoint,                  // Should be resolved to "https://api.example.com/v1/users"
			workflow.Timeout(timeout), // Pass IntRef directly (proper SDK pattern)
		)

		// Use apiVersion to avoid "declared but not used" error
		_ = apiVersion

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	// Verify manifest was created
	manifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
	assertFileExists(t, manifestPath)

	// Read and parse manifest
	var manifest workflowv1.WorkflowManifest
	readProtoManifest(t, manifestPath, &manifest)

	if len(manifest.Workflows) != 1 {
		t.Fatalf("Expected 1 workflow, got %d", len(manifest.Workflows))
	}

	workflow := manifest.Workflows[0]
	if workflow.Spec == nil {
		t.Fatal("Workflow spec is nil")
	}

	// CRITICAL TEST: Verify NO __stigmer_init_context task exists
	// With compile-time resolution, this task should NOT be generated
	for _, task := range workflow.Spec.Tasks {
		if task.Name == "__stigmer_init_context" {
			t.Fatal("Found __stigmer_init_context task - compile-time resolution should eliminate this!")
		}
	}

	// Should have only 1 user task (fetchAPI)
	if len(workflow.Spec.Tasks) != 1 {
		t.Fatalf("Expected 1 task (NO context init), got %d", len(workflow.Spec.Tasks))
	}

	fetchTask := workflow.Spec.Tasks[0]
	if fetchTask.Name != "fetchAPI" {
		t.Errorf("Task name = %v, want fetchAPI", fetchTask.Name)
	}

	// Verify task config has interpolated values
	if fetchTask.TaskConfig == nil {
		t.Fatal("Task config is nil")
	}

	// Check endpoint field
	endpointField, ok := fetchTask.TaskConfig.Fields["endpoint"]
	if !ok {
		t.Fatal("Task should have 'endpoint' field")
	}

	endpointStruct := endpointField.GetStructValue()
	if endpointStruct == nil {
		t.Fatal("Endpoint should be a struct")
	}

	// Verify URI was interpolated at compile-time
	uriField, ok := endpointStruct.Fields["uri"]
	if !ok {
		t.Fatal("Endpoint should have 'uri' field")
	}

	uriValue := uriField.GetStringValue()
	t.Logf("Interpolated URI: %s", uriValue)

	// The URI should be fully resolved: "https://api.example.com/v1/users"
	// NOT "${baseURL}/${version}/users"
	expectedURI := "https://api.example.com/v1/users"
	if uriValue != expectedURI {
		t.Errorf("URI = %v, want %v (compile-time interpolation failed)", uriValue, expectedURI)
	}

	// Verify timeout was passed correctly (IntRef → number)
	timeoutField, ok := fetchTask.TaskConfig.Fields["timeout_seconds"]
	if ok {
		timeoutValue := timeoutField.GetNumberValue()
		if timeoutValue != 30 {
			t.Errorf("Timeout = %v, want 30", timeoutValue)
		} else {
			t.Logf("✅ Timeout interpolated as number: %v", timeoutValue)
		}
	} else {
		t.Error("timeout_seconds field not found in config")
	}

	t.Log("✅ Compile-time variable resolution VERIFIED:")
	t.Log("   - NO __stigmer_init_context task generated")
	t.Log("   - Variables interpolated into task configs")
	t.Log("   - URI fully resolved: https://api.example.com/v1/users")
	t.Log("   - Type preservation working (numbers stay numbers)")
}

// Helper function to run an example and verify output
func runExampleTest(t *testing.T, exampleFile string, verify func(*testing.T, string)) {
	t.Helper()

	// Create temporary output directory
	outputDir := t.TempDir()

	// Get the path to the example file
	examplePath := filepath.Join(".", exampleFile)

	// Check if example file exists
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Fatalf("Example file not found: %s", examplePath)
	}

	// Run the example with STIGMER_OUT_DIR set
	cmd := exec.Command("go", "run", examplePath)
	cmd.Env = append(os.Environ(), "STIGMER_OUT_DIR="+outputDir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run example %s: %v\nOutput: %s", exampleFile, err, string(output))
	}

	t.Logf("Example %s output:\n%s", exampleFile, string(output))

	// Run verification function
	verify(t, outputDir)
}

// Helper function to assert a file exists
func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Expected file does not exist: %s", path)
	}
}

// TestExample14_WorkflowWithRuntimeSecrets tests the workflow with runtime secrets example
// This test validates the CRITICAL SECURITY PATTERN:
// 1. Runtime secrets appear as placeholders in manifest: "${.secrets.KEY}"
// 2. Runtime env vars appear as placeholders: "${.env_vars.VAR}"
// 3. NO actual secret values appear anywhere in the manifest
// 4. Placeholders are correctly formatted and validated
func TestExample14_WorkflowWithRuntimeSecrets(t *testing.T) {
	runExampleTest(t, "14_workflow_with_runtime_secrets.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest workflowv1.WorkflowManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Workflows) != 1 {
			t.Fatalf("Expected 1 workflow, got %d", len(manifest.Workflows))
		}

		wf := manifest.Workflows[0]
		if wf.Spec == nil || wf.Spec.Document == nil {
			t.Fatal("Workflow spec or document is nil")
		}

		if wf.Spec.Document.Name != "secure-api-workflow" {
			t.Errorf("Workflow name = %v, want secure-api-workflow", wf.Spec.Document.Name)
		}

		// Verify workflow has tasks
		if len(wf.Spec.Tasks) == 0 {
			t.Fatal("Workflow should have tasks")
		}

		t.Logf("Found %d tasks in workflow", len(wf.Spec.Tasks))

		// ============================================================================
		// SECURITY TEST 1: Verify Runtime Secret Placeholders
		// ============================================================================

		// Find callOpenAI task - should use RuntimeSecret for API key
		var openaiTask *workflowv1.WorkflowTask
		for _, task := range wf.Spec.Tasks {
			if task.Name == "callOpenAI" {
				openaiTask = task
				break
			}
		}

		if openaiTask == nil {
			t.Fatal("callOpenAI task not found")
		}

		// Verify Authorization header contains runtime secret placeholder
		if openaiTask.TaskConfig == nil {
			t.Fatal("callOpenAI task config is nil")
		}

		// Check headers field
		headersField, ok := openaiTask.TaskConfig.Fields["headers"]
		if !ok {
			t.Fatal("callOpenAI should have 'headers' field")
		}

		headersStruct := headersField.GetStructValue()
		if headersStruct == nil {
			t.Fatal("Headers should be a struct")
		}

		authHeader, ok := headersStruct.Fields["Authorization"]
		if !ok {
			t.Fatal("Should have Authorization header")
		}

		authValue := authHeader.GetStringValue()
		t.Logf("Authorization header: %s", authValue)

		// CRITICAL: Verify it's a PLACEHOLDER, not the actual secret
		// Expected: "Bearer ${.secrets.OPENAI_API_KEY}"
		// NOT: "Bearer sk-proj-abc123"
		if !containsRuntimeRef(authValue, "secrets", "OPENAI_API_KEY") {
			t.Errorf("Authorization header should contain runtime secret placeholder ${.secrets.OPENAI_API_KEY}, got: %s", authValue)
		}

		// Verify NO actual secret values anywhere
		if containsSecretValue(authValue) {
			t.Errorf("❌ SECURITY VIOLATION: Authorization header contains what looks like an actual secret: %s", authValue)
		}

		// ============================================================================
		// SECURITY TEST 2: Verify Runtime Environment Variable Placeholders
		// ============================================================================

		// Find processData task - should use RuntimeEnv for environment config
		var processTask *workflowv1.WorkflowTask
		for _, task := range wf.Spec.Tasks {
			if task.Name == "processData" {
				processTask = task
				break
			}
		}

		if processTask == nil {
			t.Fatal("processData task not found")
		}

		// Check endpoint field for environment-specific URL
		endpointField, ok := processTask.TaskConfig.Fields["endpoint"]
		if !ok {
			t.Fatal("processData should have 'endpoint' field")
		}

		endpointStruct := endpointField.GetStructValue()
		if endpointStruct == nil {
			t.Fatal("Endpoint should be a struct")
		}

		uriField, ok := endpointStruct.Fields["uri"]
		if !ok {
			t.Fatal("Endpoint should have 'uri' field")
		}

		uriValue := uriField.GetStringValue()
		t.Logf("Endpoint URI: %s", uriValue)

		// Verify environment variable placeholder
		// Expected: "https://api-${.env_vars.ENVIRONMENT}.example.com/process"
		// NOT: "https://api-production.example.com/process"
		if !containsRuntimeRef(uriValue, "env_vars", "ENVIRONMENT") {
			t.Errorf("URI should contain runtime env var placeholder ${.env_vars.ENVIRONMENT}, got: %s", uriValue)
		}

		// Check X-Region header should have runtime env var
		processHeaders, ok := processTask.TaskConfig.Fields["headers"]
		if ok {
			processHeadersStruct := processHeaders.GetStructValue()
			if processHeadersStruct != nil {
				regionHeader, hasRegion := processHeadersStruct.Fields["X-Region"]
				if hasRegion {
					regionValue := regionHeader.GetStringValue()
					t.Logf("X-Region header: %s", regionValue)

					if !containsRuntimeRef(regionValue, "env_vars", "AWS_REGION") {
						t.Errorf("X-Region header should contain ${.env_vars.AWS_REGION}, got: %s", regionValue)
					}
				}
			}
		}

		// ============================================================================
		// SECURITY TEST 3: Verify Multiple Secrets Pattern
		// ============================================================================

		// Find chargePayment task - should have multiple runtime secrets
		var stripeTask *workflowv1.WorkflowTask
		for _, task := range wf.Spec.Tasks {
			if task.Name == "chargePayment" {
				stripeTask = task
				break
			}
		}

		if stripeTask == nil {
			t.Fatal("chargePayment task not found")
		}

		// Verify Stripe API key is a runtime secret
		stripeHeaders, ok := stripeTask.TaskConfig.Fields["headers"]
		if ok {
			stripeHeadersStruct := stripeHeaders.GetStructValue()
			if stripeHeadersStruct != nil {
				stripeAuth, hasAuth := stripeHeadersStruct.Fields["Authorization"]
				if hasAuth {
					stripeAuthValue := stripeAuth.GetStringValue()
					t.Logf("Stripe Authorization: %s", stripeAuthValue)

					if !containsRuntimeRef(stripeAuthValue, "secrets", "STRIPE_API_KEY") {
						t.Errorf("Stripe Authorization should contain ${.secrets.STRIPE_API_KEY}, got: %s", stripeAuthValue)
					}

					if containsSecretValue(stripeAuthValue) {
						t.Errorf("❌ SECURITY VIOLATION: Stripe Authorization contains actual secret: %s", stripeAuthValue)
					}
				}

				// Check idempotency key is also a runtime secret
				idempotencyKey, hasKey := stripeHeadersStruct.Fields["Idempotency-Key"]
				if hasKey {
					keyValue := idempotencyKey.GetStringValue()
					t.Logf("Idempotency-Key: %s", keyValue)

					if !containsRuntimeRef(keyValue, "secrets", "STRIPE_IDEMPOTENCY_KEY") {
						t.Errorf("Idempotency-Key should contain ${.secrets.STRIPE_IDEMPOTENCY_KEY}, got: %s", keyValue)
					}
				}
			}
		}

		// ============================================================================
		// SECURITY TEST 4: Database Credentials
		// ============================================================================

		// Find storeResults task - should have database password as runtime secret
		var dbTask *workflowv1.WorkflowTask
		for _, task := range wf.Spec.Tasks {
			if task.Name == "storeResults" {
				dbTask = task
				break
			}
		}

		if dbTask == nil {
			t.Fatal("storeResults task not found")
		}

		// Check gRPC metadata for db-password
		metadata, ok := dbTask.TaskConfig.Fields["metadata"]
		if ok {
			metadataStruct := metadata.GetStructValue()
			if metadataStruct != nil {
				dbPassword, hasPassword := metadataStruct.Fields["db-password"]
				if hasPassword {
					passwordValue := dbPassword.GetStringValue()
					t.Logf("Database password metadata: %s", passwordValue)

					if !containsRuntimeRef(passwordValue, "secrets", "DATABASE_PASSWORD") {
						t.Errorf("db-password should contain ${.secrets.DATABASE_PASSWORD}, got: %s", passwordValue)
					}

					if containsSecretValue(passwordValue) {
						t.Errorf("❌ SECURITY VIOLATION: Database password contains actual secret: %s", passwordValue)
					}
				}
			}
		}

		// ============================================================================
		// SECURITY TEST 5: Webhook Registration with Runtime Secrets
		// ============================================================================

		// Find registerWebhook task - should have webhook signing secret
		var webhookTask *workflowv1.WorkflowTask
		for _, task := range wf.Spec.Tasks {
			if task.Name == "registerWebhook" {
				webhookTask = task
				break
			}
		}

		if webhookTask == nil {
			t.Fatal("registerWebhook task not found")
		}

		// Check Authorization header for external API key
		webhookHeaders, ok := webhookTask.TaskConfig.Fields["headers"]
		if ok {
			webhookHeadersStruct := webhookHeaders.GetStructValue()
			if webhookHeadersStruct != nil {
				webhookAuth, hasAuth := webhookHeadersStruct.Fields["Authorization"]
				if hasAuth {
					authValue := webhookAuth.GetStringValue()
					t.Logf("Webhook Authorization: %s", authValue)

					if !containsRuntimeRef(authValue, "secrets", "EXTERNAL_API_KEY") {
						t.Errorf("Webhook Authorization should contain .secrets.EXTERNAL_API_KEY, got: %s", authValue)
					}
				}
			}
		}

		// Check body for webhook signing secret
		webhookBody, ok := webhookTask.TaskConfig.Fields["body"]
		if ok {
			bodyStruct := webhookBody.GetStructValue()
			if bodyStruct != nil {
				secret, hasSecret := bodyStruct.Fields["secret"]
				if hasSecret {
					secretValue := secret.GetStringValue()
					t.Logf("Webhook secret: %s", secretValue)

					if !containsRuntimeRef(secretValue, "secrets", "WEBHOOK_SIGNING_SECRET") {
						t.Errorf("Webhook secret should contain .secrets.WEBHOOK_SIGNING_SECRET, got: %s", secretValue)
					}

					if containsSecretValue(secretValue) {
						t.Errorf("❌ SECURITY VIOLATION: Webhook secret contains actual secret: %s", secretValue)
					}
				}
			}
		}

		// ============================================================================
		// FINAL VERIFICATION
		// ============================================================================

		t.Log("✅ Runtime Secret Security Verified:")
		t.Log("   - All API keys use RuntimeSecret() placeholders")
		t.Log("   - Environment config uses RuntimeEnv() placeholders")
		t.Log("   - NO actual secret values found in manifest")
		t.Log("   - Placeholders correctly embedded: .secrets.KEY and .env_vars.VAR")
		t.Log("   - Multiple secrets in single task work correctly")
		t.Log("   - Database credentials properly secured")
		t.Log("   - Webhook signing secrets properly secured")
		t.Log("   - Environment-specific URLs work with runtime env vars")
	})
}

// Helper function to check if a string contains a runtime reference
func containsRuntimeRef(value string, refType string, keyName string) bool {
	// Check for both exact match and interpolation patterns
	placeholder := fmt.Sprintf(".%s.%s", refType, keyName)

	// The placeholder might appear in different formats:
	// - ${.secrets.KEY} (direct)
	// - ${ "Bearer " + .secrets.KEY } (Interpolate() format)
	// - .secrets.KEY (without ${ })
	
	return containsSubstring(value, placeholder)
}

// Helper function to check if string contains another string
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || recursiveContains(s, substr))
}

// Recursive helper for substring checking
func recursiveContains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	if s[:len(substr)] == substr {
		return true
	}
	return recursiveContains(s[1:], substr)
}

// Helper function to detect if a value looks like an actual secret
// (starts with common secret prefixes)
func containsSecretValue(value string) bool {
	secretPrefixes := []string{
		"sk-proj-",      // OpenAI project keys
		"sk-",           // OpenAI/Stripe keys
		"sk_live_",      // Stripe live keys
		"sk_test_",      // Stripe test keys
		"AKIA",          // AWS access keys
		"glpat-",        // GitLab personal access tokens
		"ghp_",          // GitHub personal access tokens
		"xoxb-",         // Slack bot tokens
		"rk_live_",      // Stripe restricted keys
	}

	for _, prefix := range secretPrefixes {
		if len(value) >= len(prefix) && value[:len(prefix)] == prefix {
			return true
		}
		// Check anywhere in the string
		if containsSubstring(value, prefix) {
			return true
		}
	}

	return false
}

// Helper function to read and unmarshal a protobuf manifest
func readProtoManifest(t *testing.T, path string, message proto.Message) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read manifest file %s: %v", path, err)
	}

	if err := proto.Unmarshal(data, message); err != nil {
		t.Fatalf("Failed to unmarshal manifest %s: %v", path, err)
	}
}
