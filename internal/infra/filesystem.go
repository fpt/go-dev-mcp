package infra

import (
	"context"
	"os"
	"path/filepath"

	"fujlog.net/godev-mcp/internal/repository"
	"github.com/pkg/errors"
)

func IsFileExist(workdir, filename string) bool {
	_, err := os.Stat(filepath.Join(workdir, filename))
	return err == nil
}

type DirWalker struct{}

func NewDirWalker() repository.DirWalker {
	return &DirWalker{}
}

func (dw *DirWalker) Walk(ctx context.Context, function repository.WalkDirFunc, prefixFunc repository.WalkDirNextPrefixFunc, prefix, path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return errors.Wrap(err, "failed to read directory")
	}

	// Filter out .git directory
	filteredEntries := make([]os.DirEntry, 0)
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() != ".git" {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	for i, entry := range filteredEntries {
		isLastEntry := (i == len(filteredEntries)-1)

		// Call the function for each entry
		if err := function(entry.Name(), prefix, isLastEntry); err != nil {
			return err
		}

		if entry.IsDir() {
			nextPrefix := prefixFunc(prefix, isLastEntry)

			subpath := filepath.Join(path, entry.Name())
			err := dw.Walk(ctx, function, prefixFunc, nextPrefix, subpath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
