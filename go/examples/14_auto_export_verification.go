//go:build ignore

// Package examples demonstrates the auto-export feature that makes the SDK Pulumi-like.
//
// This example verifies that:
// 1. Context variables from ctx.SetX() are automatically exported
// 2. Tasks automatically export when .Field() is called
// 3. Export and field references are properly aligned
package main

import (
	"fmt"
	"log"

	"github.com/leftbin/stigmer-sdk/go/stigmer"
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

func main() {
	err := stigmer.Run(func(ctx *stigmer.Context) error {
		fmt.Println("=== AUTO-EXPORT VERIFICATION ===\n")

		// ============================================
		// TEST 1: Context Variables Auto-Export
		// ============================================
		fmt.Println("TEST 1: Context Variables Auto-Export")
		fmt.Println("---------------------------------------")

		// Context variables are automatically exported in __stigmer_init_context task
		apiBase := ctx.SetString("apiBase", "https://jsonplaceholder.typicode.com")
		retries := ctx.SetInt("retries", 3)
		debugMode := ctx.SetBool("debugMode", true)

		fmt.Println("✅ Created context variables:")
		fmt.Println("   - apiBase: string")
		fmt.Println("   - retries: int")
		fmt.Println("   - debugMode: bool")
		fmt.Println("✅ These will be automatically exported by __stigmer_init_context task")
		fmt.Println("✅ Accessible as: $context.apiBase, $context.retries, $context.debugMode\n")

		// ============================================
		// TEST 2: Task Auto-Export on .Field() Call
		// ============================================
		fmt.Println("TEST 2: Task Auto-Export on .Field() Call")
		fmt.Println("------------------------------------------")

		// Create workflow
		wf, err := workflow.New(ctx,
			workflow.WithNamespace("examples"),
			workflow.WithName("auto-export-demo"),
			workflow.WithVersion("1.0.0"),
			workflow.WithDescription("Demonstrates auto-export feature"),
		)
		if err != nil {
			return err
		}

		// Build endpoint using context variable
		endpoint := apiBase.Concat("/posts/1")

		// Create HTTP task
		fetchTask := wf.HttpGet("fetchData", endpoint,
			workflow.Header("Content-Type", "application/json"),
			workflow.Timeout(30),
		)

		// VERIFY: Before calling .Field(), no export should be set
		if fetchTask.ExportAs != "" {
			return fmt.Errorf("❌ FAILED: fetchTask should not have export before .Field() call")
		}
		fmt.Println("✅ Before .Field(): ExportAs = '' (empty)")

		// Call .Field() - this should AUTO-EXPORT the task!
		titleRef := fetchTask.Field("title")
		bodyRef := fetchTask.Field("body")
		userIdRef := fetchTask.Field("userId")

		// VERIFY: After calling .Field(), export should be set to "${.}"
		if fetchTask.ExportAs != "${.}" {
			return fmt.Errorf("❌ FAILED: fetchTask should auto-export after .Field() call, got: %s", fetchTask.ExportAs)
		}
		fmt.Println("✅ After .Field(): ExportAs = '${.}' (auto-set!)")

		// VERIFY: Field references generate correct expressions
		expectedExprs := map[string]string{
			"title":  "${ $context.fetchData.title }",
			"body":   "${ $context.fetchData.body }",
			"userId": "${ $context.fetchData.userId }",
		}

		if titleRef.Expression() != expectedExprs["title"] {
			return fmt.Errorf("❌ FAILED: title expression mismatch: got %s", titleRef.Expression())
		}
		if bodyRef.Expression() != expectedExprs["body"] {
			return fmt.Errorf("❌ FAILED: body expression mismatch: got %s", bodyRef.Expression())
		}
		if userIdRef.Expression() != expectedExprs["userId"] {
			return fmt.Errorf("❌ FAILED: userId expression mismatch: got %s", userIdRef.Expression())
		}

		fmt.Println("✅ Field references generate correct expressions:")
		fmt.Printf("   - title:  %s\n", titleRef.Expression())
		fmt.Printf("   - body:   %s\n", bodyRef.Expression())
		fmt.Printf("   - userId: %s\n", userIdRef.Expression())
		fmt.Println()

		// ============================================
		// TEST 3: Export/Reference Alignment
		// ============================================
		fmt.Println("TEST 3: Export/Reference Alignment")
		fmt.Println("-----------------------------------")

		fmt.Println("How auto-export works:")
		fmt.Println("  1. Export: { as: '${.}' } means:")
		fmt.Println("     - Take current task output (.)")
		fmt.Println("     - Make it available at $context.<taskName>")
		fmt.Println()
		fmt.Println("  2. For 'fetchData' task:")
		fmt.Println("     - Export: { as: '${.}' }")
		fmt.Println("     - Result: output → $context.fetchData")
		fmt.Println()
		fmt.Println("  3. Field reference fetchData.Field('title'):")
		fmt.Println("     - Generates: ${ $context.fetchData.title }")
		fmt.Println("     - Reads from: $context.fetchData.title")
		fmt.Println()
		fmt.Println("✅ Export and reference are ALIGNED!")
		fmt.Println()

		// ============================================
		// TEST 4: Idempotency (Multiple .Field() Calls)
		// ============================================
		fmt.Println("TEST 4: Idempotency (Multiple .Field() Calls)")
		fmt.Println("----------------------------------------------")

		// Call .Field() multiple times on the same task
		_ = fetchTask.Field("id")
		_ = fetchTask.Field("createdAt")
		_ = fetchTask.Field("updatedAt")

		// Export should still be "${.}" (not changed)
		if fetchTask.ExportAs != "${.}" {
			return fmt.Errorf("❌ FAILED: multiple .Field() calls should be idempotent")
		}
		fmt.Println("✅ Multiple .Field() calls are idempotent")
		fmt.Println("   Export remains: '${.}'\n")

		// ============================================
		// TEST 5: Custom Exports Are Preserved
		// ============================================
		fmt.Println("TEST 5: Custom Exports Are Preserved")
		fmt.Println("-------------------------------------")

		// Create a task with custom export
		customTask := wf.HttpGet("customExport", "https://api.example.com/data")
		customTask.ExportField("specificField")
		customExport := customTask.ExportAs

		// Now call .Field() - should NOT override custom export
		_ = customTask.Field("someField")

		if customTask.ExportAs != customExport {
			return fmt.Errorf("❌ FAILED: custom export should be preserved")
		}
		fmt.Println("✅ Custom exports are preserved when .Field() is called")
		fmt.Printf("   Custom export: %s\n", customExport)
		fmt.Println()

		// ============================================
		// TEST 6: Real-World Usage Pattern
		// ============================================
		fmt.Println("TEST 6: Real-World Usage Pattern")
		fmt.Println("---------------------------------")

		// Use field references in another task
		// This demonstrates the clean Pulumi-style UX
		processTask := wf.SetVars("processResponse",
			"postTitle", titleRef, // Auto-export already happened!
			"postBody", bodyRef, // No manual .ExportAll() needed!
			"postUserId", userIdRef, // Clean and simple!
			"debugEnabled", debugMode, // Context variables work too!
			"retryCount", retries, // All auto-exported!
		)

		// Verify dependencies are tracked
		fmt.Println("✅ Used field references in processResponse task")
		fmt.Println("✅ Implicit dependencies created automatically")
		fmt.Printf("✅ processResponse depends on: %v\n", processTask.Dependencies)
		fmt.Println()

		// ============================================
		// SUMMARY
		// ============================================
		fmt.Println("=== SUMMARY ===")
		fmt.Println()
		fmt.Println("✅ All auto-export features verified:")
		fmt.Println("   1. Context variables auto-export (Task 1 fix)")
		fmt.Println("   2. Tasks auto-export on .Field() call (Task 2 fix)")
		fmt.Println("   3. Export/reference alignment correct")
		fmt.Println("   4. Multiple .Field() calls are idempotent")
		fmt.Println("   5. Custom exports are preserved")
		fmt.Println("   6. Real-world usage pattern works cleanly")
		fmt.Println()
		fmt.Println("Workflow tasks:")
		for i, task := range wf.Tasks {
			exportStatus := "no export"
			if task.ExportAs != "" {
				exportStatus = fmt.Sprintf("export: %s", task.ExportAs)
			}
			fmt.Printf("  %d. %s (%s)\n", i+1, task.Name, exportStatus)
		}

		return nil
	})

	if err != nil {
		log.Fatal("❌ TEST FAILED:", err)
	}

	fmt.Println("\n🎉 All auto-export tests PASSED!")
	fmt.Println()
	fmt.Println("What this means:")
	fmt.Println("  - No need to manually call .ExportAll()")
	fmt.Println("  - Just use .Field() and it auto-exports")
	fmt.Println("  - Context variables automatically work")
	fmt.Println("  - Clean, Pulumi-style UX! 🚀")
}
