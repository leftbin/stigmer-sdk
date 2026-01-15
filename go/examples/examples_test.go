package examples_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	agentv1 "buf.build/gen/go/leftbin/stigmer/protocolbuffers/go/ai/stigmer/agentic/agent/v1"
	workflowv1 "buf.build/gen/go/leftbin/stigmer/protocolbuffers/go/ai/stigmer/agentic/workflow/v1"
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

		// Should have 3 tasks: initialize, fetchData, processResponse
		if len(workflow.Spec.Tasks) != 3 {
			t.Errorf("Expected 3 tasks, got %d", len(workflow.Spec.Tasks))
		}
	})
}

// TestExample08_WorkflowWithConditionals tests the workflow with conditionals example
func TestExample08_WorkflowWithConditionals(t *testing.T) {
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
func TestExample09_WorkflowWithLoops(t *testing.T) {
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
func TestExample10_WorkflowWithErrorHandling(t *testing.T) {
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
func TestExample11_WorkflowWithParallelExecution(t *testing.T) {
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
