package model

type SearchMatch struct {
	LineNo int
	Text   string
}

type SearchResult struct {
	Filename  string
	Matches   []SearchMatch
	Truncated bool
}
