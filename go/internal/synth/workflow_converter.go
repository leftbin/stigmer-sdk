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
func ToWorkflowManifest(workflowInterfaces ...interface{}) (*workflowv1.WorkflowManifest, error) {
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

		// Convert to proto
		protoWorkflow, err := workflowToProto(wf)
		if err != nil {
			return nil, fmt.Errorf("workflow[%d] %s: %w", wfIdx, wf.Document.Name, err)
		}

		// Add to manifest
		manifest.Workflows = append(manifest.Workflows, protoWorkflow)
	}

	return manifest, nil
}

// workflowToProto converts a workflow.Workflow to a workflowv1.Workflow proto.
func workflowToProto(wf *workflow.Workflow) (*workflowv1.Workflow, error) {
	// Create workflow proto
	protoWorkflow := &workflowv1.Workflow{
		ApiVersion: "agentic.stigmer.ai/v1",
		Kind:       "Workflow",
	}

	// Convert metadata (placeholder - would need actual metadata proto structure)
	// For now, we'll focus on the spec

	// Convert spec
	spec, err := workflowSpecToProto(wf)
	if err != nil {
		return nil, fmt.Errorf("converting spec: %w", err)
	}
	protoWorkflow.Spec = spec

	return protoWorkflow, nil
}

// workflowSpecToProto converts workflow spec to proto.
func workflowSpecToProto(wf *workflow.Workflow) (*workflowv1.WorkflowSpec, error) {
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

	// Convert tasks
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
		for i, c := range cfg.Cases {
			cases[i] = map[string]interface{}{
				"condition": c.Condition,
				"then":      c.Then,
			}
		}
		configMap = map[string]interface{}{
			"cases":   mapSliceToInterfaceSlice(cases),
			"default": cfg.DefaultTask,
		}

	case workflow.TaskKindFor:
		cfg := task.Config.(*workflow.ForTaskConfig)
		doTasks := make([]map[string]interface{}, len(cfg.Do))
		for i, t := range cfg.Do {
			doTasks[i] = map[string]interface{}{
				"name": t.Name,
				"kind": string(t.Kind), // Convert TaskKind enum to string
			}
		}
		configMap = map[string]interface{}{
			"in": cfg.In,
			"do": mapSliceToInterfaceSlice(doTasks),
		}

	case workflow.TaskKindFork:
		cfg := task.Config.(*workflow.ForkTaskConfig)
		branches := make([]map[string]interface{}, len(cfg.Branches))
		for i, b := range cfg.Branches {
			branches[i] = map[string]interface{}{
				"name": b.Name,
			}
		}
		configMap = map[string]interface{}{
			"branches": mapSliceToInterfaceSlice(branches),
		}

	case workflow.TaskKindTry:
		cfg := task.Config.(*workflow.TryTaskConfig)
		catchBlocks := make([]map[string]interface{}, len(cfg.Catch))
		for i, c := range cfg.Catch {
			catchBlocks[i] = map[string]interface{}{
				"errors": stringSliceToInterfaceSlice(c.Errors), // Convert []string to []interface{}
				"as":     c.As,
			}
		}
		configMap = map[string]interface{}{
			"catch": mapSliceToInterfaceSlice(catchBlocks),
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
