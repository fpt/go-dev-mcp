package app

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"

	"github.com/fpt/go-dev-mcp/internal/repository"
)

const headerSize = 16

type SearchMatch struct {
	LineNo int
	Text   string
}

type SearchResult struct {
	Filename  string
	Matches   []SearchMatch
	Truncated bool
}

func SearchLocalFiles(
	ctx context.Context, fw repository.FileWalker, path, extension, query string, maxMatches int,
) ([]SearchResult, error) {
	var results []SearchResult
	err := fw.Walk(ctx, func(filePath string) error {
		matches, truncated, err := searchInFile(filePath, query, maxMatches)
		if err != nil {
			return errors.Wrap(err, "failed to search in file")
		}
		if len(matches) > 0 {
			results = append(results, SearchResult{
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

func searchInFile(filename, query string, maxMatches int) ([]SearchMatch, bool, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return nil, false, err
	}
	defer fp.Close()

	reader := bufio.NewReader(fp)
	if !isTextFile(reader) {
		return nil, false, nil
	}

	lineNo := 1
	var matches []SearchMatch
	truncated := false
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if searchInLine(line, query) {
			// Check if we've reached the maximum number of matches
			if len(matches) >= maxMatches {
				truncated = true
				break
			}
			m := SearchMatch{LineNo: lineNo, Text: line}
			matches = append(matches, m)
		}
		lineNo++
	}
	if err := scanner.Err(); err != nil {
		return nil, false, err
	}

	return matches, truncated, nil
}

func searchInLine(line, query string) bool {
	return strings.Contains(line, query)
}

func isTextFile(reader *bufio.Reader) bool {
	buf, err := reader.Peek(headerSize)
	if err != nil && !errors.Is(err, io.EOF) {
		return false
	}

	return utf8.ValidString(string(buf))
}
