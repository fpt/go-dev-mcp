package app

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"

	"fujlog.net/godev-mcp/internal/repository"
)

const headerSize = 16

type SearchMatch struct {
	LineNo int
	Text   string
}

type SearchResult struct {
	Filename string
	Matches  []SearchMatch
}

func SearchLocalFiles(
	ctx context.Context, fw repository.FileWalker, path, extension, query string,
) ([]SearchResult, error) {
	var results []SearchResult
	err := fw.Walk(ctx, func(filePath string) error {
		matches, err := searchInFile(filePath, query)
		if err != nil {
			return errors.Wrap(err, "failed to search in file")
		}
		if len(matches) > 0 {
			results = append(results, SearchResult{Filename: filePath, Matches: matches})
		}

		return nil
	}, path, extension, true)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func searchInFile(filename, query string) ([]SearchMatch, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	reader := bufio.NewReader(fp)
	if !isTextFile(reader) {
		return nil, nil
	}

	lineNo := 1
	var matches []SearchMatch
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if searchInLine(line, query) {
			m := SearchMatch{LineNo: lineNo, Text: line}
			matches = append(matches, m)
		}
		lineNo++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return matches, nil
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
