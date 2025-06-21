package app

import (
	"bufio"
	"context"
	"io"
	"os"
	"unicode/utf8"

	"github.com/pkg/errors"

	"github.com/fpt/go-dev-mcp/internal/contentsearch"
	"github.com/fpt/go-dev-mcp/internal/model"
	"github.com/fpt/go-dev-mcp/internal/repository"
)

const headerSize = 16

func SearchLocalFiles(
	ctx context.Context, fw repository.FileWalker, path, extension, query string, maxMatches int,
) ([]model.SearchResult, error) {
	var results []model.SearchResult
	err := fw.Walk(ctx, func(filePath string) error {
		matches, truncated, err := searchInFile(filePath, query, maxMatches)
		if err != nil {
			return errors.Wrap(err, "failed to search in file")
		}
		if len(matches) > 0 {
			results = append(results, model.SearchResult{
				Filename:  filePath,
				Matches:   matches,
				Truncated: truncated,
			})
		}

		return nil
	}, path, extension, true)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func searchInFile(filename, query string, maxMatches int) ([]model.SearchMatch, bool, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return nil, false, err
	}
	defer fp.Close()

	reader := bufio.NewReader(fp)
	if !isTextFile(reader) {
		return nil, false, nil
	}

	matches, truncated, err := contentsearch.SearchInContent(reader, query, maxMatches)
	if err != nil {
		return nil, false, err
	}

	return matches, truncated, nil
}

func isTextFile(reader *bufio.Reader) bool {
	buf, err := reader.Peek(headerSize)
	if err != nil && !errors.Is(err, io.EOF) {
		return false
	}

	return utf8.ValidString(string(buf))
}
