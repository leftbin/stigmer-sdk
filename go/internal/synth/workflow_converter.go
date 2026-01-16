package synth

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	// Import Buf-generated proto packages
	apiresource "buf.build/gen/go/leftbin/stigmer/protocolbuffers/go/ai/stigmer/commons/apiresource"
	workflowv1 "buf.build/gen/go/leftbin/stigmer/protocolbuffers/go/ai/stigmer/agentic/workflow/v1"
	sdk "buf.build/gen/go/leftbin/stigmer/protocolbuffers/go/ai/stigmer/commons/sdk"

	// Import SDK types
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

// ToWorkflowManifest converts one or more SDK Workflows to a WorkflowManifest proto message.
//
// This is the core converter that transforms the Go SDK's type-safe workflow
// configurations into the protocol buffer format that the CLI expects.
//
// Conversion mapping:
//   - workflow.Workflow → workflowv1.Workflow
//   - workflow.Task → workflowv1.WorkflowTask
//
// Returns an error if any nested conversion fails.
//
// Note: This version does not inject context variables. Use ToWorkflowManifestWithContext
// if you need automatic context variable injection.
func ToWorkflowManifest(workflowInterfaces ...interface{}) (*workflowv1.WorkflowManifest, error) {
	return ToWorkflowManifestWithContext(nil, workflowInterfaces...)
}

// ToWorkflowManifestWithContext converts SDK Workflows to a WorkflowManifest proto message
// with automatic context variable injection.
//
// When contextVars is provided, a SET task is automatically injected as the first task
// in each workflow to initialize the workflow context with the provided variables.
//
// This implements the Pulumi-style pattern where context variables defined via
// ctx.SetString(), ctx.SetInt(), etc. are automatically available in the workflow runtime.
//
// Example:
//   ctx.SetString("apiURL", "https://api.example.com")
//   ctx.SetInt("retries", 3)
//
// Will generate a SET task:
//   - name: __stigmer_init_context
//     kind: SET
//     task_config:
//       variables:
//         apiURL: "https://api.example.com"
//         retries: 3
func ToWorkflowManifestWithContext(contextVars map[string]interface{}, workflowInterfaces ...interface{}) (*workflowv1.WorkflowManifest, error) {
	if len(workflowInterfaces) == 0 {
		return nil, fmt.Errorf("at least one workflow is required")
	}

	// Create SDK metadata
	sdkMetadata := &sdk.SdkMetadata{
		Language:    "go",
		Version:     "0.1.0", // TODO: Get from build info
		GeneratedAt: time.Now().Unix(),
	}

	// Create manifest with empty workflows list
	manifest := &workflowv1.WorkflowManifest{
		SdkMetadata: sdkMetadata,
		Workflows:   []*workflowv1.Workflow{},
	}

	// Convert each workflow
	for wfIdx, workflowInterface := range workflowInterfaces {
		// Type assert to *workflow.Workflow
		wf, ok := workflowInterface.(*workflow.Workflow)
		if !ok {
			return nil, fmt.Errorf("workflow[%d]: invalid type %T, expected *workflow.Workflow", wfIdx, workflowInterface)
		}

		// Convert to proto with context variable injection
		protoWorkflow, err := workflowToProtoWithContext(wf, contextVars)
		if err != nil {
			return nil, fmt.Errorf("workflow[%d] %s: %w", wfIdx, wf.Document.Name, err)
		}

		// Add to manifest
		manifest.Workflows = append(manifest.Workflows, protoWorkflow)
	}

	return manifest, nil
}

// workflowToProto converts a workflow.Workflow to a workflowv1.Workflow proto.
// This version does not inject context variables.
func workflowToProto(wf *workflow.Workflow) (*workflowv1.Workflow, error) {
	return workflowToProtoWithContext(wf, nil)
}

// workflowToProtoWithContext converts a workflow.Workflow to a workflowv1.Workflow proto
// with automatic context variable injection.
func workflowToProtoWithContext(wf *workflow.Workflow, contextVars map[string]interface{}) (*workflowv1.Workflow, error) {
	// Create workflow proto
	protoWorkflow := &workflowv1.Workflow{
		ApiVersion: "agentic.stigmer.ai/v1",
		Kind:       "Workflow",
	}

	// Convert metadata (placeholder - would need actual metadata proto structure)
	// For now, we'll focus on the spec

	// Convert spec with context variable injection
	spec, err := workflowSpecToProtoWithContext(wf, contextVars)
	if err != nil {
		return nil, fmt.Errorf("converting spec: %w", err)
	}
	protoWorkflow.Spec = spec

	return protoWorkflow, nil
}

