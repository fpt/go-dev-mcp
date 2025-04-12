package infra

import (
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name           string
		cmd            string
		args           []string
		expectedOutput string
		expectedCode   int
		expectError    bool
	}{
		{
			name:           "echo command",
			cmd:            "echo",
			args:           []string{"hello", "world"},
			expectedOutput: "hello world\n",
			expectedCode:   0,
			expectError:    false,
		},
		{
			name:           "ls command",
			cmd:            "ls",
			args:           []string{"-la"},
			expectedOutput: "", // actual output will vary by environment
			expectedCode:   0,
			expectError:    false,
		},
		{
			name:           "command not found",
			cmd:            "nonexistentcommand",
			args:           []string{},
			expectedOutput: "",
			expectedCode:   1,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _, exitCode, err := Run(".", tt.cmd, tt.args...)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if exitCode != tt.expectedCode {
				t.Errorf("expected exit code %d, got %d", tt.expectedCode, exitCode)
			}

			if tt.cmd == "echo" && !strings.Contains(stdout, "hello world") {
				t.Errorf("expected output to contain 'hello world', got: %s", stdout)
			}

			if tt.cmd == "ls" && err == nil {
				// Just verify we got some output for ls
				if len(stdout) == 0 {
					t.Errorf("expected non-empty output for ls command")
				}
			}
		})
	}
}
