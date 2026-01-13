// Package synth provides automatic manifest synthesis for the Stigmer SDK.
//
// The synthesis model works as follows:
//  1. User defines agent using agent.New()
//  2. Agent automatically registers in global registry
//  3. User calls defer synth.AutoSynth() in main()
//  4. On exit, AutoSynth() checks STIGMER_OUT_DIR env var:
//     - If not set: Dry-run mode (print message, exit)
//     - If set: Synthesis mode (convert to proto, write manifest.pb)
//  5. CLI reads manifest.pb and deploys
package synth

import (
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/proto"

	"github.com/leftbin/stigmer-sdk/go/internal/registry"
)

// AutoSynth performs automatic manifest synthesis based on the STIGMER_OUT_DIR environment variable.
//
// This should be called via defer in the user's main() function:
//
//	func main() {
//	    defer synth.AutoSynth()
//
//	    agent, _ := agent.New(...)
//	    // ... configure agent
//	}
//
// Behavior:
//   - If STIGMER_OUT_DIR is not set: Dry-run mode
//     Prints a message indicating successful dry-run and returns.
//     This is useful for testing agent definitions without deploying.
//
//   - If STIGMER_OUT_DIR is set: Synthesis mode
//     Converts the registered agent to a manifest proto and writes it to:
//     $STIGMER_OUT_DIR/manifest.pb
//
// Error Handling:
//   - If no agent is registered: Prints warning and returns (exit 0)
//   - If conversion fails: Prints error and exits with code 1
//   - If write fails: Prints error and exits with code 1
func AutoSynth() {
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