// workflowSpecToProto converts workflow spec to proto.
// This version does not inject context variables.
func workflowSpecToProto(wf *workflow.Workflow) (*workflowv1.WorkflowSpec, error) {
	return workflowSpecToProtoWithContext(wf, nil)
}

// workflowSpecToProtoWithContext converts workflow spec to proto with context variable injection.
//
// If contextVars is provided and non-empty, a SET task named "__stigmer_init_context" is
// automatically injected as the first task to initialize the workflow context.
//
// This follows the Pulumi pattern where SDK context variables are automatically available
// at runtime without manual wiring.
func workflowSpecToProtoWithContext(wf *workflow.Workflow, contextVars map[string]interface{}) (*workflowv1.WorkflowSpec, error) {
	spec := &workflowv1.WorkflowSpec{
		Description: wf.Description,
		Document: &workflowv1.WorkflowDocument{
			Dsl:         wf.Document.DSL,
			Namespace:   wf.Document.Namespace,
			Name:        wf.Document.Name,
			Version:     wf.Document.Version,
			Description: wf.Document.Description,
		},
		Tasks: []*workflowv1.WorkflowTask{},
	}

	// Inject context initialization task if context variables exist
	if len(contextVars) > 0 {
		contextInitTask, err := createContextInitTask(contextVars)
		if err != nil {
			return nil, fmt.Errorf("creating context init task: %w", err)
		}
		spec.Tasks = append(spec.Tasks, contextInitTask)
	}

	// Convert user-defined tasks
	for i, task := range wf.Tasks {
		protoTask, err := taskToProto(task)
		if err != nil {
			return nil, fmt.Errorf("converting task[%d] %s: %w", i, task.Name, err)
		}
		spec.Tasks = append(spec.Tasks, protoTask)
	}

	// Convert environment variables (if any)
	// Note: Environment spec conversion is deferred as the proto structure may not be finalized
	// For now, we'll skip env spec conversion
	// TODO: Implement environmentVariablesToEnvSpec when proto structure is finalized

	return spec, nil
}

// createContextInitTask creates a SET task that initializes workflow context variables.
//
// This task is automatically injected as the first task in workflows when context variables
// are defined via ctx.SetString(), ctx.SetInt(), etc.
//
// The task sets all context variables to their initial values, making them available
// for use in subsequent tasks via JQ expressions like "${ $context.variableName }".
//
// Task structure:
//   - name: __stigmer_init_context
//   - kind: SET
//   - task_config:
//       variables:
//         variableName1: value1
//         variableName2: value2
//
// Note: The contextVars map values must implement the Ref interface with ToValue() method.
func createContextInitTask(contextVars map[string]interface{}) (*workflowv1.WorkflowTask, error) {
	// Import the Ref interface to access ToValue()
	// We need to import the parent package, but to avoid circular imports,
	// we'll use type assertion with interface{} and call ToValue via reflection
	// Actually, we can't import stigmer package here due to circular dependency
	// So we need to pass already-serialized values from the context

	// Build variables map for the SET task
	variables := make(map[string]interface{}, len(contextVars))
	for name, refInterface := range contextVars {
		// The contextVars map contains Ref interface values
		// We need to call ToValue() on each one
		// Use type assertion to access the ToValue() method
		type valueExtractor interface {
			ToValue() interface{}
		}
		
		if ref, ok := refInterface.(valueExtractor); ok {
			variables[name] = ref.ToValue()
		} else {
			// Fallback: use the value as-is (shouldn't happen if called correctly)
			variables[name] = refInterface
		}
	}

	// Create SET task config
	setConfig := map[string]interface{}{
		"variables": variables,
	}

	// Convert to protobuf Struct
	taskConfigStruct, err := structpb.NewStruct(setConfig)
	if err != nil {
		return nil, fmt.Errorf("creating task config struct: %w", err)
	}

	// Build the context init task
	task := &workflowv1.WorkflowTask{
		Name:       "__stigmer_init_context",
		Kind:       apiresource.WorkflowTaskKind_WORKFLOW_TASK_KIND_SET,
		TaskConfig: taskConfigStruct,
		// No export needed - this task just sets variables
		// No flow control - tasks will execute sequentially
	}

	return task, nil
}

