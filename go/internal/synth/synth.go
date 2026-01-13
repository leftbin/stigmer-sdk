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

	// Get all registered agents from global registry
	agentInterfaces := registry.Global().GetAgents()
	if len(agentInterfaces) == 0 {
		fmt.Println("⚠ Stigmer SDK: No agents defined. Nothing to synthesize.")
		return
	}

	agentCount := len(agentInterfaces)
	if agentCount == 1 {
		fmt.Println("→ Stigmer SDK: Synthesizing manifest for 1 agent...")
	} else {
		fmt.Printf("→ Stigmer SDK: Synthesizing manifest for %d agents...\n", agentCount)
	}

	// Convert all SDK agents to manifest proto
	manifest, err := ToManifest(agentInterfaces...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ Stigmer SDK: Synthesis failed: %v\n", err)
		os.Exit(1)
	}

	// Serialize to binary protobuf
	data, err := proto.Marshal(manifest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ Stigmer SDK: Failed to serialize manifest: %v\n", err)
		os.Exit(1)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Stigmer SDK: Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// Write manifest.pb to output directory
	manifestPath := filepath.Join(outputDir, "manifest.pb")
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Stigmer SDK: Failed to write manifest: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Stigmer SDK: Manifest written to: %s\n", manifestPath)
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
