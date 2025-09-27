package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithContext(t *testing.T) {
	tests := []struct {
		name       string
		definition ProjectDefinition
	}{
		{
			name: "empty project definition",
			definition: ProjectDefinition{},
		},
		{
			name: "complete project definition",
			definition: ProjectDefinition{
				Name:        "test-project",
				Description: "A test project",
				Version:     "1.0.0",
				RepoUrl:     "https://github.com/test/project",
				Codebase: Codebase{
					Language:     "go",
					Dependencies: "go.mod",
					Install: Operation{
						Steps: []string{"go mod download"},
					},
					Build: Operation{
						Steps: []string{"go build ./..."},
					},
				},
			},
		},
		{
			name: "project with nested operations",
			definition: ProjectDefinition{
				Name: "complex-project",
				Codebase: Codebase{
					Language: "python",
					Install: Operation{
						FailFast: true,
						Env: map[string]string{
							"PYTHONPATH": "/custom/path",
						},
						Steps: []string{"pip install -r requirements.txt"},
					},
					Build: Operation{
						FailFast: false,
						Env: map[string]string{
							"BUILD_ENV": "production",
						},
						Steps: []string{"python setup.py build", "python -m pytest"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			newCtx := WithContext(ctx, tt.definition)

			// Verify that the context is different
			assert.NotEqual(t, ctx, newCtx)

			// Verify that the definition can be retrieved
			retrieved := FromContext(newCtx)
			assert.Equal(t, tt.definition, retrieved)
		})
	}
}

func TestFromContext(t *testing.T) {
	tests := []struct {
		name        string
		setupCtx    func() context.Context
		expectPanic bool
		expected    ProjectDefinition
	}{
		{
			name: "context with valid project definition",
			setupCtx: func() context.Context {
				definition := ProjectDefinition{
					Name:    "test-project",
					Version: "1.0.0",
				}
				return WithContext(context.Background(), definition)
			},
			expectPanic: false,
			expected: ProjectDefinition{
				Name:    "test-project",
				Version: "1.0.0",
			},
		},
		{
			name: "context without project definition",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectPanic: true,
		},
		{
			name: "context with wrong type",
			setupCtx: func() context.Context {
				ctx := context.WithValue(context.Background(), configKey, "not a ProjectDefinition")
				return ctx
			},
			expectPanic: true,
		},
		{
			name: "context with nil value",
			setupCtx: func() context.Context {
				ctx := context.WithValue(context.Background(), configKey, nil)
				return ctx
			},
			expectPanic: true,
		},
		{
			name: "nested context with project definition",
			setupCtx: func() context.Context {
				definition := ProjectDefinition{
					Name: "nested-project",
					Codebase: Codebase{
						Language: "rust",
					},
				}
				ctx := WithContext(context.Background(), definition)
				// Add another value to create a nested context
				ctx = context.WithValue(ctx, "other_key", "other_value")
				return ctx
			},
			expectPanic: false,
			expected: ProjectDefinition{
				Name: "nested-project",
				Codebase: Codebase{
					Language: "rust",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()

			if tt.expectPanic {
				assert.Panics(t, func() {
					FromContext(ctx)
				}, "Expected FromContext to panic")
			} else {
				assert.NotPanics(t, func() {
					result := FromContext(ctx)
					assert.Equal(t, tt.expected, result)
				}, "Expected FromContext not to panic")
			}
		})
	}
}

func TestContextKey(t *testing.T) {
	// Test that the context key is properly defined
	assert.Equal(t, contextKey("config"), configKey)
	assert.IsType(t, contextKey(""), configKey)
}

func TestWithContext_Chaining(t *testing.T) {
	// Test that WithContext can be chained
	ctx := context.Background()

	definition1 := ProjectDefinition{Name: "project1"}
	ctx1 := WithContext(ctx, definition1)

	definition2 := ProjectDefinition{Name: "project2"}
	ctx2 := WithContext(ctx1, definition2)

	// The last definition should be retrieved
	result := FromContext(ctx2)
	assert.Equal(t, definition2, result)

	// The original context should still have the first definition
	result1 := FromContext(ctx1)
	assert.Equal(t, definition1, result1)
}

func TestFromContext_TypeAssertion(t *testing.T) {
	// Test the type assertion behavior more thoroughly
	ctx := context.Background()

	// Test with different types
	ctxWithString := context.WithValue(ctx, configKey, "string value")
	ctxWithInt := context.WithValue(ctx, configKey, 42)
	ctxWithStruct := context.WithValue(ctx, configKey, struct{ Name string }{"test"})

	// All should panic except the correct type
	assert.Panics(t, func() { FromContext(ctxWithString) })
	assert.Panics(t, func() { FromContext(ctxWithInt) })
	assert.Panics(t, func() { FromContext(ctxWithStruct) })

	// Only the correct type should work
	ctxWithProject := WithContext(ctx, ProjectDefinition{Name: "test"})
	assert.NotPanics(t, func() { FromContext(ctxWithProject) })
}