// taskToProto converts a workflow.Task to a workflowv1.WorkflowTask proto.
func taskToProto(task *workflow.Task) (*workflowv1.WorkflowTask, error) {
	// Convert task config to google.protobuf.Struct
	taskConfig, err := taskConfigToStruct(task)
	if err != nil {
		return nil, fmt.Errorf("converting task config: %w", err)
	}

	protoTask := &workflowv1.WorkflowTask{
		Name:       task.Name,
		Kind:       taskKindToProtoKind(task.Kind),
		TaskConfig: taskConfig,
	}

	// Convert export if present
	if task.ExportAs != "" {
		protoTask.Export = &workflowv1.Export{
			As: task.ExportAs,
		}
	}

	// Convert flow control if present
	if task.ThenTask != "" {
		protoTask.Flow = &workflowv1.FlowControl{
			Then: task.ThenTask,
		}
	}

	return protoTask, nil
}

// taskKindToProtoKind converts SDK task kind to proto enum value.
func taskKindToProtoKind(kind workflow.TaskKind) apiresource.WorkflowTaskKind {
	// Map SDK task kind string to proto enum value
	// These values must match the WorkflowTaskKind enum in ai/stigmer/commons/apiresource/enum.proto
	kindMap := map[workflow.TaskKind]apiresource.WorkflowTaskKind{
		workflow.TaskKindSet:          apiresource.WorkflowTaskKind_WORKFLOW_TASK_KIND_SET,
		workflow.TaskKindHttpCall:     apiresource.WorkflowTaskKind_WORKFLOW_TASK_KIND_HTTP_CALL,
		workflow.TaskKindGrpcCall:     apiresource.WorkflowTaskKind_WORKFLOW_TASK_KIND_GRPC_CALL,
		workflow.TaskKindCallActivity: apiresource.WorkflowTaskKind_WORKFLOW_TASK_KIND_CALL_ACTIVITY,
		workflow.TaskKindSwitch:       apiresource.WorkflowTaskKind_WORKFLOW_TASK_KIND_SWITCH,
		workflow.TaskKindFor:          apiresource.WorkflowTaskKind_WORKFLOW_TASK_KIND_FOR,
		workflow.TaskKindFork:         apiresource.WorkflowTaskKind_WORKFLOW_TASK_KIND_FORK,
		workflow.TaskKindTry:          apiresource.WorkflowTaskKind_WORKFLOW_TASK_KIND_TRY,
		workflow.TaskKindListen:       apiresource.WorkflowTaskKind_WORKFLOW_TASK_KIND_LISTEN,
		workflow.TaskKindWait:         apiresource.WorkflowTaskKind_WORKFLOW_TASK_KIND_WAIT,
		workflow.TaskKindRaise:        apiresource.WorkflowTaskKind_WORKFLOW_TASK_KIND_RAISE,
		workflow.TaskKindRun:          apiresource.WorkflowTaskKind_WORKFLOW_TASK_KIND_RUN,
	}
	return kindMap[kind]
}

