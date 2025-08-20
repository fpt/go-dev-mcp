package subcmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/fpt/go-dev-mcp/internal/app"
	"github.com/google/subcommands"
)

type ValidateCmd struct {
	directory string
}

func (*ValidateCmd) Name() string     { return "validate" }
func (*ValidateCmd) Synopsis() string { return "Validate Go code using multiple static analysis tools" }
func (*ValidateCmd) Usage() string {
	return `validate [-directory <path>]:
  Validate Go code using go vet, build checks, formatting validation, and module tidiness.
  Provides comprehensive code quality assessment.
`
}

func (p *ValidateCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.directory, "directory", ".", "Directory containing Go code to validate")
}

func (p *ValidateCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	// Convert to absolute path
	absPath, err := filepath.Abs(p.directory)
	if err != nil {
		fmt.Printf("Error: Failed to resolve absolute path: %v\n", err)
		return subcommands.ExitFailure
	}

	// Run validation
	report, err := app.ValidateGoCode(ctx, absPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return subcommands.ExitFailure
	}

	// Print results in JSON format
	output, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting output: %v\n", err)
		return subcommands.ExitFailure
	}

	fmt.Println(string(output))

	// Return appropriate exit code
	for _, result := range report.Results {
		if result.Status == "fail" || result.Status == "error" {
			return subcommands.ExitFailure
		}
	}

	return subcommands.ExitSuccess
}
