package synth

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"

	// Import Buf-generated proto packages
	workflowv1 "buf.build/gen/go/leftbin/stigmer/protocolbuffers/go/ai/stigmer/agentic/workflow/v1"
	sdk "buf.build/gen/go/leftbin/stigmer/protocolbuffers/go/ai/stigmer/commons/sdk"

	// Import SDK types
	"github.com/leftbin/stigmer-sdk/go/environment"
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
//   - environment.Variable → workflowv1.EnvironmentSpec
//
// Returns an error if any nested conversion fails.
func ToWorkflowManifest(workflowInterfaces ...interface{}) (*sdk.WorkflowManifest, error) {
	if len(workflowInterfaces) == 0 {
		return nil, fmt.Errorf("at least one workflow is required")
	}

	// Create manifest with empty workflows list
	manifest := &sdk.WorkflowManifest{
		Workflows: []*workflowv1.Workflow{},
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
	}

	// Convert document
	spec.Document = &workflowv1.WorkflowDocument{
		Dsl:         wf.Document.DSL,
		Namespace:   wf.Document.Namespace,
		Name:        wf.Document.Name,
		Version:     wf.Document.Version,
		Description: wf.Document.Description,
	}

	// Convert tasks
	for i, task := range wf.Tasks {
		protoTask, err := taskToProto(task)
		if err != nil {
			return nil, fmt.Errorf("converting task[%d] %s: %w", i, task.Name, err)
		}
		spec.Tasks = append(spec.Tasks, protoTask)
	}

	// Convert environment variables
	if len(wf.EnvironmentVariables) > 0 {
		envSpec, err := environmentVariablesToEnvSpec(wf.EnvironmentVariables)
		if err != nil {
			return nil, fmt.Errorf("converting environment variables: %w", err)
		}
		spec.EnvSpec = envSpec
	}

	return spec, nil
}

// taskToProto converts a workflow.Task to a workflowv1.WorkflowTask proto.
func taskToProto(task *workflow.Task) (*workflowv1.WorkflowTask, error) {
	protoTask := &workflowv1.WorkflowTask{
		Name: task.Name,
		// Kind will be set based on task type
	}

	// Convert task config to google.protobuf.Struct
	taskConfig, taskKind, err := taskConfigToStruct(task)
	if err != nil {
		return nil, fmt.Errorf("converting task config: %w", err)
	}

	protoTask.Kind = taskKind
	protoTask.TaskConfig = taskConfig

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

// taskConfigToStruct converts task configuration to google.protobuf.Struct.
func taskConfigToStruct(task *workflow.Task) (*structpb.Struct, int32, error) {
	// Map SDK task kind to proto enum
	var protoKind int32
	var configMap map[string]interface{}

	switch task.Kind {
	case workflow.TaskKindSet:
		protoKind = 1 // SET
		cfg := task.Config.(*workflow.SetTaskConfig)
		configMap = map[string]interface{}{
			"variables": cfg.Variables,
		}

	case workflow.TaskKindHttpCall:
		protoKind = 2 // HTTP_CALL
		cfg := task.Config.(*workflow.HttpCallTaskConfig)
		configMap = map[string]interface{}{
			"method": cfg.Method,
			"endpoint": map[string]interface{}{
				"uri": cfg.URI,
			},
			"headers":        cfg.Headers,
			"body":           cfg.Body,
			"timeout_seconds": cfg.TimeoutSeconds,
		}

	case workflow.TaskKindGrpcCall:
		protoKind = 3 // GRPC_CALL
		cfg := task.Config.(*workflow.GrpcCallTaskConfig)
		configMap = map[string]interface{}{
			"service": cfg.Service,
			"method":  cfg.Method,
			"body":    cfg.Body,
		}

	case workflow.TaskKindSwitch:
		protoKind = 4 // SWITCH
		cfg := task.Config.(*workflow.SwitchTaskConfig)
		cases := make([]map[string]interface{}, len(cfg.Cases))
		for i, c := range cfg.Cases {
			cases[i] = map[string]interface{}{
				"condition": c.Condition,
				"then":      c.Then,
			}
		}
		configMap = map[string]interface{}{
			"cases":   cases,
			"default": cfg.DefaultTask,
		}

	case workflow.TaskKindFor:
		protoKind = 5 // FOR
		cfg := task.Config.(*workflow.ForTaskConfig)
		doTasks := make([]map[string]interface{}, len(cfg.Do))
		for i, t := range cfg.Do {
			doTasks[i] = map[string]interface{}{
				"name": t.Name,
				"kind": t.Kind,
			}
		}
		configMap = map[string]interface{}{
			"in": cfg.In,
			"do": doTasks,
		}

	case workflow.TaskKindFork:
		protoKind = 6 // FORK
		cfg := task.Config.(*workflow.ForkTaskConfig)
		branches := make([]map[string]interface{}, len(cfg.Branches))
		for i, b := range cfg.Branches {
			branches[i] = map[string]interface{}{
				"name": b.Name,
			}
		}
		configMap = map[string]interface{}{
			"branches": branches,
		}

	case workflow.TaskKindTry:
		protoKind = 7 // TRY
		cfg := task.Config.(*workflow.TryTaskConfig)
		catchBlocks := make([]map[string]interface{}, len(cfg.Catch))
		for i, c := range cfg.Catch {
			catchBlocks[i] = map[string]interface{}{
				"errors": c.Errors,
				"as":     c.As,
			}
		}
		configMap = map[string]interface{}{
			"catch": catchBlocks,
		}

	case workflow.TaskKindListen:
		protoKind = 8 // LISTEN
		cfg := task.Config.(*workflow.ListenTaskConfig)
		configMap = map[string]interface{}{
			"event": cfg.Event,
		}

	case workflow.TaskKindWait:
		protoKind = 9 // WAIT
		cfg := task.Config.(*workflow.WaitTaskConfig)
		configMap = map[string]interface{}{
			"duration": cfg.Duration,
		}

	case workflow.TaskKindCallActivity:
		protoKind = 10 // CALL_ACTIVITY
		cfg := task.Config.(*workflow.CallActivityTaskConfig)
		configMap = map[string]interface{}{
			"activity": cfg.Activity,
			"input":    cfg.Input,
		}

	case workflow.TaskKindRaise:
		protoKind = 11 // RAISE
		cfg := task.Config.(*workflow.RaiseTaskConfig)
		configMap = map[string]interface{}{
			"error":   cfg.Error,
			"message": cfg.Message,
			"data":    cfg.Data,
		}

	case workflow.TaskKindRun:
		protoKind = 12 // RUN
		cfg := task.Config.(*workflow.RunTaskConfig)
		configMap = map[string]interface{}{
			"workflow": cfg.WorkflowName,
			"input":    cfg.Input,
		}

	default:
		return nil, 0, fmt.Errorf("unknown task kind: %s", task.Kind)
	}

	// Convert to protobuf Struct
	pbStruct, err := structpb.NewStruct(configMap)
	if err != nil {
		return nil, 0, fmt.Errorf("creating protobuf struct: %w", err)
	}

	return pbStruct, protoKind, nil
}

// environmentVariablesToEnvSpec converts environment variables to EnvironmentSpec proto.
func environmentVariablesToEnvSpec(vars []environment.Variable) (*workflowv1.EnvironmentSpec, error) {
	// This would need the actual EnvironmentSpec proto definition
	// For now, return a placeholder
	// TODO: Implement when EnvironmentSpec proto is available
	return nil, nil
}
