//go:build ignore

// Package examples demonstrates the UX improvement from auto-export.
//
// This example shows the BEFORE and AFTER code patterns to highlight
// how auto-export makes the SDK cleaner and more Pulumi-like.
package main

import (
	"fmt"
	"log"

	"github.com/leftbin/stigmer-sdk/go/stigmer"
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

func main() {
	fmt.Println("=== AUTO-EXPORT: BEFORE vs AFTER ===\n")

	err := stigmer.Run(func(ctx *stigmer.Context) error {
		// ============================================
		// BEFORE: Manual Export Required (Old Pattern)
		// ============================================
		fmt.Println("BEFORE Auto-Export:")
		fmt.Println("-------------------")
		fmt.Println("```go")
		fmt.Println("// Context variables didn't work")
		fmt.Println("apiBase := ctx.SetString(\"apiBase\", \"https://api.example.com\")")
		fmt.Println("endpoint := apiBase.Concat(\"/data\")  // ❌ Would fail at runtime!")
		fmt.Println("")
		fmt.Println("// Manual export required")
		fmt.Println("fetchTask := wf.HttpGet(\"fetch\", endpoint)")
		fmt.Println("fetchTask.ExportAll()  // ❌ Easy to forget!")
		fmt.Println("title := fetchTask.Field(\"title\")")
		fmt.Println("```")
		fmt.Println()
		fmt.Println("Problems:")
		fmt.Println("  ❌ Context variables didn't work")
		fmt.Println("  ❌ Had to remember to call .ExportAll()")
		fmt.Println("  ❌ Easy to forget and get runtime errors")
		fmt.Println("  ❌ Not discoverable - unclear when export is needed")
		fmt.Println()

		// ============================================
		// AFTER: Auto-Export (New Pattern)
		// ============================================
		fmt.Println("AFTER Auto-Export:")
		fmt.Println("------------------")
		fmt.Println("```go")
		fmt.Println("// Context variables just work")
		fmt.Println("apiBase := ctx.SetString(\"apiBase\", \"https://api.example.com\")")
		fmt.Println("endpoint := apiBase.Concat(\"/data\")  // ✅ Works!")
		fmt.Println("")
		fmt.Println("// Auto-export - no manual call needed")
		fmt.Println("fetchTask := wf.HttpGet(\"fetch\", endpoint)")
		fmt.Println("title := fetchTask.Field(\"title\")  // ✅ Auto-exports fetchTask!")
		fmt.Println("```")
		fmt.Println()
		fmt.Println("Benefits:")
		fmt.Println("  ✅ Context variables work automatically")
		fmt.Println("  ✅ No need to call .ExportAll()")
		fmt.Println("  ✅ .Field() call automatically exports")
		fmt.Println("  ✅ Clean, Pulumi-style implicit dependencies")
		fmt.Println("  ✅ Discoverable - just use .Field() naturally")
		fmt.Println()

		// ============================================
		// LIVE DEMONSTRATION
		// ============================================
		fmt.Println("=== LIVE DEMONSTRATION ===\n")

		// Context variables - now they work!
		apiBase := ctx.SetString("apiBase", "https://jsonplaceholder.typicode.com")
		orgName := ctx.SetString("org", "demo-org")

		fmt.Println("✅ Context variables created:")
		fmt.Printf("   apiBase: %s\n", apiBase.Value())
		fmt.Printf("   orgName: %s\n", orgName.Value())
		fmt.Println()

		// Create workflow
		wf, err := workflow.New(ctx,
			workflow.WithNamespace("examples"),
			workflow.WithName("auto-export-demo"),
			workflow.WithVersion("1.0.0"),
			workflow.WithOrg(orgName), // Context variable works!
		)
		if err != nil {
			return err
		}

		// Build endpoint using context variable - works automatically!
		endpoint := apiBase.Concat("/posts/1")

		// Create HTTP task - NO .ExportAll() call!
		fetchTask := wf.HttpGet("fetchData", endpoint,
			workflow.Header("Content-Type", "application/json"),
		)

		fmt.Println("✅ Created fetchData task (no .ExportAll() called)")
		fmt.Printf("   Export before .Field(): '%s'\n", fetchTask.ExportAs)
		fmt.Println()

		// Use .Field() - auto-export happens here!
		title := fetchTask.Field("title")
		body := fetchTask.Field("body")
		userId := fetchTask.Field("userId")

		fmt.Println("✅ Called .Field() three times")
		fmt.Printf("   Export after .Field(): '%s' (auto-set!)\n", fetchTask.ExportAs)
		fmt.Println()

		// Use field references in another task - clean and simple!
		processTask := wf.SetVars("processData",
			"postTitle", title,
			"postBody", body,
			"postUserId", userId,
			"organization", orgName, // Context variable reference
		)

		fmt.Println("✅ Used field references in processData task")
		fmt.Println("   No manual export, no manual dependencies!")
		fmt.Println()

		// ============================================
		// COMPARISON SUMMARY
		// ============================================
		fmt.Println("=== COMPARISON SUMMARY ===\n")

		fmt.Println("Lines of Code Comparison:")
		fmt.Println("-------------------------")
		fmt.Println("BEFORE:")
		fmt.Println("  fetchTask := wf.HttpGet(...)")
		fmt.Println("  fetchTask.ExportAll()        // Extra line!")
		fmt.Println("  title := fetchTask.Field(\"title\")")
		fmt.Println()
		fmt.Println("AFTER:")
		fmt.Println("  fetchTask := wf.HttpGet(...)")
		fmt.Println("  title := fetchTask.Field(\"title\")  // One less line!")
		fmt.Println()

		fmt.Println("Developer Experience:")
		fmt.Println("---------------------")
		fmt.Println("BEFORE: ❌ Manual export required, easy to forget")
		fmt.Println("AFTER:  ✅ Auto-export, just works!")
		fmt.Println()

		fmt.Println("Pulumi Alignment:")
		fmt.Println("-----------------")
		fmt.Println("BEFORE: ❌ Not Pulumi-like")
		fmt.Println("AFTER:  ✅ Matches Pulumi's implicit dependency pattern")
		fmt.Println()

		// Show final workflow structure
		fmt.Println("Final Workflow Structure:")
		fmt.Println("-------------------------")
		for i, task := range wf.Tasks {
			exportInfo := "no export"
			if task.ExportAs != "" {
				exportInfo = fmt.Sprintf("export: %s", task.ExportAs)
			}
			dependsInfo := ""
			if len(task.Dependencies) > 0 {
				dependsInfo = fmt.Sprintf(", depends on: %v", task.Dependencies)
			}
			fmt.Printf("  %d. %s (%s%s)\n", i+1, task.Name, exportInfo, dependsInfo)
		}

		_ = processTask // avoid unused warning

		return nil
	})

	if err != nil {
		log.Fatal("❌ FAILED:", err)
	}

	fmt.Println("\n🎉 Auto-Export Demo Complete!")
	fmt.Println()
	fmt.Println("Key Takeaways:")
	fmt.Println("  1. Context variables work automatically")
	fmt.Println("  2. .Field() call auto-exports the task")
	fmt.Println("  3. No manual .ExportAll() needed")
	fmt.Println("  4. Cleaner code, fewer lines")
	fmt.Println("  5. Pulumi-style UX achieved! 🚀")
}
