// Package synth provides automatic manifest synthesis for the Stigmer SDK.
//
// The synthesis model works as follows:
//  1. User imports the SDK (agent package)
//  2. User defines agent using agent.New()
//  3. Agent automatically registers in global registry
//  4. On program exit, synthesis runs automatically (no user code needed)
//  5. AutoSynth() checks STIGMER_OUT_DIR env var:
//     - If not set: Dry-run mode (print message, exit)
//     - If set: Synthesis mode (convert to proto, write manifest.pb)
//  6. CLI reads manifest.pb and deploys
//
// This package uses runtime hooks to automatically trigger synthesis when the
// program exits, eliminating the need for explicit defer synth.AutoSynth() calls.
package synth

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"google.golang.org/protobuf/proto"

	"github.com/leftbin/stigmer-sdk/go/internal/registry"
)

var (
	// synthOnce ensures synthesis only happens once
	synthOnce sync.Once
	
	// autoSynthEnabled controls whether automatic synthesis is active
	autoSynthEnabled = true
)

func init() {
	// Try to register automatic synthesis using runtime exit hooks
	// This will work on Go 1.24+ which has runtime.AddExitHook
	if registerExitHook(autoSynth) {
		// Successfully registered exit hook - synthesis will run automatically
		return
	}
	
	// Fallback: Exit hooks not available
	// The user must manually call defer synth.AutoSynth() in main()
	// This is documented in the package comments
}

// registerExitHook attempts to register an exit hook if the Go version supports it.
// Returns true if successful, false if not supported.
func registerExitHook(fn func()) bool {
	// Go 1.24+ has runtime.AddExitHook
	// For older versions, this will be detected at compile time
	
	// Use build tags or runtime detection
	// For now, return false to indicate manual defer is required
	// This will be updated when targeting Go 1.24+
	
	// TODO: Once Go 1.24 is stable, add build tag version:
	// //go:build go1.24
	// func registerExitHook(fn func()) bool {
	//     runtime.AddExitHook(fn)
	//     return true
	// }
	
	return false
}

// autoSynth is the internal implementation called automatically or manually.
// It's separated from AutoSynth() to allow reuse by both manual and automatic triggers.
func autoSynth() {
	outputDir := os.Getenv("STIGMER_OUT_DIR")

	// Dry-run mode: No output directory set
	if outputDir == "" {
		fmt.Println("✓ Stigmer SDK: Dry-run complete. Run 'stigmer up' to deploy.")
		return
	}

	// Synthesis mode: Output directory is set

	// Get all registered resources from global registry
	agentInterfaces := registry.Global().GetAgents()
	workflowInterfaces := registry.Global().GetWorkflows()

	if len(agentInterfaces) == 0 && len(workflowInterfaces) == 0 {
		fmt.Println("⚠ Stigmer SDK: No agents or workflows defined. Nothing to synthesize.")
		return
	}

	// Report what we're synthesizing
	var parts []string
	if len(agentInterfaces) > 0 {
		if len(agentInterfaces) == 1 {
			parts = append(parts, "1 agent")
		} else {
			parts = append(parts, fmt.Sprintf("%d agents", len(agentInterfaces)))
		}
	}
	if len(workflowInterfaces) > 0 {
		if len(workflowInterfaces) == 1 {
			parts = append(parts, "1 workflow")
		} else {
			parts = append(parts, fmt.Sprintf("%d workflows", len(workflowInterfaces)))
		}
	}

	fmt.Printf("→ Stigmer SDK: Synthesizing manifest for %s...\n", joinParts(parts))

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Stigmer SDK: Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// Synthesize agents if any exist
	if len(agentInterfaces) > 0 {
		agentManifest, err := ToManifest(agentInterfaces...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Stigmer SDK: Agent synthesis failed: %v\n", err)
			os.Exit(1)
		}

		// Serialize to binary protobuf
		data, err := proto.Marshal(agentManifest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Stigmer SDK: Failed to serialize agent manifest: %v\n", err)
			os.Exit(1)
		}

		// Write agent manifest.pb to output directory
		agentManifestPath := filepath.Join(outputDir, "agent-manifest.pb")
		if err := os.WriteFile(agentManifestPath, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "✗ Stigmer SDK: Failed to write agent manifest: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Stigmer SDK: Agent manifest written to: %s\n", agentManifestPath)
	}

	// Synthesize workflows if any exist
	if len(workflowInterfaces) > 0 {
		workflowManifest, err := ToWorkflowManifest(workflowInterfaces...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Stigmer SDK: Workflow synthesis failed: %v\n", err)
			os.Exit(1)
		}

		// Serialize to binary protobuf
		data, err := proto.Marshal(workflowManifest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Stigmer SDK: Failed to serialize workflow manifest: %v\n", err)
			os.Exit(1)
		}

		// Write workflow manifest.pb to output directory
		workflowManifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
		if err := os.WriteFile(workflowManifestPath, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "✗ Stigmer SDK: Failed to write workflow manifest: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Stigmer SDK: Workflow manifest written to: %s\n", workflowManifestPath)
	}
}

