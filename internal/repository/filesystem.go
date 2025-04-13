package repository

import "context"

type (
	WalkDirFunc           func(name, prefix string, isLastEntry bool) error
	WalkDirNextPrefixFunc func(prefix string, isLastEntry bool) string
)

type DirWalker interface {
	Walk(ctx context.Context, function WalkDirFunc, prefixFunc WalkDirNextPrefixFunc, prefix, path string) error
}
