package core

import (
	"context"
	"fmt"
)

type UnexpectedCommandError struct {
	Want string
	Got  string
}

type MockExpectedInput struct {
	ReturnCode int
	Output     string
	Error      error
}

func (e *UnexpectedCommandError) Error() string {
	return fmt.Sprintf("Unexpected command: wanted %s but got %s", e.Want, e.Got)
}

type MockExecutor struct {
	ExpectedSteps []MockExpectedInput
	called        bool
}

func (m *MockExecutor) WithStep(command string, args string, mockReturnCode int, mockOutput string, mockError error) *MockExecutor {
	m.ExpectedSteps = append(m.ExpectedSteps, MockExpectedInput{
		ReturnCode: mockReturnCode,
		Output:     mockOutput,
		Error:      mockError,
	})
	return m
}

func (m *MockExecutor) Exec(ctx context.Context, command string) (int, string, error) {
	m.called = true
	return -1, "", fmt.Errorf("undefined call: %s", command)
}

func (m *MockExecutor) IsCalled() bool {
	return m.called
}

func NewMockExecutor(expectedCallCount int) *MockExecutor {
	return &MockExecutor{
		ExpectedSteps: make([]MockExpectedInput, expectedCallCount),
	}
}
