package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/fpt/go-dev-mcp/internal/infra"
)

// ValidationResult represents the result of a single validation check
type ValidationResult struct {
	Check   string `json:"check"`
	Status  string `json:"status"` // "pass", "fail", "error"
	Output  string `json:"output,omitempty"`
	Summary string `json:"summary"`
}

// ValidationReport represents the complete validation report
type ValidationReport struct {
	Directory string             `json:"directory"`
	Results   []ValidationResult `json:"results"`
	Summary   string             `json:"summary"`
}

// ValidateGoCode runs multiple Go validation checks on the specified directory
func ValidateGoCode(ctx context.Context, directory string) (*ValidationReport, error) {
	report := &ValidationReport{
		Directory: directory,
		Results:   []ValidationResult{},
	}

	// Define validation checks
	checks := []struct {
		name        string
		cmd         string
		args        []string
		description string
	}{
		{
			name:        "go vet",
			cmd:         "go",
			args:        []string{"vet", "./..."},
			description: "Static analysis to find suspicious constructs",
		},
		{
			name:        "go build (dry-run)",
			cmd:         "go",
			args:        []string{"build", "-n", "./..."},
			description: "Check if code compiles without building",
		},
		{
			name:        "go mod tidy (check)",
			cmd:         "go",
			args:        []string{"mod", "tidy", "-diff"},
			description: "Check if go.mod is tidy",
		},
		{
			name:        "gofmt check",
			cmd:         "gofmt",
			args:        []string{"-l", "."},
			description: "Check if code is properly formatted",
		},
	}

	// Run each check
	for _, check := range checks {
		result := runValidationCheck(ctx, directory, check.name, check.cmd, check.args, check.description)
		report.Results = append(report.Results, result)
	}

	// Generate summary
	passed := 0
	failed := 0
	errors := 0
	for _, result := range report.Results {
		switch result.Status {
		case "pass":
			passed++
		case "fail":
			failed++
		case "error":
			errors++
		}
	}

	if errors > 0 {
		report.Summary = fmt.Sprintf("Validation completed with %d errors, %d failures, %d passed", errors, failed, passed)
	} else if failed > 0 {
		report.Summary = fmt.Sprintf("Validation failed: %d checks failed, %d passed", failed, passed)
	} else {
		report.Summary = fmt.Sprintf("All %d validation checks passed ✓", passed)
	}

	return report, nil
}

func runValidationCheck(ctx context.Context, workDir, name, cmdName string, args []string, description string) ValidationResult {
	result := ValidationResult{
		Check: fmt.Sprintf("%s - %s", name, description),
	}

	// Use infra.Run to execute command with proper stdout/stderr separation
	stdout, stderr, exitCode, err := infra.Run(workDir, cmdName, args...)

	if err != nil {
		// Command couldn't run at all
		result.Status = "error"
		result.Output = err.Error()
		result.Summary = fmt.Sprintf("Could not run %s: %v", name, err)
	} else if exitCode != 0 {
		// Command ran but returned non-zero exit code
		result.Status = "fail"

		// Use stderr for error information, stdout for normal output
		if stderr != "" {
			result.Output = stderr
		} else if stdout != "" {
			result.Output = stdout
		}

		switch name {
		case "gofmt check":
			if stdout != "" {
				result.Summary = fmt.Sprintf("Files need formatting: %s", strings.ReplaceAll(stdout, "\n", ", "))
			} else {
				result.Summary = "Files need formatting"
			}
		case "go mod tidy (check)":
			result.Summary = "go.mod needs tidying"
		case "go vet":
			// go vet typically outputs to stderr
			issues := stderr
			if issues == "" {
				issues = stdout
			}
			if issues != "" {
				lines := strings.Split(issues, "\n")
				result.Summary = fmt.Sprintf("Found %d vet issues", len(lines))
			} else {
				result.Summary = "Vet found issues"
			}
		case "go build (dry-run)":
			result.Summary = "Build would fail - compilation errors found"
		default:
			result.Summary = fmt.Sprintf("Check failed: %s", name)
		}
	} else {
		// Command succeeded (exit code 0)
		result.Status = "pass"

		switch name {
		case "gofmt check":
			// gofmt -l returns 0 even when files need formatting, but lists files to stdout
			if stdout != "" {
				result.Status = "fail"
				result.Output = stdout
				result.Summary = fmt.Sprintf("Files need formatting: %s", strings.ReplaceAll(stdout, "\n", ", "))
			} else {
				result.Summary = "All files are properly formatted"
			}
		case "go vet":
			result.Summary = "No vet issues found"
		case "go build (dry-run)":
			// Don't include verbose build output when successful
			result.Summary = "Code compiles successfully"
		case "go mod tidy (check)":
			result.Summary = "go.mod is tidy"
		default:
			if stdout != "" {
				result.Output = stdout
			}
			result.Summary = fmt.Sprintf("%s passed", name)
		}
	}

	return result
}
