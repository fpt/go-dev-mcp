package infra

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func Run(workdir, cmd string, args []string) (string, string, int, error) {
	stdout := strings.Builder{}
	stderr := strings.Builder{}

	// Create the command
	command := exec.Command(cmd, args...)
	command.Dir = workdir
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()

	stdoutStr := strings.TrimSpace(stdout.String())
	stderrStr := strings.TrimSpace(stderr.String())

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return stdoutStr, stderrStr, exitError.ExitCode(), errors.Wrap(err, "command execution failed")
		}
		return stdoutStr, stderrStr, 1, errors.Wrap(err, "command execution failed")
	}

	// Print the output
	exitCode := command.ProcessState.ExitCode()
	return stdoutStr, stderrStr, exitCode, nil
}
