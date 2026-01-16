package stigmer

import "github.com/leftbin/stigmer-sdk/go/internal/synth"

// Complete triggers manifest synthesis and should be called via defer in main().
//
// This is the synthesis hook for the Stigmer SDK. It enables the "synthesis model"
// where agent definitions are automatically converted to protobuf manifests when
// the program exits.
//
// Usage:
//
//	import "github.com/leftbin/stigmer-sdk/go/stigmer"
//	import "github.com/leftbin/stigmer-sdk/go/agent"
//	
//	func main() {
//	    defer stigmer.Complete()
//	
//	    agent.New(
//	        agent.WithName("code-reviewer"),
//	        agent.WithInstructions("Review code and suggest improvements"),
//	    )
//	}
//
// Why is this needed?
//
// Unlike Python (which has atexit hooks), Go doesn't provide automatic program exit hooks.
// This single line of code enables the "synthesis model" where:
//  1. You define agents using agent.New()
//  2. Agents auto-register in the global registry
//  3. On program exit, Complete() synthesizes manifest.pb
//  4. The CLI reads manifest.pb and deploys
//
// This is the cleanest API possible given Go's limitations.
//
// See docs/architecture/synthesis-model.md for a detailed explanation of why
// Go requires this approach, and docs/implementation/synthesis-api-improvement.md
// for the evolution of this API.
func Complete() {
	synth.AutoSynth()
}