// stringMapToInterface converts map[string]string to map[string]interface{}.
// This is needed because structpb.NewStruct cannot handle map[string]string directly.
func stringMapToInterface(m map[string]string) map[string]interface{} {
	if m == nil {
		return nil
	}
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// mapSliceToInterfaceSlice converts []map[string]interface{} to []interface{}.
// This is needed because structpb.NewStruct cannot handle []map[string]interface{} directly.
func mapSliceToInterfaceSlice(slice []map[string]interface{}) []interface{} {
	if slice == nil {
		return nil
	}
	result := make([]interface{}, len(slice))
	for i, m := range slice {
		result[i] = m
	}
	return result
}

// stringSliceToInterfaceSlice converts []string to []interface{}.
// This is needed because structpb.NewStruct cannot handle []string directly in nested structures.
func stringSliceToInterfaceSlice(slice []string) []interface{} {
	if slice == nil {
		return nil
	}
	result := make([]interface{}, len(slice))
	for i, s := range slice {
		result[i] = s
	}
	return result
}

// convertNestedTasksToMaps recursively converts a slice of Tasks to proto-compatible maps.
// This ensures nested tasks (in FOR, FORK, TRY) have all required fields, not just name/kind.
//
// Each task map includes:
// - name: task identifier
// - kind: task type (converted to proto enum string)
// - task_config: task configuration as map (not Struct, to avoid nested Struct issues)
// - export: export configuration (if present)
// - flow: flow control (if present)
func convertNestedTasksToMaps(tasks []workflow.Task) ([]interface{}, error) {
	if tasks == nil {
		return nil, nil
	}
	
	result := make([]interface{}, len(tasks))
	for i, task := range tasks {
		// Convert task config to Struct
		taskConfig, err := taskConfigToStruct(&task)
		if err != nil {
			return nil, fmt.Errorf("converting nested task[%d] %s config: %w", i, task.Name, err)
		}
		
		// Convert the Struct back to a map to avoid nested Struct issues
		// structpb.NewStruct() cannot handle *structpb.Struct as a value
		taskConfigMap := taskConfig.AsMap()
		
		// Build task map with all required proto fields
		taskMap := map[string]interface{}{
			"name":        task.Name,
			"kind":        taskKindToProtoKind(task.Kind).String(),
			"task_config": taskConfigMap,
		}
		
		// Add export if present
		if task.ExportAs != "" {
			taskMap["export"] = map[string]interface{}{
				"as": task.ExportAs,
			}
		}
		
		// Add flow control if present
		if task.ThenTask != "" {
			taskMap["flow"] = map[string]interface{}{
				"then": task.ThenTask,
			}
		}
		
		result[i] = taskMap
	}
	
	return result, nil
}

// taskConfigToStruct converts task configuration to google.protobuf.Struct.
func taskConfigToStruct(task *workflow.Task) (*structpb.Struct, error) {
	var configMap map[string]interface{}

	switch task.Kind {
	case workflow.TaskKindSet:
		cfg := task.Config.(*workflow.SetTaskConfig)
		configMap = map[string]interface{}{
			"variables": stringMapToInterface(cfg.Variables),
		}

	case workflow.TaskKindHttpCall:
		cfg := task.Config.(*workflow.HttpCallTaskConfig)
		configMap = map[string]interface{}{
			"method": cfg.Method,
			"endpoint": map[string]interface{}{
				"uri": cfg.URI,
			},
			"headers":         stringMapToInterface(cfg.Headers),
			"body":            cfg.Body,
			"timeout_seconds": cfg.TimeoutSeconds,
		}

	case workflow.TaskKindGrpcCall:
		cfg := task.Config.(*workflow.GrpcCallTaskConfig)
		configMap = map[string]interface{}{
			"service": cfg.Service,
			"method":  cfg.Method,
			"body":    cfg.Body,
		}

	case workflow.TaskKindSwitch:
		cfg := task.Config.(*workflow.SwitchTaskConfig)
		cases := make([]map[string]interface{}, len(cfg.Cases))
		
		// Track if we have a default case (empty condition)
		hasExplicitDefault := false
		
		for i, c := range cfg.Cases {
			caseMap := map[string]interface{}{
				// Generate case name (proto requires it)
				"name": fmt.Sprintf("case%d", i+1),
				// Map Go "Condition" → Proto "when"
				"when": c.Condition,
				"then": c.Then,
			}
			
			// Check if this is a default case (empty condition)
			if c.Condition == "" {
				hasExplicitDefault = true
			}
			
			cases[i] = caseMap
		}
		
		// If DefaultTask is specified and we don't have an explicit default case,
		// add it as the last case with empty "when"
		if cfg.DefaultTask != "" && !hasExplicitDefault {
			defaultCase := map[string]interface{}{
				"name": "default",
				"when": "",  // Empty condition = default case
				"then": cfg.DefaultTask,
			}
			cases = append(cases, defaultCase)
		}
		
		configMap = map[string]interface{}{
			"cases": mapSliceToInterfaceSlice(cases),
		}

	case workflow.TaskKindFor:
		cfg := task.Config.(*workflow.ForTaskConfig)
		
		// Convert nested tasks fully (not just name/kind)
		doTasks, err := convertNestedTasksToMaps(cfg.Do)
		if err != nil {
			return nil, fmt.Errorf("converting FOR task nested tasks: %w", err)
		}
		
		configMap = map[string]interface{}{
			// Default "each" to "item" for now
			// TODO: Add "Each" field to ForTaskConfig Go struct for better UX
			"each": "item",
			"in":   cfg.In,
			"do":   doTasks,
		}

	case workflow.TaskKindFork:
		cfg := task.Config.(*workflow.ForkTaskConfig)
		branches := make([]map[string]interface{}, len(cfg.Branches))
		
		for i, b := range cfg.Branches {
			// Convert nested tasks in each branch
			doTasks, err := convertNestedTasksToMaps(b.Tasks)
			if err != nil {
				return nil, fmt.Errorf("converting FORK branch[%d] %s tasks: %w", i, b.Name, err)
			}
			
			branches[i] = map[string]interface{}{
				"name": b.Name,
				"do":   doTasks,
			}
		}
		
		configMap = map[string]interface{}{
			"branches": mapSliceToInterfaceSlice(branches),
			// Default "compete" to false (all branches must complete)
			// TODO: Add "Compete" field to ForkTaskConfig Go struct for race mode support
			"compete": false,
		}

	case workflow.TaskKindTry:
		cfg := task.Config.(*workflow.TryTaskConfig)
		
		// Convert "try" tasks (proto uses "try", not "tasks")
		tryTasks, err := convertNestedTasksToMaps(cfg.Tasks)
		if err != nil {
			return nil, fmt.Errorf("converting TRY task 'try' tasks: %w", err)
		}
		
		configMap = map[string]interface{}{
			"try": tryTasks,
		}
		
		// Handle catch blocks (proto expects singular "catch", not array)
		// If multiple catch blocks exist in Go, use the first one
		// TODO: Update TryTaskConfig Go struct to use singular Catch for proto alignment
		if len(cfg.Catch) > 0 {
			firstCatch := cfg.Catch[0]
			
			// Convert catch tasks
			catchTasks, err := convertNestedTasksToMaps(firstCatch.Tasks)
			if err != nil {
				return nil, fmt.Errorf("converting TRY task 'catch' tasks: %w", err)
			}
			
			catchBlock := map[string]interface{}{
				"as": firstCatch.As,
				"do": catchTasks,
				// Note: Proto doesn't have "errors" field for filtering by error type
				// The Go struct has it for UX, but we can't map it to proto
				// TODO: Discuss with team if proto should support error type filtering
			}
			
			configMap["catch"] = catchBlock
		}

	case workflow.TaskKindListen:
		cfg := task.Config.(*workflow.ListenTaskConfig)
		configMap = map[string]interface{}{
			"event": cfg.Event,
		}

	case workflow.TaskKindWait:
		cfg := task.Config.(*workflow.WaitTaskConfig)
		configMap = map[string]interface{}{
			"duration": cfg.Duration,
		}

	case workflow.TaskKindCallActivity:
		cfg := task.Config.(*workflow.CallActivityTaskConfig)
		configMap = map[string]interface{}{
			"activity": cfg.Activity,
			"input":    cfg.Input,
		}

	case workflow.TaskKindRaise:
		cfg := task.Config.(*workflow.RaiseTaskConfig)
		configMap = map[string]interface{}{
			"error":   cfg.Error,
			"message": cfg.Message,
			"data":    cfg.Data,
		}

	case workflow.TaskKindRun:
		cfg := task.Config.(*workflow.RunTaskConfig)
		configMap = map[string]interface{}{
			"workflow": cfg.WorkflowName,
			"input":    cfg.Input,
		}

	default:
		return nil, fmt.Errorf("unknown task kind: %s", task.Kind)
	}

	// Convert to protobuf Struct
	pbStruct, err := structpb.NewStruct(configMap)
	if err != nil {
		return nil, fmt.Errorf("creating protobuf struct: %w", err)
	}

	return pbStruct, nil
}
