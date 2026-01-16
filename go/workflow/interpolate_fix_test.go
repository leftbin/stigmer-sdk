package workflow

import (
	"testing"
)

// TestInterpolateFix verifies that Interpolate generates valid Serverless Workflow expressions
func TestInterpolateFix(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "variable with path",
			input:    []string{VarRef("apiURL"), "/posts/1"},
			expected: "${ .apiURL + \"/posts/1\" }",
		},
		{
			name:     "bearer token",
			input:    []string{"Bearer ", VarRef("token")},
			expected: "${ \"Bearer \" + .token }",
		},
		{
			name:     "complex URL",
			input:    []string{"https://", VarRef("domain"), "/api/", FieldRef("version")},
			expected: "${ \"https://\" + .domain + \"/api/\" + .version }",
		},
		{
			name:     "plain string (no expressions)",
			input:    []string{"https://api.example.com/data"},
			expected: "https://api.example.com/data",
		},
		{
			name:     "only expression",
			input:    []string{VarRef("fullURL")},
			expected: "${.fullURL}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Interpolate(tt.input...)
			if result != tt.expected {
				t.Errorf("Interpolate(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestInterpolateRealWorldExample tests the actual use case from basic-data-fetch workflow
func TestInterpolateRealWorldExample(t *testing.T) {
	// This is what the user's workflow does:
	// workflow.WithURI(workflow.Interpolate(workflow.VarRef("apiURL"), "/posts/1"))
	
	uri := Interpolate(VarRef("apiURL"), "/posts/1")
	expected := "${ .apiURL + \"/posts/1\" }"
	
	if uri != expected {
		t.Errorf("Real-world example failed: got %q, want %q", uri, expected)
	}
	
	t.Logf("✅ Generated valid expression: %s", uri)
	t.Logf("   This will be evaluated at runtime to: https://jsonplaceholder.typicode.com/posts/1")
}