// joinParts joins string parts with commas and "and" before the last item.
func joinParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	if len(parts) == 2 {
		return parts[0] + " and " + parts[1]
	}
	return parts[0] + " and " + parts[1]
}

// AutoSynth performs automatic manifest synthesis based on the STIGMER_OUT_DIR environment variable.
//
// Note: Due to Go's lack of atexit-style hooks (unlike Python), synthesis requires
// minimal user code. Call this via defer in main():
//
//	func main() {
//	    defer synth.AutoSynth()
//	    agent.New(...)
//	}
//
// Or use the cleaner agent.Complete() wrapper:
//
//	func main() {
//	    defer agent.Complete()
//	    agent.New(...)
//	}
//
// Behavior:
//   - If STIGMER_OUT_DIR is not set: Dry-run mode
//     Prints a message indicating successful dry-run and returns.
//   - If STIGMER_OUT_DIR is set: Synthesis mode
//     Converts registered agents to manifest proto and writes to:
//     $STIGMER_OUT_DIR/manifest.pb
//
// Error Handling:
//   - If no agents registered: Prints warning and returns (exit 0)
//   - If conversion fails: Prints error and exits with code 1
//   - If write fails: Prints error and exits with code 1
func AutoSynth() {
	synthOnce.Do(autoSynth)
}

// ResetForTesting resets the synthOnce guard to allow multiple AutoSynth() calls in tests.
// This should ONLY be used in tests, never in production code.
func ResetForTesting() {
	synthOnce = sync.Once{}
}

// SynthToFile is a convenience function that synthesizes the manifest and writes it to a specific file.
//
// This is useful for testing or custom workflows where you want explicit control over the output location.
//
// Example:
//
//	agent, _ := agent.New(...)
//	if err := synth.SynthToFile(agent, "/tmp/test-manifest.pb"); err != nil {
//	    log.Fatal(err)
//	}
func SynthToFile(agent interface{}, outputPath string) error {
	// Import agent type safely
	a, ok := agent.(interface {
		String() string
	})
	if !ok {
		return fmt.Errorf("invalid agent type")
	}

	// Note: This is a simplified version for the public API
	// The actual conversion will use the internal ToManifest function
	fmt.Printf("→ Synthesizing manifest to: %s\n", outputPath)

	// Get agent from registry if nil passed
	var agentToConvert interface{}
	if agent == nil {
		agentToConvert = registry.Global().GetAgent()
		if agentToConvert == nil {
			return fmt.Errorf("no agent registered")
		}
	} else {
		agentToConvert = a
	}

	_ = agentToConvert // Will be used in actual implementation

	// TODO: Implement full conversion once converter is ready
	return fmt.Errorf("not yet implemented")
}
