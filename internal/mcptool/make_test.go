package tool

import (
	"strings"
	"testing"
)

func TestSmartFilterOutput(t *testing.T) {
	// Test case 1: Build with errors
	stderr := `gcc -o main main.c
main.c:5:1: error: 'return' with no value, in function returning non-void
main.c:10:15: warning: unused variable 'x'
collect2: error: ld returned 1 exit status
make: *** [main] Error 1`

	stdout := `Starting build...
Compiling main.c
Processing dependencies...
Build step 1 of 5 completed
Build step 2 of 5 completed`

	result := smartFilterOutput(stdout, stderr)

	// Should contain error indicators
	if !strings.Contains(result, "🚨 Errors/Critical Issues:") {
		t.Error("Expected error section in output")
	}

	// Should contain the actual error
	if !strings.Contains(result, "error: 'return' with no value") {
		t.Error("Expected specific error message")
	}

	// Should contain warning section
	if !strings.Contains(result, "⚠️  Warnings:") {
		t.Error("Expected warning section")
	}

	// Should contain line numbers
	if !strings.Contains(result, "Line") {
		t.Error("Expected line numbers in output")
	}

	// Should show omitted line count
	if !strings.Contains(result, "lines omitted") {
		t.Error("Expected omitted line count")
	}
}

func TestSmartFilterOutputSuccess(t *testing.T) {
	// Test case 2: Successful build
	stdout := `Starting build...
Compiling main.c
Linking...
Build completed successfully
Tests passed: 15/15
All checks completed`

	stderr := ""

	result := smartFilterOutput(stdout, stderr)

	// Should contain summary section
	if !strings.Contains(result, "ℹ️  Summary:") {
		t.Error("Expected summary section for successful build")
	}

	// Should contain success indicators
	if !strings.Contains(result, "completed") || !strings.Contains(result, "passed") {
		t.Error("Expected success keywords in output")
	}
}

func TestSmartFilterOutputNoKeywords(t *testing.T) {
	// Test case 3: Output with no keywords (fallback to truncation)
	stdout := `Step 1
Step 2
Step 3
Step 4
Step 5`

	stderr := ""

	result := smartFilterOutput(stdout, stderr)

	// Should fall back to showing the output directly (since it's short)
	if !strings.Contains(result, "Step") {
		t.Error("Expected fallback output to contain steps")
	}
}

func TestGetLinePriority(t *testing.T) {
	testCases := []struct {
		line     string
		expected int
	}{
		{"error: undefined reference to main", 1}, // High priority
		{"make: *** [target] Error 1", 1},         // High priority
		{"warning: unused variable", 2},           // Medium priority
		{"deprecated function call", 2},           // Medium priority
		{"build completed successfully", 3},       // Info
		{"tests passed", 3},                       // Info
		{"some regular output", 0},                // No priority
	}

	for _, tc := range testCases {
		result := getLinePriority(tc.line)
		if result != tc.expected {
			t.Errorf("getLinePriority(%q) = %d, expected %d", tc.line, result, tc.expected)
		}
	}
}
