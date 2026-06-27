package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/fpt/go-dev-mcp/internal/infra"
)

// Validation check names.
const (
	checkGoVet     = "go vet"
	checkGoBuild   = "go build (dry-run)"
	checkGoModTidy = "go mod tidy (check)"
	checkGofmt     = "gofmt check"
)

// Validation result statuses.
const (
	statusPass  = "pass"
	statusFail  = "fail"
	statusError = "error"
)

// ValidationResult represents the result of a single validation check
type ValidationResult struct {
	Check   string `json:"check"`
	Status  string `json:"status"` // pass, fail, or error
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
			name:        checkGoVet,
			cmd:         "go",
			args:        []string{"vet", "./..."},
			description: "Static analysis to find suspicious constructs",
		},
		{
			name:        checkGoBuild,
			cmd:         "go",
			args:        []string{"build", "-n", "./..."},
			description: "Check if code compiles without building",
		},
		{
			name:        checkGoModTidy,
			cmd:         "go",
			args:        []string{"mod", "tidy", "-diff"},
			description: "Check if go.mod is tidy",
		},
		{
			name:        checkGofmt,
			cmd:         "gofmt",
			args:        []string{"-l", "."},
			description: "Check if code is properly formatted",
		},
	}

	// Run each check
	for _, check := range checks {
		result := runValidationCheck(
			ctx,
			directory,
			check.name,
			check.cmd,
			check.args,
			check.description,
		)
		report.Results = append(report.Results, result)
	}

	// Generate summary
	passed := 0
	failed := 0
	errors := 0
	for _, result := range report.Results {
		switch result.Status {
		case statusPass:
			passed++
		case statusFail:
			failed++
		case statusError:
			errors++
		}
	}

	if errors > 0 {
		report.Summary = fmt.Sprintf(
			"Validation completed with %d errors, %d failures, %d passed",
			errors,
			failed,
			passed,
		)
	} else if failed > 0 {
		report.Summary = fmt.Sprintf(
			"Validation failed: %d checks failed, %d passed",
			failed,
			passed,
		)
	} else {
		report.Summary = fmt.Sprintf("All %d validation checks passed ✓", passed)
	}

	return report, nil
}

func runValidationCheck(
	_ context.Context,
	workDir, name, cmdName string,
	args []string,
	description string,
) ValidationResult {
	result := ValidationResult{
		Check: fmt.Sprintf("%s - %s", name, description),
	}

	// Use infra.Run to execute command with proper stdout/stderr separation
	stdout, stderr, exitCode, err := infra.Run(workDir, cmdName, args...)

	if err != nil {
		// Command couldn't run at all
		result.Status = statusError
		result.Output = err.Error()
		result.Summary = fmt.Sprintf("Could not run %s: %v", name, err)
	} else if exitCode != 0 {
		// Command ran but returned non-zero exit code
		result.Status = statusFail

		// Use stderr for error information, stdout for normal output
		if stderr != "" {
			result.Output = stderr
		} else if stdout != "" {
			result.Output = stdout
		}

		switch name {
		case checkGofmt:
			if stdout != "" {
				result.Summary = fmt.Sprintf(
					"Files need formatting: %s",
					strings.ReplaceAll(stdout, "\n", ", "),
				)
			} else {
				result.Summary = "Files need formatting"
			}
		case checkGoModTidy:
			result.Summary = "go.mod needs tidying"
		case checkGoVet:
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
		case checkGoBuild:
			result.Summary = "Build would fail - compilation errors found"
		default:
			result.Summary = fmt.Sprintf("Check failed: %s", name)
		}
	} else {
		// Command succeeded (exit code 0)
		result.Status = statusPass

		switch name {
		case checkGofmt:
			// gofmt -l returns 0 even when files need formatting, but lists files to stdout
			if stdout != "" {
				result.Status = statusFail
				result.Output = stdout
				result.Summary = fmt.Sprintf(
					"Files need formatting: %s",
					strings.ReplaceAll(stdout, "\n", ", "),
				)
			} else {
				result.Summary = "All files are properly formatted"
			}
		case checkGoVet:
			result.Summary = "No vet issues found"
		case checkGoBuild:
			// Don't include verbose build output when successful
			result.Summary = "Code compiles successfully"
		case checkGoModTidy:
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
