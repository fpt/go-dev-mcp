package contentsearch

import (
	"bufio"
	"io"
	"strings"

	"github.com/fpt/go-dev-mcp/internal/model"
)

func SearchInContent(
	reader io.Reader,
	query string,
	maxMatches int,
) ([]model.SearchMatch, bool, error) {
	lineNo := 1
	var matches []model.SearchMatch
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
			m := model.SearchMatch{LineNo: lineNo, Text: line}
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
