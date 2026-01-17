package examples_test

import (
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
