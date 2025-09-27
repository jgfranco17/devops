package fileutils

import (
	"context"
	"io/fs"
)

type contextKey string

const (
	rootDirKey contextKey = "rootDir"
)

func ApplyRootDirToContext(ctx context.Context, files fs.FS) context.Context {
	ctx = context.WithValue(ctx, rootDirKey, files)
	return ctx
}
func RootDirFromContext(ctx context.Context) fs.FS {
	rootDir, ok := ctx.Value(rootDirKey).(fs.FS)
	if !ok {
		panic("No root dir found in context, bad code path")
	}
	return rootDir
}
