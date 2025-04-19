package app

import (
	"bufio"
	"context"
	"os"
	"strings"

	"github.com/pkg/errors"

	"fujlog.net/godev-mcp/internal/repository"
)

type SearchResult struct {
	Filename string
	Content  string
}

func SearchLocalFiles(ctx context.Context, fw repository.FileWalker, path, extension, query string) ([]SearchResult, error) {
	var results []SearchResult
	err := fw.Walk(ctx, func(filePath string) error {
		content, err := searchInFile(filePath, query)
		if err != nil {
			return errors.Wrap(err, "failed to search in file")
		}
		if content != "" {
			results = append(results, SearchResult{Filename: filePath, Content: content})
		}

		return nil
	}, path, extension, true)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func searchInFile(filename, query string) (string, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	var content strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, query) {
			content.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}
	if content.Len() > 0 {
		return content.String(), nil
	}
	return "", nil
}
