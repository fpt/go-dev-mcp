package infra

import (
	"context"
	"os"
	"path/filepath"
	"strings"

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

func (dw *DirWalker) Walk(
	_ context.Context, function repository.WalkDirFunc, prefixFunc repository.WalkDirNextPrefixFunc,
	prefix, path string, ignoreDot bool, maxDepth int,
) error {
	return dw.walkWithDepth(function, prefixFunc, prefix, path, ignoreDot, maxDepth, 0)
}

func (dw *DirWalker) walkWithDepth(
	function repository.WalkDirFunc, prefixFunc repository.WalkDirNextPrefixFunc,
	prefix, path string, ignoreDot bool, maxDepth, currentDepth int,
) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return errors.Wrap(err, "failed to read directory")
	}

	// Filter out .git directory and optionally other dot files/directories
	filteredEntries := make([]os.DirEntry, 0)
	for _, entry := range entries {
		// Always filter out .git directory
		if entry.IsDir() && entry.Name() == ".git" {
			continue
		}
		// Optionally filter out other dot files/directories
		if ignoreDot && strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		filteredEntries = append(filteredEntries, entry)
	}

	for i, entry := range filteredEntries {
		isLastEntry := (i == len(filteredEntries)-1)

		// Call the function for each entry
		if err := function(entry.Name(), prefix, isLastEntry); err != nil {
			return err
		}

		if entry.IsDir() {
			// Check if we've reached the max depth
			if currentDepth >= maxDepth {
				continue
			}

			nextPrefix := prefixFunc(prefix, isLastEntry)

			subpath := filepath.Join(path, entry.Name())
			err := dw.walkWithDepth(
				function,
				prefixFunc,
				nextPrefix,
				subpath,
				ignoreDot,
				maxDepth,
				currentDepth+1,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type FileWalker struct{}

func NewFileWalker() repository.FileWalker {
	return &FileWalker{}
}

func (fw *FileWalker) Walk(
	ctx context.Context, function repository.WalkFileFunc, path, extension string, ignoreDot bool,
) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return errors.Wrap(err, "failed to read directory")
	}

	// Filter out . entry
	filteredEntries := make([]os.DirEntry, 0)
	for _, entry := range entries {
		if !ignoreDot || !strings.HasPrefix(entry.Name(), ".") {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	for _, entry := range filteredEntries {
		if !entry.IsDir() {
			if extension != "" && filepath.Ext(entry.Name()) != extension {
				continue
			}

			filePath := filepath.Join(path, entry.Name())
			if err := function(filePath); err != nil {
				return errors.Wrap(err, "failed to process file: "+filePath)
			}
		} else {
			nextPath := filepath.Join(path, entry.Name())
			err := fw.Walk(ctx, function, nextPath, extension, ignoreDot)
			if err != nil {
				return errors.Wrap(err, "failed to walk into directory: "+nextPath)
			}
		}
	}

	return nil
}
