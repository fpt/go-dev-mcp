package app

import (
	"fmt"
	"strings"

	"fujlog.net/godev-mcp/internal/repository"

	"github.com/pkg/errors"
)

func PrintTree(b *strings.Builder, walker repository.DirWalker, path string, isLast bool) error {
	err := walker.Walk(
		func(name, prefix string, isLastEntry bool) error {
			// Determine the current line's connector
			connector := "|-- "
			if isLastEntry {
				connector = "└-- "
			}

			// Print the current item
			b.WriteString(fmt.Sprintf("%s%s%s\n", prefix, connector, name))

			return nil
		},
		func(prefix string, isLastEntry bool) string {
			nextPrefix := prefix + "|   "
			if isLastEntry {
				nextPrefix = prefix + "    "
			}
			return nextPrefix
		},
		"",
		path)

	if err != nil {
		return errors.Wrap(err, "failed to walk directory")
	}

	return nil
}
