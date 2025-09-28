package config

import (
	"context"
)

type contextKey string

const (
	configKey contextKey = "config"
)

func WithContext(ctx context.Context, definition ProjectDefinition) context.Context {
	ctx = context.WithValue(ctx, configKey, definition)
	return ctx
}

func FromContext(ctx context.Context) ProjectDefinition {
	config, ok := ctx.Value(configKey).(ProjectDefinition)
	if !ok {
		panic("No root dir found in context, bad code path")
	}
	return config
}
