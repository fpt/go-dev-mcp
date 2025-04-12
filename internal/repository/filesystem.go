package repository

type WalkDirFunc func(name, prefix string, isLastEntry bool) error
type WalkDirNextPrefixFunc func(prefix string, isLastEntry bool) string

type DirWalker interface {
	Walk(function WalkDirFunc, prefixFunc WalkDirNextPrefixFunc, prefix, path string) error
}
