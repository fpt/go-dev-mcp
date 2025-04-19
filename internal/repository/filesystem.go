package repository

import "context"

type (
	WalkDirFunc           func(name, prefix string, isLastEntry bool) error
	WalkDirNextPrefixFunc func(prefix string, isLastEntry bool) string
	WalkFileFunc          func(path string) error
)

type DirWalker interface {
	Walk(ctx context.Context, function WalkDirFunc, prefixFunc WalkDirNextPrefixFunc, prefix, path string) error
}

type FileWalker interface {
	Walk(ctx context.Context, function WalkFileFunc, path, extension string, ignoreDot bool) error
}
